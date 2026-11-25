package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"makeitiny/api/internal/db"
	"makeitiny/api/internal/qr"
)

const (
	ClickDeltasKey = "clicks:deltas"
	LogsQueueKey   = "logs:queue"
	SyncInterval   = 30 * time.Second
	LogBatchSize   = 500
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
// periyodik (30 sn) olarak PostgreSQL'e yazar; ayrıca süresi dolan anonim linkleri temizler.
type SyncWorker struct {
	pool    *pgxpool.Pool
	q       *db.Queries
	rdb     *redis.Client
	qrStore *qr.Store
}

func NewSyncWorker(pool *pgxpool.Pool, q *db.Queries, rdb *redis.Client, qrStore *qr.Store) *SyncWorker {
	return &SyncWorker{
		pool:    pool,
		q:       q,
		rdb:     rdb,
		qrStore: qrStore,
	}
}

func (w *SyncWorker) Start(ctx context.Context) {
	// Başlangıçta süresi geçmiş anonim linkleri ve eski refresh token'ları temizle
	w.CleanupExpiredAnonymousLinks(ctx)
	_ = w.q.DeleteExpiredRefreshTokens(ctx)

	go func() {
		ticker := time.NewTicker(SyncInterval)
		defer ticker.Stop()

		tickCount := 0
		for {
			select {
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				w.FlushClickDeltas(shutdownCtx)
				w.FlushLogsQueue(shutdownCtx)
				cancel()
				return
			case <-ticker.C:
				tickCount++
				w.FlushClickDeltas(ctx)
				w.FlushLogsQueue(ctx)

				// Her 5 dakikada bir (10 tick) süresi dolmuş anonim linkleri ve eski refresh token'ları temizle
				if tickCount%10 == 0 {
					w.CleanupExpiredAnonymousLinks(ctx)
					if err := w.q.DeleteExpiredRefreshTokens(ctx); err != nil {
						log.Printf("worker expired refresh tokens temizleme hatası: %v", err)
					}
				}
			}
		}
	}()
}

// FlushClickDeltas Redis'teki clicks:deltas Hash'ini atomik RENAME ile izole eder ve PostgreSQL'e batch UPDATE yapar.
func (w *SyncWorker) FlushClickDeltas(ctx context.Context) {
	if w.rdb == nil || w.pool == nil {
		return
	}

	procKey := fmt.Sprintf("%s:proc:%d", ClickDeltasKey, time.Now().UnixNano())
	if err := w.rdb.Rename(ctx, ClickDeltasKey, procKey).Err(); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "no such key") {
			log.Printf("worker click deltas RENAME uyarısı: %v", err)
		}
		return
	}

	deltas, err := w.rdb.HGetAll(ctx, procKey).Result()
	if err != nil {
		log.Printf("worker click deltas HGETALL hatası: %v", err)
		return
	}
	if len(deltas) == 0 {
		_ = w.rdb.Del(ctx, procKey).Err()
		return
	}

	batch := &pgx.Batch{}
	count := 0
	for alias, valStr := range deltas {
		var delta int64
		if _, err := fmt.Sscan(valStr, &delta); err != nil || delta <= 0 {
			continue
		}
		batch.Queue(`UPDATE links SET clicks = clicks + $2 WHERE alias = $1`, alias, delta)
		count++
	}

	if count > 0 {
		br := w.pool.SendBatch(ctx, batch)
		for i := 0; i < count; i++ {
			if _, err := br.Exec(); err != nil {
				log.Printf("worker click delta batch UPDATE hatası: %v", err)
			}
		}
		if err := br.Close(); err != nil {
			log.Printf("worker click delta batch close hatası: %v", err)
			return
		}
	}

	_ = w.rdb.Del(ctx, procKey).Err()
}

// FlushLogsQueue Redis logs:queue listesindeki JSON log kayıtlarını batch olarak PostgreSQL logs tablosuna yazar.
func (w *SyncWorker) FlushLogsQueue(ctx context.Context) {
	if w.rdb == nil || w.pool == nil {
		return
	}

	for iter := 0; iter < 5; iter++ {
		rawItems, err := w.rdb.LPopCount(ctx, LogsQueueKey, LogBatchSize).Result()
		if err != nil {
			if err != redis.Nil {
				log.Printf("worker logs LPOPCOUNT hatası: %v", err)
			}
			return
		}
		if len(rawItems) == 0 {
			return
		}

		batch := &pgx.Batch{}
		queued := 0
		for _, raw := range rawItems {
			var entry LogQueueEntry
			if err := json.Unmarshal([]byte(raw), &entry); err != nil {
				log.Printf("worker log JSON parse uyarısı: %v", err)
				continue
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

			eventType := entry.EventType
			if eventType == "" {
				eventType = "redirect"
			}

			batch.Queue(
				`INSERT INTO logs (
					link_id, link_alias, link_public_id,
					ip_address, user_agent, device_type, browser, os,
					referer, accept_language, country_code, event_type, clicked_at
				) VALUES ((SELECT id FROM links WHERE id = $1), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
				linkID,
				entry.LinkAlias,
				entry.LinkPublicID,
				entry.IPAddress,
				entry.UserAgent,
				entry.DeviceType,
				entry.Browser,
				entry.OS,
				entry.Referer,
				entry.AcceptLanguage,
				entry.CountryCode,
				eventType,
				pgtype.Timestamptz{Time: clickedAt, Valid: true},
			)
			queued++
		}

		if queued > 0 {
			br := w.pool.SendBatch(ctx, batch)
			for i := 0; i < queued; i++ {
				if _, err := br.Exec(); err != nil {
					log.Printf("worker log batch INSERT hatası: %v", err)
				}
			}
			_ = br.Close()
		}

		if len(rawItems) < LogBatchSize {
			return
		}
	}
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

	keys := make([]string, 0, len(deleted))
	for _, item := range deleted {
		keys = append(keys, "link:"+item.Alias)
		if w.qrStore != nil && item.QrCodePath != "" {
			w.qrStore.Delete(item.QrCodePath)
		}
	}

	if w.rdb != nil && len(keys) > 0 {
		_ = w.rdb.Del(ctx, keys...).Err()
	}

	log.Printf("worker: %d adet süresi dolmuş anonim link temizlendi (loglar korundu)", len(deleted))
}
