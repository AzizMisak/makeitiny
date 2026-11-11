package handler

import (
	"encoding/json"
	"errors"
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

// Redirect: Şifresiz linklerde 302 Found + Cache-Control: no-store, no-cache ile hedef URL'e yönlendirir.
// Şifreli linklerde kullanıcıyı Frontend üzerindeki /unlock/{link} ekranına yönlendirir.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request, link string) {
	target, err := h.redirects.Resolve(r.Context(), link)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkLocked):
			unlockURL := h.frontendURL + "/unlock/" + link
			w.Header().Set("Cache-Control", "no-store, no-cache")
			w.Header().Set("Location", unlockURL)
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
		Alias:  rec.Alias,
		Target: target,
		Locked: locked,
	})
}
