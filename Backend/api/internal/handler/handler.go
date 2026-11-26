package handler

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) isHTTPS(r *http.Request) bool {
	base := ""
	if h.links != nil {
		base = h.links.PublicBaseURL()
	}
	return auth.IsHTTPS(r, base)
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

func isTrustedProxyIP(addr netip.Addr) bool {
	addr = addr.WithZone("").Unmap()
	return addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast()
}

func extractClientIP(r *http.Request) string {
	peerIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		peerIP = host
	}
	peerAddr, err := netip.ParseAddr(strings.Trim(peerIP, "[]"))
	if err != nil || !isTrustedProxyIP(peerAddr) {
		return peerIP
	}

	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		if parsed, err := netip.ParseAddr(strings.Trim(ip, "[]")); err == nil {
			return parsed.WithZone("").Unmap().String()
		}
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if p := strings.TrimSpace(parts[0]); p != "" {
			if parsed, err := netip.ParseAddr(strings.Trim(p, "[]")); err == nil {
				return parsed.WithZone("").Unmap().String()
			}
		}
	}
	return peerAddr.WithZone("").Unmap().String()
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

func toPlanItem(p service.PlanInfo) gen.PlanItem {
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
	if p.User.PlanExpiresAt.Valid && time.Now().Before(p.User.PlanExpiresAt.Time) {
		s := p.User.PlanExpiresAt.Time.Format("2006-01-02")
		planExpiresAt = &s
	}
	var bannedUntil *string
	if p.User.BannedUntil.Valid {
		s := p.User.BannedUntil.Time.Format(time.RFC3339)
		bannedUntil = &s
	}
	roles := p.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}
	id := p.User.PublicID
	if id == "" {
		id = strconv.FormatInt(p.User.ID, 10)
	}
	return gen.User{
		Id:            id,
		Name:          p.User.Name,
		Email:         openapi_types.Email(p.User.Email),
		IsActive:      p.User.IsActive,
		BanReason:     p.User.BanReason,
		BannedUntil:   bannedUntil,
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
	var planGraceUntil *string
	if l.PlanGraceUntil.Valid {
		s := l.PlanGraceUntil.Time.Format("2006-01-02")
		planGraceUntil = &s
	}
	hasQr := l.QrCodePath != ""
	qrUrl := ""
	if hasQr {
		qrUrl = "/api/links/" + l.PublicID + "/qr"
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
		HasQr:          hasQr,
		QrUrl:          qrUrl,
		IsAnonymous:    l.IsAnonymous,
		PlanGraceUntil: planGraceUntil,
		CreatedAt:      createdAt,
	}
}

func linkFromAdminRow(row db.ListAdminLinksPaginatedRow) db.Link {
	return db.Link{
		ID:                row.ID,
		UserID:            row.UserID,
		Alias:             row.Alias,
		TargetUrl:         row.TargetUrl,
		PasswordHash:      row.PasswordHash,
		Clicks:            row.Clicks,
		CreatedAt:         row.CreatedAt,
		WaitEnabled:       row.WaitEnabled,
		WaitSeconds:       row.WaitSeconds,
		PublicID:          row.PublicID,
		Title:             row.Title,
		IsActive:          row.IsActive,
		ExpiresAt:         row.ExpiresAt,
		MaxClicks:         row.MaxClicks,
		AdEnabled:         row.AdEnabled,
		SecurityStatus:    row.SecurityStatus,
		SecurityCheckedAt: row.SecurityCheckedAt,
		QrCodePath:        row.QrCodePath,
		UpdatedAt:         row.UpdatedAt,
		IsAnonymous:       row.IsAnonymous,
		CreatedByIp:       row.CreatedByIp,
	}
}

func toAdminLinkItem(row db.ListAdminLinksPaginatedRow) gen.AdminLinkItem {
	base := toLinkItem(linkFromAdminRow(row))
	userID := row.UserPublicID
	if userID == "" && row.UserID.Valid && row.UserID.Int64 > 0 {
		userID = strconv.FormatInt(row.UserID.Int64, 10)
	}
	if userID == "" {
		userID = "0"
	}
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
		HasQr:          base.HasQr,
		QrUrl:          base.QrUrl,
		IsAnonymous:    base.IsAnonymous,
		CreatedAt:      base.CreatedAt,
		UserId:         userID,
		UserEmail:      row.UserEmail,
		UserName:       row.UserName,
	}
}

func toAdminLinkDetailItem(row db.GetAdminLinkByPublicIDRow) gen.AdminLinkItem {
	base := toLinkItem(db.Link{
		ID:                row.ID,
		UserID:            row.UserID,
		Alias:             row.Alias,
		TargetUrl:         row.TargetUrl,
		PasswordHash:      row.PasswordHash,
		Clicks:            row.Clicks,
		CreatedAt:         row.CreatedAt,
		WaitEnabled:       row.WaitEnabled,
		WaitSeconds:       row.WaitSeconds,
		PublicID:          row.PublicID,
		Title:             row.Title,
		IsActive:          row.IsActive,
		ExpiresAt:         row.ExpiresAt,
		MaxClicks:         row.MaxClicks,
		AdEnabled:         row.AdEnabled,
		SecurityStatus:    row.SecurityStatus,
		SecurityCheckedAt: row.SecurityCheckedAt,
		QrCodePath:        row.QrCodePath,
		UpdatedAt:         row.UpdatedAt,
		IsAnonymous:       row.IsAnonymous,
		CreatedByIp:       row.CreatedByIp,
	})
	userID := row.UserPublicID
	if userID == "" && row.UserID.Valid && row.UserID.Int64 > 0 {
		userID = strconv.FormatInt(row.UserID.Int64, 10)
	}
	if userID == "" {
		userID = "0"
	}
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
		HasQr:          base.HasQr,
		QrUrl:          base.QrUrl,
		IsAnonymous:    base.IsAnonymous,
		CreatedAt:      base.CreatedAt,
		UserId:         userID,
		UserEmail:      row.UserEmail,
		UserName:       row.UserName,
	}
}

func toLinkLogItem(lg db.Log, linkPublicID string) gen.LinkLogItem {
	var clickedAt time.Time
	if lg.ClickedAt.Valid {
		clickedAt = lg.ClickedAt.Time.UTC()
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

// writeDecodeError gövde çözümlenemediğinde hatayı yazar. E-posta biçimi hatalıysa bunu açıkça söyler.
func writeDecodeError(w http.ResponseWriter, err error) {
	if errors.Is(err, openapi_types.ErrValidationEmail) {
		writeError(w, http.StatusBadRequest, "invalid_email", service.ErrInvalidEmail.Error())
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req gen.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	token, refreshToken, profile, err := h.auth.Register(r.Context(), req.Name, string(req.Email), req.Password, r.UserAgent(), extractClientIP(r))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			writeError(w, http.StatusBadRequest, "invalid_email", err.Error())
		case errors.Is(err, service.ErrInvalidName):
			writeError(w, http.StatusBadRequest, "invalid_name", err.Error())
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

	auth.SetRefreshCookie(w, refreshToken, int(auth.RefreshTokenTTL.Seconds()), h.isHTTPS(r))
	writeJSON(w, http.StatusCreated, gen.AuthResponse{
		Token: token,
		User:  toUserModel(profile),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req gen.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	token, refreshToken, profile, err := h.auth.Login(r.Context(), string(req.Email), req.Password, r.UserAgent(), extractClientIP(r))
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

	auth.SetRefreshCookie(w, refreshToken, int(auth.RefreshTokenTTL.Seconds()), h.isHTTPS(r))
	writeJSON(w, http.StatusOK, gen.AuthResponse{
		Token: token,
		User:  toUserModel(profile),
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	rawRefresh := auth.ExtractRefreshCookie(r)
	if rawRefresh == "" {
		var req gen.RefreshTokenRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RefreshToken != nil {
				rawRefresh = strings.TrimSpace(*req.RefreshToken)
			}
		}
	}

	token, newRefresh, profile, err := h.auth.RefreshSession(r.Context(), rawRefresh, r.UserAgent(), extractClientIP(r))
	if err != nil {
		auth.ClearRefreshCookie(w, h.isHTTPS(r))
		if errors.Is(err, service.ErrUserSuspended) {
			writeError(w, http.StatusForbidden, "account_suspended", err.Error())
			return
		}
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum yenilenemedi. Lütfen tekrar giriş yapın.")
		return
	}

	auth.SetRefreshCookie(w, newRefresh, int(auth.RefreshTokenTTL.Seconds()), h.isHTTPS(r))
	writeJSON(w, http.StatusOK, gen.AuthResponse{
		Token: token,
		User:  toUserModel(profile),
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenStr, _ := auth.ExtractToken(r)
	rawRefresh := auth.ExtractRefreshCookie(r)
	h.auth.Logout(r.Context(), tokenStr, rawRefresh)
	auth.ClearRefreshCookie(w, h.isHTTPS(r))

	writeJSON(w, http.StatusOK, gen.MessageResponse{
		Message: "Oturum başarıyla sonlandırıldı.",
	})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.authorize(w, r)
	if !ok {
		return
	}

	var req gen.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	curPw := ""
	if req.CurrentPassword != nil {
		curPw = *req.CurrentPassword
	}

	targetEmail := strings.TrimSpace(string(req.Email))
	isSelf := strings.EqualFold(targetEmail, actor.User.Email)

	if err := h.auth.ResetPassword(r.Context(), actor, string(req.Email), curPw, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			msg := "Bu kullanıcının parolasını değiştirme yetkiniz bulunmuyor."
			if !service.IsPrivileged(actor.Roles) && !isSelf {
				msg = "Yalnızca kendi şifrenizi değiştirebilirsiniz."
			}
			writeError(w, http.StatusForbidden, "forbidden", msg)
		case errors.Is(err, service.ErrInvalidCurrentPassword):
			writeError(w, http.StatusBadRequest, "invalid_current_password", err.Error())
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

	msg := "Kullanıcının parolası başarıyla güncellendi."
	if isSelf {
		msg = "Parolanız başarıyla güncellendi."
	}
	writeJSON(w, http.StatusOK, gen.MessageResponse{
		Message: msg,
	})
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, toUserModel(profile))
}

func (h *Handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	var req gen.UpdateCurrentUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	updated, err := h.auth.UpdateName(r.Context(), profile, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidName):
			writeError(w, http.StatusBadRequest, "invalid_name", err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			writeError(w, http.StatusUnauthorized, "user_not_found", "Kullanıcı bulunamadı.")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Ad güncellenemedi.")
		}
		return
	}

	writeJSON(w, http.StatusOK, toUserModel(updated))
}

func (h *Handler) ChangeCurrentUserPlan(w http.ResponseWriter, r *http.Request) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	var req gen.ChangePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	updated, err := h.auth.ChangeOwnPlan(r.Context(), profile, req.PlanId)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlanNotFound):
			writeError(w, http.StatusBadRequest, "plan_not_found", err.Error())
		case errors.Is(err, service.ErrPlanNotApplicable):
			writeError(w, http.StatusBadRequest, "plan_not_applicable", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Paket değiştirilemedi.")
		}
		return
	}

	writeJSON(w, http.StatusOK, toUserModel(updated))
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

	auth.ClearRefreshCookie(w, h.isHTTPS(r))
	w.WriteHeader(http.StatusNoContent)
}

// Links

func (h *Handler) CreateAnonymousLink(w http.ResponseWriter, r *http.Request) {
	var req gen.CreateAnonymousLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Geçersiz JSON gövdesi.")
		return
	}

	link, err := h.links.CreateAnonymousLink(r.Context(), service.CreateAnonymousLinkInput{
		Target:      req.Target,
		Password:    req.Password,
		WaitEnabled: req.WaitEnabled,
		WaitSeconds: req.WaitSeconds,
		TTLHours:    req.TtlHours,
		ClientIP:    extractClientIP(r),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAnonRateLimited):
			writeError(w, http.StatusTooManyRequests, "anon_rate_limited", err.Error())
		case errors.Is(err, service.ErrInvalidTargetURL),
			errors.Is(err, service.ErrInternalTargetURL),
			errors.Is(err, service.ErrBlankPassword),
			errors.Is(err, service.ErrInvalidWaitSeconds),
			errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Anonim link oluşturulurken bir hata oluştu.")
		}
		return
	}

	expiresAt := ""
	if link.ExpiresAt.Valid {
		expiresAt = link.ExpiresAt.Time.Format(time.RFC3339)
	}
	hasQr := link.QrCodePath != ""
	qrUrl := ""
	if hasQr {
		qrUrl = "/api/links/" + link.PublicID + "/qr"
	}

	writeJSON(w, http.StatusCreated, gen.AnonymousLinkResponse{
		Id:          link.PublicID,
		Alias:       link.Alias,
		ShortUrl:    h.links.PublicBaseURL() + "/r/" + link.Alias,
		Target:      link.TargetUrl,
		Locked:      link.PasswordHash.Valid && link.PasswordHash.String != "",
		WaitEnabled: link.WaitEnabled,
		WaitSeconds: link.WaitSeconds,
		ExpiresAt:   expiresAt,
		HasQr:       hasQr,
		QrUrl:       qrUrl,
	})
}

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
			errors.Is(err, service.ErrInternalTargetURL),
			errors.Is(err, service.ErrBlankPassword),
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
	if errors.Is(err, service.ErrAliasTaken) {
		available, err = false, nil
	}
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
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
		case errors.Is(err, service.ErrPlanCustomAliasRestricted),
			errors.Is(err, service.ErrPlanPasswordRestricted),
			errors.Is(err, service.ErrPlanAdsRestricted):
			writeError(w, http.StatusForbidden, "plan_feature_restricted", err.Error())
		case errors.Is(err, service.ErrInvalidTargetURL),
			errors.Is(err, service.ErrInternalTargetURL),
			errors.Is(err, service.ErrBlankPassword),
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

func (h *Handler) GenerateLinkQr(w http.ResponseWriter, r *http.Request, id string) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	updated, err := h.links.GenerateLinkQR(r.Context(), profile, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
		case errors.Is(err, service.ErrPlanQRRestricted):
			writeError(w, http.StatusForbidden, "plan_feature_restricted", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "QR kod üretilirken bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusOK, toLinkItem(updated))
}

func (h *Handler) GetLinkQr(w http.ResponseWriter, r *http.Request, id string, params gen.GetLinkQrParams) {
	alias, filePath, err := h.links.GetLinkQRFile(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "qr_not_found", "QR kod bulunamadı.")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store, no-cache")
	if params.Download != nil && (*params.Download == "1" || strings.EqualFold(*params.Download, "true")) {
		w.Header().Set("Content-Disposition", `attachment; filename="qr-`+alias+`.png"`)
	}
	http.ServeFile(w, r, filePath)
}

func (h *Handler) DeleteLinkQr(w http.ResponseWriter, r *http.Request, id string) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	if err := h.links.DeleteLinkQR(r.Context(), profile, id); err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "QR kod silinirken bir hata oluştu.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
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

func (h *Handler) AdminGetLink(w http.ResponseWriter, r *http.Request, id string) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	row, err := h.links.AdminGetLink(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Link bilgisi alınamadı.")
		return
	}

	writeJSON(w, http.StatusOK, toAdminLinkDetailItem(row))
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
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", "Engellenmiş veya zararlı bir link yayına alınamaz.")
		case errors.Is(err, service.ErrInvalidTargetURL),
			errors.Is(err, service.ErrInternalTargetURL),
			errors.Is(err, service.ErrBlankPassword),
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

func (h *Handler) AdminGetUser(w http.ResponseWriter, r *http.Request, id string) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	prof, err := h.auth.AdminGetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Kullanıcı bilgisi alınamadı.")
		return
	}

	writeJSON(w, http.StatusOK, toUserModel(prof))
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

	updated, err := h.auth.AdminUpdateUser(r.Context(), actor, id, req.IsActive, req.BanReason, req.BannedUntil, req.PlanId, req.PlanExpiresAt)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user_not_found", err.Error())
		case errors.Is(err, service.ErrPlanNotFound):
			writeError(w, http.StatusBadRequest, "plan_not_found", err.Error())
		case errors.Is(err, service.ErrPlanNotApplicable):
			writeError(w, http.StatusBadRequest, "plan_not_applicable", err.Error())
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
		case errors.Is(err, service.ErrInvalidPlanValues):
			writeError(w, http.StatusBadRequest, "invalid_plan_values", err.Error())
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Paket güncellenemedi.")
		}
		return
	}

	writeJSON(w, http.StatusOK, toPlanItem(updated))
}
