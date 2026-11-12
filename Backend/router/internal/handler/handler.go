package handler

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
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

// Redirect: Şifresiz ve bekletmesiz linklerde 302 Found + Cache-Control: no-store, no-cache ile hedef URL'e yönlendirir.
// Şifreli linklerde kullanıcıyı Frontend üzerindeki /unlock/{link} ekranına yönlendirir.
// Bekletme açık olan linklerde kullanıcıyı Frontend üzerindeki /w/{link} ekranına yönlendirir.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request, link string) {
	meta := extractClickMeta(r)
	target, err := h.redirects.Resolve(r.Context(), link, meta)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkLocked):
			unlockURL := h.frontendURL + "/unlock/" + link
			w.Header().Set("Cache-Control", "no-store, no-cache")
			w.Header().Set("Location", unlockURL)
			w.WriteHeader(http.StatusFound)
		case errors.Is(err, service.ErrLinkWaiting):
			waitURL := h.frontendURL + "/w/" + link
			w.Header().Set("Cache-Control", "no-store, no-cache")
			w.Header().Set("Location", waitURL)
			w.WriteHeader(http.StatusFound)
		case errors.Is(err, service.ErrLinkNotFound):
			writeError(w, http.StatusNotFound, "link_not_found", "Aradığınız kısa link bulunamadı.")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Yönlendirme sırasında bir hata oluştu.")
		}
		return
	}

	w.Header().Set("Cache-Control", "no-store, no-cache")
	w.Header().Set("Location", target)
	w.WriteHeader(http.StatusFound)
}

func (h *Handler) Unlock(w http.ResponseWriter, r *http.Request, link string) {
	var req gen.UnlockLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Lütfen geçerli bir parola giriniz.")
		return
	}

	target, err := h.redirects.Unlock(r.Context(), link, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPassword):
			writeError(w, http.StatusUnauthorized, "invalid_password", "Girdiğiniz parola hatalı.")
		case errors.Is(err, service.ErrLinkNotFound):
			writeError(w, http.StatusNotFound, "link_not_found", "Aradığınız kısa link bulunamadı.")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Kilit açma işlemi sırasında bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusOK, gen.UnlockLinkResponse{
		Target: target,
	})
}

func (h *Handler) Preview(w http.ResponseWriter, r *http.Request, link string) {
	rec, err := h.redirects.Preview(r.Context(), link)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", "Aradığınız kısa link bulunamadı.")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Önizleme bilgisi alınamadı.")
		return
	}

	locked := rec.PasswordHash != ""
	target := rec.TargetURL
	if locked {
		target = ""
	}

	writeJSON(w, http.StatusOK, gen.LinkPreview{
		Alias:       rec.Alias,
		Target:      target,
		Locked:      locked,
		WaitEnabled: rec.WaitEnabled,
		WaitSeconds: rec.WaitSeconds,
	})
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

	return service.ClickLogMeta{
		IPAddress:      ip,
		UserAgent:      ua,
		DeviceType:     deviceType,
		Browser:        browser,
		OS:             osName,
		Referer:        referer,
		AcceptLanguage: lang,
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
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ios"):
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
