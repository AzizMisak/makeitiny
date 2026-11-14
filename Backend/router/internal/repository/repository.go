package repository

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"makeitiny/router/internal/db"
)

const linkCacheTTL = 24 * time.Hour

func linkCacheKey(alias string) string {
	return "link:" + alias
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
}

// LinkRepository link verisine erişim katmanı (Redis Hash + PostgreSQL).
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
		IsActive:       link.IsActive,
		ExpiresAtUnix:  expiresAtUnix,
		MaxClicks:      maxClicks,
		AdEnabled:      link.AdEnabled,
		SecurityStatus: link.SecurityStatus,
	}

	if r.redis != nil {
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
		}).Err()
		if err != nil {
			log.Printf("router redis HSET uyarısı (%s): %v", key, err)
		} else {
			_ = r.redis.Expire(ctx, key, linkCacheTTL).Err()
		}
	}

	return rec, nil
}

// IncrementClicks önce PostgreSQL üzerinde tıklama sayısını artırır,
// hemen ardından Redis Hash içindeki clicks alanını atomik olarak (HINCRBY) artırır.
func (r *LinkRepository) IncrementClicks(ctx context.Context, alias string) error {
	if err := r.q.IncrementLinkClicks(ctx, alias); err != nil {
		return err
	}

	if r.redis != nil {
		key := linkCacheKey(alias)
		exists, err := r.redis.Exists(ctx, key).Result()
		if err == nil && exists > 0 {
			if err := r.redis.HIncrBy(ctx, key, "clicks", 1).Err(); err != nil {
				log.Printf("router redis HINCRBY uyarısı (%s): %v", key, err)
			}
		}
	}

	return nil
}

// CreateLog tıklayan cihazın bilgilerini PostgreSQL logs tablosuna kaydeder.
func (r *LinkRepository) CreateLog(ctx context.Context, params db.CreateLogParams) error {
	_, err := r.q.CreateLog(ctx, params)
	if err != nil {
		log.Printf("router log kaydı uyarısı (link_id=%d): %v", params.LinkID, err)
	}
	return err
}
