package handler

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"makeitiny/router/internal/gen"
	"makeitiny/router/internal/service"
)

// Handler gen.ServerInterface arayüzünü uygular (HTTP katmanı).
type Handler struct {
	redirects   *service.RedirectService
	frontendURL string
}

var _ gen.ServerInterface = (*Handler)(nil)

func New(redirects *service.RedirectService, frontendURL string) *Handler {
	if frontendURL == "" {
		frontendURL = "http://localhost"
	}
	return &Handler{
		redirects:   redirects,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, gen.ApiError{
		Code:    code,
		Message: message,
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Redirect GET isteklerinde her durumda 302 döner. HEAD isteklerinde (önizleyici botlar, curl -I)
// tıklama saymadan ve bekleme bileti üretmeden salt-okunur durum döner.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request, link string) {
	w.Header().Set("Cache-Control", "no-store, no-cache")

	if r.Method == http.MethodHead {
		rec, err := h.redirects.Preview(r.Context(), link)
		if err != nil {
			w.Header().Set("Location", h.frontendURL+"/error/"+url.PathEscape(link)+"?reason="+reasonFor(err))
			w.WriteHeader(http.StatusFound)
			return
		}
		if rec.PasswordHash != "" {
			w.Header().Set("Location", h.frontendURL+"/unlock/"+url.PathEscape(link))
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	target, ticket, err := h.redirects.Resolve(r.Context(), link, extractClickMeta(r))

	if err == nil {
		w.Header().Set("Location", target)
		w.WriteHeader(http.StatusFound)
		return
	}

	escaped := url.PathEscape(link)
	location := h.frontendURL + "/error/" + escaped + "?reason=" + reasonFor(err)
	switch {
	case errors.Is(err, service.ErrLinkLocked):
		location = h.frontendURL + "/unlock/" + escaped
		if origRef := strings.TrimSpace(r.Referer()); origRef != "" && !strings.HasPrefix(origRef, h.frontendURL+"/unlock/") {
			if len(origRef) > 512 {
				origRef = origRef[:512]
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "mit_unlock_ref",
				Value:    url.QueryEscape(link + "|" + origRef),
				Path:     "/r/",
				MaxAge:   600,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   h.isSecureRequest(r),
			})
		}
	case errors.Is(err, service.ErrLinkWaiting):
		location = h.frontendURL + "/w/" + escaped + "?t=" + url.QueryEscape(ticket)
	}
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusFound)
}

func (h *Handler) isSecureRequest(r *http.Request) bool {
	return strings.HasPrefix(strings.ToLower(h.frontendURL), "https://") ||
		r.TLS != nil ||
		strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// RootNotFound kısa kod içermeyen istekleri (GET /r/) hata ekranına gönderir.
func (h *Handler) RootNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache")
	w.Header().Set("Location", h.frontendURL+"/error?reason=not_found")
	w.WriteHeader(http.StatusFound)
}

// reasonFor hata ekranındaki ?reason= değerini üretir.
func reasonFor(err error) string {
	switch {
	case errors.Is(err, service.ErrLinkBlocked):
		return "blocked"
	case errors.Is(err, service.ErrLinkDisabled):
		return "disabled"
	case errors.Is(err, service.ErrLinkExpired):
		return "expired"
	case errors.Is(err, service.ErrLinkLimitReached):
		return "limit_reached"
	case errors.Is(err, service.ErrLinkNotFound):
		return "not_found"
	case errors.Is(err, service.ErrTargetUnreachable):
		return "unreachable"
	default:
		return "internal"
	}
}

// writeServiceError servis hatalarını JSON uç noktaları için ortak HTTP durumlarına çevirir.
func writeServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrInvalidPassword):
		writeError(w, http.StatusUnauthorized, "invalid_password", "Girdiğiniz parola hatalı.")
	case errors.Is(err, service.ErrTooManyRequests):
		writeError(w, http.StatusTooManyRequests, "too_many_requests", err.Error())
	case errors.Is(err, service.ErrLinkLocked):
		writeError(w, http.StatusForbidden, "link_locked", err.Error())
	case errors.Is(err, service.ErrLinkNotLocked):
		writeError(w, http.StatusBadRequest, "link_not_locked", err.Error())
	case errors.Is(err, service.ErrLinkBlocked):
		writeError(w, http.StatusForbidden, "link_blocked", err.Error())
	case errors.Is(err, service.ErrLinkDisabled):
		writeError(w, http.StatusGone, "link_disabled", err.Error())
	case errors.Is(err, service.ErrLinkExpired):
		writeError(w, http.StatusGone, "link_expired", err.Error())
	case errors.Is(err, service.ErrLinkLimitReached):
		writeError(w, http.StatusGone, "link_limit_reached", err.Error())
	case errors.Is(err, service.ErrTicketInvalid):
		writeError(w, http.StatusGone, "ticket_invalid", err.Error())
	case errors.Is(err, service.ErrWaitNotFinished):
		writeError(w, http.StatusTooEarly, "wait_not_finished", err.Error())
	case errors.Is(err, service.ErrLinkNotFound):
		writeError(w, http.StatusNotFound, "link_not_found", "Aradığınız kısa link bulunamadı.")
	case errors.Is(err, service.ErrTargetUnreachable):
		writeError(w, http.StatusBadGateway, "target_unreachable", "Bu kısa linkin yönlendirdiği hedef adrese şu anda ulaşılamıyor.")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", fallback)
	}
}

func (h *Handler) Unlock(w http.ResponseWriter, r *http.Request, link string) {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	var req gen.UnlockLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Lütfen geçerli bir parola giriniz.")
		return
	}

	meta := extractClickMeta(r)
	restoredRef := false
	if c, err := r.Cookie("mit_unlock_ref"); err == nil && c.Value != "" {
		if raw, err := url.QueryUnescape(c.Value); err == nil {
			if cookieLink, origRef, ok := strings.Cut(raw, "|"); ok && cookieLink == link && strings.TrimSpace(origRef) != "" {
				meta.Referer = origRef
				restoredRef = true
				http.SetCookie(w, &http.Cookie{
					Name:     "mit_unlock_ref",
					Value:    "",
					Path:     "/r/",
					MaxAge:   -1,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
					Secure:   h.isSecureRequest(r),
				})
			}
		}
	}
	if !restoredRef && strings.HasPrefix(meta.Referer, h.frontendURL+"/unlock/") {
		meta.Referer = "Doğrudan"
	}

	target, err := h.redirects.Unlock(r.Context(), link, req.Password, meta)
	if err != nil {
		writeServiceError(w, err, "Kilit açma işlemi sırasında bir hata oluştu.")
		return
	}

	writeJSON(w, http.StatusOK, gen.TargetResponse{Target: target})
}

func (h *Handler) Preview(w http.ResponseWriter, r *http.Request, link string) {
	rec, err := h.redirects.Preview(r.Context(), link)
	if err != nil {
		writeServiceError(w, err, "Önizleme bilgisi alınamadı.")
		return
	}

	writeJSON(w, http.StatusOK, gen.LinkPreview{
		Alias:       rec.Alias,
		Locked:      rec.PasswordHash != "",
		WaitEnabled: rec.WaitEnabled,
		WaitSeconds: rec.WaitSeconds,
		AdEnabled:   rec.AdEnabled,
	})
}

func (h *Handler) WaitInfo(w http.ResponseWriter, r *http.Request, link string, params gen.WaitInfoParams) {
	if strings.TrimSpace(params.T) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Bekleme bileti eksik.")
		return
	}

	state, err := h.redirects.WaitInfo(r.Context(), link, params.T)
	if err != nil {
		writeServiceError(w, err, "Bekleme bilgisi alınamadı.")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, gen.WaitInfo{
		Alias:            state.Alias,
		WaitSeconds:      state.WaitSeconds,
		RemainingSeconds: state.RemainingSeconds,
		AdEnabled:        state.AdEnabled,
	})
}

func (h *Handler) ContinueWait(w http.ResponseWriter, r *http.Request, link string) {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	var req gen.ContinueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Ticket) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Bekleme bileti eksik.")
		return
	}

	target, err := h.redirects.Continue(r.Context(), link, req.Ticket)
	if err != nil {
		writeServiceError(w, err, "Hedef adres alınamadı.")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, gen.TargetResponse{Target: target})
}

var langToCountryMap = map[string]string{
	"tr": "TR", "en": "US", "de": "DE", "fr": "FR", "es": "ES",
	"it": "IT", "nl": "NL", "ru": "RU", "az": "AZ", "ar": "SA",
	"pt": "PT", "zh": "CN", "ja": "JP", "ko": "KR", "pl": "PL",
	"uk": "UA", "el": "GR", "bg": "BG", "ro": "RO",
}

func resolveCountryCode(r *http.Request, acceptLang string) string {
	for _, hdr := range []string{"CF-IPCountry", "X-Country-Code"} {
		if c := strings.ToUpper(strings.TrimSpace(r.Header.Get(hdr))); len(c) == 2 && c != "XX" {
			return c
		}
	}
	if acceptLang == "" || acceptLang == "-" {
		return ""
	}
	normalized := strings.ReplaceAll(acceptLang, "_", "-")
	parts := strings.Split(normalized, "-")
	if len(parts) >= 2 && len(parts[1]) == 2 {
		return strings.ToUpper(parts[1])
	}
	if len(parts) >= 1 {
		if code, ok := langToCountryMap[strings.ToLower(parts[0])]; ok {
			return code
		}
	}
	return ""
}

func extractClickMeta(r *http.Request) service.ClickLogMeta {
	ip := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if ip == "" {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			ip = strings.TrimSpace(parts[0])
		}
	}
	if ip == "" {
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}

	ua := strings.TrimSpace(r.UserAgent())
	deviceType, browser, osName := parseUserAgent(ua)

	referer := strings.TrimSpace(r.Referer())
	if referer == "" {
		referer = "Doğrudan"
	}

	lang := strings.TrimSpace(r.Header.Get("Accept-Language"))
	if lang != "" {
		if idx := strings.IndexAny(lang, ",;"); idx > 0 {
			lang = strings.TrimSpace(lang[:idx])
		}
	} else {
		lang = "-"
	}

	countryCode := resolveCountryCode(r, lang)

	return service.ClickLogMeta{
		IPAddress:      ip,
		UserAgent:      ua,
		DeviceType:     deviceType,
		Browser:        browser,
		OS:             osName,
		Referer:        referer,
		AcceptLanguage: lang,
		CountryCode:    countryCode,
	}
}

func parseUserAgent(ua string) (deviceType, browser, osName string) {
	lower := strings.ToLower(ua)

	// Device Type
	switch {
	case strings.Contains(lower, "bot") || strings.Contains(lower, "crawler") || strings.Contains(lower, "spider") || strings.Contains(lower, "curl") || strings.Contains(lower, "postman"):
		deviceType = "Bot / Araç"
	case strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet"):
		deviceType = "Tablet"
	case strings.Contains(lower, "mobi") || strings.Contains(lower, "iphone") || strings.Contains(lower, "android"):
		deviceType = "Mobil"
	case ua == "":
		deviceType = "Bilinmiyor"
	default:
		deviceType = "Masaüstü"
	}

	// Browser
	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera"):
		browser = "Opera"
	case strings.Contains(lower, "chrome/") || strings.Contains(lower, "crios/"):
		browser = "Chrome"
	case strings.Contains(lower, "firefox/") || strings.Contains(lower, "fxios/"):
		browser = "Firefox"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome"):
		browser = "Safari"
	case strings.Contains(lower, "curl/"):
		browser = "curl"
	default:
		browser = "Diğer"
	}

	// OS
	switch {
	case strings.Contains(lower, "android"):
		osName = "Android"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ipod") || strings.Contains(lower, "cpu os"):
		osName = "iOS"
	case strings.Contains(lower, "windows"):
		osName = "Windows"
	case strings.Contains(lower, "mac os x") || strings.Contains(lower, "macintosh"):
		osName = "macOS"
	case strings.Contains(lower, "linux"):
		osName = "Linux"
	default:
		osName = "Diğer"
	}

	return deviceType, browser, osName
}
