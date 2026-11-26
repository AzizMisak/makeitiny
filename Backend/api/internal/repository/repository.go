package repository

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"makeitiny/api/internal/db"
	"makeitiny/api/internal/qr"
)

const (
	linkCacheTTL        = 24 * time.Hour
	clickDeltasProcsKey = "clicks:deltas:procs"
)

func linkCacheKey(alias string) string {
	return "link:" + alias
}

// UserRepository kullanıcı, rol, paket ve refresh token verisine erişim katmanı.
type UserRepository struct {
	pool    *pgxpool.Pool
	q       *db.Queries
	redis   *redis.Client
	qrStore *qr.Store
}

func NewUserRepository(pool *pgxpool.Pool, q *db.Queries, rdb *redis.Client, qrStore *qr.Store) *UserRepository {
	return &UserRepository{pool: pool, q: q, redis: rdb, qrStore: qrStore}
}

func (r *UserRepository) Create(ctx context.Context, publicID, name, email, passwordHash string) (db.User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		PublicID:     publicID,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return db.User{}, err
	}

	_ = r.q.AssignRoleByName(ctx, db.AssignRoleByNameParams{
		UserID: u.ID,
		Name:   "user",
	})

	return u, nil
}

func (r *UserRepository) UpdateName(ctx context.Context, userID int64, name string) (db.User, error) {
	return r.q.UpdateUserName(ctx, db.UpdateUserNameParams{ID: userID, Name: name})
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (db.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *UserRepository) GetByPublicID(ctx context.Context, publicID string) (db.User, error) {
	return r.q.GetUserByPublicID(ctx, publicID)
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
	if r.pool == nil {
		if err := r.q.ClearUserNonOwnerRoles(ctx, userID); err != nil {
			return err
		}
		return r.q.AssignRoleByName(ctx, db.AssignRoleByNameParams{
			UserID: userID,
			Name:   roleName,
		})
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.q.WithTx(tx)
	if err := qtx.ClearUserNonOwnerRoles(ctx, userID); err != nil {
		return err
	}
	if err := qtx.AssignRoleByName(ctx, db.AssignRoleByNameParams{
		UserID: userID,
		Name:   roleName,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *UserRepository) GetPlanByID(ctx context.Context, planID int32) (db.GetPlanByIDRow, error) {
	return r.q.GetPlanByID(ctx, planID)
}

func (r *UserRepository) ListPlans(ctx context.Context) ([]db.ListPlansRow, error) {
	return r.q.ListPlans(ctx)
}

func (r *UserRepository) UpdatePlan(ctx context.Context, params db.UpdatePlanParams) (db.UpdatePlanRow, error) {
	return r.q.UpdatePlan(ctx, params)
}

func (r *UserRepository) CountLinks(ctx context.Context, userID int64) (int64, error) {
	return r.q.CountLinksByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
}

// ChangePlan kullanıcının paketini değiştirir ve linklerinin askı durumunu yeni sınıra göre aynı işlemde
// yeniden hesaplar. CreateWithQuota ile aynı danışma kilidi alınır; eşzamanlı link oluşturma araya giremez.
// Askı yalnızca PostgreSQL'de tutulur (yönlendirmeyi etkilemez), bu yüzden Redis güncellenmez.
func (r *UserRepository) ChangePlan(ctx context.Context, userID int64, planID, maxLinks int32, graceUntil time.Time) (db.User, error) {
	apply := func(q *db.Queries) (db.User, error) {
		user, err := q.UpdateUserPlan(ctx, db.UpdateUserPlanParams{ID: userID, PlanID: planID})
		if err != nil {
			return db.User{}, err
		}
		err = q.ApplyLinkPlanGrace(ctx, db.ApplyLinkPlanGraceParams{
			UserID:     userID,
			MaxLinks:   maxLinks,
			GraceUntil: pgtype.Timestamptz{Time: graceUntil, Valid: true},
		})
		return user, err
	}
	if r.pool == nil {
		return apply(r.q)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return db.User{}, err
	}

	user, err := apply(r.q.WithTx(tx))
	if err != nil {
		return db.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return db.User{}, err
	}
	return user, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, email, passwordHash string) (int64, error) {
	return r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
}

// InvalidateUserLinksCache kullanıcının bütün link önbelleklerini siler; router bir sonraki ziyarette
// sahibin güncel durumunu (ban dahil) PostgreSQL'den okur.
func (r *UserRepository) InvalidateUserLinksCache(ctx context.Context, userID int64) {
	if r.redis == nil {
		return
	}
	links, err := r.q.ListLinkIDsByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil || len(links) == 0 {
		return
	}
	keys := make([]string, 0, len(links)*2)
	for _, l := range links {
		keys = append(keys, linkCacheKey(l.Alias), "reach:"+l.Alias)
	}
	if err := r.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf("redis kullanıcı linkleri önbellek temizleme uyarısı: %v", err)
	}
}

// Delete önce kullanıcının linklerini tespit eder, PostgreSQL'den kullanıcıyı siler,
// hemen ardından Redis üzerindeki link Hash kayıtlarını ve diskteki QR dosyalarını temizler.
func (r *UserRepository) Delete(ctx context.Context, id int64) (int64, error) {
	links, err := r.q.ListLinksByUser(ctx, pgtype.Int8{Int64: id, Valid: true})
	if err != nil {
		return 0, err
	}

	rows, err := r.q.DeleteUser(ctx, id)
	if err != nil {
		return 0, err
	}
	if rows == 0 {
		return 0, nil
	}

	if r.redis != nil && len(links) > 0 {
		keys := make([]string, 0, len(links)*2)
		for _, l := range links {
			keys = append(keys, linkCacheKey(l.Alias), "reach:"+l.Alias)
		}
		if err := r.redis.Del(ctx, keys...).Err(); err != nil {
			log.Printf("redis kullanıcı linkleri silme uyarısı: %v", err)
		}
	}

	if r.qrStore != nil {
		for _, l := range links {
			if l.QrCodePath != "" {
				r.qrStore.Delete(l.QrCodePath)
			}
		}
	}

	return rows, nil
}

func (r *UserRepository) CountAdminUsers(ctx context.Context, search string) (int64, error) {
	return r.q.CountAdminUsers(ctx, search)
}

func (r *UserRepository) ListAdminUsersPaginated(ctx context.Context, search string, limit, offset int32) ([]db.User, error) {
	return r.q.ListAdminUsersPaginated(ctx, db.ListAdminUsersPaginatedParams{
		Column1: search,
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *UserRepository) AdminUpdateUser(ctx context.Context, params db.AdminUpdateUserParams) (db.User, error) {
	u, err := r.q.AdminUpdateUser(ctx, params)
	if err != nil {
		return db.User{}, err
	}
	r.InvalidateUserLinksCache(ctx, u.ID)
	return u, nil
}

func (r *UserRepository) UnbanUser(ctx context.Context, userID int64) (db.User, error) {
	u, err := r.q.UnbanUser(ctx, userID)
	if err != nil {
		return db.User{}, err
	}
	r.InvalidateUserLinksCache(ctx, u.ID)
	return u, nil
}

// --- JWT / Refresh Token Revocation ---

func (r *UserRepository) CreateRefreshToken(ctx context.Context, params db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	return r.q.CreateRefreshToken(ctx, params)
}

func (r *UserRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	return r.q.GetRefreshTokenByHash(ctx, tokenHash)
}

// RevokeActiveRefreshToken token henüz kullanılmamışsa tek sorguda iptal edip döner (atomik rotasyon).
func (r *UserRepository) RevokeActiveRefreshToken(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	return r.q.RevokeActiveRefreshToken(ctx, tokenHash)
}

// DeleteRefreshToken çıkışta token'ı siler. İptal işareti yalnızca rotasyon içindir; silinen token
// sonradan gelirse "bilinmeyen token" sayılır ve tekrar kullanım (çalıntı) alarmı tetiklenmez.
func (r *UserRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	return r.q.DeleteRefreshTokenByHash(ctx, tokenHash)
}

// DeleteAllRefreshTokens kullanıcının bütün oturum yenileme token'larını siler (parola değişimi, askı,
// tekrar kullanım tespiti). Eski cihazlardan gelen token'lar sonradan yeni oturumu düşürmez.
func (r *UserRepository) DeleteAllRefreshTokens(ctx context.Context, userID int64) error {
	return r.q.DeleteUserRefreshTokens(ctx, userID)
}

func (r *UserRepository) BlacklistJTI(ctx context.Context, jti string, ttl time.Duration) {
	if r.redis == nil || jti == "" || ttl <= 0 {
		return
	}
	if err := r.redis.Set(ctx, "jwt:blacklist:"+jti, "1", ttl).Err(); err != nil {
		log.Printf("redis JTI blacklist uyarısı: %v", err)
	}
}

func (r *UserRepository) IsJTIBlacklisted(ctx context.Context, jti string) bool {
	if r.redis == nil || jti == "" {
		return false
	}
	n, err := r.redis.Exists(ctx, "jwt:blacklist:"+jti).Result()
	return err == nil && n > 0
}

func (r *UserRepository) RevokeUserTokensBefore(ctx context.Context, userID int64) {
	if r.redis == nil || userID <= 0 {
		return
	}
	key := "user:revoked_before:" + strconv.FormatInt(userID, 10)
	if err := r.redis.Set(ctx, key, strconv.FormatInt(time.Now().UnixNano(), 10), 20*time.Minute).Err(); err != nil {
		log.Printf("redis user revoked_before uyarısı: %v", err)
	}
}

func (r *UserRepository) GetUserRevokedBefore(ctx context.Context, userID int64) int64 {
	if r.redis == nil || userID <= 0 {
		return 0
	}
	val, err := r.redis.Get(ctx, "user:revoked_before:"+strconv.FormatInt(userID, 10)).Int64()
	if err != nil {
		return 0
	}
	return val
}

// LinkRepository link verisine erişim katmanı (PostgreSQL + Redis Hash).
type LinkRepository struct {
	pool    *pgxpool.Pool
	q       *db.Queries
	redis   *redis.Client
	qrStore *qr.Store
}

func NewLinkRepository(pool *pgxpool.Pool, q *db.Queries, rdb *redis.Client, qrStore *qr.Store) *LinkRepository {
	return &LinkRepository{pool: pool, q: q, redis: rdb, qrStore: qrStore}
}

// syncToRedis link verisini NoSQL dokümanı/tablosu gibi Redis Hash (HSET) olarak kaydeder.
func (r *LinkRepository) syncToRedis(ctx context.Context, link db.Link) {
	if r.redis == nil {
		return
	}

	ttl := linkCacheTTL
	if link.IsAnonymous && link.ExpiresAt.Valid {
		rem := time.Until(link.ExpiresAt.Time)
		if rem <= 0 {
			return
		}
		if rem < ttl {
			ttl = rem
		}
	}

	key := linkCacheKey(link.Alias)
	pwHash := ""
	if link.PasswordHash.Valid {
		pwHash = link.PasswordHash.String
	}

	var userID int64
	if link.UserID.Valid {
		userID = link.UserID.Int64
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
		"user_id":         strconv.FormatInt(userID, 10),
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
		"is_anonymous":    strconv.FormatBool(link.IsAnonymous),
	}).Err()
	if err != nil {
		log.Printf("redis HSET uyarısı (%s): %v", key, err)
		return
	}

	if err := r.redis.Expire(ctx, key, ttl).Err(); err != nil {
		log.Printf("redis EXPIRE uyarısı (%s): %v", key, err)
	}
}

func (r *LinkRepository) removeFromRedis(ctx context.Context, alias string) {
	if r.redis == nil || alias == "" {
		return
	}
	if err := r.redis.Del(ctx, linkCacheKey(alias), "reach:"+alias).Err(); err != nil {
		log.Printf("redis DEL uyarısı (%s): %v", alias, err)
	}
}

// ClickDeltaField router'ın clicks:deltas içindeki alan adı. Farklar link kimliğiyle tutulur;
// takma ad değişse de tıklamalar doğru linke yazılır.
func ClickDeltaField(linkID int64) string {
	return "id:" + strconv.FormatInt(linkID, 10)
}

// LinkRef bekleyen tıklama sorgusu için link kimliği ve (eski kayıtlar için) takma adı.
type LinkRef struct {
	ID    int64
	Alias string
}

func (r *LinkRepository) clickDeltaKeys(ctx context.Context) []string {
	keys := []string{"clicks:deltas"}
	if r.redis == nil {
		return keys
	}
	procKeys, err := r.redis.SMembers(ctx, clickDeltasProcsKey).Result()
	if err != nil {
		return keys
	}
	for _, k := range procKeys {
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

// PendingClicks henüz worker ile PostgreSQL'e yazılmamış (ve aktarım sürecindeki) tıklama farklarını okur.
// Eski sürümden kalan takma ad anahtarlı farklar da toplanır.
func (r *LinkRepository) PendingClicks(ctx context.Context, refs []LinkRef) map[int64]int64 {
	result := make(map[int64]int64, len(refs))
	if r.redis == nil || len(refs) == 0 {
		return result
	}
	fields := make([]string, 0, len(refs)*2)
	for _, ref := range refs {
		fields = append(fields, ClickDeltaField(ref.ID), ref.Alias)
	}
	for _, key := range r.clickDeltaKeys(ctx) {
		vals, err := r.redis.HMGet(ctx, key, fields...).Result()
		if err != nil {
			continue
		}
		for i, v := range vals {
			s, ok := v.(string)
			if !ok {
				continue
			}
			if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
				result[refs[i/2].ID] += n
			}
		}
	}
	return result
}

// TotalPendingClicks platform genelinde henüz yazılmamış bütün tıklama farklarının toplamı.
func (r *LinkRepository) TotalPendingClicks(ctx context.Context) int64 {
	if r.redis == nil {
		return 0
	}
	var total int64
	for _, key := range r.clickDeltaKeys(ctx) {
		vals, err := r.redis.HVals(ctx, key).Result()
		if err != nil {
			continue
		}
		for _, v := range vals {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
				total += n
			}
		}
	}
	return total
}

func (r *LinkRepository) ListIDsByUser(ctx context.Context, userID int64) ([]LinkRef, error) {
	rows, err := r.q.ListLinkIDsByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	refs := make([]LinkRef, len(rows))
	for i, row := range rows {
		refs[i] = LinkRef{ID: row.ID, Alias: row.Alias}
	}
	return refs, nil
}

// CheckAndSetAnonRateLimit anonim link için IP başına 24 saatlik hız sınırını kontrol eder.
func (r *LinkRepository) CheckAndSetAnonRateLimit(ctx context.Context, ip string) (bool, error) {
	if ip == "" {
		ip = "unknown"
	}

	if r.redis != nil {
		key := "anon:limit:" + ip
		ok, err := r.redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
		if err == nil {
			return ok, nil
		}
		log.Printf("redis anon rate limit uyarısı: %v", err)
	}

	count, err := r.q.CountRecentAnonymousLinksByIP(ctx, ip)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *LinkRepository) ReleaseAnonRateLimit(ctx context.Context, ip string) {
	if r.redis == nil {
		return
	}
	if ip == "" {
		ip = "unknown"
	}
	_ = r.redis.Del(ctx, "anon:limit:"+ip).Err()
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

// CreateWithQuota aynı kullanıcının eşzamanlı isteklerinin paket link sınırını aşmaması için
// işlem içi danışma kilidiyle kotayı denetleyip linki oluşturur.
func (r *LinkRepository) CreateWithQuota(ctx context.Context, userID int64, maxLinks int32, params db.CreateLinkParams) (db.Link, bool, error) {
	if r.pool == nil || maxLinks < 0 {
		link, err := r.Create(ctx, params)
		return link, err == nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Link{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return db.Link{}, false, err
	}

	qtx := r.q.WithTx(tx)
	count, err := qtx.CountLinksByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return db.Link{}, false, err
	}
	if count >= int64(maxLinks) {
		return db.Link{}, false, nil
	}

	link, err := qtx.CreateLink(ctx, params)
	if err != nil {
		return db.Link{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return db.Link{}, false, err
	}

	r.syncToRedis(ctx, link)
	return link, true, nil
}

// CreateAnonymous önce PostgreSQL'e anonim linki kaydeder, hemen ardından Redis Hash'e yazar.
func (r *LinkRepository) CreateAnonymous(ctx context.Context, params db.CreateAnonymousLinkParams) (db.Link, error) {
	link, err := r.q.CreateAnonymousLink(ctx, params)
	if err != nil {
		return db.Link{}, err
	}

	r.syncToRedis(ctx, link)
	return link, nil
}

func (r *LinkRepository) UpdateQRPath(ctx context.Context, linkID int64, qrPath string) (db.Link, error) {
	return r.q.UpdateLinkQRPath(ctx, db.UpdateLinkQRPathParams{
		ID:         linkID,
		QrCodePath: qrPath,
	})
}

func (r *LinkRepository) ListByUser(ctx context.Context, userID int64) ([]db.Link, error) {
	return r.q.ListLinksByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
}

func (r *LinkRepository) ListByUserPaginated(ctx context.Context, userID int64, limit, offset int32) ([]db.Link, error) {
	return r.q.ListLinksByUserPaginated(ctx, db.ListLinksByUserPaginatedParams{
		UserID: pgtype.Int8{Int64: userID, Valid: true},
		Limit:  limit,
		Offset: offset,
	})
}

func (r *LinkRepository) GetUserStats(ctx context.Context, userID int64) (db.GetUserLinkStatsRow, error) {
	return r.q.GetUserLinkStats(ctx, pgtype.Int8{Int64: userID, Valid: true})
}

func (r *LinkRepository) ListTopByUser(ctx context.Context, userID int64, limit int32) ([]db.Link, error) {
	return r.q.ListTopLinksByUser(ctx, db.ListTopLinksByUserParams{
		UserID: pgtype.Int8{Int64: userID, Valid: true},
		Limit:  limit,
	})
}

func (r *LinkRepository) GetByIDAndUser(ctx context.Context, publicID string, userID int64) (db.Link, error) {
	return r.q.GetLinkByIDAndUser(ctx, db.GetLinkByIDAndUserParams{
		PublicID: publicID,
		UserID:   pgtype.Int8{Int64: userID, Valid: true},
	})
}

func (r *LinkRepository) GetByPublicID(ctx context.Context, publicID string) (db.Link, error) {
	return r.q.GetLinkByPublicID(ctx, publicID)
}

func (r *LinkRepository) AliasExists(ctx context.Context, alias string) (bool, error) {
	return r.q.AliasExists(ctx, alias)
}

// Update önce PostgreSQL'de günceller, hemen ardından Redis'teki link ve erişilebilirlik kayıtlarını siler.
// Kayıt yeniden yazılmaz: router bir sonraki ziyarette sahibin ban durumunu ve bekleyen tıklamaları da
// hesaba katarak PostgreSQL'den okur. Böylece eski tıklama sayısı ya da ban dışı bir durum önbelleğe yazılmaz.
func (r *LinkRepository) Update(ctx context.Context, oldAlias string, params db.UpdateLinkParams) (db.Link, error) {
	updated, err := r.q.UpdateLink(ctx, params)
	if err != nil {
		return db.Link{}, err
	}

	r.removeFromRedis(ctx, oldAlias)
	r.removeFromRedis(ctx, updated.Alias)

	return updated, nil
}

// Delete önce PostgreSQL'den siler, hemen ardından Redis Hash kaydını ve diskteki QR dosyasını siler.
func (r *LinkRepository) Delete(ctx context.Context, publicID string, userID int64) (int64, error) {
	existing, err := r.GetByIDAndUser(ctx, publicID, userID)
	if err != nil {
		return 0, err
	}

	rows, err := r.q.DeleteLink(ctx, db.DeleteLinkParams{
		ID:     existing.ID,
		UserID: pgtype.Int8{Int64: userID, Valid: true},
	})
	if err != nil {
		return 0, err
	}

	if rows > 0 {
		r.removeFromRedis(ctx, existing.Alias)
		if r.qrStore != nil && existing.QrCodePath != "" {
			r.qrStore.Delete(existing.QrCodePath)
		}
	}

	return rows, nil
}

func (r *LinkRepository) CountLogsByLink(ctx context.Context, linkID int64, since pgtype.Timestamptz) (int64, error) {
	return r.q.CountLogsByLink(ctx, db.CountLogsByLinkParams{
		LinkID:    pgtype.Int8{Int64: linkID, Valid: true},
		ClickedAt: since,
	})
}

func (r *LinkRepository) ListLogsByLinkPaginated(ctx context.Context, linkID int64, since pgtype.Timestamptz, limit, offset int32) ([]db.Log, error) {
	return r.q.ListLogsByLinkPaginated(ctx, db.ListLogsByLinkPaginatedParams{
		LinkID:    pgtype.Int8{Int64: linkID, Valid: true},
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
		Column1: search,
		Column2: status,
	})
}

func (r *LinkRepository) ListAdminLinksPaginated(ctx context.Context, search, status string, limit, offset int32) ([]db.ListAdminLinksPaginatedRow, error) {
	return r.q.ListAdminLinksPaginated(ctx, db.ListAdminLinksPaginatedParams{
		Column1: search,
		Column2: status,
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *LinkRepository) GetAdminLinkByPublicID(ctx context.Context, publicID string) (db.GetAdminLinkByPublicIDRow, error) {
	return r.q.GetAdminLinkByPublicID(ctx, publicID)
}

func (r *LinkRepository) AdminUpdate(ctx context.Context, oldAlias string, params db.AdminUpdateLinkParams) (db.Link, error) {
	updated, err := r.q.AdminUpdateLink(ctx, params)
	if err != nil {
		return db.Link{}, err
	}

	// Kullanıcı güncellemesiyle aynı: önbellek silinir, router doğru durumu PostgreSQL'den okur.
	r.removeFromRedis(ctx, oldAlias)
	r.removeFromRedis(ctx, updated.Alias)

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
		if r.qrStore != nil && existing.QrCodePath != "" {
			r.qrStore.Delete(existing.QrCodePath)
		}
	}

	return rows, nil
}
