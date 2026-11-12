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
	ErrInvalidEmail       = errors.New("geçerli bir e-posta adresi giriniz")
	ErrPasswordTooShort   = errors.New("parola en az 4 karakter olmalıdır")
	ErrInvalidInput       = errors.New("geçersiz istek verisi")

	ErrAliasTaken         = errors.New("seçilen takma ad (alias) zaten kullanımda")
	ErrInvalidAlias       = errors.New("takma ad 3-32 karakter olmalı ve yalnızca harf, rakam, tire (-) veya alt çizgi (_) içermelidir")
	ErrInvalidTargetURL   = errors.New("geçerli bir hedef web adresi giriniz (örn. ornek.com)")
	ErrInvalidWaitSeconds = errors.New("bekletme açıkken bekleme süresi en az 1 saniye olmalıdır")
	ErrLinkNotFound       = errors.New("link bulunamadı")
)

const (
	base62Chars       = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	defaultAliasLen   = 6
	maxAliasCollision = 5
)

var aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

// AuthService kimlik doğrulama iş mantığı.
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

func (s *AuthService) Register(ctx context.Context, name *string, email, password string) (string, db.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(normalizedEmail, "@") {
		return "", db.User{}, ErrInvalidEmail
	}
	if len(password) < 4 {
		return "", db.User{}, ErrPasswordTooShort
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
		return "", db.User{}, ErrEmailTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", db.User{}, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", db.User{}, err
	}

	user, err := s.users.Create(ctx, userName, normalizedEmail, hash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", db.User{}, ErrEmailTaken
		}
		return "", db.User{}, err
	}

	token, err := auth.GenerateToken(s.jwtSecret, user.ID, user.Email)
	if err != nil {
		return "", db.User{}, err
	}

	return token, user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, db.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" || password == "" {
		return "", db.User{}, ErrInvalidInput
	}

	user, err := s.users.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", db.User{}, ErrInvalidCredentials
		}
		return "", db.User{}, err
	}

	ok, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return "", db.User{}, ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(s.jwtSecret, user.ID, user.Email)
	if err != nil {
		return "", db.User{}, err
	}

	return token, user, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (db.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, err
	}
	return user, nil
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
	rows, err := s.users.Delete(ctx, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
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

func (s *LinkService) CreateLink(ctx context.Context, userID int64, target string, customAlias *string, password *string, waitEnabled *bool, waitSeconds *int32) (db.Link, error) {
	cleanTarget, err := validateTargetURL(target)
	if err != nil {
		return db.Link{}, err
	}

	var finalAlias string
	if customAlias != nil && strings.TrimSpace(*customAlias) != "" {
		validAlias, err := validateAlias(*customAlias)
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
	if password != nil && *password != "" {
		hashed, err := auth.HashPassword(*password)
		if err != nil {
			return db.Link{}, err
		}
		pwHash = pgtype.Text{String: hashed, Valid: true}
	}

	// Şifre açıksa bekletme kesinlikle kapalı olur.
	var finalWaitEnabled bool
	var finalWaitSeconds int32
	if !pwHash.Valid && waitEnabled != nil && *waitEnabled {
		finalWaitEnabled = true
		if waitSeconds == nil || *waitSeconds <= 0 {
			return db.Link{}, ErrInvalidWaitSeconds
		}
		finalWaitSeconds = *waitSeconds
	}

	link, err := s.links.Create(ctx, db.CreateLinkParams{
		UserID:       userID,
		Alias:        finalAlias,
		TargetUrl:    cleanTarget,
		PasswordHash: pwHash,
		WaitEnabled:  finalWaitEnabled,
		WaitSeconds:  finalWaitSeconds,
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

func (s *LinkService) ListLinks(ctx context.Context, userID int64) ([]db.Link, error) {
	return s.links.ListByUser(ctx, userID)
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
	id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
	if err != nil {
		return db.Link{}, ErrLinkNotFound
	}
	link, err := s.links.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrLinkNotFound
		}
		return db.Link{}, err
	}
	return link, nil
}

func (s *LinkService) UpdateLink(ctx context.Context, idStr string, userID int64, target, alias, password *string, waitEnabled *bool, waitSeconds *int32) (db.Link, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
	if err != nil {
		return db.Link{}, ErrLinkNotFound
	}

	existing, err := s.links.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrLinkNotFound
		}
		return db.Link{}, err
	}

	newTarget := existing.TargetUrl
	if target != nil && strings.TrimSpace(*target) != "" {
		cleanTarget, err := validateTargetURL(*target)
		if err != nil {
			return db.Link{}, err
		}
		newTarget = cleanTarget
	}

	newAlias := existing.Alias
	if alias != nil && strings.TrimSpace(*alias) != "" {
		validAlias, err := validateAlias(*alias)
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
	if password != nil {
		if *password == "" {
			newPwHash = pgtype.Text{Valid: false}
		} else {
			hashed, err := auth.HashPassword(*password)
			if err != nil {
				return db.Link{}, err
			}
			newPwHash = pgtype.Text{String: hashed, Valid: true}
		}
	}

	newWaitEnabled := existing.WaitEnabled
	if waitEnabled != nil {
		newWaitEnabled = *waitEnabled
	}

	newWaitSeconds := existing.WaitSeconds
	if waitSeconds != nil {
		newWaitSeconds = *waitSeconds
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

	updated, err := s.links.Update(ctx, existing.Alias, db.UpdateLinkParams{
		ID:           existing.ID,
		UserID:       userID,
		Alias:        newAlias,
		TargetUrl:    newTarget,
		PasswordHash: newPwHash,
		WaitEnabled:  newWaitEnabled,
		WaitSeconds:  newWaitSeconds,
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

func (s *LinkService) ListLinkLogs(ctx context.Context, idStr string, userID int64) ([]db.Log, error) {
	link, err := s.GetLink(ctx, idStr, userID)
	if err != nil {
		return nil, err
	}
	return s.links.ListLogsByLink(ctx, link.ID)
}

func (s *LinkService) DeleteLink(ctx context.Context, idStr string, userID int64) error {
	id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
	if err != nil {
		return ErrLinkNotFound
	}

	rows, err := s.links.Delete(ctx, id, userID)
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
