package repository

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"makeitiny/api/internal/db"
)

const linkCacheTTL = 24 * time.Hour

func linkCacheKey(alias string) string {
	return "link:" + alias
}

// UserRepository kullanıcı verisine erişim katmanı.
type UserRepository struct {
	q     *db.Queries
	redis *redis.Client
}

func NewUserRepository(q *db.Queries, rdb *redis.Client) *UserRepository {
	return &UserRepository{q: q, redis: rdb}
}

func (r *UserRepository) Create(ctx context.Context, name, email, passwordHash string) (db.User, error) {
	return r.q.CreateUser(ctx, db.CreateUserParams{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	})
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (db.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *UserRepository) UpdatePassword(ctx context.Context, email, passwordHash string) (int64, error) {
	return r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
}

// Delete önce kullanıcının linklerini tespit eder, PostgreSQL'den kullanıcıyı (CASCADE ile linklerini) siler,
// hemen ardından Redis üzerindeki ilgili link Hash kayıtlarını temizler.
func (r *UserRepository) Delete(ctx context.Context, id int64) (int64, error) {
	links, _ := r.q.ListLinksByUser(ctx, id)

	rows, err := r.q.DeleteUser(ctx, id)
	if err != nil {
		return 0, err
	}

	if r.redis != nil && len(links) > 0 {
		keys := make([]string, 0, len(links))
		for _, l := range links {
			keys = append(keys, linkCacheKey(l.Alias))
		}
		if err := r.redis.Del(ctx, keys...).Err(); err != nil {
			log.Printf("redis kullanıcı linkleri silme uyarısı: %v", err)
		}
	}

	return rows, nil
}

// LinkRepository link verisine erişim katmanı (PostgreSQL + Redis Hash).
type LinkRepository struct {
	q     *db.Queries
	redis *redis.Client
}

func NewLinkRepository(q *db.Queries, rdb *redis.Client) *LinkRepository {
	return &LinkRepository{q: q, redis: rdb}
}

// syncToRedis link verisini NoSQL dokümanı/tablosu gibi Redis Hash (HSET) olarak kaydeder.
func (r *LinkRepository) syncToRedis(ctx context.Context, link db.Link) {
	if r.redis == nil {
		return
	}

	key := linkCacheKey(link.Alias)
	pwHash := ""
	if link.PasswordHash.Valid {
		pwHash = link.PasswordHash.String
	}

	err := r.redis.HSet(ctx, key, map[string]any{
		"id":            strconv.FormatInt(link.ID, 10),
		"user_id":       strconv.FormatInt(link.UserID, 10),
		"alias":         link.Alias,
		"target_url":    link.TargetUrl,
		"password_hash": pwHash,
		"clicks":        strconv.FormatInt(link.Clicks, 10),
		"wait_enabled":  strconv.FormatBool(link.WaitEnabled),
		"wait_seconds":  strconv.FormatInt(int64(link.WaitSeconds), 10),
	}).Err()
	if err != nil {
		log.Printf("redis HSET uyarısı (%s): %v", key, err)
		return
	}

	if err := r.redis.Expire(ctx, key, linkCacheTTL).Err(); err != nil {
		log.Printf("redis EXPIRE uyarısı (%s): %v", key, err)
	}
}

func (r *LinkRepository) removeFromRedis(ctx context.Context, alias string) {
	if r.redis == nil || alias == "" {
		return
	}
	if err := r.redis.Del(ctx, linkCacheKey(alias)).Err(); err != nil {
		log.Printf("redis DEL uyarısı (%s): %v", alias, err)
	}
}

// Create önce PostgreSQL'e kaydeder, hemen ardından Redis Hash'e yazar.
func (r *LinkRepository) Create(ctx context.Context, params db.CreateLinkParams) (db.Link, error) {
	link, err := r.q.CreateLink(ctx, params)
	if err != nil {
		return db.Link{}, err
	}

	r.syncToRedis(ctx, link)
	return link, nil
}

func (r *LinkRepository) ListByUser(ctx context.Context, userID int64) ([]db.Link, error) {
	return r.q.ListLinksByUser(ctx, userID)
}

func (r *LinkRepository) GetByIDAndUser(ctx context.Context, id, userID int64) (db.Link, error) {
	return r.q.GetLinkByIDAndUser(ctx, db.GetLinkByIDAndUserParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *LinkRepository) AliasExists(ctx context.Context, alias string) (bool, error) {
	return r.q.AliasExists(ctx, alias)
}

// Update önce PostgreSQL'de günceller, hemen ardından Redis Hash kaydını yeniler (alias değiştiyse eskiyi siler).
func (r *LinkRepository) Update(ctx context.Context, oldAlias string, params db.UpdateLinkParams) (db.Link, error) {
	updated, err := r.q.UpdateLink(ctx, params)
	if err != nil {
		return db.Link{}, err
	}

	if oldAlias != "" && oldAlias != updated.Alias {
		r.removeFromRedis(ctx, oldAlias)
	}
	r.syncToRedis(ctx, updated)

	return updated, nil
}

// Delete önce PostgreSQL'den siler, hemen ardından Redis Hash kaydını siler.
func (r *LinkRepository) Delete(ctx context.Context, id, userID int64) (int64, error) {
	existing, err := r.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	rows, err := r.q.DeleteLink(ctx, db.DeleteLinkParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return 0, err
	}

	if rows > 0 {
		r.removeFromRedis(ctx, existing.Alias)
	}

	return rows, nil
}

func (r *LinkRepository) ListLogsByLink(ctx context.Context, linkID int64) ([]db.Log, error) {
	return r.q.ListLogsByLink(ctx, linkID)
}
