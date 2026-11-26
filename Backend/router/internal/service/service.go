package service

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"

	"makeitiny/router/internal/repository"
)

var (
	ErrLinkNotFound      = errors.New("aradığınız kısa link bulunamadı")
	ErrLinkDisabled      = errors.New("bu kısa link devre dışı bırakılmıştır")
	ErrLinkBlocked       = errors.New("bu kısa link güvenlik nedeniyle engellenmiştir")
	ErrLinkExpired       = errors.New("bu kısa linkin kullanım süresi dolmuştur")
	ErrLinkLimitReached  = errors.New("bu kısa link maksimum tıklama sınırına ulaşmıştır")
	ErrLinkLocked        = errors.New("bu kısa link parola korumalıdır")
	ErrLinkNotLocked     = errors.New("bu kısa link parola korumalı değildir")
	ErrLinkWaiting       = errors.New("bu kısa link bekletme sayfasına yönlendiriliyor")
	ErrInvalidPassword   = errors.New("girdiğiniz parola hatalı")
	ErrTooManyRequests   = errors.New("çok fazla deneme yaptınız; lütfen biraz bekleyip tekrar deneyin")
	ErrTargetUnreachable = errors.New("hedef adrese ulaşılamıyor")
	ErrTicketInvalid     = errors.New("bekleme bileti geçersiz veya kullanılmış; linki baştan açın")
	ErrWaitNotFinished   = errors.New("bekleme süresi henüz dolmadı")
)

// ClickLogMeta tıklayan cihazın HTTP isteğinden çıkarılan log verileri.
type ClickLogMeta struct {
	IPAddress      string
	UserAgent      string
	DeviceType     string
	Browser        string
	OS             string
	Referer        string
	AcceptLanguage string
	CountryCode    string
}

// WaitState bekleme sayfasının ihtiyaç duyduğu bilgi. Hedef adresi içermez.
type WaitState struct {
	Alias            string
	WaitSeconds      int32
	RemainingSeconds int32
	AdEnabled        bool
}

// RedirectService yönlendirme, kilit açma, bekleme ve önizleme iş mantığı.
type RedirectService struct {
	links     *repository.LinkRepository
	unlockSem chan struct{}
}

func NewRedirectService(links *repository.LinkRepository) *RedirectService {
	return &RedirectService{
		links:     links,
		unlockSem: make(chan struct{}, 8),
	}
}

// validateLinkState linkin yayında olup olmadığını denetler (tıklama hakkı hariç).
// Engelli durum, kapalı durumdan önce denetlenir: engellenen linkler aynı zamanda kapatıldığı için
// sıra ters olursa ziyaretçi "engellendi" yerine "kapatıldı" ekranını görür.
func validateLinkState(rec repository.LinkRecord, now time.Time) error {
	if rec.SecurityStatus == "blocked" || rec.SecurityStatus == "malicious" {
		return ErrLinkBlocked
	}
	if !rec.IsActive {
		return ErrLinkDisabled
	}
	if rec.ExpiresAtUnix > 0 && now.Unix() >= rec.ExpiresAtUnix {
		return ErrLinkExpired
	}
	return nil
}

// validateWaitState bekleme biletiyle devam edilirken tıklama hakkı hariç erişimi ve
// sonradan eklenmiş parola kilidini denetler.
func validateWaitState(rec repository.LinkRecord, now time.Time) error {
	if err := validateLinkState(rec, now); err != nil {
		return err
	}
	if rec.PasswordHash != "" {
		return ErrLinkLocked
	}
	return nil
}
func validateLinkAccess(rec repository.LinkRecord, now time.Time) error {
	if err := validateLinkState(rec, now); err != nil {
		return err
	}
	if rec.MaxClicks > 0 && rec.Clicks >= rec.MaxClicks {
		return ErrLinkLimitReached
	}
	return nil
}

func buildLogEntry(rec repository.LinkRecord, meta ClickLogMeta, eventType string) repository.LogQueueEntry {
	return repository.LogQueueEntry{
		LinkID:         rec.ID,
		LinkAlias:      rec.Alias,
		LinkPublicID:   rec.PublicID,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
		DeviceType:     meta.DeviceType,
		Browser:        meta.Browser,
		OS:             meta.OS,
		Referer:        meta.Referer,
		AcceptLanguage: meta.AcceptLanguage,
		CountryCode:    meta.CountryCode,
		EventType:      eventType,
		ClickedAt:      time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *RedirectService) load(ctx context.Context, alias string) (repository.LinkRecord, error) {
	rec, err := s.links.GetByAlias(ctx, strings.TrimSpace(alias))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.LinkRecord{}, ErrLinkNotFound
		}
		return repository.LinkRecord{}, err
	}
	return rec, nil
}

// verifyTargetReachable hedefin erişilebilirliğini denetler. Olumlu sonuç 5 dakika,
// olumsuz sonuç 30 saniye önbelleklenir; böylece ölü hedef her istekte beklemeye yol açmaz.
func (s *RedirectService) verifyTargetReachable(ctx context.Context, rec repository.LinkRecord) error {
	switch s.links.CachedReachability(ctx, rec.Alias) {
	case repository.ReachOK:
		return nil
	case repository.ReachFailed:
		return ErrTargetUnreachable
	}
	if err := checkTargetReachable(ctx, rec.TargetURL); err != nil {
		s.links.CacheReachability(ctx, rec.Alias, false)
		return ErrTargetUnreachable
	}
	s.links.CacheReachability(ctx, rec.Alias, true)
	return nil
}

// Resolve kısa linki çözümler. Şifreli linklerde tıklama sayılmaz (sayım /unlock başarısında yapılır).
// Bekletmeli linklerde tıklama burada sayılır ve hedef yerine tek kullanımlık bekleme bileti döner.
func (s *RedirectService) Resolve(ctx context.Context, alias string, meta ClickLogMeta) (target string, ticket string, err error) {
	rec, err := s.load(ctx, alias)
	if err != nil {
		return "", "", err
	}

	if err := validateLinkAccess(rec, time.Now()); err != nil {
		return "", "", err
	}

	if rec.PasswordHash != "" {
		return "", "", ErrLinkLocked
	}

	if err := s.verifyTargetReachable(ctx, rec); err != nil {
		return "", "", err
	}

	if !s.links.RecordClickAndLog(ctx, rec, buildLogEntry(rec, meta, "redirect")) {
		return "", "", ErrLinkLimitReached
	}

	if rec.WaitEnabled && rec.WaitSeconds > 0 {
		ticket, terr := s.links.CreateWaitTicket(ctx, rec.Alias, time.Duration(rec.WaitSeconds)*time.Second)
		if terr != nil {
			// Bilet saklanamazsa (Redis yok) ziyaretçiyi bekletmeden hedefe gönder; tıklama zaten sayıldı.
			log.Printf("router bekleme bileti oluşturulamadı (%s): %v", rec.Alias, terr)
			return rec.TargetURL, "", nil
		}
		return "", ticket, ErrLinkWaiting
	}

	return rec.TargetURL, "", nil
}

// Unlock yalnızca parola korumalı linkin parolasını Argon2id ile doğrular.
// Parola doğruysa tıklama sayılır ve unlock_success logu kaydedilir.
func (s *RedirectService) Unlock(ctx context.Context, alias, password string, meta ClickLogMeta) (string, error) {
	rec, err := s.load(ctx, alias)
	if err != nil {
		return "", err
	}

	if err := validateLinkAccess(rec, time.Now()); err != nil {
		return "", err
	}

	if rec.PasswordHash == "" {
		return "", ErrLinkNotLocked
	}

	if !s.links.AllowUnlockAttempt(ctx, rec.Alias, meta.IPAddress) {
		return "", ErrTooManyRequests
	}

	ok, err := s.verifyPasswordLimited(ctx, password, rec.PasswordHash)
	if err != nil {
		if errors.Is(err, ErrTooManyRequests) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
		return "", ErrInvalidPassword
	}
	if !ok {
		return "", ErrInvalidPassword
	}

	if err := s.verifyTargetReachable(ctx, rec); err != nil {
		return "", err
	}

	if !s.links.RecordClickAndLog(ctx, rec, buildLogEntry(rec, meta, "unlock_success")) {
		return "", ErrLinkLimitReached
	}

	return rec.TargetURL, nil
}

func (s *RedirectService) verifyPasswordLimited(ctx context.Context, password, encodedHash string) (bool, error) {
	if s.unlockSem != nil {
		timer := time.NewTimer(2 * time.Second)
		select {
		case s.unlockSem <- struct{}{}:
			timer.Stop()
			defer func() { <-s.unlockSem }()
		case <-timer.C:
			return false, ErrTooManyRequests
		case <-ctx.Done():
			timer.Stop()
			return false, ctx.Err()
		}
	}
	return verifyArgon2id(password, encodedHash)
}

// WaitInfo bileti doğrular ve kalan süreyi döner. Ziyaret zaten sayıldığı için tıklama hakkı denetlenmez.
func (s *RedirectService) WaitInfo(ctx context.Context, alias, ticket string) (WaitState, error) {
	t, err := s.links.GetWaitTicket(ctx, ticket)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return WaitState{}, ErrTicketInvalid
		}
		return WaitState{}, err
	}
	if t.Alias != strings.TrimSpace(alias) {
		return WaitState{}, ErrTicketInvalid
	}

	rec, err := s.load(ctx, alias)
	if err != nil {
		return WaitState{}, err
	}
	if err := validateWaitState(rec, time.Now()); err != nil {
		return WaitState{}, err
	}

	remaining := time.Until(t.ReadyAt)
	secs := int32(0)
	if remaining > 0 {
		secs = int32((remaining + time.Second - 1) / time.Second)
	}

	return WaitState{
		Alias:            rec.Alias,
		WaitSeconds:      rec.WaitSeconds,
		RemainingSeconds: secs,
		AdEnabled:        rec.AdEnabled,
	}, nil
}

// Continue süre dolduysa bileti tüketir ve hedef adresi döner. Bilet tek kullanımlıktır.
func (s *RedirectService) Continue(ctx context.Context, alias, ticket string) (string, error) {
	t, err := s.links.GetWaitTicket(ctx, ticket)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return "", ErrTicketInvalid
		}
		return "", err
	}
	if t.Alias != strings.TrimSpace(alias) {
		return "", ErrTicketInvalid
	}
	if time.Now().Before(t.ReadyAt) {
		return "", ErrWaitNotFinished
	}

	rec, err := s.load(ctx, alias)
	if err != nil {
		return "", err
	}
	if err := validateWaitState(rec, time.Now()); err != nil {
		return "", err
	}

	consumed, err := s.links.ConsumeWaitTicket(ctx, ticket)
	if err != nil {
		return "", err
	}
	if !consumed {
		return "", ErrTicketInvalid
	}
	return rec.TargetURL, nil
}

// Preview kısa linkin erişim türünü döner. Tıklama saymadığı için hedef adresi döndürmez.
func (s *RedirectService) Preview(ctx context.Context, alias string) (repository.LinkRecord, error) {
	rec, err := s.load(ctx, alias)
	if err != nil {
		return repository.LinkRecord{}, err
	}
	if err := validateLinkAccess(rec, time.Now()); err != nil {
		return repository.LinkRecord{}, err
	}
	return rec, nil
}

// checkTargetReachable hedefi SSRF'e kapalı bir istemciyle denetler. HEAD ağ hatası verirse (SSRF hariç)
// veya erişilebilir olmayan bir durum kodu dönerse (WAF 503 hariç) GET (Range: bytes=0-0) ile bir kez daha bakılır.
func checkTargetReachable(ctx context.Context, rawURL string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return ErrTargetUnreachable
	}

	headCtx, headCancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	status, headers, err := probe(headCtx, http.MethodHead, parsed.String())
	headCancel()

	if err == nil && (reachableStatus(status) || isWAFChallenge(status, headers)) {
		return nil
	}
	if err != nil && errors.Is(err, errBlockedAddress) {
		return ErrTargetUnreachable
	}

	getCtx, getCancel := context.WithTimeout(ctx, 3500*time.Millisecond)
	status, headers, err = probe(getCtx, http.MethodGet, parsed.String())
	getCancel()

	if err == nil && (reachableStatus(status) || isWAFChallenge(status, headers)) {
		return nil
	}
	return ErrTargetUnreachable
}

// isWAFChallenge Cloudflare vb. WAF 503 meydan okuma yanıtlarını ayırt eder.
func isWAFChallenge(status int, headers http.Header) bool {
	if status != http.StatusServiceUnavailable || headers == nil {
		return false
	}
	if strings.TrimSpace(headers.Get("Cf-Mitigated")) != "" {
		return true
	}
	return strings.Contains(strings.ToLower(headers.Get("Server")), "cloudflare")
}

// reachableStatus: erişim/doğrulama isteyen (401/403/406) ve hız sınırlayan (429) siteler de "ayakta" sayılır.
func reachableStatus(code int) bool {
	return code < 400 || code == 401 || code == 403 || code == 406 || code == 429
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
