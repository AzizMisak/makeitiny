package repository

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	"makeitiny/router/internal/db"
)

const (
	linkCacheTTL      = 24 * time.Hour
	reachCacheTTL     = 5 * time.Minute
	clickDeltasKey    = "clicks:deltas"
	logsQueueKey      = "logs:queue"
	incrIfExistsLua   = `if redis.call('EXISTS', KEYS[1]) == 1 then redis.call('HINCRBY', KEYS[1], 'clicks', 1) end return 1`
)

func linkCacheKey(alias string) string {
	return "link:" + alias
}

func reachCacheKey(alias string) string {
	return "reach:" + alias
}

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

	rec := LinkRecord{
		ID:             link.ID,
		PublicID:       link.PublicID,
		UserID:         link.UserID,
		Alias:          link.Alias,
		TargetURL:      link.TargetUrl,
		PasswordHash:   pwHash,
		Clicks:         link.Clicks,
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
		if link.IsAnonymous && link.ExpiresAt.Valid {
			rem := time.Until(link.ExpiresAt.Time)
			if rem > 0 && rem < ttl {
				ttl = rem
			} else if rem <= 0 {
				return rec, nil
			}
		}

		err := r.redis.HSet(ctx, key, map[string]any{
			"id":              strconv.FormatInt(rec.ID, 10),
			"public_id":       rec.PublicID,
			"user_id":         strconv.FormatInt(rec.UserID, 10),
			"alias":           rec.Alias,
			"target_url":      rec.TargetURL,
			"password_hash":   rec.PasswordHash,
			"clicks":          strconv.FormatInt(rec.Clicks, 10),
			"wait_enabled":    strconv.FormatBool(rec.WaitEnabled),
			"wait_seconds":    strconv.FormatInt(int64(rec.WaitSeconds), 10),
			"is_active":       strconv.FormatBool(rec.IsActive),
			"expires_at":      strconv.FormatInt(rec.ExpiresAtUnix, 10),
			"max_clicks":      strconv.FormatInt(rec.MaxClicks, 10),
			"ad_enabled":      strconv.FormatBool(rec.AdEnabled),
			"security_status": rec.SecurityStatus,
			"is_anonymous":    strconv.FormatBool(rec.IsAnonymous),
		}).Err()
		if err != nil {
			log.Printf("router redis HSET uyarısı (%s): %v", key, err)
		} else {
			_ = r.redis.Expire(ctx, key, ttl).Err()
		}
	}

	return rec, nil
}

// IsTargetReachableCached son 5 dakikada hedef URL erişilebilir olarak doğrulandıysa true döner.
func (r *LinkRepository) IsTargetReachableCached(ctx context.Context, alias string) bool {
	if r.redis == nil || alias == "" {
		return false
	}
	n, err := r.redis.Exists(ctx, reachCacheKey(alias)).Result()
	return err == nil && n > 0
}

func (r *LinkRepository) CacheTargetReachable(ctx context.Context, alias string) {
	if r.redis == nil || alias == "" {
		return
	}
	_ = r.redis.Set(ctx, reachCacheKey(alias), "ok", reachCacheTTL).Err()
}

// RecordClickAndLog hot-path üzerinde PostgreSQL'e yazmadan tek Redis Pipeline ile:
// 1) link:{alias} Hash clicks sayacını günceller,
// 2) clicks:deltas Hash'ine +1 fark yazar,
// 3) logs:queue listesine cihaz log JSON'ını ekler.
// Redis erişilemezse veri kaybını önlemek için doğrudan PostgreSQL fallback kullanır.
func (r *LinkRepository) RecordClickAndLog(ctx context.Context, rec LinkRecord, entry LogQueueEntry) {
	if r.redis != nil {
		payload, err := json.Marshal(entry)
		if err == nil {
			pipe := r.redis.Pipeline()
			pipe.Eval(ctx, incrIfExistsLua, []string{linkCacheKey(rec.Alias)})
			pipe.HIncrBy(ctx, clickDeltasKey, rec.Alias, 1)
			pipe.RPush(ctx, logsQueueKey, string(payload))
			if _, pipeErr := pipe.Exec(ctx); pipeErr == nil {
				return
			} else {
				log.Printf("router redis pipeline uyarısı (%s), postgres fallback: %v", rec.Alias, pipeErr)
			}
		}
	}

	// Fallback: Redis devre dışıysa doğrudan PostgreSQL'e yaz
	_ = r.q.IncrementLinkClicks(ctx, rec.Alias)
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
}
