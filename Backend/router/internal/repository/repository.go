package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	"makeitiny/router/internal/db"
)

const (
	linkCacheTTL        = 24 * time.Hour
	flushRaceCacheTTL   = 5 * time.Second
	reachCacheTTL       = 5 * time.Minute
	reachFailTTL        = 30 * time.Second
	waitTicketGrace     = 10 * time.Minute
	unlockRateWindow    = 60 * time.Second
	unlockMaxAttempts   = 10
	clickDeltasKey      = "clicks:deltas"
	clickDeltasProcsKey = "clicks:deltas:procs"
	logsQueueKey        = "logs:queue"
	setIfMissingLua     = `if redis.call('EXISTS', KEYS[1]) == 0 then redis.call('HSET', KEYS[1], unpack(ARGV, 2)) redis.call('EXPIRE', KEYS[1], tonumber(ARGV[1])) end return 1`
	recordClickLua      = `local max_clicks = tonumber(ARGV[1])
local exists = redis.call('EXISTS', KEYS[1])
if max_clicks > 0 then
  if exists == 0 then
    return -1
  end
  local cur = tonumber(redis.call('HGET', KEYS[1], 'clicks') or '0') or 0
  if cur >= max_clicks then
    return 0
  end
end
if exists == 1 then
  redis.call('HINCRBY', KEYS[1], 'clicks', 1)
end
redis.call('HINCRBY', KEYS[2], ARGV[2], 1)
redis.call('RPUSH', KEYS[3], ARGV[3])
return 1`
)

func linkCacheKey(alias string) string {
	return "link:" + alias
}

func reachCacheKey(alias string) string {
	return "reach:" + alias
}

// ClickDeltaField tıklama farkının clicks:deltas içindeki alanı. Farklar takma adla değil link kimliğiyle
// tutulur; böylece takma ad değişse de tıklamalar doğru linke yazılır.
func ClickDeltaField(linkID int64) string {
	return "id:" + strconv.FormatInt(linkID, 10)
}

func waitTicketKey(ticket string) string {
	return "wait:" + ticket
}

// Reachability hedef erişilebilirliğinin önbellekteki durumu.
type Reachability int

const (
	ReachUnknown Reachability = iota
	ReachOK
	ReachFailed
)

// WaitTicket bekletmeli linke yapılan sayılmış bir ziyaretin bileti.
type WaitTicket struct {
	Alias   string
	ReadyAt time.Time
}

var (
	ErrNoRedis        = errors.New("redis kullanılamıyor")
	ErrTicketNotFound = errors.New("bekleme bileti bulunamadı")
)

type LinkRecord struct {
	ID             int64
	PublicID       string
	UserID         int64
	Alias          string
	TargetURL      string
	PasswordHash   string
	Clicks         int64
	WaitEnabled    bool
	WaitSeconds    int32
	IsActive       bool
	ExpiresAtUnix  int64
	MaxClicks      int64
	AdEnabled      bool
	SecurityStatus string
	IsAnonymous    bool
}

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

// LinkRepository link verisine erişim katmanı (Redis Hash + Pipeline + PostgreSQL fallback).
type LinkRepository struct {
	q     *db.Queries
	redis *redis.Client
}

func NewLinkRepository(q *db.Queries, rdb *redis.Client) *LinkRepository {
	return &LinkRepository{q: q, redis: rdb}
}

// GetByAlias önce Redis Hash (HGETALL link:{alias}) üzerinden okur;
// bulamazsa PostgreSQL'den sorgular ve Redis Hash'e yazar.
func (r *LinkRepository) GetByAlias(ctx context.Context, alias string) (LinkRecord, error) {
	key := linkCacheKey(alias)

	if r.redis != nil {
		vals, err := r.redis.HGetAll(ctx, key).Result()
		if err == nil && len(vals) > 0 && vals["target_url"] != "" {
			id, _ := strconv.ParseInt(vals["id"], 10, 64)
			userID, _ := strconv.ParseInt(vals["user_id"], 10, 64)
			clicks, _ := strconv.ParseInt(vals["clicks"], 10, 64)
			waitEnabled, _ := strconv.ParseBool(vals["wait_enabled"])
			waitSeconds, _ := strconv.ParseInt(vals["wait_seconds"], 10, 32)
			isActive := true
			if vals["is_active"] != "" {
				isActive, _ = strconv.ParseBool(vals["is_active"])
			}
			expiresAtUnix, _ := strconv.ParseInt(vals["expires_at"], 10, 64)
			maxClicks, _ := strconv.ParseInt(vals["max_clicks"], 10, 64)
			adEnabled, _ := strconv.ParseBool(vals["ad_enabled"])
			isAnonymous, _ := strconv.ParseBool(vals["is_anonymous"])
			secStatus := vals["security_status"]
			if secStatus == "" {
				secStatus = "unchecked"
			}

			return LinkRecord{
				ID:             id,
				PublicID:       vals["public_id"],
				UserID:         userID,
				Alias:          alias,
				TargetURL:      vals["target_url"],
				PasswordHash:   vals["password_hash"],
				Clicks:         clicks,
				WaitEnabled:    waitEnabled,
				WaitSeconds:    int32(waitSeconds),
				IsActive:       isActive,
				ExpiresAtUnix:  expiresAtUnix,
				MaxClicks:      maxClicks,
				AdEnabled:      adEnabled,
				SecurityStatus: secStatus,
				IsAnonymous:    isAnonymous,
			}, nil
		}
	}

	link, err := r.q.GetLinkByAlias(ctx, alias)
	if err != nil {
		return LinkRecord{}, err
	}

	pwHash := ""
	if link.PasswordHash.Valid {
		pwHash = link.PasswordHash.String
	}

	var expiresAtUnix int64
	if link.ExpiresAt.Valid {
		expiresAtUnix = link.ExpiresAt.Time.Unix()
	}

	var maxClicks int64
	if link.MaxClicks.Valid {
		maxClicks = link.MaxClicks.Int64
	}

	isActive := true
	if link.IsActive.Valid {
		isActive = link.IsActive.Bool
	}

	pending, flushing := r.pendingClicks(ctx, link.ID, link.Alias)
	rec := LinkRecord{
		ID:             link.ID,
		PublicID:       link.PublicID,
		UserID:         link.UserID,
		Alias:          link.Alias,
		TargetURL:      link.TargetUrl,
		PasswordHash:   pwHash,
		Clicks:         link.Clicks + pending,
		WaitEnabled:    link.WaitEnabled,
		WaitSeconds:    link.WaitSeconds,
		IsActive:       isActive,
		ExpiresAtUnix:  expiresAtUnix,
		MaxClicks:      maxClicks,
		AdEnabled:      link.AdEnabled,
		SecurityStatus: link.SecurityStatus,
		IsAnonymous:    link.IsAnonymous,
	}

	if r.redis != nil {
		ttl := linkCacheTTL
		if flushing && flushRaceCacheTTL < ttl {
			ttl = flushRaceCacheTTL
		}
		if link.IsAnonymous && link.ExpiresAt.Valid {
			rem := time.Until(link.ExpiresAt.Time)
			if rem > 0 && rem < ttl {
				ttl = rem
			} else if rem <= 0 {
				return rec, nil
			}
		}

		ttlSec := int64(ttl / time.Second)
		if ttlSec < 1 {
			ttlSec = 1
		}
		err := r.redis.Eval(ctx, setIfMissingLua, []string{key},
			strconv.FormatInt(ttlSec, 10),
			"id", strconv.FormatInt(rec.ID, 10),
			"public_id", rec.PublicID,
			"user_id", strconv.FormatInt(rec.UserID, 10),
			"alias", rec.Alias,
			"target_url", rec.TargetURL,
			"password_hash", rec.PasswordHash,
			"clicks", strconv.FormatInt(rec.Clicks, 10),
			"wait_enabled", strconv.FormatBool(rec.WaitEnabled),
			"wait_seconds", strconv.FormatInt(int64(rec.WaitSeconds), 10),
			"is_active", strconv.FormatBool(rec.IsActive),
			"expires_at", strconv.FormatInt(rec.ExpiresAtUnix, 10),
			"max_clicks", strconv.FormatInt(rec.MaxClicks, 10),
			"ad_enabled", strconv.FormatBool(rec.AdEnabled),
			"security_status", rec.SecurityStatus,
			"is_anonymous", strconv.FormatBool(rec.IsAnonymous),
		).Err()
		if err != nil {
			log.Printf("router redis HSET uyarısı (%s): %v", key, err)
		}
	}

	return rec, nil
}

// CachedReachability hedefin son denetim sonucunu döner (önbellekte yoksa ReachUnknown).
func (r *LinkRepository) CachedReachability(ctx context.Context, alias string) Reachability {
	if r.redis == nil || alias == "" {
		return ReachUnknown
	}
	val, err := r.redis.Get(ctx, reachCacheKey(alias)).Result()
	switch {
	case err != nil:
		return ReachUnknown
	case val == "fail":
		return ReachFailed
	default:
		return ReachOK
	}
}

// CacheReachability olumlu sonucu 5 dakika, olumsuz sonucu 30 saniye saklar.
func (r *LinkRepository) CacheReachability(ctx context.Context, alias string, ok bool) {
	if r.redis == nil || alias == "" {
		return
	}
	if ok {
		_ = r.redis.Set(ctx, reachCacheKey(alias), "ok", reachCacheTTL).Err()
		return
	}
	_ = r.redis.Set(ctx, reachCacheKey(alias), "fail", reachFailTTL).Err()
}

// CreateWaitTicket tek kullanımlık bekleme bileti üretir; bilet beklemeden 10 dakika sonra kendiliğinden silinir.
func (r *LinkRepository) CreateWaitTicket(ctx context.Context, alias string, wait time.Duration) (string, error) {
	if r.redis == nil {
		return "", ErrNoRedis
	}
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	readyAt := time.Now().Add(wait).UnixMilli()
	value := alias + "|" + strconv.FormatInt(readyAt, 10)
	if err := r.redis.Set(ctx, waitTicketKey(ticket), value, wait+waitTicketGrace).Err(); err != nil {
		return "", err
	}
	return ticket, nil
}

func parseWaitTicket(value string) (WaitTicket, bool) {
	idx := strings.LastIndex(value, "|")
	if idx <= 0 {
		return WaitTicket{}, false
	}
	ms, err := strconv.ParseInt(value[idx+1:], 10, 64)
	if err != nil {
		return WaitTicket{}, false
	}
	return WaitTicket{Alias: value[:idx], ReadyAt: time.UnixMilli(ms)}, true
}

// GetWaitTicket bileti tüketmeden okur.
func (r *LinkRepository) GetWaitTicket(ctx context.Context, ticket string) (WaitTicket, error) {
	if r.redis == nil {
		return WaitTicket{}, ErrNoRedis
	}
	if ticket == "" {
		return WaitTicket{}, ErrTicketNotFound
	}
	val, err := r.redis.Get(ctx, waitTicketKey(ticket)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return WaitTicket{}, ErrTicketNotFound
		}
		return WaitTicket{}, err
	}
	t, ok := parseWaitTicket(val)
	if !ok {
		return WaitTicket{}, ErrTicketNotFound
	}
	return t, nil
}

// ConsumeWaitTicket bileti atomik olarak siler; yalnızca ilk çağıran true alır.
func (r *LinkRepository) ConsumeWaitTicket(ctx context.Context, ticket string) (bool, error) {
	if r.redis == nil {
		return false, ErrNoRedis
	}
	if ticket == "" {
		return false, nil
	}
	n, err := r.redis.Del(ctx, waitTicketKey(ticket)).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// AllowUnlockAttempt belirli bir IP ve link takma adı için 60 saniyede en fazla 10 parola denemesine izin verir.
// Redis erişilemezse fail-open davranır (true döner).
func (r *LinkRepository) AllowUnlockAttempt(ctx context.Context, alias, ip string) bool {
	if r == nil || r.redis == nil {
		return true
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "unknown"
	}
	key := "unlock:rl:" + alias + ":" + ip
	count, err := r.redis.Incr(ctx, key).Result()
	if err != nil {
		return true
	}
	if count == 1 {
		_ = r.redis.Expire(ctx, key, unlockRateWindow).Err()
	}
	return count <= unlockMaxAttempts
}

// pendingClicks henüz worker ile PostgreSQL'e yazılmamış (veya aktarım sürecindeki) tıklama farkını
// ve bir flush işleminin aktif olup olmadığını döner.
func (r *LinkRepository) pendingClicks(ctx context.Context, linkID int64, alias string) (int64, bool) {
	if r.redis == nil {
		return 0, false
	}
	keys := []string{clickDeltasKey}
	procKeys, err := r.redis.SMembers(ctx, clickDeltasProcsKey).Result()
	flushing := err == nil && len(procKeys) > 0
	if flushing {
		keys = append(keys, procKeys...)
	}

	var total int64
	fieldID := ClickDeltaField(linkID)
	for _, key := range keys {
		vals, err := r.redis.HMGet(ctx, key, fieldID, alias).Result()
		if err != nil {
			continue
		}
		for _, v := range vals {
			if str, ok := v.(string); ok {
				if n, err := strconv.ParseInt(str, 10, 64); err == nil && n > 0 {
					total += n
				}
			}
		}
	}
	return total, flushing
}

// RecordClickAndLog hot-path üzerinde atomik Lua betiğiyle max_clicks sınırını aşmadan:
// 1) link:{alias} Hash clicks sayacını günceller,
// 2) clicks:deltas Hash'ine +1 fark yazar,
// 3) logs:queue listesine cihaz log JSON'ını ekler.
// Sınır eşzamanlı isteklerle dolduysa false döner; Redis anahtarı yoksa veya erişilemezse PostgreSQL fallback kullanır.
func (r *LinkRepository) RecordClickAndLog(ctx context.Context, rec LinkRecord, entry LogQueueEntry) bool {
	if r.redis != nil {
		payload, err := json.Marshal(entry)
		if err == nil {
			field := rec.Alias
			if rec.ID > 0 {
				field = ClickDeltaField(rec.ID)
			}
			res, evalErr := r.redis.Eval(
				ctx,
				recordClickLua,
				[]string{linkCacheKey(rec.Alias), clickDeltasKey, logsQueueKey},
				strconv.FormatInt(rec.MaxClicks, 10),
				field,
				string(payload),
			).Int64()
			if evalErr == nil {
				if res == 1 {
					return true
				}
				if res == 0 {
					return false
				}
				// res == -1: önbellekteki link anahtarı eksik; max_clicks denetimi için PostgreSQL fallback'e düş.
			} else {
				log.Printf("router redis lua uyarısı (%s), postgres fallback: %v", rec.Alias, evalErr)
			}
		}
	}

	// Fallback: Redis devre dışıysa veya sınırlı link anahtarı eksikse doğrudan PostgreSQL'e yaz.
	// Sınır denetimine henüz worker ile aktarılmamış tıklamalar da katılır; yoksa sınır aşılabilirdi.
	var pending int64
	if rec.MaxClicks > 0 {
		pending, _ = r.pendingClicks(ctx, rec.ID, rec.Alias)
	}
	rows, err := r.q.IncrementLinkClicks(ctx, db.IncrementLinkClicksParams{Alias: rec.Alias, Clicks: pending})
	if err != nil {
		log.Printf("router postgres tıklama artırma hatası (%s): %v", rec.Alias, err)
		if rec.MaxClicks > 0 {
			return false
		}
	} else if rows == 0 && rec.MaxClicks > 0 {
		return false
	}
	if rec.ID > 0 {
		_, err := r.q.CreateLog(ctx, db.CreateLogParams{
			LinkID:         pgtype.Int8{Int64: rec.ID, Valid: true},
			LinkAlias:      entry.LinkAlias,
			LinkPublicID:   entry.LinkPublicID,
			IpAddress:      entry.IPAddress,
			UserAgent:      entry.UserAgent,
			DeviceType:     entry.DeviceType,
			Browser:        entry.Browser,
			Os:             entry.OS,
			Referer:        entry.Referer,
			AcceptLanguage: entry.AcceptLanguage,
			CountryCode:    entry.CountryCode,
			EventType:      entry.EventType,
		})
		if err != nil {
			log.Printf("router postgres log fallback hatası (link_id=%d): %v", rec.ID, err)
		}
	}
	return true
}
