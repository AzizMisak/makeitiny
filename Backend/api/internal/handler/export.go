package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"makeitiny/api/internal/db"
	"makeitiny/api/internal/gen"
	"makeitiny/api/internal/service"
)

// csvSafe Excel ve benzerlerinin hücreyi formül olarak çalıştırmasını önler (CSV injection).
// Ziyaretçiden gelen referer ve tarayıcı bilgisi de dosyaya yazıldığı için gereklidir.
func csvSafe(v string) string {
	if v == "" {
		return v
	}
	switch v[0] {
	case '=', '+', '-', '@', '\t', '\r', '\n':
		return "'" + v
	}
	trimmed := strings.TrimLeft(v, " \t\r\n")
	if trimmed != "" {
		switch trimmed[0] {
		case '=', '+', '-', '@':
			return "'" + v
		}
	}
	return v
}

func exportFormat(raw *string) (string, bool) {
	if raw == nil || *raw == "" {
		return "csv", true
	}
	f := strings.ToLower(*raw)
	return f, f == "csv" || f == "json"
}

// writeCSV UTF-8 BOM ile başlayan, virgülle ayrılmış bir dosya yazar; Excel Türkçe karakterleri doğru açar.
func writeCSV(w http.ResponseWriter, filename string, header []string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})

	cw := csv.NewWriter(w)
	_ = cw.Write(header)
	for _, row := range rows {
		for i := range row {
			row[i] = csvSafe(row[i])
		}
		_ = cw.Write(row)
	}
	cw.Flush()
}

func writeAttachmentJSON(w http.ResponseWriter, filename string, v any) {
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, v)
}

func yesNo(b bool) string {
	if b {
		return "Evet"
	}
	return "Hayır"
}

func optString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (h *Handler) ExportLinks(w http.ResponseWriter, r *http.Request, params gen.ExportLinksParams) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	var raw *string
	if params.Format != nil {
		v := string(*params.Format)
		raw = &v
	}
	format, valid := exportFormat(raw)
	if !valid {
		writeError(w, http.StatusBadRequest, "invalid_format", "Biçim csv ya da json olmalıdır.")
		return
	}

	links, err := h.links.ExportLinks(r.Context(), profile.User.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Linkler dışa aktarılamadı.")
		return
	}

	stamp := time.Now().Format("2006-01-02")
	items := make([]gen.LinkItem, 0, len(links))
	for _, l := range links {
		items = append(items, toLinkItem(l))
	}

	if format == "json" {
		writeAttachmentJSON(w, "makeitiny-linkler-"+stamp+".json", items)
		return
	}

	base := h.links.PublicBaseURL() + "/r/"
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		rows = append(rows, linkItemCSVRow(base, it))
	}
	writeCSV(w, "makeitiny-linkler-"+stamp+".csv", []string{
		"Kısa adres", "Takma ad", "Hedef adres", "Başlık", "Tıklama", "Yönlendirme", "İnceleme",
		"Erişim", "Bekleme (sn)", "Son kullanma", "Tıklama sınırı", "Reklam", "QR kodu", "Oluşturulma",
	}, rows)
}

func linkItemCSVRow(base string, it gen.LinkItem) []string {
	access := "Herkese açık"
	if it.Locked {
		access = "Parolalı"
	} else if it.WaitEnabled {
		access = "Bekletmeli"
	}
	maxClicks := ""
	if it.MaxClicks != nil {
		maxClicks = strconv.FormatInt(*it.MaxClicks, 10)
	}
	return []string{
		base + it.Alias,
		it.Alias,
		it.Target,
		it.Title,
		strconv.FormatInt(it.Clicks, 10),
		map[bool]string{true: "Açık", false: "Kapalı"}[it.IsActive],
		it.SecurityStatus,
		access,
		strconv.Itoa(int(it.WaitSeconds)),
		optString(it.ExpiresAt),
		maxClicks,
		yesNo(it.AdEnabled),
		yesNo(it.HasQr),
		it.CreatedAt,
	}
}

func adminLinkItemCSVRow(base string, it gen.AdminLinkItem) []string {
	access := "Herkese açık"
	if it.Locked {
		access = "Parolalı"
	} else if it.WaitEnabled {
		access = "Bekletmeli"
	}
	maxClicks := ""
	if it.MaxClicks != nil {
		maxClicks = strconv.FormatInt(*it.MaxClicks, 10)
	}
	return []string{
		base + it.Alias,
		it.Alias,
		it.Target,
		it.Title,
		it.UserName,
		it.UserEmail,
		it.UserId,
		yesNo(it.IsAnonymous),
		strconv.FormatInt(it.Clicks, 10),
		map[bool]string{true: "Açık", false: "Kapalı"}[it.IsActive],
		it.SecurityStatus,
		access,
		strconv.Itoa(int(it.WaitSeconds)),
		optString(it.ExpiresAt),
		maxClicks,
		yesNo(it.AdEnabled),
		yesNo(it.HasQr),
		it.CreatedAt,
	}
}

func (h *Handler) writeAdminLinksExport(w http.ResponseWriter, format, filenameBase string, items []gen.AdminLinkItem) {
	if format == "json" {
		writeAttachmentJSON(w, filenameBase+".json", items)
		return
	}
	base := h.links.PublicBaseURL() + "/r/"
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		rows = append(rows, adminLinkItemCSVRow(base, it))
	}
	writeCSV(w, filenameBase+".csv", []string{
		"Kısa adres", "Takma ad", "Hedef adres", "Başlık", "Sahip adı", "Sahip e-postası", "Sahip kimliği",
		"Üyeliksiz", "Tıklama", "Yönlendirme", "İnceleme", "Erişim", "Bekleme (sn)",
		"Son kullanma", "Tıklama sınırı", "Reklam", "QR kodu", "Oluşturulma",
	}, rows)
}

func toAdminLinkItemFromUser(l db.Link, owner db.User) gen.AdminLinkItem {
	base := toLinkItem(l)
	userID := owner.PublicID
	if userID == "" && owner.ID > 0 {
		userID = strconv.FormatInt(owner.ID, 10)
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
		UserEmail:      owner.Email,
		UserName:       owner.Name,
	}
}

func (h *Handler) AdminExportLinks(w http.ResponseWriter, r *http.Request, params gen.AdminExportLinksParams) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	var raw *string
	if params.Format != nil {
		v := string(*params.Format)
		raw = &v
	}
	format, valid := exportFormat(raw)
	if !valid {
		writeError(w, http.StatusBadRequest, "invalid_format", "Biçim csv ya da json olmalıdır.")
		return
	}

	var statusStr *string
	if params.Status != nil {
		s := string(*params.Status)
		statusStr = &s
	}

	rows, err := h.links.AdminExportLinks(r.Context(), params.Search, statusStr)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Linkler dışa aktarılamadı.")
		return
	}

	items := make([]gen.AdminLinkItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAdminLinkItem(row))
	}

	stamp := time.Now().Format("2006-01-02")
	h.writeAdminLinksExport(w, format, "makeitiny-yonetim-linkler-"+stamp, items)
}

func (h *Handler) AdminExportUsers(w http.ResponseWriter, r *http.Request, params gen.AdminExportUsersParams) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	var raw *string
	if params.Format != nil {
		v := string(*params.Format)
		raw = &v
	}
	format, valid := exportFormat(raw)
	if !valid {
		writeError(w, http.StatusBadRequest, "invalid_format", "Biçim csv ya da json olmalıdır.")
		return
	}

	profiles, err := h.auth.AdminExportUsers(r.Context(), params.Search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Kullanıcılar dışa aktarılamadı.")
		return
	}

	items := make([]gen.User, 0, len(profiles))
	for _, prof := range profiles {
		items = append(items, toUserModel(prof))
	}

	stamp := time.Now().Format("2006-01-02")
	name := "makeitiny-kullanicilar-" + stamp
	if format == "json" {
		writeAttachmentJSON(w, name+".json", items)
		return
	}

	rows := make([][]string, 0, len(items))
	for _, u := range items {
		planName := u.Plan.Name
		planExp := optString(u.PlanExpiresAt)
		if service.IsPrivileged(u.Roles) {
			planName = "Pakete tabi değil"
			planExp = ""
		}
		rows = append(rows, []string{
			u.Id,
			u.Name,
			string(u.Email),
			strings.Join(u.Roles, ", "),
			map[bool]string{true: "Aktif", false: "Askıda"}[u.IsActive],
			u.BanReason,
			optString(u.BannedUntil),
			planName,
			planExp,
			strconv.FormatInt(u.LinkCount, 10),
			u.CreatedAt,
		})
	}
	writeCSV(w, name+".csv", []string{
		"Kimlik", "Ad", "E-posta", "Rol", "Durum", "Askı sebebi", "Askı bitişi",
		"Paket", "Paket bitişi", "Link sayısı", "Kayıt tarihi",
	}, rows)
}

func (h *Handler) AdminExportUserLinks(w http.ResponseWriter, r *http.Request, id string, params gen.AdminExportUserLinksParams) {
	if _, ok := h.authorize(w, r, "admin", "owner"); !ok {
		return
	}

	var raw *string
	if params.Format != nil {
		v := string(*params.Format)
		raw = &v
	}
	format, valid := exportFormat(raw)
	if !valid {
		writeError(w, http.StatusBadRequest, "invalid_format", "Biçim csv ya da json olmalıdır.")
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

	links, err := h.links.ExportLinks(r.Context(), prof.User.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Kullanıcının linkleri dışa aktarılamadı.")
		return
	}

	items := make([]gen.AdminLinkItem, 0, len(links))
	for _, l := range links {
		items = append(items, toAdminLinkItemFromUser(l, prof.User))
	}

	slug := prof.User.PublicID
	if slug == "" {
		slug = strconv.FormatInt(prof.User.ID, 10)
	}
	stamp := time.Now().Format("2006-01-02")
	h.writeAdminLinksExport(w, format, fmt.Sprintf("makeitiny-%s-linkler-%s", slug, stamp), items)
}

func (h *Handler) ExportLinkLogs(w http.ResponseWriter, r *http.Request, id string, params gen.ExportLinkLogsParams) {
	profile, ok := h.authorize(w, r)
	if !ok {
		return
	}

	var raw *string
	if params.Format != nil {
		v := string(*params.Format)
		raw = &v
	}
	format, valid := exportFormat(raw)
	if !valid {
		writeError(w, http.StatusBadRequest, "invalid_format", "Biçim csv ya da json olmalıdır.")
		return
	}

	link, logs, err := h.links.ExportLinkLogs(r.Context(), profile, id)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			writeError(w, http.StatusNotFound, "link_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Ziyaret kayıtları dışa aktarılamadı.")
		return
	}

	name := fmt.Sprintf("makeitiny-%s-ziyaretler-%s", link.Alias, time.Now().Format("2006-01-02"))
	items := make([]gen.LinkLogItem, 0, len(logs))
	for _, lg := range logs {
		items = append(items, toLinkLogItem(lg, id))
	}

	if format == "json" {
		writeAttachmentJSON(w, name+".json", items)
		return
	}

	rows := make([][]string, 0, len(items))
	for _, it := range items {
		rows = append(rows, []string{
			it.ClickedAt.UTC().Format(time.RFC3339),
			it.DeviceType,
			it.Browser,
			it.Os,
			it.Referer,
			it.AcceptLanguage,
			it.CountryCode,
			it.IpAddress,
			it.UserAgent,
		})
	}
	writeCSV(w, name+".csv", []string{
		"Zaman (UTC)", "Cihaz", "Tarayıcı", "İşletim sistemi", "Geldiği yer", "Tarayıcı dili", "Ülke kodu", "IP adresi", "Tarayıcı bilgisi",
	}, rows)
}
