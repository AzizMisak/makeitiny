package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"makeitiny/api/internal/auth"
	"makeitiny/api/internal/db"
	"makeitiny/api/internal/gen"
	"makeitiny/api/internal/service"
)

// Handler gen.ServerInterface arayüzünü uygular (HTTP katmanı).
type Handler struct {
	auth  *service.AuthService
	links *service.LinkService
}

var _ gen.ServerInterface = (*Handler)(nil)

func New(authSvc *service.AuthService, linkSvc *service.LinkService) *Handler {
	return &Handler{
		auth:  authSvc,
		links: linkSvc,
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

func (h *Handler) authenticate(r *http.Request) (*auth.Claims, error) {
	tokenStr, err := auth.ExtractToken(r)
	if err != nil {
		return nil, err
	}
	return h.auth.VerifyToken(tokenStr)
}

func toUserModel(u db.User) gen.User {
	return gen.User{
		Id:    strconv.FormatInt(u.ID, 10),
		Name:  u.Name,
		Email: openapi_types.Email(u.Email),
	}
}

func toLinkItem(l db.Link) gen.LinkItem {
	createdAt := ""
	if l.CreatedAt.Valid {
		createdAt = l.CreatedAt.Time.Format("2006-01-02")
	}
	return gen.LinkItem{
		Id:        strconv.FormatInt(l.ID, 10),
		Alias:     l.Alias,
		Target:    l.TargetUrl,
		Clicks:    l.Clicks,
		Locked:    l.PasswordHash.Valid && l.PasswordHash.String != "",
		CreatedAt: createdAt,
	}
}

// Health

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Healthy"))
}

// Auth

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req gen.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	token, user, err := h.auth.Register(r.Context(), req.Name, string(req.Email), req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			writeError(w, http.StatusBadRequest, "invalid_email", err.Error())
		case errors.Is(err, service.ErrPasswordTooShort):
			writeError(w, http.StatusBadRequest, "password_too_short", err.Error())
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		case errors.Is(err, service.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email_taken", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Kayıt işlemi sırasında bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusCreated, gen.AuthResponse{
		Token: token,
		User:  toUserModel(user),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req gen.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	token, user, err := h.auth.Login(r.Context(), string(req.Email), req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Giriş işlemi sırasında bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusOK, gen.AuthResponse{
		Token: token,
		User:  toUserModel(user),
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	writeJSON(w, http.StatusOK, gen.MessageResponse{
		Message: "Oturum başarıyla sonlandırıldı.",
	})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req gen.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	if err := h.auth.ResetPassword(r.Context(), string(req.Email), req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrPasswordTooShort):
			writeError(w, http.StatusBadRequest, "password_too_short", err.Error())
		case errors.Is(err, service.ErrInvalidEmail):
			writeError(w, http.StatusBadRequest, "invalid_email", err.Error())
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user_not_found", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Şifre sıfırlama sırasında bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusOK, gen.MessageResponse{
		Message: "Parolanız başarıyla güncellendi.",
	})
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	user, err := h.auth.GetCurrentUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "user_not_found", "Kullanıcı bulunamadı.")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Kullanıcı bilgileri alınamadı.")
		return
	}

	writeJSON(w, http.StatusOK, toUserModel(user))
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	if err := h.auth.DeleteUser(r.Context(), claims.UserID); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "user_not_found", "Kullanıcı bulunamadı.")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Hesap silinirken bir hata oluştu.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Links

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	var req gen.CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	link, err := h.links.CreateLink(r.Context(), claims.UserID, req.Target, req.Alias, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTargetURL), errors.Is(err, service.ErrInvalidAlias):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		case errors.Is(err, service.ErrAliasTaken):
			writeError(w, http.StatusConflict, "alias_taken", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Link oluşturulurken bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toLinkItem(link))
}

func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	links, err := h.links.ListLinks(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Linkler listelenirken bir hata oluştu.")
		return
	}

	items := make([]gen.LinkItem, 0, len(links))
	for _, l := range links {
		items = append(items, toLinkItem(l))
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) CheckAlias(w http.ResponseWriter, r *http.Request, alias string) {
	if _, err := h.authenticate(r); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	available, err := h.links.CheckAlias(r.Context(), alias)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAlias) {
			writeError(w, http.StatusBadRequest, "invalid_alias", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Takma ad kontrolü sırasında bir hata oluştu.")
		return
	}

	writeJSON(w, http.StatusOK, gen.AliasCheckResult{
		Alias:     alias,
		Available: available,
	})
}

func (h *Handler) GetLink(w http.ResponseWriter, r *http.Request, id string) {
	claims, err := h.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	link, err := h.links.GetLink(r.Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Link bilgisi alınamadı.")
		return
	}

	writeJSON(w, http.StatusOK, toLinkItem(link))
}

func (h *Handler) UpdateLink(w http.ResponseWriter, r *http.Request, id string) {
	claims, err := h.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	var req gen.UpdateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	updated, err := h.links.UpdateLink(r.Context(), id, claims.UserID, req.Target, req.Alias, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
		case errors.Is(err, service.ErrInvalidTargetURL), errors.Is(err, service.ErrInvalidAlias):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		case errors.Is(err, service.ErrAliasTaken):
			writeError(w, http.StatusConflict, "alias_taken", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Link güncellenirken bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusOK, toLinkItem(updated))
}

func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request, id string) {
	claims, err := h.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return
	}

	if err := h.links.DeleteLink(r.Context(), id, claims.UserID); err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Link silinirken bir hata oluştu.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
