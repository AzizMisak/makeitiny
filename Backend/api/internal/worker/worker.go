package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"makeitiny/api/internal/db"
	"makeitiny/api/internal/qr"
)

const (
	ClickDeltasKey      = "clicks:deltas"
	ClickProcPrefix     = "clicks:deltas:proc:"
	clickDeltasProcsKey = "clicks:deltas:procs"
	LogsQueueKey        = "logs:queue"
	SyncInterval        = 30 * time.Second
	LogBatchSize        = 500
	// procKeyTTL: işlenmeyi bekleyen tıklama kopyası en fazla bu kadar korunur (yalnızca kurtarma içindir).
	procKeyTTL = 7 * 24 * time.Hour
)

type LogQueueEntry struct {
	LinkID         int64  `json:"link_id"`
	LinkAlias      string `json:"link_alias"`
	LinkPublicID   string `json:"link_public_id"`
	IPAddress      string `json:"ip_address"`
	UserAgent      string `json:"user_agent"`
	DeviceType     string `json:"device_type"`
	Browser        string `json:"browser"`
	OS             string `json:"os"`
	Referer        string `json:"referer"`
	AcceptLanguage string `json:"accept_language"`
	CountryCode    string `json:"country_code"`
	EventType      string `json:"event_type"`
	ClickedAt      string `json:"clicked_at"`
}

// SyncWorker Redis üzerindeki birikmiş tıklama farklarını ve cihaz loglarını
// periyodik (30 sn) olarak PostgreSQL'e yazar; süresi dolan anonim linkleri ve askıları temizler.
type SyncWorker struct {
	pool    *pgxpool.Pool
	q       *db.Queries
	rdb     *redis.Client
	qrStore *qr.Store
	wg      sync.WaitGroup
}

func NewSyncWorker(pool *pgxpool.Pool, q *db.Queries, rdb *redis.Client, qrStore *qr.Store) *SyncWorker {
	return &SyncWorker{
		pool:    pool,
		q:       q,
		rdb:     rdb,
		qrStore: qrStore,
	}
}

// Start arka plan döngüsünü başlatır. ctx iptal edilince son bir aktarım yapılır; Wait bunun bitmesini bekler.
func (w *SyncWorker) Start(ctx context.Context) {
	w.CleanupExpiredAnonymousLinks(ctx)
	w.CleanupExpiredGraceLinks(ctx)
	w.UnbanExpiredUsers(ctx)
	_ = w.q.DeleteExpiredRefreshTokens(ctx)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(SyncInterval)
		defer ticker.Stop()

		tickCount := 0
		for {
			select {
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				w.FlushClickDeltas(shutdownCtx)
				w.FlushLogsQueue(shutdownCtx)
				cancel()
				return
			case <-ticker.C:
				tickCount++
				w.FlushClickDeltas(ctx)
				w.FlushLogsQueue(ctx)
				w.UnbanExpiredUsers(ctx)

				// Her 5 dakikada bir (10 tick) süresi dolmuş anonim linkleri ve eski refresh token'ları temizle
				if tickCount%10 == 0 {
					w.CleanupExpiredAnonymousLinks(ctx)
					w.CleanupExpiredGraceLinks(ctx)
					if err := w.q.DeleteExpiredRefreshTokens(ctx); err != nil {
						log.Printf("worker expired refresh tokens temizleme hatası: %v", err)
					}
				}
			}
		}
	}()
}

// Wait son aktarım bitene kadar bekler. Uygulama kapanırken Redis ve PostgreSQL bağlantıları kapatılmadan önce çağrılır.
func (w *SyncWorker) Wait() {
	w.wg.Wait()
}

// FlushClickDeltas önce yarıda kalmış eski kopyaları, sonra güncel farkları PostgreSQL'e yazar.
func (w *SyncWorker) FlushClickDeltas(ctx context.Context) {
	if w.rdb == nil || w.pool == nil {
		return
	}

	// Önceki turda yazılamayan ya da işlem ortasında kalan kopyalar.
	var cursor uint64
	for {
		keys, next, err := w.rdb.Scan(ctx, cursor, ClickProcPrefix+"*", 100).Result()
		if err != nil {
			log.Printf("worker click deltas SCAN uyarısı: %v", err)
			break
		}
		for _, key := range keys {
			_ = w.rdb.SAdd(ctx, clickDeltasProcsKey, key).Err()
			w.flushProcKey(ctx, key)
		}
		if next == 0 {
			break
		}
		cursor = next
	}

	procKey := fmt.Sprintf("%s%d", ClickProcPrefix, time.Now().UnixNano())
	// Kopya, adı değişmeden önce izleme kümesine yazılır. Böylece router, RENAME ile SADD arasındaki
	// anda bekleyen tıklamaları gözden kaçırıp tıklama sınırını aşacak eksik bir sayı önbelleklemez.
	_ = w.rdb.SAdd(ctx, clickDeltasProcsKey, procKey).Err()
	if err := w.rdb.Rename(ctx, ClickDeltasKey, procKey).Err(); err != nil {
		_ = w.rdb.SRem(ctx, clickDeltasProcsKey, procKey).Err()
		if !strings.Contains(strings.ToLower(err.Error()), "no such key") {
			log.Printf("worker click deltas RENAME uyarısı: %v", err)
		}
		return
	}
	// Kopya, yazılamazsa sonraki turda yeniden denenir; sonsuza kadar sahipsiz kalmasın diye süresi vardır.
	_ = w.rdb.Expire(ctx, procKey, procKeyTTL).Err()
	w.flushProcKey(ctx, procKey)
}

// flushProcKey bir kopyadaki bütün farkları tek bir işlemde yazar. İşlem başarısızsa kopya silinmez,
// böylece tıklamalar kaybolmaz; başarılıysa silinir.
func (w *SyncWorker) flushProcKey(ctx context.Context, procKey string) {
	deltas, err := w.rdb.HGetAll(ctx, procKey).Result()
	if err != nil {
		log.Printf("worker click deltas HGETALL hatası (%s): %v", procKey, err)
		return
	}
	if len(deltas) == 0 {
		_ = w.rdb.SRem(ctx, clickDeltasProcsKey, procKey).Err()
		return
	}

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		log.Printf("worker click deltas işlem başlatılamadı: %v", err)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	batch := &pgx.Batch{}
	queued := 0
	for field, valStr := range deltas {
		delta, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil || delta <= 0 {
			continue
		}
		if idStr, ok := strings.CutPrefix(field, "id:"); ok {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				continue
			}
			batch.Queue(`UPDATE links SET clicks = clicks + $2 WHERE id = $1`, id, delta)
		} else {
			// Eski sürümden kalan takma ad anahtarlı farklar
			batch.Queue(`UPDATE links SET clicks = clicks + $2 WHERE alias = $1`, field, delta)
		}
		queued++
	}

	if queued > 0 {
		br := tx.SendBatch(ctx, batch)
		for i := 0; i < queued; i++ {
			if _, err := br.Exec(); err != nil {
				log.Printf("worker click delta UPDATE hatası, kopya korunuyor (%s): %v", procKey, err)
				_ = br.Close()
				return
			}
		}
		if err := br.Close(); err != nil {
			log.Printf("worker click delta batch kapanış hatası, kopya korunuyor (%s): %v", procKey, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("worker click delta commit hatası, kopya korunuyor (%s): %v", procKey, err)
		return
	}
	if err := w.rdb.Del(ctx, procKey).Err(); err != nil {
		log.Printf("worker click delta kopyası silinemedi (%s): %v", procKey, err)
	} else {
		_ = w.rdb.SRem(ctx, clickDeltasProcsKey, procKey).Err()
	}
}

const logInsertSQL = `INSERT INTO logs (
	link_id, link_alias, link_public_id,
	ip_address, user_agent, device_type, browser, os,
	referer, accept_language, country_code, event_type, clicked_at
) VALUES ((SELECT id FROM links WHERE id = $1), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

// sanitizePGText PostgreSQL TEXT alanlarının reddettiği NUL (\x00) baytlarını temizler.
func sanitizePGText(s string) string {
	return strings.ReplaceAll(s, "\x00", "")
}

// isPermanentPGError geçici bağlantı kesintisinden değil, geçersiz veri (22xxx) veya kısıt (23xxx)
// hatasından kaynaklanan kalıcı PostgreSQL hatalarını ayırt eder.
func isPermanentPGError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && len(pgErr.Code) >= 2 {
		cls := pgErr.Code[:2]
		return cls == "22" || cls == "23"
	}
	return false
}

func parseLogInsertArgs(raw string) ([]any, error) {
	var entry LogQueueEntry
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		return nil, err
	}

	var linkID pgtype.Int8
	if entry.LinkID > 0 {
		linkID = pgtype.Int8{Int64: entry.LinkID, Valid: true}
	}

	clickedAt := time.Now()
	if entry.ClickedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, entry.ClickedAt); err == nil {
			clickedAt = parsed
		}
	}

	eventType := sanitizePGText(entry.EventType)
	if eventType == "" {
		eventType = "redirect"
	}

	return []any{
		linkID,
		sanitizePGText(entry.LinkAlias),
		sanitizePGText(entry.LinkPublicID),
		sanitizePGText(entry.IPAddress),
		sanitizePGText(entry.UserAgent),
		sanitizePGText(entry.DeviceType),
		sanitizePGText(entry.Browser),
		sanitizePGText(entry.OS),
		sanitizePGText(entry.Referer),
		sanitizePGText(entry.AcceptLanguage),
		sanitizePGText(entry.CountryCode),
		eventType,
		pgtype.Timestamptz{Time: clickedAt, Valid: true},
	}, nil
}

// FlushLogsQueue kuyruğun başındaki kayıtları silmeden okur, tek işlemde PostgreSQL'e yazar ve
// yalnızca işlem başarılı olunca kuyruktan çıkarır. Tek bir bozuk kayıt kalıcı PG hatası verirse
// satır satır işlenerek yalnızca bozuk kayıt atlanır; geçici kesintide kayıtlar kuyrukta kalır.
func (w *SyncWorker) FlushLogsQueue(ctx context.Context) {
	if w.rdb == nil || w.pool == nil {
		return
	}

	for iter := 0; iter < 5; iter++ {
		rawItems, err := w.rdb.LRange(ctx, LogsQueueKey, 0, LogBatchSize-1).Result()
		if err != nil {
			if err != redis.Nil {
				log.Printf("worker logs LRANGE hatası: %v", err)
			}
			return
		}
		if len(rawItems) == 0 {
			return
		}

		trimmed := len(rawItems)
		if err := w.insertLogs(ctx, rawItems); err != nil {
			if !isPermanentPGError(err) {
				log.Printf("worker log aktarımı başarısız, kayıtlar kuyrukta bekliyor: %v", err)
				return
			}
			log.Printf("worker log batch kalıcı veri hatası, tekil moda geçiliyor: %v", err)
			processed := w.insertLogsOneByOne(ctx, rawItems)
			if processed == 0 {
				return
			}
			trimmed = processed
		}

		// Kuyruğa yalnızca sondan ekleme yapıldığı için baştaki işlenmiş kayıtlar güvenle kırpılır.
		if err := w.rdb.LTrim(ctx, LogsQueueKey, int64(trimmed), -1).Err(); err != nil {
			log.Printf("worker logs LTRIM hatası: %v", err)
			return
		}

		if trimmed < LogBatchSize {
			return
		}
	}
}

func (w *SyncWorker) insertLogs(ctx context.Context, rawItems []string) error {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	batch := &pgx.Batch{}
	queued := 0
	for _, raw := range rawItems {
		args, err := parseLogInsertArgs(raw)
		if err != nil {
			log.Printf("worker bozuk log kaydı atlandı: %v", err)
			continue
		}
		batch.Queue(logInsertSQL, args...)
		queued++
	}

	if queued > 0 {
		br := tx.SendBatch(ctx, batch)
		for i := 0; i < queued; i++ {
			if _, err := br.Exec(); err != nil {
				_ = br.Close()
				return err
			}
		}
		if err := br.Close(); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// insertLogsOneByOne zehirli (poison-pill) bir kayıt yüzünden toplu aktarım durduğunda her kaydı
// tek tek dener; kalıcı veri hatası veren satırı atlar, geçici DB kesintisinde durup kaldığı yeri döner.
func (w *SyncWorker) insertLogsOneByOne(ctx context.Context, rawItems []string) int {
	processed := 0
	for _, raw := range rawItems {
		args, err := parseLogInsertArgs(raw)
		if err != nil {
			log.Printf("worker bozuk log kaydı atlandı: %v", err)
			processed++
			continue
		}
		if _, err := w.pool.Exec(ctx, logInsertSQL, args...); err != nil {
			if isPermanentPGError(err) {
				log.Printf("worker kalıcı hatalı log kaydı atlandı: %v", err)
				processed++
				continue
			}
			log.Printf("worker tekil log aktarımında geçici hata, kalanlar kuyrukta bekliyor: %v", err)
			return processed
		}
		processed++
	}
	return processed
}

// UnbanExpiredUsers süresi dolan askıları kullanıcı giriş yapmasını beklemeden kaldırır ve
// linklerinin önbelleğini siler; linkler en geç bir tur (30 sn) içinde yeniden çalışır.
func (w *SyncWorker) UnbanExpiredUsers(ctx context.Context) {
	ids, err := w.q.UnbanExpiredUsers(ctx)
	if err != nil {
		log.Printf("worker süresi dolan askıları kaldırma hatası: %v", err)
		return
	}
	if len(ids) == 0 || w.rdb == nil {
		return
	}
	for _, id := range ids {
		links, err := w.q.ListLinkIDsByUser(ctx, pgtype.Int8{Int64: id, Valid: true})
		if err != nil || len(links) == 0 {
			continue
		}
		keys := make([]string, 0, len(links)*2)
		for _, l := range links {
			keys = append(keys, "link:"+l.Alias, "reach:"+l.Alias)
		}
		_ = w.rdb.Del(ctx, keys...).Err()
	}
	log.Printf("worker: %d kullanıcının süresi dolan askısı kaldırıldı", len(ids))
}

// CleanupExpiredAnonymousLinks süresi dolan anonim linkleri DB'den ve Redis'ten siler;
// bağlı loglar ON DELETE SET NULL sayesinde PostgreSQL'de korunur.
func (w *SyncWorker) CleanupExpiredAnonymousLinks(ctx context.Context) {
	deleted, err := w.q.DeleteExpiredAnonymousLinks(ctx)
	if err != nil {
		log.Printf("worker anonim link temizleme hatası: %v", err)
		return
	}
	if len(deleted) == 0 {
		return
	}

	keys := make([]string, 0, len(deleted)*2)
	for _, item := range deleted {
		keys = append(keys, "link:"+item.Alias, "reach:"+item.Alias)
		if w.qrStore != nil && item.QrCodePath != "" {
			w.qrStore.Delete(item.QrCodePath)
		}
	}

	if w.rdb != nil && len(keys) > 0 {
		_ = w.rdb.Del(ctx, keys...).Err()
	}

	log.Printf("worker: %d adet süresi dolmuş anonim link temizlendi (loglar korundu)", len(deleted))
}

// CleanupExpiredGraceLinks paket düşürüldüğü için askıya alınıp bekleme süresi (2 ay) dolan linkleri
// önce PostgreSQL'den, ardından Redis'ten ve diskteki QR dosyalarından siler; loglar korunur.
func (w *SyncWorker) CleanupExpiredGraceLinks(ctx context.Context) {
	deleted, err := w.q.DeleteExpiredGraceLinks(ctx)
	if err != nil {
		log.Printf("worker askıdaki link temizleme hatası: %v", err)
		return
	}
	if len(deleted) == 0 {
		return
	}

	keys := make([]string, 0, len(deleted)*2)
	for _, item := range deleted {
		keys = append(keys, "link:"+item.Alias, "reach:"+item.Alias)
		if w.qrStore != nil && item.QrCodePath != "" {
			w.qrStore.Delete(item.QrCodePath)
		}
	}
	if w.rdb != nil {
		_ = w.rdb.Del(ctx, keys...).Err()
	}

	log.Printf("worker: askı süresi dolan %d link silindi (loglar korundu)", len(deleted))
}
