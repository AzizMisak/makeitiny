package service

import (
	"context"
	"crypto/rand"
	"errors"
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

	ErrAliasTaken                = errors.New("seçilen takma ad (alias) zaten kullanımda")
	ErrInvalidAlias              = errors.New("takma ad 3-32 karakter olmalı ve yalnızca harf, rakam, tire (-) veya alt çizgi (_) içermelidir")
	ErrInvalidTargetURL          = errors.New("geçerli bir hedef web adresi giriniz (örn. ornek.com)")
	ErrInvalidWaitSeconds        = errors.New("bekletme açıkken bekleme süresi en az 1 saniye olmalıdır")
	ErrLinkNotFound              = errors.New("link bulunamadı")
	ErrPlanLinkLimitReached      = errors.New("paketinizin maksimum link oluşturma sınırına ulaştınız")
	ErrPlanCustomAliasRestricted = errors.New("özel takma ad (custom alias) belirlemek için paketinizi yükseltmelisiniz")
	ErrPlanPasswordRestricted    = errors.New("şifreli link oluşturmak için paketinizi yükseltmelisiniz")
	ErrPlanAdsRestricted         = errors.New("reklam/gelir özelliğini kullanmak için paketinizi yükseltmelisiniz")
)

const (
	base62Chars        = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	defaultAliasLen    = 6
	defaultPublicIDLen = 16
	maxAliasCollision  = 5
)

var aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

type UserProfile struct {
	User      db.User
	Roles     []string
	Plan      db.Plan
	LinkCount int64
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

// AuthService kimlik doğrulama, yetkilendirme ve kullanıcı/paket yönetimi iş mantığı.
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

	plan, err := s.users.GetPlanByID(ctx, effectivePlanID)
	if err != nil {
		plan = db.Plan{
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
	}

	linkCount, _ := s.users.CountLinks(ctx, user.ID)

	return UserProfile{
		User:      user,
		Roles:     roles,
		Plan:      plan,
		LinkCount: linkCount,
	}, nil
}

// Authorize gelen JWT token'ı doğrular, kullanıcının aktifliğini ve (varsa) gereken rolleri kontrol eder.
func (s *AuthService) Authorize(ctx context.Context, tokenStr string, requiredRoles ...string) (UserProfile, error) {
	claims, err := s.VerifyToken(tokenStr)
	if err != nil {
		return UserProfile{}, err
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserProfile{}, ErrUserNotFound
		}
		return UserProfile{}, err
	}

	if !user.IsActive {
		return UserProfile{}, ErrUserSuspended
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

func (s *AuthService) Register(ctx context.Context, name *string, email, password string) (string, UserProfile, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(normalizedEmail, "@") {
		return "", UserProfile{}, ErrInvalidEmail
	}
	if len(password) < 4 {
		return "", UserProfile{}, ErrPasswordTooShort
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
		return "", UserProfile{}, ErrEmailTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", UserProfile{}, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", UserProfile{}, err
	}

	user, err := s.users.Create(ctx, userName, normalizedEmail, hash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", UserProfile{}, ErrEmailTaken
		}
		return "", UserProfile{}, err
	}

	profile, err := s.BuildUserProfile(ctx, user)
	if err != nil {
		return "", UserProfile{}, err
	}

	token, err := auth.GenerateToken(s.jwtSecret, user.ID, user.Email, profile.Roles)
	if err != nil {
		return "", UserProfile{}, err
	}

	return token, profile, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, UserProfile, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" || password == "" {
		return "", UserProfile{}, ErrInvalidInput
	}

	user, err := s.users.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", UserProfile{}, ErrInvalidCredentials
		}
		return "", UserProfile{}, err
	}

	ok, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return "", UserProfile{}, ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", UserProfile{}, ErrUserSuspended
	}

	profile, err := s.BuildUserProfile(ctx, user)
	if err != nil {
		return "", UserProfile{}, err
	}

	token, err := auth.GenerateToken(s.jwtSecret, user.ID, user.Email, profile.Roles)
	if err != nil {
		return "", UserProfile{}, err
	}

	return token, profile, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, newPassword string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(normalizedEmail, "@") {
		return ErrInvalidEmail
	}
	if len(newPassword) < 4 {
		return ErrPasswordTooShort
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

func (s *AuthService) ListPlans(ctx context.Context) ([]db.Plan, error) {
	return s.users.ListPlans(ctx)
}

// --- Admin User & Plan Management ---

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

	profiles := make([]UserProfile, 0, len(users))
	for _, u := range users {
		prof, err := s.BuildUserProfile(ctx, u)
		if err != nil {
			continue
		}
		profiles = append(profiles, prof)
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

func (s *AuthService) AdminUpdateUser(ctx context.Context, actor UserProfile, targetIDStr string, isActive *bool, planID *int32, planExpiresAt *string) (UserProfile, error) {
	targetID, err := strconv.ParseInt(strings.TrimSpace(targetIDStr), 10, 64)
	if err != nil {
		return UserProfile{}, ErrUserNotFound
	}

	targetUser, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserProfile{}, ErrUserNotFound
		}
		return UserProfile{}, err
	}

	targetRoles, _ := s.users.GetRoles(ctx, targetUser.ID)

	// Owner hesabının aktifliği veya paketi başka biri tarafından kısıtlanamaz;
	// ayrıca bir admin başka bir admin üzerinde işlem yapamaz (yalnızca owner yapabilir).
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
	})
	if err != nil {
		return UserProfile{}, err
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

	targetID, err := strconv.ParseInt(strings.TrimSpace(targetIDStr), 10, 64)
	if err != nil {
		return UserProfile{}, ErrUserNotFound
	}

	targetUser, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserProfile{}, ErrUserNotFound
		}
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

func (s *AuthService) AdminUpdatePlan(ctx context.Context, actor UserProfile, planID int32, in UpdatePlanInput) (db.Plan, error) {
	if !HasAnyRole(actor.Roles, "owner") {
		return db.Plan{}, ErrForbidden
	}

	existing, err := s.users.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Plan{}, ErrPlanNotFound
		}
		return db.Plan{}, err
	}

	params := db.UpdatePlanParams{
		ID:               existing.ID,
		Name:             existing.Name,
		PriceMonthly:     existing.PriceMonthly,
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
		params.PriceMonthly = *in.PriceMonthly
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

	return s.users.UpdatePlan(ctx, params)
}

// LinkService link iş mantığı.
type LinkService struct {
	links *repository.LinkRepository
}

func NewLinkService(links *repository.LinkRepository) *LinkService {
	return &LinkService{links: links}
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
		for i := 0; i < maxAliasCollision; i++ {
			candidate, err := generateBase62(defaultAliasLen)
			if err != nil {
				return db.Link{}, err
			}
			exists, err := s.links.AliasExists(ctx, candidate)
			if err != nil {
				return db.Link{}, err
			}
			if !exists {
				finalAlias = candidate
				break
			}
		}
		if finalAlias == "" {
			return db.Link{}, errors.New("benzersiz kısa kod üretilemedi")
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

	// Şifre açıksa bekletme kesinlikle kapalı olur.
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
		UserID:       profile.User.ID,
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

	return PaginatedLinksResult{
		Items:       items,
		Page:        p,
		Limit:       l,
		Total:       stats.TotalLinks,
		TotalPages:  totalPages,
		TotalClicks: stats.TotalClicks,
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

	// Şifre + bekletme aynı anda açık olamaz: şifre aktifse bekletme kapatılır.
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
		UserID:       profile.User.ID,
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

// --- Admin Link Operations ---

type PaginatedAdminLinksResult struct {
	Items      []db.AdminLinkRow
	Page       int32
	Limit      int32
	Total      int64
	TotalPages int32
}

func (s *LinkService) AdminGetStats(ctx context.Context) (db.GetPlatformStatsRow, error) {
	return s.links.GetPlatformStats(ctx)
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
