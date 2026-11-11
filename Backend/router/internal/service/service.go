package service

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"

	"makeitiny/router/internal/repository"
)

var (
	ErrLinkNotFound    = errors.New("aradığınız kısa link bulunamadı")
	ErrLinkLocked      = errors.New("bu kısa link parola korumalıdır")
	ErrInvalidPassword = errors.New("girdiğiniz parola hatalı")
)

// RedirectService yönlendirme, kilit açma ve önizleme iş mantığı.
type RedirectService struct {
	links *repository.LinkRepository
}

func NewRedirectService(links *repository.LinkRepository) *RedirectService {
	return &RedirectService{links: links}
}

// Resolve kısa linki çözümler. Şifreliyse ErrLinkLocked döner; şifresizse tıklama sayısını artırıp hedef URL'i döner.
func (s *RedirectService) Resolve(ctx context.Context, alias string) (string, error) {
	rec, err := s.links.GetByAlias(ctx, strings.TrimSpace(alias))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrLinkNotFound
		}
		return "", err
	}

	if rec.PasswordHash != "" {
		return "", ErrLinkLocked
	}

	_ = s.links.IncrementClicks(ctx, rec.Alias)
	return rec.TargetURL, nil
}

// Unlock şifreli kısa linkin parolasını Argon2id ile doğrular, doğruysa tıklama sayısını artırıp hedef URL'i döner.
func (s *RedirectService) Unlock(ctx context.Context, alias, password string) (string, error) {
	rec, err := s.links.GetByAlias(ctx, strings.TrimSpace(alias))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrLinkNotFound
		}
		return "", err
	}

	if rec.PasswordHash != "" {
		ok, err := verifyArgon2id(password, rec.PasswordHash)
		if err != nil || !ok {
			return "", ErrInvalidPassword
		}
	}

	_ = s.links.IncrementClicks(ctx, rec.Alias)
	return rec.TargetURL, nil
}

// Preview kısa linkin önizleme bilgisini döner (şifreliyse hedef adresi gizler).
func (s *RedirectService) Preview(ctx context.Context, alias string) (repository.LinkRecord, error) {
	rec, err := s.links.GetByAlias(ctx, strings.TrimSpace(alias))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.LinkRecord{}, ErrLinkNotFound
		}
		return repository.LinkRecord{}, err
	}
	return rec, nil
}

func verifyArgon2id(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("geçersiz argon2id hash formatı")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}

	var memory, timeCost uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(expectedHash)))
	return subtle.ConstantTimeCompare(expectedHash, computedHash) == 1, nil
}
