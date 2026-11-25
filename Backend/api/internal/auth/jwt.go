package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenTTL     = 15 * time.Minute
	RefreshTokenTTL    = 7 * 24 * time.Hour
	RefreshCookieName  = "mit_refresh_token"
	AccessCookieName   = "mit_token"
)

var (
	ErrMissingToken = errors.New("yetkilendirme token'ı bulunamadı")
	ErrInvalidToken = errors.New("geçersiz veya süresi dolmuş token")
)

type Claims struct {
	UserID       int64    `json:"user_id"`
	PublicID     string   `json:"public_id"`
	Email        string   `json:"email"`
	Roles        []string `json:"roles,omitempty"`
	IssuedAtNano int64    `json:"iat_ns,omitempty"`
	jwt.RegisteredClaims
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken kullanıcı için 15 dakika geçerli, JTI içeren HS256 Access JWT üretir.
func GenerateToken(secret string, userID int64, publicID, email string, roles []string) (string, error) {
	now := time.Now()
	jti, err := randomHex(16)
	if err != nil {
		return "", err
	}
	claims := Claims{
		UserID:       userID,
		PublicID:     publicID,
		Email:        email,
		Roles:        roles,
		IssuedAtNano: now.UnixNano(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   publicID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken JWT token'ı doğrular ve içindeki Claims bilgisini döner.
func ParseToken(secret, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ExtractToken HTTP isteğindeki Authorization: Bearer <token> başlığından veya mit_token çerezinden token'ı çıkarır.
func ExtractToken(r *http.Request) (string, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && strings.TrimSpace(parts[1]) != "" {
			return strings.TrimSpace(parts[1]), nil
		}
		return "", ErrInvalidToken
	}

	if cookie, err := r.Cookie(AccessCookieName); err == nil && strings.TrimSpace(cookie.Value) != "" {
		return strings.TrimSpace(cookie.Value), nil
	}

	return "", ErrMissingToken
}

// GenerateRefreshToken rastgele 32 byte kriptografik token ve DB için SHA-256 hash'ini üretir.
func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	hash = HashRefreshToken(raw)
	return raw, hash, nil
}

// HashRefreshToken refresh token'ın SHA-256 hex özetini hesaplar.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func SetRefreshCookie(w http.ResponseWriter, rawToken string, maxAgeSeconds int) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    rawToken,
		Path:     "/",
		MaxAge:   maxAgeSeconds,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ExtractRefreshCookie(r *http.Request) string {
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}
