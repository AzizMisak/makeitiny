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
	ID           int64
	UserID       int64
	Alias        string
	TargetURL    string
	PasswordHash string
	Clicks       int64
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
			return LinkRecord{
				ID:           id,
				UserID:       userID,
				Alias:        alias,
				TargetURL:    vals["target_url"],
				PasswordHash: vals["password_hash"],
				Clicks:       clicks,
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

	rec := LinkRecord{
		ID:           link.ID,
		UserID:       link.UserID,
		Alias:        link.Alias,
		TargetURL:    link.TargetUrl,
		PasswordHash: pwHash,
		Clicks:       link.Clicks,
	}

	if r.redis != nil {
		err := r.redis.HSet(ctx, key, map[string]any{
			"id":            strconv.FormatInt(rec.ID, 10),
			"user_id":       strconv.FormatInt(rec.UserID, 10),
			"alias":         rec.Alias,
			"target_url":    rec.TargetURL,
			"password_hash": rec.PasswordHash,
			"clicks":        strconv.FormatInt(rec.Clicks, 10),
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
