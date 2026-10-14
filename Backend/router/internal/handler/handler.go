package handler

import (
	"net/http"

	"makeitiny/router/internal/gen"
	"makeitiny/router/internal/service"
)

// Handler gen.ServerInterface arayüzünü uygular (HTTP katmanı).
type Handler struct {
	redirects *service.RedirectService
}

var _ gen.ServerInterface = (*Handler)(nil)

func New(redirects *service.RedirectService) *Handler {
	return &Handler{redirects: redirects}
}

func notImplemented(w http.ResponseWriter) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Redirect: 302 Found + Cache-Control: no-store, no-cache ile yönlendirecek.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request, link string) {
	notImplemented(w)
}

func (h *Handler) Unlock(w http.ResponseWriter, r *http.Request, link string)  { notImplemented(w) }
func (h *Handler) Preview(w http.ResponseWriter, r *http.Request, link string) { notImplemented(w) }
