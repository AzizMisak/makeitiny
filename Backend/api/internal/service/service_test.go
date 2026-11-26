package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"makeitiny/api/internal/auth"
	"makeitiny/api/internal/db"
	"makeitiny/api/internal/repository"
)

type stubRow struct {
	scanFn func(dest ...any) error
}

func (r stubRow) Scan(dest ...any) error {
	return r.scanFn(dest...)
}

type stubRows struct {
	pgx.Rows
	items []string
	idx   int
	err   error
}

func (r *stubRows) Close()                        {}
func (r *stubRows) Err() error                    { return r.err }
func (r *stubRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }
func (r *stubRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}
func (r *stubRows) Next() bool {
	if r.idx < len(r.items) {
		r.idx++
		return true
	}
	return false
}
func (r *stubRows) Scan(dest ...any) error {
	if p, ok := dest[0].(*string); ok {
		*p = r.items[r.idx-1]
	}
	return nil
}
func (r *stubRows) Values() ([]any, error) { return nil, nil }
func (r *stubRows) RawValues() [][]byte    { return nil }
func (r *stubRows) Conn() *pgx.Conn        { return nil }

type stubDBTX struct {
	usersByEmail map[string]db.User
	rolesByUser  map[int64][]string
	linksByID    map[string]db.Link
	lastAdminUpd *db.AdminUpdateLinkParams
	updatedPass  map[string]string
}

func (s *stubDBTX) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "UPDATE users") && strings.Contains(sql, "password_hash") {
		email := args[0].(string)
		hash := args[1].(string)
		if _, ok := s.usersByEmail[email]; !ok {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		}
		if s.updatedPass == nil {
			s.updatedPass = map[string]string{}
		}
		s.updatedPass[email] = hash
		return pgconn.NewCommandTag("UPDATE 1"), nil
	}
	return pgconn.NewCommandTag("DELETE 0"), nil
}

func (s *stubDBTX) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM roles") {
		uid := args[0].(int64)
		return &stubRows{items: s.rolesByUser[uid]}, nil
	}
	return &stubRows{}, nil
}

func (s *stubDBTX) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	switch {
	case strings.Contains(sql, "FROM links") && strings.Contains(sql, "WHERE public_id = $1"):
		pid := args[0].(string)
		return stubRow{scanFn: func(dest ...any) error {
			l, ok := s.linksByID[pid]
			if !ok {
				return pgx.ErrNoRows
			}
			assignLinkDest(dest, l)
			return nil
		}}
	case strings.Contains(sql, "UPDATE links"):
		return stubRow{scanFn: func(dest ...any) error {
			p := db.AdminUpdateLinkParams{
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
			}
			s.lastAdminUpd = &p
			out := db.Link{
				ID:                p.ID,
				Alias:             p.Alias,
				TargetUrl:         p.TargetUrl,
				PasswordHash:      p.PasswordHash,
				WaitEnabled:       p.WaitEnabled,
				WaitSeconds:       p.WaitSeconds,
				Title:             p.Title,
				IsActive:          p.IsActive,
				ExpiresAt:         p.ExpiresAt,
				MaxClicks:         p.MaxClicks,
				AdEnabled:         p.AdEnabled,
				SecurityStatus:    p.SecurityStatus,
				SecurityCheckedAt: p.SecurityCheckedAt,
				PublicID:          "lnk123",
			}
			assignLinkDest(dest, out)
			return nil
		}}
	case strings.Contains(sql, "FROM users") && strings.Contains(sql, "WHERE email = $1"):
		email := args[0].(string)
		return stubRow{scanFn: func(dest ...any) error {
			u, ok := s.usersByEmail[email]
			if !ok {
				return pgx.ErrNoRows
			}
			assignUserDest(dest, u)
			return nil
		}}
	}
	return stubRow{scanFn: func(dest ...any) error { return pgx.ErrNoRows }}
}

func assignLinkDest(dest []any, l db.Link) {
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

func assignUserDest(dest []any, u db.User) {
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

func TestAdminUpdateLinkUnblockAndReactivate(t *testing.T) {
	str := func(v string) *string { return &v }
	b := func(v bool) *bool { return &v }
	i32 := func(v int32) *int32 { return &v }

	blockedLink := db.Link{
		ID:             10,
		PublicID:       "lnk123",
		Alias:          "ornek1",
		TargetUrl:      "https://example.com",
		IsActive:       false,
		SecurityStatus: "blocked",
	}
	safeLink := db.Link{
		ID:             11,
		PublicID:       "lnkSafe",
		Alias:          "safe11",
		TargetUrl:      "https://example.com",
		IsActive:       true,
		SecurityStatus: "safe",
	}

	stub := &stubDBTX{
		linksByID: map[string]db.Link{
			"lnk123":  blockedLink,
			"lnkSafe": safeLink,
		},
	}
	repo := repository.NewLinkRepository(nil, db.New(stub), nil, nil)
	svc := NewLinkService(repo, nil, "https://makeitiny.com")

	// Engeli kaldırıp aynı istekte yayına alma başarılı olmalı
	updated, err := svc.AdminUpdateLink(context.Background(), "lnk123", UpdateLinkInput{
		SecurityStatus: str("safe"),
		IsActive:       b(true),
	})
	if err != nil {
		t.Fatalf("engeli kaldır + aktif et başarısız olmamalıydı: %v", err)
	}
	if updated.SecurityStatus != "safe" || !updated.IsActive {
		t.Fatalf("beklenen safe + active=true, gelen status=%q active=%v", updated.SecurityStatus, updated.IsActive)
	}

	// Engelli linki engel kalkmadan aktif etmeye çalışmak ErrForbidden dönmeli
	if _, err := svc.AdminUpdateLink(context.Background(), "lnk123", UpdateLinkInput{
		IsActive: b(true),
	}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("engelli link aktif edilirken ErrForbidden beklendi, gelen: %v", err)
	}

	// Linki blocked yaparken aynı anda isActive=true göndermek ErrForbidden dönmeli
	if _, err := svc.AdminUpdateLink(context.Background(), "lnkSafe", UpdateLinkInput{
		SecurityStatus: str("blocked"),
		IsActive:       b(true),
	}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("blocked + active=true için ErrForbidden beklendi, gelen: %v", err)
	}

	// Linki blocked yaparken isActive gönderilmezse otomatik false olmalı
	blockedOut, err := svc.AdminUpdateLink(context.Background(), "lnkSafe", UpdateLinkInput{
		SecurityStatus: str("blocked"),
	})
	if err != nil || blockedOut.IsActive || blockedOut.SecurityStatus != "blocked" {
		t.Fatalf("blocked sonrası active=false olmalıydı, gelen err=%v active=%v status=%q", err, blockedOut.IsActive, blockedOut.SecurityStatus)
	}

	// Boş target veya alias reddedilmeli
	if _, err := svc.AdminUpdateLink(context.Background(), "lnkSafe", UpdateLinkInput{Target: str("   ")}); !errors.Is(err, ErrInvalidTargetURL) {
		t.Fatalf("boş target için ErrInvalidTargetURL beklendi, gelen: %v", err)
	}
	if _, err := svc.AdminUpdateLink(context.Background(), "lnkSafe", UpdateLinkInput{Alias: str("")}); !errors.Is(err, ErrInvalidAlias) {
		t.Fatalf("boş alias için ErrInvalidAlias beklendi, gelen: %v", err)
	}

	// waitSeconds üst sınırı (300 sn) aşılırsa ErrInvalidWaitSeconds dönmeli
	if _, err := svc.AdminUpdateLink(context.Background(), "lnkSafe", UpdateLinkInput{
		WaitEnabled: b(true),
		WaitSeconds: i32(301),
	}); !errors.Is(err, ErrInvalidWaitSeconds) {
		t.Fatalf("waitSeconds=301 için ErrInvalidWaitSeconds beklendi, gelen: %v", err)
	}
}

func TestResetPasswordCurrentPassword(t *testing.T) {
	hash, err := auth.HashPassword("EskiSifre123")
	if err != nil {
		t.Fatalf("hash hatası: %v", err)
	}

	userA := db.User{
		ID:           1,
		PublicID:     "usr1",
		Email:        "user@example.com",
		PasswordHash: hash,
		IsActive:     true,
		CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	adminUser := db.User{
		ID:           2,
		PublicID:     "adm2",
		Email:        "admin@example.com",
		PasswordHash: hash,
		IsActive:     true,
	}

	stub := &stubDBTX{
		usersByEmail: map[string]db.User{
			userA.Email:     userA,
			adminUser.Email: adminUser,
		},
		rolesByUser: map[int64][]string{
			userA.ID:     {"user"},
			adminUser.ID: {"user", "admin"},
		},
	}
	userRepo := repository.NewUserRepository(nil, db.New(stub), nil, nil)
	authSvc := NewAuthService(userRepo, "test-secret")

	actorSelf := UserProfile{User: userA, Roles: []string{"user"}}
	actorAdmin := UserProfile{User: adminUser, Roles: []string{"user", "admin"}}

	// Mevcut parola boşsa reddedilmeli
	if err := authSvc.ResetPassword(context.Background(), actorSelf, userA.Email, "", "YeniSifre123"); !errors.Is(err, ErrInvalidCurrentPassword) {
		t.Fatalf("boş currentPassword için ErrInvalidCurrentPassword beklendi, gelen: %v", err)
	}
	// Mevcut parola yanlışsa reddedilmeli
	if err := authSvc.ResetPassword(context.Background(), actorSelf, userA.Email, "YanlisSifre", "YeniSifre123"); !errors.Is(err, ErrInvalidCurrentPassword) {
		t.Fatalf("yanlış currentPassword için ErrInvalidCurrentPassword beklendi, gelen: %v", err)
	}
	// 2 rune (4 byte UTF-8) parola rune sayımıyla reddedilmeli
	if err := authSvc.ResetPassword(context.Background(), actorSelf, userA.Email, "EskiSifre123", "şç"); !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("2 rune parola için ErrPasswordTooShort beklendi, gelen: %v", err)
	}
	// Doğru mevcut parola ile güncelleme başarılı olmalı
	if err := authSvc.ResetPassword(context.Background(), actorSelf, userA.Email, "EskiSifre123", "YeniSifre123"); err != nil {
		t.Fatalf("doğru currentPassword ile parola güncellenmeliydi: %v", err)
	}
	if stub.updatedPass[userA.Email] == "" {
		t.Fatal("güncellenmiş parola hash'i yazılmadı")
	}

	// Admin başka kullanıcının parolasını currentPassword olmadan sıfırlayabilmeli
	if err := authSvc.ResetPassword(context.Background(), actorAdmin, userA.Email, "", "AdminSifre99"); err != nil {
		t.Fatalf("admin başkasının parolasını currentPassword olmadan sıfırlayabilmeliydi: %v", err)
	}
}

func TestAdminListLinksInvalidStatus(t *testing.T) {
	repo := repository.NewLinkRepository(nil, db.New(&stubDBTX{}), nil, nil)
	svc := NewLinkService(repo, nil, "https://makeitiny.com")

	bad := "bogus"
	if _, err := svc.AdminListLinks(context.Background(), nil, nil, nil, &bad); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("geçersiz status için ErrInvalidInput beklendi, gelen: %v", err)
	}
}
