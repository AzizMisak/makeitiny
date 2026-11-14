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

// authorize gelen isteğin JWT token'ını doğrular, kullanıcının aktifliğini ve (varsa) gereken rolleri kontrol eder.
func (h *Handler) authorize(w http.ResponseWriter, r *http.Request, requiredRoles ...string) (service.UserProfile, bool) {
	tokenStr, err := auth.ExtractToken(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		return service.UserProfile{}, false
	}

	profile, err := h.auth.Authorize(r.Context(), tokenStr, requiredRoles...)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserSuspended):
			writeError(w, http.StatusForbidden, "account_suspended", err.Error())
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
		default:
			writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum doğrulaması başarısız.")
		}
		return service.UserProfile{}, false
	}

	return profile, true
}

func toPlanItem(p db.Plan) gen.PlanItem {
	return gen.PlanItem{
		Id:               p.ID,
		Slug:             p.Slug,
		Name:             p.Name,
		PriceMonthly:     p.PriceMonthly,
		MaxLinks:         p.MaxLinks,
		AllowCustomAlias: p.AllowCustomAlias,
		AllowPassword:    p.AllowPassword,
		AllowQr:          p.AllowQr,
		AllowAds:         p.AllowAds,
		AnalyticsDays:    p.AnalyticsDays,
	}
}

func toUserModel(p service.UserProfile) gen.User {
	createdAt := ""
	if p.User.CreatedAt.Valid {
		createdAt = p.User.CreatedAt.Time.Format("2006-01-02")
	}
	var planExpiresAt *string
	if p.User.PlanExpiresAt.Valid {
		s := p.User.PlanExpiresAt.Time.Format("2006-01-02")
		planExpiresAt = &s
	}
	roles := p.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}
	return gen.User{
		Id:            strconv.FormatInt(p.User.ID, 10),
		Name:          p.User.Name,
		Email:         openapi_types.Email(p.User.Email),
		IsActive:      p.User.IsActive,
		Roles:         roles,
		Plan:          toPlanItem(p.Plan),
		PlanExpiresAt: planExpiresAt,
		LinkCount:     p.LinkCount,
		CreatedAt:     createdAt,
	}
}

func toLinkItem(l db.Link) gen.LinkItem {
	createdAt := ""
	if l.CreatedAt.Valid {
		createdAt = l.CreatedAt.Time.Format("2006-01-02")
	}
	var expiresAt *string
	if l.ExpiresAt.Valid {
		s := l.ExpiresAt.Time.Format("2006-01-02")
		expiresAt = &s
	}
	var maxClicks *int64
	if l.MaxClicks.Valid && l.MaxClicks.Int64 > 0 {
		mc := l.MaxClicks.Int64
		maxClicks = &mc
	}
	secStatus := l.SecurityStatus
	if secStatus == "" {
		secStatus = "unchecked"
	}
	return gen.LinkItem{
		Id:             l.PublicID,
		Alias:          l.Alias,
		Target:         l.TargetUrl,
		Title:          l.Title,
		Clicks:         l.Clicks,
		Locked:         l.PasswordHash.Valid && l.PasswordHash.String != "",
		WaitEnabled:    l.WaitEnabled,
		WaitSeconds:    l.WaitSeconds,
		IsActive:       l.IsActive,
		ExpiresAt:      expiresAt,
		MaxClicks:      maxClicks,
		AdEnabled:      l.AdEnabled,
		SecurityStatus: secStatus,
		CreatedAt:      createdAt,
	}
}

func toAdminLinkItem(row db.AdminLinkRow) gen.AdminLinkItem {
	base := toLinkItem(row.Link)
	return gen.AdminLinkItem{
		Id:             base.Id,
		Alias:          base.Alias,
		Target:         base.Target,
		Title:          base.Title,
		Clicks:         base.Clicks,
		Locked:         base.Locked,
		WaitEnabled:    base.WaitEnabled,
		WaitSeconds:    base.WaitSeconds,
		IsActive:       base.IsActive,
		ExpiresAt:      base.ExpiresAt,
		MaxClicks:      base.MaxClicks,
		AdEnabled:      base.AdEnabled,
		SecurityStatus: base.SecurityStatus,
		CreatedAt:      base.CreatedAt,
		UserId:         strconv.FormatInt(row.UserID, 10),
		UserEmail:      row.UserEmail,
		UserName:       row.UserName,
	}
}

func toLinkLogItem(lg db.Log, linkPublicID string) gen.LinkLogItem {
	clickedAt := ""
	if lg.ClickedAt.Valid {
		clickedAt = lg.ClickedAt.Time.Local().Format("2006-01-02 15:04:05")
	}
	return gen.LinkLogItem{
		Id:             strconv.FormatInt(lg.ID, 10),
		LinkId:         linkPublicID,
		IpAddress:      lg.IpAddress,
		UserAgent:      lg.UserAgent,
		DeviceType:     lg.DeviceType,
		Browser:        lg.Browser,
		Os:             lg.Os,
		Referer:        lg.Referer,
		AcceptLanguage: lg.AcceptLanguage,
		CountryCode:    lg.CountryCode,
		ClickedAt:      clickedAt,
	}
}

// Health & Plans

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Healthy"))
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.auth.ListPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Üyelik paketleri listelenemedi.")
		return
	}
	items := make([]gen.PlanItem, 0, len(plans))
	for _, p := range plans {
		items = append(items, toPlanItem(p))
	}
	writeJSON(w, http.StatusOK, items)
}

// Auth

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req gen.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	token, profile, err := h.auth.Register(r.Context(), req.Name, string(req.Email), req.Password)
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
		User:  toUserModel(profile),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req gen.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	token, profile, err := h.auth.Login(r.Context(), string(req.Email), req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
		case errors.Is(err, service.ErrUserSuspended):
			writeError(w, http.StatusForbidden, "account_suspended", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Giriş işlemi sırasında bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusOK, gen.AuthResponse{
		Token: token,
		User:  toUserModel(profile),
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authorize(w, r); !ok {
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
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, toUserModel(profile))
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	if err := h.auth.DeleteUser(r.Context(), profile.User.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", "Owner hesabı silinemez.")
		case errors.Is(err, service.ErrUserNotFound):
			writeError(w, http.StatusUnauthorized, "user_not_found", "Kullanıcı bulunamadı.")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Hesap silinirken bir hata oluştu.")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Links

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	var req gen.CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	link, err := h.links.CreateLink(r.Context(), profile, service.CreateLinkInput{
		Target:      req.Target,
		Alias:       req.Alias,
		Title:       req.Title,
		Password:    req.Password,
		WaitEnabled: req.WaitEnabled,
		WaitSeconds: req.WaitSeconds,
		ExpiresAt:   req.ExpiresAt,
		MaxClicks:   req.MaxClicks,
		AdEnabled:   req.AdEnabled,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlanLinkLimitReached):
			writeError(w, http.StatusForbidden, "plan_limit_reached", err.Error())
		case errors.Is(err, service.ErrPlanCustomAliasRestricted),
			errors.Is(err, service.ErrPlanPasswordRestricted),
			errors.Is(err, service.ErrPlanAdsRestricted):
			writeError(w, http.StatusForbidden, "plan_feature_restricted", err.Error())
		case errors.Is(err, service.ErrInvalidTargetURL),
			errors.Is(err, service.ErrInvalidAlias),
			errors.Is(err, service.ErrInvalidWaitSeconds),
			errors.Is(err, service.ErrInvalidInput):
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

func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request, params gen.ListLinksParams) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	result, err := h.links.ListLinks(r.Context(), profile.User.ID, params.Page, params.Limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Linkler listelenirken bir hata oluştu.")
		return
	}

	items := make([]gen.LinkItem, 0, len(result.Items))
	for _, l := range result.Items {
		items = append(items, toLinkItem(l))
	}

	topItems := make([]gen.LinkItem, 0, len(result.TopLinks))
	for _, l := range result.TopLinks {
		topItems = append(topItems, toLinkItem(l))
	}

	writeJSON(w, http.StatusOK, gen.PaginatedLinksResponse{
		Items:       items,
		Page:        result.Page,
		Limit:       result.Limit,
		Total:       result.Total,
		TotalPages:  result.TotalPages,
		TotalClicks: result.TotalClicks,
		LockedCount: result.LockedCount,
		TopLinks:    topItems,
	})
}

func (h *Handler) CheckAlias(w http.ResponseWriter, r *http.Request, alias string) {
	if _, ok := h.authorize(w, r); !ok {
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
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	link, err := h.links.GetLink(r.Context(), id, profile.User.ID)
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
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	var req gen.UpdateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	updated, err := h.links.UpdateLink(r.Context(), profile, id, service.UpdateLinkInput{
		Target:      req.Target,
		Alias:       req.Alias,
		Title:       req.Title,
		Password:    req.Password,
		WaitEnabled: req.WaitEnabled,
		WaitSeconds: req.WaitSeconds,
		IsActive:    req.IsActive,
		ExpiresAt:   req.ExpiresAt,
		MaxClicks:   req.MaxClicks,
		AdEnabled:   req.AdEnabled,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
		case errors.Is(err, service.ErrPlanCustomAliasRestricted),
			errors.Is(err, service.ErrPlanPasswordRestricted),
			errors.Is(err, service.ErrPlanAdsRestricted):
			writeError(w, http.StatusForbidden, "plan_feature_restricted", err.Error())
		case errors.Is(err, service.ErrInvalidTargetURL),
			errors.Is(err, service.ErrInvalidAlias),
			errors.Is(err, service.ErrInvalidWaitSeconds),
			errors.Is(err, service.ErrInvalidInput):
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

func (h *Handler) ListLinkLogs(w http.ResponseWriter, r *http.Request, id string, params gen.ListLinkLogsParams) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	result, err := h.links.ListLinkLogs(r.Context(), profile, id, params.Page, params.Limit)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Tıklama logları alınamadı.")
		return
	}

	items := make([]gen.LinkLogItem, 0, len(result.Items))
	for _, lg := range result.Items {
		items = append(items, toLinkLogItem(lg, id))
	}

	writeJSON(w, http.StatusOK, gen.PaginatedLogsResponse{
		Items:      items,
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request, id string) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	if err := h.links.DeleteLink(r.Context(), id, profile.User.ID); err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Link silinirken bir hata oluştu.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Admin & Owner Endpoints ---

func (h *Handler) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	stats, err := h.links.AdminGetStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Platform istatistikleri alınamadı.")
		return
	}

	writeJSON(w, http.StatusOK, gen.AdminStatsResponse{
		TotalUsers:   stats.TotalUsers,
		ActiveUsers:  stats.ActiveUsers,
		TotalLinks:   stats.TotalLinks,
		ActiveLinks:  stats.ActiveLinks,
		BlockedLinks: stats.BlockedLinks,
		TotalClicks:  stats.TotalClicks,
	})
}

func (h *Handler) AdminListLinks(w http.ResponseWriter, r *http.Request, params gen.AdminListLinksParams) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	var statusStr *string
	if params.Status != nil {
		s := string(*params.Status)
		statusStr = &s
	}

	result, err := h.links.AdminListLinks(r.Context(), params.Page, params.Limit, params.Search, statusStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Linkler listelenemedi.")
		return
	}

	items := make([]gen.AdminLinkItem, 0, len(result.Items))
	for _, row := range result.Items {
		items = append(items, toAdminLinkItem(row))
	}

	writeJSON(w, http.StatusOK, gen.AdminPaginatedLinksResponse{
		Items:      items,
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *Handler) AdminUpdateLink(w http.ResponseWriter, r *http.Request, id string) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	var req gen.AdminUpdateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	var secStatus *string
	if req.SecurityStatus != nil {
		s := string(*req.SecurityStatus)
		secStatus = &s
	}

	updated, err := h.links.AdminUpdateLink(r.Context(), id, service.UpdateLinkInput{
		Target:         req.Target,
		Alias:          req.Alias,
		Title:          req.Title,
		Password:       req.Password,
		WaitEnabled:    req.WaitEnabled,
		WaitSeconds:    req.WaitSeconds,
		IsActive:       req.IsActive,
		ExpiresAt:      req.ExpiresAt,
		MaxClicks:      req.MaxClicks,
		AdEnabled:      req.AdEnabled,
		SecurityStatus: secStatus,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
		case errors.Is(err, service.ErrInvalidTargetURL),
			errors.Is(err, service.ErrInvalidAlias),
			errors.Is(err, service.ErrInvalidWaitSeconds),
			errors.Is(err, service.ErrInvalidInput):
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

func (h *Handler) AdminDeleteLink(w http.ResponseWriter, r *http.Request, id string) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	if err := h.links.AdminDeleteLink(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Link silinirken bir hata oluştu.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminListLinkLogs(w http.ResponseWriter, r *http.Request, id string, params gen.AdminListLinkLogsParams) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	result, err := h.links.AdminListLinkLogs(r.Context(), id, params.Page, params.Limit)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Tıklama logları alınamadı.")
		return
	}

	items := make([]gen.LinkLogItem, 0, len(result.Items))
	for _, lg := range result.Items {
		items = append(items, toLinkLogItem(lg, id))
	}

	writeJSON(w, http.StatusOK, gen.PaginatedLogsResponse{
		Items:      items,
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request, params gen.AdminListUsersParams) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	result, err := h.auth.AdminListUsers(r.Context(), params.Page, params.Limit, params.Search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Kullanıcılar listelenemedi.")
		return
	}

	items := make([]gen.User, 0, len(result.Items))
	for _, prof := range result.Items {
		items = append(items, toUserModel(prof))
	}

	writeJSON(w, http.StatusOK, gen.AdminPaginatedUsersResponse{
		Items:      items,
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *Handler) AdminUpdateUser(w http.ResponseWriter, r *http.Request, id string) {
	actor, ok := h.authorize(w, r, "admin", "owner")
	if !ok {
		return
	}

	var req gen.AdminUpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	updated, err := h.auth.AdminUpdateUser(r.Context(), actor, id, req.IsActive, req.PlanId, req.PlanExpiresAt)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user_not_found", err.Error())
		case errors.Is(err, service.ErrPlanNotFound):
			writeError(w, http.StatusBadRequest, "plan_not_found", err.Error())
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Kullanıcı güncellenemedi.")
		}
		return
	}

	writeJSON(w, http.StatusOK, toUserModel(updated))
}

func (h *Handler) AdminUpdateUserRole(w http.ResponseWriter, r *http.Request, id string) {
	actor, ok := h.authorize(w, r, "owner")
	if !ok {
		return
	}

	var req gen.AdminUpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	updated, err := h.auth.AdminUpdateUserRole(r.Context(), actor, id, string(req.Role))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user_not_found", err.Error())
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid_role", "Geçersiz rol (yalnızca 'user' veya 'admin' atanabilir).")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Kullanıcı rolü güncellenemedi.")
		}
		return
	}

	writeJSON(w, http.StatusOK, toUserModel(updated))
}

func (h *Handler) AdminUpdatePlan(w http.ResponseWriter, r *http.Request, id int32) {
	actor, ok := h.authorize(w, r, "owner")
	if !ok {
		return
	}

	var req gen.AdminUpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	updated, err := h.auth.AdminUpdatePlan(r.Context(), actor, id, service.UpdatePlanInput{
		Name:             req.Name,
		PriceMonthly:     req.PriceMonthly,
		MaxLinks:         req.MaxLinks,
		AllowCustomAlias: req.AllowCustomAlias,
		AllowPassword:    req.AllowPassword,
		AllowQr:          req.AllowQr,
		AllowAds:         req.AllowAds,
		AnalyticsDays:    req.AnalyticsDays,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlanNotFound):
			writeError(w, http.StatusNotFound, "plan_not_found", err.Error())
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Paket güncellenemedi.")
		}
		return
	}

	writeJSON(w, http.StatusOK, toPlanItem(updated))
}
