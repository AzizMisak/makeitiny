package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"makeitiny/api/internal/auth"
	"makeitiny/api/internal/db"
	"makeitiny/api/internal/qr"
	"makeitiny/api/internal/repository"
)

var (
	ErrEmailTaken         = errors.New("bu e-posta adresi zaten kullanımda")
	ErrInvalidCredentials = errors.New("e-posta veya parola hatalı")
	ErrUserNotFound       = errors.New("kullanıcı bulunamadı")
	ErrUserSuspended      = errors.New("hesabınız askıya alınmıştır")
	ErrForbidden          = errors.New("bu işlem için yetkiniz bulunmuyor")
	ErrInvalidEmail       = errors.New("geçerli bir e-posta adresi giriniz")
	ErrPasswordTooShort   = errors.New("parola en az 4 karakter olmalıdır")
	ErrInvalidInput       = errors.New("geçersiz istek verisi")
	ErrPlanNotFound       = errors.New("üyelik paketi bulunamadı")
	ErrInvalidToken       = errors.New("geçersiz veya süresi dolmuş oturum")

	ErrAliasTaken                = errors.New("seçilen takma ad (alias) zaten kullanımda")
	ErrInvalidAlias              = errors.New("takma ad 3-32 karakter olmalı ve yalnızca harf, rakam, tire (-) veya alt çizgi (_) içermelidir")
	ErrInvalidTargetURL          = errors.New("geçerli bir hedef web adresi giriniz (örn. ornek.com)")
	ErrInvalidWaitSeconds        = errors.New("bekletme açıkken bekleme süresi en az 1 saniye olmalıdır")
	ErrLinkNotFound              = errors.New("link bulunamadı")
	ErrQRNotFound                = errors.New("QR kod görseli bulunamadı")
	ErrAnonRateLimited           = errors.New("günde yalnızca 1 adet anonim kısa link oluşturabilirsiniz; daha fazlası için ücretsiz hesap açın")
	ErrPlanLinkLimitReached      = errors.New("paketinizin maksimum link oluşturma sınırına ulaştınız")
	ErrPlanCustomAliasRestricted = errors.New("özel takma ad (custom alias) belirlemek için paketinizi yükseltmelisiniz")
	ErrPlanPasswordRestricted    = errors.New("şifreli link oluşturmak için paketinizi yükseltmelisiniz")
	ErrPlanQRRestricted          = errors.New("QR kod oluşturmak için paketinizi yükseltmelisiniz")
	ErrPlanAdsRestricted         = errors.New("reklam/gelir özelliğini kullanmak için paketinizi yükseltmelisiniz")
)

const (
	base62Chars        = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	defaultAliasLen    = 6
	defaultPublicIDLen = 16
	maxAliasCollision  = 5
)

var aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

type PlanInfo struct {
	ID               int32
	Slug             string
	Name             string
	PriceMonthly     float64
	MaxLinks         int32
	AllowCustomAlias bool
	AllowPassword    bool
	AllowQr          bool
	AllowAds         bool
	AnalyticsDays    int32
}

type UserProfile struct {
	User      db.User
	Roles     []string
	Plan      PlanInfo
	LinkCount int64
}

func planFromRow(id int32, slug, name string, price float64, maxLinks int32, customAlias, password, qrAllow, ads bool, days int32) PlanInfo {
	return PlanInfo{
		ID:               id,
		Slug:             slug,
		Name:             name,
		PriceMonthly:     price,
		MaxLinks:         maxLinks,
		AllowCustomAlias: customAlias,
		AllowPassword:    password,
		AllowQr:          qrAllow,
		AllowAds:         ads,
		AnalyticsDays:    days,
	}
}

func HasAnyRole(userRoles []string, allowed ...string) bool {
	for _, r := range userRoles {
		for _, a := range allowed {
			if r == a {
				return true
			}
		}
	}
	return false
}

func IsPrivileged(userRoles []string) bool {
	return HasAnyRole(userRoles, "admin", "owner")
}

// AuthService kimlik doğrulama, yetkilendirme, ban yönetimi ve token rotasyonu iş mantığı.
type AuthService struct {
	users     *repository.UserRepository
	jwtSecret string
}

func NewAuthService(users *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) VerifyToken(tokenStr string) (*auth.Claims, error) {
	return auth.ParseToken(s.jwtSecret, tokenStr)
}

func (s *AuthService) BuildUserProfile(ctx context.Context, user db.User) (UserProfile, error) {
	roles, err := s.users.GetRoles(ctx, user.ID)
	if err != nil {
		roles = []string{"user"}
	}

	effectivePlanID := user.PlanID
	if effectivePlanID <= 0 {
		effectivePlanID = 1
	}
	if user.PlanExpiresAt.Valid && time.Now().After(user.PlanExpiresAt.Time) {
		effectivePlanID = 1
	}

	var plan PlanInfo
	pRow, err := s.users.GetPlanByID(ctx, effectivePlanID)
	if err != nil {
		plan = PlanInfo{
			ID:               1,
			Slug:             "free",
			Name:             "Ücretsiz Paket",
			PriceMonthly:     0,
			MaxLinks:         50,
			AllowCustomAlias: false,
			AllowPassword:    true,
			AllowQr:          false,
			AllowAds:         false,
			AnalyticsDays:    7,
		}
	} else {
		plan = planFromRow(pRow.ID, pRow.Slug, pRow.Name, pRow.PriceMonthly, pRow.MaxLinks, pRow.AllowCustomAlias, pRow.AllowPassword, pRow.AllowQr, pRow.AllowAds, pRow.AnalyticsDays)
	}

	linkCount, _ := s.users.CountLinks(ctx, user.ID)

	return UserProfile{
		User:      user,
		Roles:     roles,
		Plan:      plan,
		LinkCount: linkCount,
	}, nil
}

func formatBanMessage(u db.User) error {
	msg := "Hesabınız askıya alınmıştır"
	if u.BanReason != "" {
		msg += " (Sebep: " + u.BanReason + ")"
	}
	if u.BannedUntil.Valid {
		msg += " — Bitiş: " + u.BannedUntil.Time.Local().Format("02.01.2006 15:04")
	} else {
		msg += " — Süresiz"
	}
	return fmt.Errorf("%w: %s", ErrUserSuspended, msg)
}

func (s *AuthService) checkAndAutoUnban(ctx context.Context, user db.User) (db.User, error) {
	if user.IsActive {
		return user, nil
	}
	if user.BannedUntil.Valid && time.Now().After(user.BannedUntil.Time) {
		unbanned, err := s.users.UnbanUser(ctx, user.ID)
		if err == nil {
			return unbanned, nil
		}
	}
	return user, formatBanMessage(user)
}

func (s *AuthService) issueSessionTokens(ctx context.Context, user db.User, roles []string, userAgent, ip string) (accessToken string, refreshToken string, err error) {
	accessToken, err = auth.GenerateToken(s.jwtSecret, user.ID, user.PublicID, user.Email, roles)
	if err != nil {
		return "", "", err
	}

	rawRefresh, hashRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	_, err = s.users.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: hashRefresh,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(auth.RefreshTokenTTL), Valid: true},
		UserAgent: userAgent,
		IpAddress: ip,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, rawRefresh, nil
}

// Authorize gelen JWT token'ı doğrular, JTI/kullanıcı iptali, ban durumu ve gereken rolleri kontrol eder.
func (s *AuthService) Authorize(ctx context.Context, tokenStr string, requiredRoles ...string) (UserProfile, error) {
	claims, err := s.VerifyToken(tokenStr)
	if err != nil {
		return UserProfile{}, err
	}

	if claims.ID != "" && s.users.IsJTIBlacklisted(ctx, claims.ID) {
		return UserProfile{}, ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserProfile{}, ErrUserNotFound
		}
		return UserProfile{}, err
	}

	user, err = s.checkAndAutoUnban(ctx, user)
	if err != nil {
		return UserProfile{}, err
	}

	if revokedBefore := s.users.GetUserRevokedBefore(ctx, claims.UserID); revokedBefore > 0 {
		issuedNs := claims.IssuedAtNano
		if issuedNs == 0 && claims.IssuedAt != nil {
			issuedNs = claims.IssuedAt.Time.UnixNano()
		}
		if issuedNs <= revokedBefore {
			return UserProfile{}, ErrInvalidToken
		}
	}

	profile, err := s.BuildUserProfile(ctx, user)
	if err != nil {
		return UserProfile{}, err
	}

	if len(requiredRoles) > 0 && !HasAnyRole(profile.Roles, requiredRoles...) {
		return UserProfile{}, ErrForbidden
	}

	return profile, nil
}

func (s *AuthService) Register(ctx context.Context, name *string, email, password, userAgent, ip string) (string, string, UserProfile, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(normalizedEmail, "@") {
		return "", "", UserProfile{}, ErrInvalidEmail
	}
	if len(password) < 4 {
		return "", "", UserProfile{}, ErrPasswordTooShort
	}

	userName := ""
	if name != nil && strings.TrimSpace(*name) != "" {
		userName = strings.TrimSpace(*name)
	} else {
		parts := strings.Split(normalizedEmail, "@")
		userName = parts[0]
		if userName == "" {
			userName = "kullanici"
		}
	}

	if _, err := s.users.GetByEmail(ctx, normalizedEmail); err == nil {
		return "", "", UserProfile{}, ErrEmailTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", "", UserProfile{}, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	var user db.User
	for i := 0; i < maxAliasCollision; i++ {
		publicID, err := generateBase62(defaultPublicIDLen)
		if err != nil {
			return "", "", UserProfile{}, err
		}
		user, err = s.users.Create(ctx, publicID, userName, normalizedEmail, hash)
		if err == nil {
			break
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "email") {
				return "", "", UserProfile{}, ErrEmailTaken
			}
			continue
		}
		return "", "", UserProfile{}, err
	}
	if user.ID == 0 {
		return "", "", UserProfile{}, errors.New("kullanıcı oluşturulamadı")
	}

	profile, err := s.BuildUserProfile(ctx, user)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	accessToken, refreshToken, err := s.issueSessionTokens(ctx, user, profile.Roles, userAgent, ip)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	return accessToken, refreshToken, profile, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ip string) (string, string, UserProfile, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" || password == "" {
		return "", "", UserProfile{}, ErrInvalidInput
	}

	user, err := s.users.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", UserProfile{}, ErrInvalidCredentials
		}
		return "", "", UserProfile{}, err
	}

	ok, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return "", "", UserProfile{}, ErrInvalidCredentials
	}

	user, err = s.checkAndAutoUnban(ctx, user)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	profile, err := s.BuildUserProfile(ctx, user)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	accessToken, refreshToken, err := s.issueSessionTokens(ctx, user, profile.Roles, userAgent, ip)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	return accessToken, refreshToken, profile, nil
}

// RefreshSession refresh token rotasyonu uygular; iptal edilmiş token tekrar kullanılırsa tüm oturumları kapatır.
func (s *AuthService) RefreshSession(ctx context.Context, rawRefreshToken, userAgent, ip string) (string, string, UserProfile, error) {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return "", "", UserProfile{}, ErrInvalidToken
	}

	hash := auth.HashRefreshToken(rawRefreshToken)
	rt, err := s.users.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return "", "", UserProfile{}, ErrInvalidToken
	}

	if rt.RevokedAt.Valid {
		// Token replay detection: daha önce kullanılmış token tekrar gönderildi -> tüm oturumları iptal et!
		_ = s.users.RevokeAllUserRefreshTokens(ctx, rt.UserID)
		s.users.RevokeUserTokensBefore(ctx, rt.UserID)
		return "", "", UserProfile{}, ErrInvalidToken
	}

	if !rt.ExpiresAt.Valid || time.Now().After(rt.ExpiresAt.Time) {
		return "", "", UserProfile{}, ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", "", UserProfile{}, ErrInvalidToken
	}

	user, err = s.checkAndAutoUnban(ctx, user)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	_ = s.users.RevokeRefreshTokenByHash(ctx, hash)

	profile, err := s.BuildUserProfile(ctx, user)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	accessToken, newRefresh, err := s.issueSessionTokens(ctx, user, profile.Roles, userAgent, ip)
	if err != nil {
		return "", "", UserProfile{}, err
	}

	return accessToken, newRefresh, profile, nil
}

func (s *AuthService) Logout(ctx context.Context, tokenStr, rawRefreshToken string) {
	if claims, err := s.VerifyToken(tokenStr); err == nil && claims.ID != "" {
		ttl := 15 * time.Minute
		if claims.ExpiresAt != nil {
			if rem := time.Until(claims.ExpiresAt.Time); rem > 0 {
				ttl = rem
			}
		}
		s.users.BlacklistJTI(ctx, claims.ID, ttl)
	}
	if strings.TrimSpace(rawRefreshToken) != "" {
		_ = s.users.RevokeRefreshTokenByHash(ctx, auth.HashRefreshToken(rawRefreshToken))
	}
}

func (s *AuthService) ResetPassword(ctx context.Context, actor UserProfile, email, newPassword string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(normalizedEmail, "@") {
		return ErrInvalidEmail
	}
	if len(newPassword) < 4 {
		return ErrPasswordTooShort
	}

	if actor.User.Email != normalizedEmail && !IsPrivileged(actor.Roles) {
		return ErrForbidden
	}

	targetUser, err := s.users.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	rows, err := s.users.UpdatePassword(ctx, normalizedEmail, hash)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	_ = s.users.RevokeAllUserRefreshTokens(ctx, targetUser.ID)
	s.users.RevokeUserTokensBefore(ctx, targetUser.ID)
	return nil
}

func (s *AuthService) DeleteUser(ctx context.Context, userID int64) error {
	roles, _ := s.users.GetRoles(ctx, userID)
	if HasAnyRole(roles, "owner") {
		return ErrForbidden
	}

	rows, err := s.users.Delete(ctx, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *AuthService) ListPlans(ctx context.Context) ([]PlanInfo, error) {
	rows, err := s.users.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PlanInfo, 0, len(rows))
	for _, p := range rows {
		out = append(out, planFromRow(p.ID, p.Slug, p.Name, p.PriceMonthly, p.MaxLinks, p.AllowCustomAlias, p.AllowPassword, p.AllowQr, p.AllowAds, p.AnalyticsDays))
	}
	return out, nil
}

// --- Admin User & Plan Management ---

func (s *AuthService) resolveTargetUser(ctx context.Context, idStr string) (db.User, error) {
	clean := strings.TrimSpace(idStr)
	if clean == "" {
		return db.User{}, ErrUserNotFound
	}

	user, err := s.users.GetByPublicID(ctx, clean)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, err
	}

	if numID, parseErr := strconv.ParseInt(clean, 10, 64); parseErr == nil {
		u2, err2 := s.users.GetByID(ctx, numID)
		if err2 == nil {
			return u2, nil
		}
		if !errors.Is(err2, pgx.ErrNoRows) {
			return db.User{}, err2
		}
	}

	return db.User{}, ErrUserNotFound
}

func (s *AuthService) AdminGetUser(ctx context.Context, targetIDStr string) (UserProfile, error) {
	targetUser, err := s.resolveTargetUser(ctx, targetIDStr)
	if err != nil {
		return UserProfile{}, err
	}
	return s.BuildUserProfile(ctx, targetUser)
}

type PaginatedUsersResult struct {
	Items      []UserProfile
	Page       int32
	Limit      int32
	Total      int64
	TotalPages int32
}

func (s *AuthService) AdminListUsers(ctx context.Context, page, limit *int32, search *string) (PaginatedUsersResult, error) {
	p, l := normalizePagination(page, limit, 20, 100)
	q := ""
	if search != nil {
		q = strings.TrimSpace(*search)
	}

	total, err := s.users.CountAdminUsers(ctx, q)
	if err != nil {
		return PaginatedUsersResult{}, err
	}

	totalPages := calcTotalPages(total, l)
	if p > totalPages {
		p = totalPages
	}
	offset := (p - 1) * l

	users, err := s.users.ListAdminUsersPaginated(ctx, q, l, offset)
	if err != nil {
		return PaginatedUsersResult{}, err
	}

	plansList, _ := s.ListPlans(ctx)
	planMap := make(map[int32]PlanInfo, len(plansList))
	for _, pl := range plansList {
		planMap[pl.ID] = pl
	}

	profiles := make([]UserProfile, 0, len(users))
	for _, u := range users {
		roles, err := s.users.GetRoles(ctx, u.ID)
		if err != nil {
			roles = []string{"user"}
		}
		effPlanID := u.PlanID
		if effPlanID <= 0 || (u.PlanExpiresAt.Valid && time.Now().After(u.PlanExpiresAt.Time)) {
			effPlanID = 1
		}
		pl, ok := planMap[effPlanID]
		if !ok {
			pl = planMap[1]
		}
		linkCount, _ := s.users.CountLinks(ctx, u.ID)
		profiles = append(profiles, UserProfile{
			User:      u,
			Roles:     roles,
			Plan:      pl,
			LinkCount: linkCount,
		})
	}

	return PaginatedUsersResult{
		Items:      profiles,
		Page:       p,
		Limit:      l,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func parseOptionalTimestamp(raw *string, existing pgtype.Timestamptz) (pgtype.Timestamptz, error) {
	if raw == nil {
		return existing, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return pgtype.Timestamptz{Valid: false}, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, trimmed); err == nil {
			if layout == "2006-01-02" {
				t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
			}
			return pgtype.Timestamptz{Time: t, Valid: true}, nil
		}
	}
	return pgtype.Timestamptz{}, ErrInvalidInput
}

func (s *AuthService) AdminUpdateUser(
	ctx context.Context,
	actor UserProfile,
	targetIDStr string,
	isActive *bool,
	banReason *string,
	bannedUntil *string,
	planID *int32,
	planExpiresAt *string,
) (UserProfile, error) {
	targetUser, err := s.resolveTargetUser(ctx, targetIDStr)
	if err != nil {
		return UserProfile{}, err
	}

	targetRoles, _ := s.users.GetRoles(ctx, targetUser.ID)

	if HasAnyRole(targetRoles, "owner") {
		if isActive != nil && !*isActive {
			return UserProfile{}, ErrForbidden
		}
		if !HasAnyRole(actor.Roles, "owner") {
			return UserProfile{}, ErrForbidden
		}
	} else if HasAnyRole(targetRoles, "admin") && !HasAnyRole(actor.Roles, "owner") {
		return UserProfile{}, ErrForbidden
	}

	newIsActive := targetUser.IsActive
	if isActive != nil {
		newIsActive = *isActive
	}

	newBanReason := targetUser.BanReason
	newBannedAt := targetUser.BannedAt
	newBannedUntil := targetUser.BannedUntil

	if !newIsActive {
		if targetUser.IsActive {
			newBannedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		}
		if banReason != nil {
			newBanReason = strings.TrimSpace(*banReason)
		}
		parsedUntil, err := parseOptionalTimestamp(bannedUntil, targetUser.BannedUntil)
		if err != nil {
			return UserProfile{}, err
		}
		newBannedUntil = parsedUntil
	} else if isActive != nil && *isActive {
		// Unban: temizle
		newBanReason = ""
		newBannedAt = pgtype.Timestamptz{Valid: false}
		newBannedUntil = pgtype.Timestamptz{Valid: false}
	}

	newPlanID := targetUser.PlanID
	if planID != nil {
		if _, err := s.users.GetPlanByID(ctx, *planID); err != nil {
			return UserProfile{}, ErrPlanNotFound
		}
		newPlanID = *planID
	}

	newExpiresAt, err := parseOptionalTimestamp(planExpiresAt, targetUser.PlanExpiresAt)
	if err != nil {
		return UserProfile{}, err
	}

	updated, err := s.users.AdminUpdateUser(ctx, db.AdminUpdateUserParams{
		ID:            targetUser.ID,
		PlanID:        newPlanID,
		PlanExpiresAt: newExpiresAt,
		IsActive:      newIsActive,
		BanReason:     newBanReason,
		BannedAt:      newBannedAt,
		BannedUntil:   newBannedUntil,
	})
	if err != nil {
		return UserProfile{}, err
	}

	if !newIsActive {
		_ = s.users.RevokeAllUserRefreshTokens(ctx, updated.ID)
		s.users.RevokeUserTokensBefore(ctx, updated.ID)
	}

	return s.BuildUserProfile(ctx, updated)
}

func (s *AuthService) AdminUpdateUserRole(ctx context.Context, actor UserProfile, targetIDStr string, newRole string) (UserProfile, error) {
	if !HasAnyRole(actor.Roles, "owner") {
		return UserProfile{}, ErrForbidden
	}

	roleClean := strings.ToLower(strings.TrimSpace(newRole))
	if roleClean != "user" && roleClean != "admin" {
		return UserProfile{}, ErrInvalidInput
	}

	targetUser, err := s.resolveTargetUser(ctx, targetIDStr)
	if err != nil {
		return UserProfile{}, err
	}

	targetRoles, _ := s.users.GetRoles(ctx, targetUser.ID)
	if HasAnyRole(targetRoles, "owner") {
		return UserProfile{}, ErrForbidden
	}

	if err := s.users.SetRole(ctx, targetUser.ID, roleClean); err != nil {
		return UserProfile{}, err
	}

	return s.BuildUserProfile(ctx, targetUser)
}

type UpdatePlanInput struct {
	Name             *string
	PriceMonthly     *float64
	MaxLinks         *int32
	AllowCustomAlias *bool
	AllowPassword    *bool
	AllowQr          *bool
	AllowAds         *bool
	AnalyticsDays    *int32
}

func (s *AuthService) AdminUpdatePlan(ctx context.Context, actor UserProfile, planID int32, in UpdatePlanInput) (PlanInfo, error) {
	if !HasAnyRole(actor.Roles, "owner") {
		return PlanInfo{}, ErrForbidden
	}

	existing, err := s.users.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PlanInfo{}, ErrPlanNotFound
		}
		return PlanInfo{}, err
	}

	var priceNumeric pgtype.Numeric
	_ = priceNumeric.Scan(fmt.Sprintf("%.2f", existing.PriceMonthly))

	params := db.UpdatePlanParams{
		ID:               existing.ID,
		Name:             existing.Name,
		PriceMonthly:     priceNumeric,
		MaxLinks:         existing.MaxLinks,
		AllowCustomAlias: existing.AllowCustomAlias,
		AllowPassword:    existing.AllowPassword,
		AllowQr:          existing.AllowQr,
		AllowAds:         existing.AllowAds,
		AnalyticsDays:    existing.AnalyticsDays,
	}

	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		params.Name = strings.TrimSpace(*in.Name)
	}
	if in.PriceMonthly != nil && *in.PriceMonthly >= 0 {
		var updatedPrice pgtype.Numeric
		_ = updatedPrice.Scan(fmt.Sprintf("%.2f", *in.PriceMonthly))
		params.PriceMonthly = updatedPrice
	}
	if in.MaxLinks != nil && *in.MaxLinks >= -1 {
		params.MaxLinks = *in.MaxLinks
	}
	if in.AllowCustomAlias != nil {
		params.AllowCustomAlias = *in.AllowCustomAlias
	}
	if in.AllowPassword != nil {
		params.AllowPassword = *in.AllowPassword
	}
	if in.AllowQr != nil {
		params.AllowQr = *in.AllowQr
	}
	if in.AllowAds != nil {
		params.AllowAds = *in.AllowAds
	}
	if in.AnalyticsDays != nil && *in.AnalyticsDays > 0 {
		params.AnalyticsDays = *in.AnalyticsDays
	}

	updated, err := s.users.UpdatePlan(ctx, params)
	if err != nil {
		return PlanInfo{}, err
	}
	return planFromRow(updated.ID, updated.Slug, updated.Name, updated.PriceMonthly, updated.MaxLinks, updated.AllowCustomAlias, updated.AllowPassword, updated.AllowQr, updated.AllowAds, updated.AnalyticsDays), nil
}

// LinkService link, anonim link ve QR kod iş mantığı.
type LinkService struct {
	links         *repository.LinkRepository
	qrStore       *qr.Store
	publicBaseURL string
}

func NewLinkService(links *repository.LinkRepository, qrStore *qr.Store, publicBaseURL string) *LinkService {
	if strings.TrimSpace(publicBaseURL) == "" {
		publicBaseURL = "http://localhost"
	}
	return &LinkService{
		links:         links,
		qrStore:       qrStore,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

func (s *LinkService) PublicBaseURL() string {
	return s.publicBaseURL
}

func validateTargetURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidTargetURL
	}
	lower := strings.ToLower(trimmed)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		trimmed = "https://" + trimmed
	}
	u, err := url.ParseRequestURI(trimmed)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", ErrInvalidTargetURL
	}
	if !strings.Contains(u.Host, ".") && !strings.HasPrefix(u.Host, "localhost") {
		return "", ErrInvalidTargetURL
	}
	return trimmed, nil
}

func validateAlias(alias string) (string, error) {
	trimmed := strings.TrimSpace(alias)
	if !aliasRegex.MatchString(trimmed) {
		return "", ErrInvalidAlias
	}
	if strings.EqualFold(trimmed, "health") {
		return "", ErrAliasTaken
	}
	return trimmed, nil
}

func generateBase62(length int) (string, error) {
	b := make([]byte, length)
	max := big.NewInt(int64(len(base62Chars)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b), nil
}

func (s *LinkService) generateUniqueAlias(ctx context.Context) (string, error) {
	for i := 0; i < maxAliasCollision; i++ {
		candidate, err := generateBase62(defaultAliasLen)
		if err != nil {
			return "", err
		}
		exists, err := s.links.AliasExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("benzersiz kısa kod üretilemedi")
}

func (s *LinkService) overlayPendingClicks(ctx context.Context, items []db.Link) []db.Link {
	if len(items) == 0 {
		return items
	}
	aliases := make([]string, len(items))
	for i, l := range items {
		aliases[i] = l.Alias
	}
	deltas := s.links.GetPendingClickDeltas(ctx, aliases)
	if len(deltas) == 0 {
		return items
	}
	out := make([]db.Link, len(items))
	copy(out, items)
	for i := range out {
		if d, ok := deltas[out[i].Alias]; ok {
			out[i].Clicks += d
		}
	}
	return out
}

type CreateLinkInput struct {
	Target      string
	Alias       *string
	Title       *string
	Password    *string
	WaitEnabled *bool
	WaitSeconds *int32
	ExpiresAt   *string
	MaxClicks   *int64
	AdEnabled   *bool
}

func (s *LinkService) CreateLink(ctx context.Context, profile UserProfile, in CreateLinkInput) (db.Link, error) {
	privileged := IsPrivileged(profile.Roles)

	if !privileged && profile.Plan.MaxLinks >= 0 && profile.LinkCount >= int64(profile.Plan.MaxLinks) {
		return db.Link{}, ErrPlanLinkLimitReached
	}

	cleanTarget, err := validateTargetURL(in.Target)
	if err != nil {
		return db.Link{}, err
	}

	var finalAlias string
	if in.Alias != nil && strings.TrimSpace(*in.Alias) != "" {
		if !privileged && !profile.Plan.AllowCustomAlias {
			return db.Link{}, ErrPlanCustomAliasRestricted
		}
		validAlias, err := validateAlias(*in.Alias)
		if err != nil {
			return db.Link{}, err
		}
		exists, err := s.links.AliasExists(ctx, validAlias)
		if err != nil {
			return db.Link{}, err
		}
		if exists {
			return db.Link{}, ErrAliasTaken
		}
		finalAlias = validAlias
	} else {
		finalAlias, err = s.generateUniqueAlias(ctx)
		if err != nil {
			return db.Link{}, err
		}
	}

	var pwHash pgtype.Text
	if in.Password != nil && *in.Password != "" {
		if !privileged && !profile.Plan.AllowPassword {
			return db.Link{}, ErrPlanPasswordRestricted
		}
		hashed, err := auth.HashPassword(*in.Password)
		if err != nil {
			return db.Link{}, err
		}
		pwHash = pgtype.Text{String: hashed, Valid: true}
	}

	var finalAdEnabled bool
	if in.AdEnabled != nil && *in.AdEnabled {
		if !privileged && !profile.Plan.AllowAds {
			return db.Link{}, ErrPlanAdsRestricted
		}
		finalAdEnabled = true
	}

	var finalWaitEnabled bool
	var finalWaitSeconds int32
	if !pwHash.Valid && in.WaitEnabled != nil && *in.WaitEnabled {
		finalWaitEnabled = true
		if in.WaitSeconds == nil || *in.WaitSeconds <= 0 {
			return db.Link{}, ErrInvalidWaitSeconds
		}
		finalWaitSeconds = *in.WaitSeconds
	}

	title := ""
	if in.Title != nil {
		title = strings.TrimSpace(*in.Title)
	}

	expiresAt, err := parseOptionalTimestamp(in.ExpiresAt, pgtype.Timestamptz{})
	if err != nil {
		return db.Link{}, err
	}

	var maxClicks pgtype.Int8
	if in.MaxClicks != nil && *in.MaxClicks > 0 {
		maxClicks = pgtype.Int8{Int64: *in.MaxClicks, Valid: true}
	}

	publicID, err := generateBase62(defaultPublicIDLen)
	if err != nil {
		return db.Link{}, err
	}

	link, err := s.links.Create(ctx, db.CreateLinkParams{
		PublicID:     publicID,
		UserID:       pgtype.Int8{Int64: profile.User.ID, Valid: true},
		Alias:        finalAlias,
		TargetUrl:    cleanTarget,
		PasswordHash: pwHash,
		WaitEnabled:  finalWaitEnabled,
		WaitSeconds:  finalWaitSeconds,
		Title:        title,
		IsActive:     true,
		ExpiresAt:    expiresAt,
		MaxClicks:    maxClicks,
		AdEnabled:    finalAdEnabled,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.Link{}, ErrAliasTaken
		}
		return db.Link{}, err
	}

	return link, nil
}

type CreateAnonymousLinkInput struct {
	Target      string
	Password    *string
	WaitEnabled *bool
	WaitSeconds *int32
	TTLHours    *int32
	ClientIP    string
}

func (s *LinkService) CreateAnonymousLink(ctx context.Context, in CreateAnonymousLinkInput) (db.Link, error) {
	cleanTarget, err := validateTargetURL(in.Target)
	if err != nil {
		return db.Link{}, err
	}

	allowed, err := s.links.CheckAndSetAnonRateLimit(ctx, in.ClientIP)
	if err != nil {
		return db.Link{}, err
	}
	if !allowed {
		return db.Link{}, ErrAnonRateLimited
	}

	ttlHours := int32(48)
	if in.TTLHours != nil {
		if *in.TTLHours < 1 {
			ttlHours = 1
		} else if *in.TTLHours > 48 {
			ttlHours = 48
		} else {
			ttlHours = *in.TTLHours
		}
	}
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(time.Duration(ttlHours) * time.Hour),
		Valid: true,
	}

	var pwHash pgtype.Text
	if in.Password != nil && *in.Password != "" {
		hashed, err := auth.HashPassword(*in.Password)
		if err != nil {
			s.links.ReleaseAnonRateLimit(ctx, in.ClientIP)
			return db.Link{}, err
		}
		pwHash = pgtype.Text{String: hashed, Valid: true}
	}

	var finalWaitEnabled bool
	var finalWaitSeconds int32
	if !pwHash.Valid && in.WaitEnabled != nil && *in.WaitEnabled {
		finalWaitEnabled = true
		if in.WaitSeconds == nil || *in.WaitSeconds <= 0 {
			s.links.ReleaseAnonRateLimit(ctx, in.ClientIP)
			return db.Link{}, ErrInvalidWaitSeconds
		}
		finalWaitSeconds = *in.WaitSeconds
	}

	finalAlias, err := s.generateUniqueAlias(ctx)
	if err != nil {
		s.links.ReleaseAnonRateLimit(ctx, in.ClientIP)
		return db.Link{}, err
	}

	publicID, err := generateBase62(defaultPublicIDLen)
	if err != nil {
		s.links.ReleaseAnonRateLimit(ctx, in.ClientIP)
		return db.Link{}, err
	}

	link, err := s.links.CreateAnonymous(ctx, db.CreateAnonymousLinkParams{
		PublicID:     publicID,
		Alias:        finalAlias,
		TargetUrl:    cleanTarget,
		PasswordHash: pwHash,
		WaitEnabled:  finalWaitEnabled,
		WaitSeconds:  finalWaitSeconds,
		ExpiresAt:    expiresAt,
		CreatedByIp:  in.ClientIP,
	})
	if err != nil {
		s.links.ReleaseAnonRateLimit(ctx, in.ClientIP)
		return db.Link{}, err
	}

	// Anonim linkler için doğrudan QR kodunu da üret ki kullanıcı kaydedebilsin
	if s.qrStore != nil {
		shortURL := s.publicBaseURL + "/r/" + link.Alias
		if qrFilename, qrErr := s.qrStore.Generate(link.PublicID, shortURL); qrErr == nil {
			if updated, upErr := s.links.UpdateQRPath(ctx, link.ID, qrFilename); upErr == nil {
				link = updated
			}
		}
	}

	return link, nil
}

type PaginatedLinksResult struct {
	Items       []db.Link
	Page        int32
	Limit       int32
	Total       int64
	TotalPages  int32
	TotalClicks int64
	LockedCount int64
	TopLinks    []db.Link
}

type PaginatedLogsResult struct {
	Items      []db.Log
	Page       int32
	Limit      int32
	Total      int64
	TotalPages int32
}

func normalizePagination(page, limit *int32, defaultLimit, maxLimit int32) (int32, int32) {
	p := int32(1)
	if page != nil && *page > 0 {
		p = *page
	}
	l := defaultLimit
	if limit != nil && *limit > 0 {
		l = *limit
		if l > maxLimit {
			l = maxLimit
		}
	}
	return p, l
}

func calcTotalPages(total int64, limit int32) int32 {
	if total <= 0 || limit <= 0 {
		return 1
	}
	pages := (total + int64(limit) - 1) / int64(limit)
	if pages < 1 {
		return 1
	}
	return int32(pages)
}

func (s *LinkService) ListLinks(ctx context.Context, userID int64, page, limit *int32) (PaginatedLinksResult, error) {
	p, l := normalizePagination(page, limit, 10, 100)

	stats, err := s.links.GetUserStats(ctx, userID)
	if err != nil {
		return PaginatedLinksResult{}, err
	}

	totalPages := calcTotalPages(stats.TotalLinks, l)
	if p > totalPages {
		p = totalPages
	}
	offset := (p - 1) * l

	items, err := s.links.ListByUserPaginated(ctx, userID, l, offset)
	if err != nil {
		return PaginatedLinksResult{}, err
	}

	topLinks, err := s.links.ListTopByUser(ctx, userID, 3)
	if err != nil {
		return PaginatedLinksResult{}, err
	}

	items = s.overlayPendingClicks(ctx, items)
	topLinks = s.overlayPendingClicks(ctx, topLinks)

	// Kullanıcının bekleyen tıklamalarının toplamını istatistiklere yansıt
	allUserLinks, err := s.links.ListByUser(ctx, userID)
	extraClicks := int64(0)
	if err == nil && len(allUserLinks) > 0 {
		aliases := make([]string, len(allUserLinks))
		for i, item := range allUserLinks {
			aliases[i] = item.Alias
		}
		for _, d := range s.links.GetPendingClickDeltas(ctx, aliases) {
			extraClicks += d
		}
	}

	return PaginatedLinksResult{
		Items:       items,
		Page:        p,
		Limit:       l,
		Total:       stats.TotalLinks,
		TotalPages:  totalPages,
		TotalClicks: stats.TotalClicks + extraClicks,
		LockedCount: stats.LockedLinks,
		TopLinks:    topLinks,
	}, nil
}

func (s *LinkService) CheckAlias(ctx context.Context, alias string) (bool, error) {
	validAlias, err := validateAlias(alias)
	if err != nil {
		return false, err
	}
	exists, err := s.links.AliasExists(ctx, validAlias)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (s *LinkService) GetLink(ctx context.Context, idStr string, userID int64) (db.Link, error) {
	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return db.Link{}, ErrLinkNotFound
	}
	link, err := s.links.GetByIDAndUser(ctx, publicID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrLinkNotFound
		}
		return db.Link{}, err
	}
	link.Clicks += s.links.GetPendingClickDelta(ctx, link.Alias)
	return link, nil
}

type UpdateLinkInput struct {
	Target         *string
	Alias          *string
	Title          *string
	Password       *string
	WaitEnabled    *bool
	WaitSeconds    *int32
	IsActive       *bool
	ExpiresAt      *string
	MaxClicks      *int64
	AdEnabled      *bool
	SecurityStatus *string
}

func (s *LinkService) UpdateLink(ctx context.Context, profile UserProfile, idStr string, in UpdateLinkInput) (db.Link, error) {
	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return db.Link{}, ErrLinkNotFound
	}

	existing, err := s.links.GetByIDAndUser(ctx, publicID, profile.User.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrLinkNotFound
		}
		return db.Link{}, err
	}

	privileged := IsPrivileged(profile.Roles)

	newTarget := existing.TargetUrl
	if in.Target != nil && strings.TrimSpace(*in.Target) != "" {
		cleanTarget, err := validateTargetURL(*in.Target)
		if err != nil {
			return db.Link{}, err
		}
		newTarget = cleanTarget
	}

	newAlias := existing.Alias
	if in.Alias != nil && strings.TrimSpace(*in.Alias) != "" {
		validAlias, err := validateAlias(*in.Alias)
		if err != nil {
			return db.Link{}, err
		}
		if validAlias != existing.Alias {
			if !privileged && !profile.Plan.AllowCustomAlias {
				return db.Link{}, ErrPlanCustomAliasRestricted
			}
			exists, err := s.links.AliasExists(ctx, validAlias)
			if err != nil {
				return db.Link{}, err
			}
			if exists {
				return db.Link{}, ErrAliasTaken
			}
			newAlias = validAlias
		}
	}

	newPwHash := existing.PasswordHash
	if in.Password != nil {
		if *in.Password == "" {
			newPwHash = pgtype.Text{Valid: false}
		} else {
			if !privileged && !profile.Plan.AllowPassword {
				return db.Link{}, ErrPlanPasswordRestricted
			}
			hashed, err := auth.HashPassword(*in.Password)
			if err != nil {
				return db.Link{}, err
			}
			newPwHash = pgtype.Text{String: hashed, Valid: true}
		}
	}

	newWaitEnabled := existing.WaitEnabled
	if in.WaitEnabled != nil {
		newWaitEnabled = *in.WaitEnabled
	}

	newWaitSeconds := existing.WaitSeconds
	if in.WaitSeconds != nil {
		newWaitSeconds = *in.WaitSeconds
	}

	if newPwHash.Valid && newPwHash.String != "" {
		newWaitEnabled = false
		newWaitSeconds = 0
	} else if !newWaitEnabled {
		newWaitSeconds = 0
	} else if newWaitSeconds <= 0 {
		return db.Link{}, ErrInvalidWaitSeconds
	}

	newTitle := existing.Title
	if in.Title != nil {
		newTitle = strings.TrimSpace(*in.Title)
	}

	newIsActive := existing.IsActive
	if in.IsActive != nil {
		newIsActive = *in.IsActive
	}

	newExpiresAt, err := parseOptionalTimestamp(in.ExpiresAt, existing.ExpiresAt)
	if err != nil {
		return db.Link{}, err
	}

	newMaxClicks := existing.MaxClicks
	if in.MaxClicks != nil {
		if *in.MaxClicks <= 0 {
			newMaxClicks = pgtype.Int8{Valid: false}
		} else {
			newMaxClicks = pgtype.Int8{Int64: *in.MaxClicks, Valid: true}
		}
	}

	newAdEnabled := existing.AdEnabled
	if in.AdEnabled != nil {
		if *in.AdEnabled && !privileged && !profile.Plan.AllowAds {
			return db.Link{}, ErrPlanAdsRestricted
		}
		newAdEnabled = *in.AdEnabled
	}

	updated, err := s.links.Update(ctx, existing.Alias, db.UpdateLinkParams{
		ID:           existing.ID,
		UserID:       pgtype.Int8{Int64: profile.User.ID, Valid: true},
		Alias:        newAlias,
		TargetUrl:    newTarget,
		PasswordHash: newPwHash,
		WaitEnabled:  newWaitEnabled,
		WaitSeconds:  newWaitSeconds,
		Title:        newTitle,
		IsActive:     newIsActive,
		ExpiresAt:    newExpiresAt,
		MaxClicks:    newMaxClicks,
		AdEnabled:    newAdEnabled,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.Link{}, ErrAliasTaken
		}
		return db.Link{}, err
	}

	// Alias değiştiyse ve QR kodu varsa yeni kısa link için QR dosyasını yenile
	if s.qrStore != nil && updated.QrCodePath != "" && existing.Alias != updated.Alias {
		shortURL := s.publicBaseURL + "/r/" + updated.Alias
		_, _ = s.qrStore.Generate(updated.PublicID, shortURL)
	}

	updated.Clicks += s.links.GetPendingClickDelta(ctx, updated.Alias)
	return updated, nil
}

func (s *LinkService) ListLinkLogs(ctx context.Context, profile UserProfile, idStr string, page, limit *int32) (PaginatedLogsResult, error) {
	link, err := s.GetLink(ctx, idStr, profile.User.ID)
	if err != nil {
		return PaginatedLogsResult{}, err
	}

	var since pgtype.Timestamptz
	if !IsPrivileged(profile.Roles) && profile.Plan.AnalyticsDays > 0 {
		since = pgtype.Timestamptz{
			Time:  time.Now().AddDate(0, 0, -int(profile.Plan.AnalyticsDays)),
			Valid: true,
		}
	} else {
		since = pgtype.Timestamptz{
			Time:  time.Unix(0, 0),
			Valid: true,
		}
	}

	p, l := normalizePagination(page, limit, 20, 100)

	total, err := s.links.CountLogsByLink(ctx, link.ID, since)
	if err != nil {
		return PaginatedLogsResult{}, err
	}

	totalPages := calcTotalPages(total, l)
	if p > totalPages {
		p = totalPages
	}
	offset := (p - 1) * l

	items, err := s.links.ListLogsByLinkPaginated(ctx, link.ID, since, l, offset)
	if err != nil {
		return PaginatedLogsResult{}, err
	}

	return PaginatedLogsResult{
		Items:      items,
		Page:       p,
		Limit:      l,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, idStr string, userID int64) error {
	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return ErrLinkNotFound
	}

	rows, err := s.links.Delete(ctx, publicID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLinkNotFound
		}
		return err
	}
	if rows == 0 {
		return ErrLinkNotFound
	}
	return nil
}

// --- QR Kod Yönetimi ---

func (s *LinkService) GenerateLinkQR(ctx context.Context, profile UserProfile, idStr string) (db.Link, error) {
	if s.qrStore == nil {
		return db.Link{}, errors.New("QR servisi aktif değil")
	}

	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return db.Link{}, ErrLinkNotFound
	}

	var link db.Link
	var err error
	privileged := IsPrivileged(profile.Roles)
	if privileged {
		link, err = s.links.GetByPublicID(ctx, publicID)
	} else {
		link, err = s.links.GetByIDAndUser(ctx, publicID, profile.User.ID)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrLinkNotFound
		}
		return db.Link{}, err
	}

	if !privileged && !profile.Plan.AllowQr {
		return db.Link{}, ErrPlanQRRestricted
	}

	shortURL := s.publicBaseURL + "/r/" + link.Alias
	filename, err := s.qrStore.Generate(link.PublicID, shortURL)
	if err != nil {
		return db.Link{}, err
	}

	updated, err := s.links.UpdateQRPath(ctx, link.ID, filename)
	if err != nil {
		return db.Link{}, err
	}
	updated.Clicks += s.links.GetPendingClickDelta(ctx, updated.Alias)
	return updated, nil
}

func (s *LinkService) GetLinkQRFile(ctx context.Context, idStr string) (alias string, filePath string, err error) {
	if s.qrStore == nil {
		return "", "", ErrQRNotFound
	}

	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return "", "", ErrLinkNotFound
	}

	link, err := s.links.GetByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrLinkNotFound
		}
		return "", "", err
	}

	if link.QrCodePath == "" {
		return "", "", ErrQRNotFound
	}

	fullPath, err := s.qrStore.ReadPath(link.QrCodePath)
	if err != nil {
		return "", "", ErrQRNotFound
	}

	return link.Alias, fullPath, nil
}

func (s *LinkService) DeleteLinkQR(ctx context.Context, profile UserProfile, idStr string) error {
	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return ErrLinkNotFound
	}

	var link db.Link
	var err error
	if IsPrivileged(profile.Roles) {
		link, err = s.links.GetByPublicID(ctx, publicID)
	} else {
		link, err = s.links.GetByIDAndUser(ctx, publicID, profile.User.ID)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLinkNotFound
		}
		return err
	}

	if s.qrStore != nil && link.QrCodePath != "" {
		s.qrStore.Delete(link.QrCodePath)
	}

	_, err = s.links.UpdateQRPath(ctx, link.ID, "")
	return err
}

// --- Admin Link Operations ---

type PaginatedAdminLinksResult struct {
	Items      []db.ListAdminLinksPaginatedRow
	Page       int32
	Limit      int32
	Total      int64
	TotalPages int32
}

func (s *LinkService) AdminGetStats(ctx context.Context) (db.GetPlatformStatsRow, error) {
	return s.links.GetPlatformStats(ctx)
}

func (s *LinkService) AdminGetLink(ctx context.Context, idStr string) (db.GetAdminLinkByPublicIDRow, error) {
	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return db.GetAdminLinkByPublicIDRow{}, ErrLinkNotFound
	}
	row, err := s.links.GetAdminLinkByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetAdminLinkByPublicIDRow{}, ErrLinkNotFound
		}
		return db.GetAdminLinkByPublicIDRow{}, err
	}
	row.Clicks += s.links.GetPendingClickDelta(ctx, row.Alias)
	return row, nil
}

func (s *LinkService) AdminListLinks(ctx context.Context, page, limit *int32, search, status *string) (PaginatedAdminLinksResult, error) {
	p, l := normalizePagination(page, limit, 20, 100)

	q := ""
	if search != nil {
		q = strings.TrimSpace(*search)
	}
	st := "all"
	if status != nil && strings.TrimSpace(*status) != "" {
		st = strings.TrimSpace(*status)
	}

	total, err := s.links.CountAdminLinks(ctx, q, st)
	if err != nil {
		return PaginatedAdminLinksResult{}, err
	}

	totalPages := calcTotalPages(total, l)
	if p > totalPages {
		p = totalPages
	}
	offset := (p - 1) * l

	items, err := s.links.ListAdminLinksPaginated(ctx, q, st, l, offset)
	if err != nil {
		return PaginatedAdminLinksResult{}, err
	}

	if len(items) > 0 {
		aliases := make([]string, len(items))
		for i, item := range items {
			aliases[i] = item.Alias
		}
		deltas := s.links.GetPendingClickDeltas(ctx, aliases)
		for i := range items {
			if d, ok := deltas[items[i].Alias]; ok {
				items[i].Clicks += d
			}
		}
	}

	return PaginatedAdminLinksResult{
		Items:      items,
		Page:       p,
		Limit:      l,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *LinkService) AdminUpdateLink(ctx context.Context, idStr string, in UpdateLinkInput) (db.Link, error) {
	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return db.Link{}, ErrLinkNotFound
	}

	existing, err := s.links.GetByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrLinkNotFound
		}
		return db.Link{}, err
	}

	newTarget := existing.TargetUrl
	if in.Target != nil && strings.TrimSpace(*in.Target) != "" {
		cleanTarget, err := validateTargetURL(*in.Target)
		if err != nil {
			return db.Link{}, err
		}
		newTarget = cleanTarget
	}

	newAlias := existing.Alias
	if in.Alias != nil && strings.TrimSpace(*in.Alias) != "" {
		validAlias, err := validateAlias(*in.Alias)
		if err != nil {
			return db.Link{}, err
		}
		if validAlias != existing.Alias {
			exists, err := s.links.AliasExists(ctx, validAlias)
			if err != nil {
				return db.Link{}, err
			}
			if exists {
				return db.Link{}, ErrAliasTaken
			}
			newAlias = validAlias
		}
	}

	newPwHash := existing.PasswordHash
	if in.Password != nil {
		if *in.Password == "" {
			newPwHash = pgtype.Text{Valid: false}
		} else {
			hashed, err := auth.HashPassword(*in.Password)
			if err != nil {
				return db.Link{}, err
			}
			newPwHash = pgtype.Text{String: hashed, Valid: true}
		}
	}

	newWaitEnabled := existing.WaitEnabled
	if in.WaitEnabled != nil {
		newWaitEnabled = *in.WaitEnabled
	}

	newWaitSeconds := existing.WaitSeconds
	if in.WaitSeconds != nil {
		newWaitSeconds = *in.WaitSeconds
	}

	if newPwHash.Valid && newPwHash.String != "" {
		newWaitEnabled = false
		newWaitSeconds = 0
	} else if !newWaitEnabled {
		newWaitSeconds = 0
	} else if newWaitSeconds <= 0 {
		return db.Link{}, ErrInvalidWaitSeconds
	}

	newTitle := existing.Title
	if in.Title != nil {
		newTitle = strings.TrimSpace(*in.Title)
	}

	newIsActive := existing.IsActive
	if in.IsActive != nil {
		newIsActive = *in.IsActive
	}

	newExpiresAt, err := parseOptionalTimestamp(in.ExpiresAt, existing.ExpiresAt)
	if err != nil {
		return db.Link{}, err
	}

	newMaxClicks := existing.MaxClicks
	if in.MaxClicks != nil {
		if *in.MaxClicks <= 0 {
			newMaxClicks = pgtype.Int8{Valid: false}
		} else {
			newMaxClicks = pgtype.Int8{Int64: *in.MaxClicks, Valid: true}
		}
	}

	newAdEnabled := existing.AdEnabled
	if in.AdEnabled != nil {
		newAdEnabled = *in.AdEnabled
	}

	newSecStatus := existing.SecurityStatus
	newSecCheckedAt := existing.SecurityCheckedAt
	if in.SecurityStatus != nil && strings.TrimSpace(*in.SecurityStatus) != "" {
		st := strings.ToLower(strings.TrimSpace(*in.SecurityStatus))
		switch st {
		case "unchecked", "safe", "suspicious", "malicious", "blocked":
			newSecStatus = st
			newSecCheckedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
			if (st == "blocked" || st == "malicious") && in.IsActive == nil {
				newIsActive = false
			}
		default:
			return db.Link{}, ErrInvalidInput
		}
	}

	updated, err := s.links.AdminUpdate(ctx, existing.Alias, db.AdminUpdateLinkParams{
		ID:                existing.ID,
		Alias:             newAlias,
		TargetUrl:         newTarget,
		PasswordHash:      newPwHash,
		WaitEnabled:       newWaitEnabled,
		WaitSeconds:       newWaitSeconds,
		Title:             newTitle,
		IsActive:          newIsActive,
		ExpiresAt:         newExpiresAt,
		MaxClicks:         newMaxClicks,
		AdEnabled:         newAdEnabled,
		SecurityStatus:    newSecStatus,
		SecurityCheckedAt: newSecCheckedAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.Link{}, ErrAliasTaken
		}
		return db.Link{}, err
	}

	if s.qrStore != nil && updated.QrCodePath != "" && existing.Alias != updated.Alias {
		shortURL := s.publicBaseURL + "/r/" + updated.Alias
		_, _ = s.qrStore.Generate(updated.PublicID, shortURL)
	}

	updated.Clicks += s.links.GetPendingClickDelta(ctx, updated.Alias)
	return updated, nil
}

func (s *LinkService) AdminDeleteLink(ctx context.Context, idStr string) error {
	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return ErrLinkNotFound
	}

	rows, err := s.links.AdminDelete(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLinkNotFound
		}
		return err
	}
	if rows == 0 {
		return ErrLinkNotFound
	}
	return nil
}

func (s *LinkService) AdminListLinkLogs(ctx context.Context, idStr string, page, limit *int32) (PaginatedLogsResult, error) {
	publicID := strings.TrimSpace(idStr)
	if publicID == "" {
		return PaginatedLogsResult{}, ErrLinkNotFound
	}

	link, err := s.links.GetByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PaginatedLogsResult{}, ErrLinkNotFound
		}
		return PaginatedLogsResult{}, err
	}

	since := pgtype.Timestamptz{Time: time.Unix(0, 0), Valid: true}
	p, l := normalizePagination(page, limit, 20, 100)

	total, err := s.links.CountLogsByLink(ctx, link.ID, since)
	if err != nil {
		return PaginatedLogsResult{}, err
	}

	totalPages := calcTotalPages(total, l)
	if p > totalPages {
		p = totalPages
	}
	offset := (p - 1) * l

	items, err := s.links.ListLogsByLinkPaginated(ctx, link.ID, since, l, offset)
	if err != nil {
		return PaginatedLogsResult{}, err
	}

	return PaginatedLogsResult{
		Items:      items,
		Page:       p,
		Limit:      l,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
