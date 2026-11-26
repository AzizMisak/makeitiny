package handler

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"makeitiny/api/internal/auth"
	"makeitiny/api/internal/db"
	"makeitiny/api/internal/gen"
	"makeitiny/api/internal/repository"
	"makeitiny/api/internal/service"
)

type hStubRow struct {
	scanFn func(dest ...any) error
}

func (r hStubRow) Scan(dest ...any) error {
	return r.scanFn(dest...)
}

type hStubRows struct {
	pgx.Rows
	items []string
	idx   int
}

func (r *hStubRows) Close()                        {}
func (r *hStubRows) Err() error                    { return nil }
func (r *hStubRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }
func (r *hStubRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}
func (r *hStubRows) Next() bool {
	if r.idx < len(r.items) {
		r.idx++
		return true
	}
	return false
}
func (r *hStubRows) Scan(dest ...any) error {
	if p, ok := dest[0].(*string); ok {
		*p = r.items[r.idx-1]
	}
	return nil
}
func (r *hStubRows) Values() ([]any, error) { return nil, nil }
func (r *hStubRows) RawValues() [][]byte    { return nil }
func (r *hStubRows) Conn() *pgx.Conn        { return nil }

type hStubDBTX struct {
	usersByID    map[int64]db.User
	usersByEmail map[string]db.User
	rolesByUser  map[int64][]string
	linksByID    map[string]db.Link
}

func (s *hStubDBTX) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "UPDATE users") && strings.Contains(sql, "password_hash") {
		email := args[0].(string)
		if _, ok := s.usersByEmail[email]; ok {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		}
		return pgconn.NewCommandTag("UPDATE 0"), nil
	}
	return pgconn.NewCommandTag("UPDATE 0"), nil
}

func (s *hStubDBTX) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM roles") {
		uid := args[0].(int64)
		return &hStubRows{items: s.rolesByUser[uid]}, nil
	}
	return &hStubRows{}, nil
}

func (s *hStubDBTX) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	switch {
	case strings.Contains(sql, "FROM users") && strings.Contains(sql, "WHERE public_id = $1"):
		pid := args[0].(string)
		return hStubRow{scanFn: func(dest ...any) error {
			for _, u := range s.usersByID {
				if u.PublicID == pid {
					hAssignUser(dest, u)
					return nil
				}
			}
			return pgx.ErrNoRows
		}}
	case strings.Contains(sql, "FROM users") && strings.Contains(sql, "WHERE id = $1"):
		uid := args[0].(int64)
		return hStubRow{scanFn: func(dest ...any) error {
			u, ok := s.usersByID[uid]
			if !ok {
				return pgx.ErrNoRows
			}
			hAssignUser(dest, u)
			return nil
		}}
	case strings.Contains(sql, "FROM users") && strings.Contains(sql, "WHERE email = $1"):
		email := args[0].(string)
		return hStubRow{scanFn: func(dest ...any) error {
			u, ok := s.usersByEmail[email]
			if !ok {
				return pgx.ErrNoRows
			}
			hAssignUser(dest, u)
			return nil
		}}
	case strings.Contains(sql, "FROM plans") && strings.Contains(sql, "WHERE id = $1"):
		return hStubRow{scanFn: func(dest ...any) error { return pgx.ErrNoRows }}
	case strings.Contains(sql, "SELECT COUNT(*) FROM links"):
		return hStubRow{scanFn: func(dest ...any) error {
			*dest[0].(*int64) = 0
			return nil
		}}
	case strings.Contains(sql, "FROM links") && strings.Contains(sql, "WHERE public_id = $1"):
		pid := args[0].(string)
		return hStubRow{scanFn: func(dest ...any) error {
			l, ok := s.linksByID[pid]
			if !ok {
				return pgx.ErrNoRows
			}
			hAssignLink(dest, l)
			return nil
		}}
	case strings.Contains(sql, "UPDATE links"):
		return hStubRow{scanFn: func(dest ...any) error {
			out := db.Link{
				ID:                args[0].(int64),
				Alias:             args[1].(string),
				TargetUrl:         args[2].(string),
				PasswordHash:      args[3].(pgtype.Text),
				WaitEnabled:       args[4].(bool),
				WaitSeconds:       args[5].(int32),
				Title:             args[6].(string),
				IsActive:          args[7].(bool),
				ExpiresAt:         args[8].(pgtype.Timestamptz),
				MaxClicks:         args[9].(pgtype.Int8),
				AdEnabled:         args[10].(bool),
				SecurityStatus:    args[11].(string),
				SecurityCheckedAt: args[12].(pgtype.Timestamptz),
				PublicID:          "lnk1",
			}
			hAssignLink(dest, out)
			return nil
		}}
	}
	return hStubRow{scanFn: func(dest ...any) error { return pgx.ErrNoRows }}
}

func hAssignUser(dest []any, u db.User) {
	*dest[0].(*int64) = u.ID
	*dest[1].(*string) = u.Name
	*dest[2].(*string) = u.Email
	*dest[3].(*string) = u.PasswordHash
	*dest[4].(*pgtype.Timestamptz) = u.CreatedAt
	*dest[5].(*int32) = u.PlanID
	*dest[6].(*pgtype.Timestamptz) = u.PlanExpiresAt
	*dest[7].(*bool) = u.IsActive
	*dest[8].(*pgtype.Timestamptz) = u.UpdatedAt
	*dest[9].(*string) = u.PublicID
	*dest[10].(*string) = u.BanReason
	*dest[11].(*pgtype.Timestamptz) = u.BannedAt
	*dest[12].(*pgtype.Timestamptz) = u.BannedUntil
}

func hAssignLink(dest []any, l db.Link) {
	*dest[0].(*int64) = l.ID
	*dest[1].(*pgtype.Int8) = l.UserID
	*dest[2].(*string) = l.Alias
	*dest[3].(*string) = l.TargetUrl
	*dest[4].(*pgtype.Text) = l.PasswordHash
	*dest[5].(*int64) = l.Clicks
	*dest[6].(*pgtype.Timestamptz) = l.CreatedAt
	*dest[7].(*bool) = l.WaitEnabled
	*dest[8].(*int32) = l.WaitSeconds
	*dest[9].(*string) = l.PublicID
	*dest[10].(*string) = l.Title
	*dest[11].(*bool) = l.IsActive
	*dest[12].(*pgtype.Timestamptz) = l.ExpiresAt
	*dest[13].(*pgtype.Int8) = l.MaxClicks
	*dest[14].(*bool) = l.AdEnabled
	*dest[15].(*string) = l.SecurityStatus
	*dest[16].(*pgtype.Timestamptz) = l.SecurityCheckedAt
	*dest[17].(*string) = l.QrCodePath
	*dest[18].(*pgtype.Timestamptz) = l.UpdatedAt
	*dest[19].(*bool) = l.IsAnonymous
	*dest[20].(*string) = l.CreatedByIp
}

func newTestHandler(t *testing.T) (*Handler, string, string) {
	t.Helper()
	const secret = "handler-test-secret"
	pwHash, err := auth.HashPassword("EskiSifre123")
	if err != nil {
		t.Fatalf("hash hatası: %v", err)
	}

	stdUser := db.User{ID: 1, PublicID: "u1", Name: "Kullanıcı", Email: "user@example.com", PasswordHash: pwHash, IsActive: true, PlanID: 1}
	adminUser := db.User{ID: 2, PublicID: "a2", Name: "Admin", Email: "admin@example.com", PasswordHash: pwHash, IsActive: true, PlanID: 1}
	ownerUser := db.User{ID: 3, PublicID: "o3", Name: "Owner", Email: "owner@example.com", PasswordHash: pwHash, IsActive: true, PlanID: 1}

	stub := &hStubDBTX{
		usersByID: map[int64]db.User{
			1: stdUser,
			2: adminUser,
			3: ownerUser,
		},
		usersByEmail: map[string]db.User{
			stdUser.Email:   stdUser,
			adminUser.Email: adminUser,
			ownerUser.Email: ownerUser,
		},
		rolesByUser: map[int64][]string{
			1: {"user"},
			2: {"user", "admin"},
			3: {"user", "owner"},
		},
		linksByID: map[string]db.Link{
			"lnk1": {
				ID:             10,
				PublicID:       "lnk1",
				Alias:          "blocked1",
				TargetUrl:      "https://example.com",
				IsActive:       false,
				SecurityStatus: "blocked",
			},
		},
	}

	queries := db.New(stub)
	userRepo := repository.NewUserRepository(nil, queries, nil, nil)
	linkRepo := repository.NewLinkRepository(nil, queries, nil, nil)
	h := New(service.NewAuthService(userRepo, secret), service.NewLinkService(linkRepo, nil, "http://localhost"))

	userTok, err := auth.GenerateToken(secret, stdUser.ID, stdUser.PublicID, stdUser.Email, []string{"user"})
	if err != nil {
		t.Fatal(err)
	}
	adminTok, err := auth.GenerateToken(secret, adminUser.ID, adminUser.PublicID, adminUser.Email, []string{"user", "admin"})
	if err != nil {
		t.Fatal(err)
	}
	return h, userTok, adminTok
}

func TestHandlerAdminUpdateLinkBlockedForbiddenAndUnblock(t *testing.T) {
	h, _, adminTok := newTestHandler(t)

	// Engelli linki engel kaldırmadan yayına alma -> 403 forbidden
	req := httptest.NewRequest(http.MethodPatch, "/admin/links/lnk1", bytes.NewBufferString(`{"isActive":true}`))
	req.Header.Set("Authorization", "Bearer "+adminTok)
	rec := httptest.NewRecorder()
	h.AdminUpdateLink(rec, req, "lnk1")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("403 beklendi, gelen: %d (%s)", rec.Code, rec.Body.String())
	}
	var apiErr gen.ApiError
	if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil || apiErr.Code != "forbidden" {
		t.Fatalf("forbidden ApiError beklendi, gelen: %s", rec.Body.String())
	}

	// Engeli kaldırıp aynı istekte yayına alma -> 200 OK
	req2 := httptest.NewRequest(http.MethodPatch, "/admin/links/lnk1", bytes.NewBufferString(`{"securityStatus":"safe","isActive":true}`))
	req2.Header.Set("Authorization", "Bearer "+adminTok)
	rec2 := httptest.NewRecorder()
	h.AdminUpdateLink(rec2, req2, "lnk1")

	if rec2.Code != http.StatusOK {
		t.Fatalf("200 beklendi, gelen: %d (%s)", rec2.Code, rec2.Body.String())
	}
	var item gen.LinkItem
	if err := json.Unmarshal(rec2.Body.Bytes(), &item); err != nil || item.SecurityStatus != "safe" || !item.IsActive {
		t.Fatalf("safe + active=true beklendi, gelen: %s", rec2.Body.String())
	}
}

func TestHandlerResetPasswordCurrentPasswordAndMessages(t *testing.T) {
	h, userTok, adminTok := newTestHandler(t)

	// 1. Kendi parolasını mevcut parola olmadan değiştirme -> 400 invalid_current_password
	req := httptest.NewRequest(http.MethodPost, "/auth/resetpass", bytes.NewBufferString(`{"email":"user@example.com","newPassword":"YeniSifre123"}`))
	req.Header.Set("Authorization", "Bearer "+userTok)
	rec := httptest.NewRecorder()
	h.ResetPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("400 beklendi, gelen: %d (%s)", rec.Code, rec.Body.String())
	}
	var apiErr gen.ApiError
	if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil || apiErr.Code != "invalid_current_password" {
		t.Fatalf("invalid_current_password beklendi, gelen: %s", rec.Body.String())
	}

	// 2. Kendi parolasını doğru mevcut parola ile değiştirme -> 200
	req2 := httptest.NewRequest(http.MethodPost, "/auth/resetpass", bytes.NewBufferString(`{"email":"user@example.com","currentPassword":"EskiSifre123","newPassword":"YeniSifre123"}`))
	req2.Header.Set("Authorization", "Bearer "+userTok)
	rec2 := httptest.NewRecorder()
	h.ResetPassword(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("200 beklendi, gelen: %d (%s)", rec2.Code, rec2.Body.String())
	}
	var msg gen.MessageResponse
	if err := json.Unmarshal(rec2.Body.Bytes(), &msg); err != nil || msg.Message != "Parolanız başarıyla güncellendi." {
		t.Fatalf("kendi parola güncelleme mesajı hatalı: %s", rec2.Body.String())
	}

	// 3. Standart kullanıcı başkasının parolasını değiştirmeye çalışırsa -> 403
	req3 := httptest.NewRequest(http.MethodPost, "/auth/resetpass", bytes.NewBufferString(`{"email":"admin@example.com","newPassword":"YeniSifre123"}`))
	req3.Header.Set("Authorization", "Bearer "+userTok)
	rec3 := httptest.NewRecorder()
	h.ResetPassword(rec3, req3)

	if rec3.Code != http.StatusForbidden {
		t.Fatalf("403 beklendi, gelen: %d", rec3.Code)
	}
	_ = json.Unmarshal(rec3.Body.Bytes(), &apiErr)
	if apiErr.Message != "Yalnızca kendi şifrenizi değiştirebilirsiniz." {
		t.Fatalf("beklenmeyen 403 mesajı: %q", apiErr.Message)
	}

	// 4. Admin, owner'ın parolasını değiştirmeye çalışırsa -> 403 yetki mesajı
	req4 := httptest.NewRequest(http.MethodPost, "/auth/resetpass", bytes.NewBufferString(`{"email":"owner@example.com","newPassword":"YeniSifre123"}`))
	req4.Header.Set("Authorization", "Bearer "+adminTok)
	rec4 := httptest.NewRecorder()
	h.ResetPassword(rec4, req4)

	if rec4.Code != http.StatusForbidden {
		t.Fatalf("403 beklendi, gelen: %d", rec4.Code)
	}
	_ = json.Unmarshal(rec4.Body.Bytes(), &apiErr)
	if apiErr.Message != "Bu kullanıcının parolasını değiştirme yetkiniz bulunmuyor." {
		t.Fatalf("beklenmeyen admin 403 mesajı: %q", apiErr.Message)
	}

	// 5. Admin, standart kullanıcının parolasını currentPassword olmadan sıfırlarsa -> 200 başkasının mesajı
	req5 := httptest.NewRequest(http.MethodPost, "/auth/resetpass", bytes.NewBufferString(`{"email":"user@example.com","newPassword":"YeniSifre123"}`))
	req5.Header.Set("Authorization", "Bearer "+adminTok)
	rec5 := httptest.NewRecorder()
	h.ResetPassword(rec5, req5)

	if rec5.Code != http.StatusOK {
		t.Fatalf("200 beklendi, gelen: %d (%s)", rec5.Code, rec5.Body.String())
	}
	_ = json.Unmarshal(rec5.Body.Bytes(), &msg)
	if msg.Message != "Kullanıcının parolası başarıyla güncellendi." {
		t.Fatalf("admin parola sıfırlama mesajı hatalı: %q", msg.Message)
	}
}

func TestIsHTTPSCookieHelper(t *testing.T) {
	plainReq := httptest.NewRequest(http.MethodGet, "http://localhost/api/auth/login", nil)
	if auth.IsHTTPS(plainReq, "http://localhost") {
		t.Error("plain HTTP localhost için IsHTTPS false olmalı")
	}
	if !auth.IsHTTPS(plainReq, "https://makeitiny.com") {
		t.Error("https publicBaseURL için IsHTTPS true olmalı")
	}

	xfpReq := httptest.NewRequest(http.MethodGet, "http://localhost/api/auth/login", nil)
	xfpReq.Header.Set("X-Forwarded-Proto", "https")
	if !auth.IsHTTPS(xfpReq, "http://localhost") {
		t.Error("X-Forwarded-Proto: https için IsHTTPS true olmalı")
	}

	tlsReq := httptest.NewRequest(http.MethodGet, "https://localhost/api/auth/login", nil)
	tlsReq.TLS = &tls.ConnectionState{}
	if !auth.IsHTTPS(tlsReq, "http://localhost") {
		t.Error("TLS bağlantısı için IsHTTPS true olmalı")
	}
}

type testSplitMux struct {
	main *http.ServeMux
	qr   *http.ServeMux
}

func (m *testSplitMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if strings.HasSuffix(pattern, "/qr") {
		m.qr.HandleFunc(pattern, handler)
		return
	}
	m.main.HandleFunc(pattern, handler)
}

func (m *testSplitMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(segs) == 3 && segs[0] == "links" && segs[2] == "qr" && segs[1] != "check-alias" && segs[1] != "logs" {
		m.qr.ServeHTTP(w, r)
		return
	}
	m.main.ServeHTTP(w, r)
}

func TestAdminExportRoutesAndFormats(t *testing.T) {
	h, userTok, adminTok := newTestHandler(t)
	mux := gen.HandlerWithOptions(h, gen.StdHTTPServerOptions{
		BaseRouter: &testSplitMux{main: http.NewServeMux(), qr: http.NewServeMux()},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		},
	})

	// Yetkisiz veya standart kullanıcı -> 403
	reqUser := httptest.NewRequest(http.MethodGet, "/admin/links/export?format=csv", nil)
	reqUser.Header.Set("Authorization", "Bearer "+userTok)
	recUser := httptest.NewRecorder()
	mux.ServeHTTP(recUser, reqUser)
	if recUser.Code != http.StatusForbidden {
		t.Fatalf("standart kullanıcı için 403 beklendi, gelen: %d", recUser.Code)
	}

	// Geçersiz biçim -> 400
	reqBad := httptest.NewRequest(http.MethodGet, "/admin/users/export?format=xml", nil)
	reqBad.Header.Set("Authorization", "Bearer "+adminTok)
	recBad := httptest.NewRecorder()
	mux.ServeHTTP(recBad, reqBad)
	if recBad.Code != http.StatusBadRequest {
		t.Fatalf("geçersiz biçim için 400 beklendi, gelen: %d (%s)", recBad.Code, recBad.Body.String())
	}

	// /admin/links/export {id} ile çakışmadan CSV döner
	reqLinks := httptest.NewRequest(http.MethodGet, "/admin/links/export?format=csv&status=all", nil)
	reqLinks.Header.Set("Authorization", "Bearer "+adminTok)
	recLinks := httptest.NewRecorder()
	mux.ServeHTTP(recLinks, reqLinks)
	if recLinks.Code != http.StatusOK || !strings.Contains(recLinks.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("/admin/links/export CSV beklendi, gelen: %d / %s", recLinks.Code, recLinks.Body.String())
	}

	// /admin/users/export {id} ile çakışmadan JSON döner
	reqUsers := httptest.NewRequest(http.MethodGet, "/admin/users/export?format=json", nil)
	reqUsers.Header.Set("Authorization", "Bearer "+adminTok)
	recUsers := httptest.NewRecorder()
	mux.ServeHTTP(recUsers, reqUsers)
	if recUsers.Code != http.StatusOK || !strings.Contains(recUsers.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("/admin/users/export JSON beklendi, gelen: %d / %s", recUsers.Code, recUsers.Body.String())
	}

	// /admin/users/{id}/links/export var olan kullanıcıda CSV, olmayanda 404 döner
	reqUserLinks := httptest.NewRequest(http.MethodGet, "/admin/users/u1/links/export?format=csv", nil)
	reqUserLinks.Header.Set("Authorization", "Bearer "+adminTok)
	recUserLinks := httptest.NewRecorder()
	mux.ServeHTTP(recUserLinks, reqUserLinks)
	if recUserLinks.Code != http.StatusOK || !strings.Contains(recUserLinks.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("/admin/users/u1/links/export CSV beklendi, gelen: %d (%s)", recUserLinks.Code, recUserLinks.Body.String())
	}

	reqMissing := httptest.NewRequest(http.MethodGet, "/admin/users/olmayan/links/export?format=csv", nil)
	reqMissing.Header.Set("Authorization", "Bearer "+adminTok)
	recMissing := httptest.NewRecorder()
	mux.ServeHTTP(recMissing, reqMissing)
	if recMissing.Code != http.StatusNotFound {
		t.Fatalf("bulunamayan kullanıcı için 404 beklendi, gelen: %d (%s)", recMissing.Code, recMissing.Body.String())
	}
}
