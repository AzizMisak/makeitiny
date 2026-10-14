package handler

import (
	"net/http"

	"makeitiny/api/internal/gen"
	"makeitiny/api/internal/service"
)

// Handler gen.ServerInterface arayüzünü uygular (HTTP katmanı).
type Handler struct {
	auth  *service.AuthService
	links *service.LinkService
}

var _ gen.ServerInterface = (*Handler)(nil)

func New(auth *service.AuthService, links *service.LinkService) *Handler {
	return &Handler{auth: auth, links: links}
}

func notImplemented(w http.ResponseWriter) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// Auth

func (h *Handler) Login(w http.ResponseWriter, r *http.Request)         { notImplemented(w) }
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request)        { notImplemented(w) }
func (h *Handler) Register(w http.ResponseWriter, r *http.Request)      { notImplemented(w) }
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	notImplemented(w)
}

// Links

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request)  { notImplemented(w) }
func (h *Handler) CheckAlias(w http.ResponseWriter, r *http.Request, alias string) {
	notImplemented(w)
}
func (h *Handler) GetLink(w http.ResponseWriter, r *http.Request, id string)    { notImplemented(w) }
func (h *Handler) UpdateLink(w http.ResponseWriter, r *http.Request, id string) { notImplemented(w) }
func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request, id string) { notImplemented(w) }
