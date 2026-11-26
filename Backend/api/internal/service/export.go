package service

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"makeitiny/api/internal/db"
	"makeitiny/api/internal/repository"
)

// MaxExportLogs bir dışa aktarmada yazılacak en fazla ziyaret kaydı.
const MaxExportLogs = 100000

// MaxExportAdminRows yönetici dışa aktarmalarında üst güvenlik sınırı.
const MaxExportAdminRows = 100000

const exportPageSize = 1000

// analyticsSince kullanıcının paketine göre okunabilecek en eski ziyaret zamanını döner.
// Yönetici ve sahip tüm geçmişi görür.
func analyticsSince(profile UserProfile) pgtype.Timestamptz {
	if !IsPrivileged(profile.Roles) && profile.Plan.AnalyticsDays > 0 {
		return pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, -int(profile.Plan.AnalyticsDays)), Valid: true}
	}
	return pgtype.Timestamptz{Time: time.Unix(0, 0), Valid: true}
}

// ExportLinks kullanıcının bütün linklerini, henüz yazılmamış tıklamalar dahil döner.
func (s *LinkService) ExportLinks(ctx context.Context, userID int64) ([]db.Link, error) {
	links, err := s.links.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.overlayPendingClicks(ctx, links), nil
}

// ExportLinkLogs linkin paket süresi içindeki ziyaret kayıtlarını (en yeniden eskiye) döner.
func (s *LinkService) ExportLinkLogs(ctx context.Context, profile UserProfile, idStr string) (db.Link, []db.Log, error) {
	link, err := s.GetLink(ctx, idStr, profile.User.ID)
	if err != nil {
		return db.Link{}, nil, err
	}

	since := analyticsSince(profile)
	logs := make([]db.Log, 0, exportPageSize)
	for offset := 0; offset < MaxExportLogs; offset += exportPageSize {
		page, err := s.links.ListLogsByLinkPaginated(ctx, link.ID, since, exportPageSize, int32(offset))
		if err != nil {
			return db.Link{}, nil, err
		}
		logs = append(logs, page...)
		if len(page) < exportPageSize {
			break
		}
	}
	return link, logs, nil
}

// AdminExportLinks seçili arama ve durum filtresine uyan tüm linkleri bekleyen tıklamalarla birlikte döner.
func (s *LinkService) AdminExportLinks(ctx context.Context, search, status *string) ([]db.ListAdminLinksPaginatedRow, error) {
	q := ""
	if search != nil {
		q = strings.TrimSpace(*search)
	}
	st, err := normalizeAdminLinkStatus(status)
	if err != nil {
		return nil, err
	}

	rows := make([]db.ListAdminLinksPaginatedRow, 0, exportPageSize)
	for offset := 0; offset < MaxExportAdminRows; offset += exportPageSize {
		page, err := s.links.ListAdminLinksPaginated(ctx, q, st, exportPageSize, int32(offset))
		if err != nil {
			return nil, err
		}
		rows = append(rows, page...)
		if len(page) < exportPageSize {
			break
		}
	}

	if len(rows) > 0 {
		refs := make([]repository.LinkRef, len(rows))
		for i, row := range rows {
			refs[i] = repository.LinkRef{ID: row.ID, Alias: row.Alias}
		}
		deltas := s.links.PendingClicks(ctx, refs)
		for i := range rows {
			rows[i].Clicks += deltas[rows[i].ID]
		}
	}
	return rows, nil
}

// AdminExportUsers arama filtresine uyan tüm kullanıcı profillerini döner.
func (s *AuthService) AdminExportUsers(ctx context.Context, search *string) ([]UserProfile, error) {
	q := ""
	if search != nil {
		q = strings.TrimSpace(*search)
	}

	users := make([]db.User, 0, exportPageSize)
	for offset := 0; offset < MaxExportAdminRows; offset += exportPageSize {
		page, err := s.users.ListAdminUsersPaginated(ctx, q, exportPageSize, int32(offset))
		if err != nil {
			return nil, err
		}
		users = append(users, page...)
		if len(page) < exportPageSize {
			break
		}
	}
	return s.buildAdminUserProfiles(ctx, users), nil
}
