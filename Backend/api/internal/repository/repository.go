package repository

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	"makeitiny/api/internal/db"
)

const linkCacheTTL = 24 * time.Hour

func linkCacheKey(alias string) string {
	return "link:" + alias
}

// UserRepository kullanıcı, rol ve paket verisine erişim katmanı.
type UserRepository struct {
	q     *db.Queries
	redis *redis.Client
}

func NewUserRepository(q *db.Queries, rdb *redis.Client) *UserRepository {
	return &UserRepository{q: q, redis: rdb}
}

func (r *UserRepository) Create(ctx context.Context, name, email, passwordHash string) (db.User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return db.User{}, err
	}

	_ = r.q.AssignRoleByName(ctx, db.AssignRoleByNameParams{
		UserID:   u.ID,
		RoleName: "user",
	})

	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (db.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *UserRepository) GetRoles(ctx context.Context, userID int64) ([]string, error) {
	roles, err := r.q.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return []string{"user"}, nil
	}
	return roles, nil
}

func (r *UserRepository) SetRole(ctx context.Context, userID int64, roleName string) error {
	if err := r.q.ClearUserNonOwnerRoles(ctx, userID); err != nil {
		return err
	}
	return r.q.AssignRoleByName(ctx, db.AssignRoleByNameParams{
		UserID:   userID,
		RoleName: roleName,
	})
}

func (r *UserRepository) GetPlanByID(ctx context.Context, planID int32) (db.Plan, error) {
	return r.q.GetPlanByID(ctx, planID)
}

func (r *UserRepository) ListPlans(ctx context.Context) ([]db.Plan, error) {
	return r.q.ListPlans(ctx)
}

func (r *UserRepository) UpdatePlan(ctx context.Context, params db.UpdatePlanParams) (db.Plan, error) {
	return r.q.UpdatePlan(ctx, params)
}

func (r *UserRepository) CountLinks(ctx context.Context, userID int64) (int64, error) {
	return r.q.CountLinksByUser(ctx, userID)
}

func (r *UserRepository) UpdatePassword(ctx context.Context, email, passwordHash string) (int64, error) {
	return r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
}

func (r *UserRepository) invalidateUserLinksCache(ctx context.Context, userID int64) {
	if r.redis == nil {
		return
	}
	links, err := r.q.ListLinksByUser(ctx, userID)
	if err != nil || len(links) == 0 {
		return
	}
	keys := make([]string, 0, len(links))
	for _, l := range links {
		keys = append(keys, linkCacheKey(l.Alias))
	}
	if err := r.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf("redis kullanıcı linkleri önbellek temizleme uyarısı: %v", err)
	}
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

func (r *UserRepository) CountAdminUsers(ctx context.Context, search string) (int64, error) {
	return r.q.CountAdminUsers(ctx, search)
}

func (r *UserRepository) ListAdminUsersPaginated(ctx context.Context, search string, limit, offset int32) ([]db.User, error) {
	return r.q.ListAdminUsersPaginated(ctx, db.ListAdminUsersPaginatedParams{
		Search: search,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *UserRepository) AdminUpdateUser(ctx context.Context, params db.AdminUpdateUserParams) (db.User, error) {
	u, err := r.q.AdminUpdateUser(ctx, params)
	if err != nil {
		return db.User{}, err
	}
	r.invalidateUserLinksCache(ctx, u.ID)
	return u, nil
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

	var expiresAtUnix int64
	if link.ExpiresAt.Valid {
		expiresAtUnix = link.ExpiresAt.Time.Unix()
	}

	var maxClicks int64
	if link.MaxClicks.Valid {
		maxClicks = link.MaxClicks.Int64
	}

	err := r.redis.HSet(ctx, key, map[string]any{
		"id":              strconv.FormatInt(link.ID, 10),
		"public_id":       link.PublicID,
		"user_id":         strconv.FormatInt(link.UserID, 10),
		"alias":           link.Alias,
		"target_url":      link.TargetUrl,
		"password_hash":   pwHash,
		"clicks":          strconv.FormatInt(link.Clicks, 10),
		"wait_enabled":    strconv.FormatBool(link.WaitEnabled),
		"wait_seconds":    strconv.FormatInt(int64(link.WaitSeconds), 10),
		"is_active":       strconv.FormatBool(link.IsActive),
		"expires_at":      strconv.FormatInt(expiresAtUnix, 10),
		"max_clicks":      strconv.FormatInt(maxClicks, 10),
		"ad_enabled":      strconv.FormatBool(link.AdEnabled),
		"security_status": link.SecurityStatus,
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

func (r *LinkRepository) ListByUserPaginated(ctx context.Context, userID int64, limit, offset int32) ([]db.Link, error) {
	return r.q.ListLinksByUserPaginated(ctx, db.ListLinksByUserPaginatedParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *LinkRepository) GetUserStats(ctx context.Context, userID int64) (db.GetUserLinkStatsRow, error) {
	return r.q.GetUserLinkStats(ctx, userID)
}

func (r *LinkRepository) ListTopByUser(ctx context.Context, userID int64, limit int32) ([]db.Link, error) {
	return r.q.ListTopLinksByUser(ctx, db.ListTopLinksByUserParams{
		UserID: userID,
		Limit:  limit,
	})
}

func (r *LinkRepository) GetByIDAndUser(ctx context.Context, publicID string, userID int64) (db.Link, error) {
	return r.q.GetLinkByIDAndUser(ctx, db.GetLinkByIDAndUserParams{
		PublicID: publicID,
		UserID:   userID,
	})
}

func (r *LinkRepository) GetByPublicID(ctx context.Context, publicID string) (db.Link, error) {
	return r.q.GetLinkByPublicID(ctx, publicID)
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
func (r *LinkRepository) Delete(ctx context.Context, publicID string, userID int64) (int64, error) {
	existing, err := r.GetByIDAndUser(ctx, publicID, userID)
	if err != nil {
		return 0, err
	}

	rows, err := r.q.DeleteLink(ctx, db.DeleteLinkParams{
		ID:     existing.ID,
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

func (r *LinkRepository) CountLogsByLink(ctx context.Context, linkID int64, since pgtype.Timestamptz) (int64, error) {
	return r.q.CountLogsByLink(ctx, db.CountLogsByLinkParams{
		LinkID:    linkID,
		ClickedAt: since,
	})
}

func (r *LinkRepository) ListLogsByLinkPaginated(ctx context.Context, linkID int64, since pgtype.Timestamptz, limit, offset int32) ([]db.Log, error) {
	return r.q.ListLogsByLinkPaginated(ctx, db.ListLogsByLinkPaginatedParams{
		LinkID:    linkID,
		ClickedAt: since,
		Limit:     limit,
		Offset:    offset,
	})
}

// --- Admin Operations ---

func (r *LinkRepository) GetPlatformStats(ctx context.Context) (db.GetPlatformStatsRow, error) {
	return r.q.GetPlatformStats(ctx)
}

func (r *LinkRepository) CountAdminLinks(ctx context.Context, search, status string) (int64, error) {
	return r.q.CountAdminLinks(ctx, db.CountAdminLinksParams{
		Search: search,
		Status: status,
	})
}

func (r *LinkRepository) ListAdminLinksPaginated(ctx context.Context, search, status string, limit, offset int32) ([]db.AdminLinkRow, error) {
	return r.q.ListAdminLinksPaginated(ctx, db.ListAdminLinksPaginatedParams{
		Search: search,
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *LinkRepository) AdminUpdate(ctx context.Context, oldAlias string, params db.AdminUpdateLinkParams) (db.Link, error) {
	updated, err := r.q.AdminUpdateLink(ctx, params)
	if err != nil {
		return db.Link{}, err
	}

	if oldAlias != "" && oldAlias != updated.Alias {
		r.removeFromRedis(ctx, oldAlias)
	}
	r.syncToRedis(ctx, updated)

	return updated, nil
}

func (r *LinkRepository) AdminDelete(ctx context.Context, publicID string) (int64, error) {
	existing, err := r.GetByPublicID(ctx, publicID)
	if err != nil {
		return 0, err
	}

	rows, err := r.q.AdminDeleteLink(ctx, existing.ID)
	if err != nil {
		return 0, err
	}

	if rows > 0 {
		r.removeFromRedis(ctx, existing.Alias)
	}

	return rows, nil
}
