package postgres

import (
	"context"
	"encoding/json"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// adminRepo implémente port.AdminRepository (suivi global superadmin).
type adminRepo struct {
	s *Store
}

// NewAdminRepository construit le repository admin.
func NewAdminRepository(s *Store) *adminRepo { return &adminRepo{s: s} }

func (r *adminRepo) Stats(ctx context.Context) (domain.AdminStats, error) {
	stats := domain.AdminStats{JobsByStatus: make(map[string]int64)}

	var err error
	if stats.Users, err = r.s.q.CountUsers(ctx); err != nil {
		return stats, err
	}
	if stats.Projects, err = r.s.q.CountProjects(ctx); err != nil {
		return stats, err
	}
	if stats.Assets, err = r.s.q.CountAssets(ctx); err != nil {
		return stats, err
	}
	if stats.Conversations, err = r.s.q.CountConversations(ctx); err != nil {
		return stats, err
	}
	if stats.CreditsConsumed, err = r.s.q.SumCreditsConsumed(ctx); err != nil {
		return stats, err
	}
	if stats.OpenTickets, err = r.s.q.CountOpenTickets(ctx); err != nil {
		return stats, err
	}
	jobs, err := r.s.q.CountJobsByStatus(ctx)
	if err != nil {
		return stats, err
	}
	for _, j := range jobs {
		stats.JobsByStatus[j.Status] = j.Total
	}
	return stats, nil
}

func (r *adminRepo) ListUsers(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.User, int64, error) {
	params := db.ListUsersFilteredParams{
		Search: f.Search, Role: f.Role, RowLimit: int32(limit), RowOffset: int32(offset),
	}
	rows, err := r.s.q.ListUsersFiltered(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountUsersFiltered(ctx, db.CountUsersFilteredParams{Search: f.Search, Role: f.Role})
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		out = append(out, toUser(row))
	}
	return out, total, nil
}

func (r *adminRepo) ListProjects(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminProject, int64, error) {
	params := db.ListAllProjectsParams{
		Status: f.Status, Search: f.Search, RowLimit: int32(limit), RowOffset: int32(offset),
	}
	rows, err := r.s.q.ListAllProjects(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountAllProjects(ctx, db.CountAllProjectsParams{Status: f.Status, Search: f.Search})
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.AdminProject, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.AdminProject{
			ID: row.ID, UserID: row.UserID, Title: row.Title, Status: row.Status,
			CreditsConsumed: int64(row.CreditsConsumed), CreatedAt: row.CreatedAt,
			UserEmail: row.UserEmail, UserName: row.UserName,
		})
	}
	return out, total, nil
}

func (r *adminRepo) ListConversations(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminConversation, int64, error) {
	params := db.ListAllConversationsParams{Search: f.Search, RowLimit: int32(limit), RowOffset: int32(offset)}
	rows, err := r.s.q.ListAllConversations(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountAllConversations(ctx, f.Search)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.AdminConversation, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.AdminConversation{
			ID: row.ID, UserID: row.UserID, Title: row.Title, Status: row.Status,
			CreatedAt: row.CreatedAt, UserEmail: row.UserEmail, UserName: row.UserName,
		})
	}
	return out, total, nil
}

func (r *adminRepo) ListAssets(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminAsset, int64, error) {
	params := db.ListAllAssetsParams{Search: f.Search, RowLimit: int32(limit), RowOffset: int32(offset)}
	rows, err := r.s.q.ListAllAssets(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountAllAssets(ctx, f.Search)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.AdminAsset, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.AdminAsset{
			ID: row.ID, ProjectID: row.ProjectID, Kind: row.Kind, Filename: row.Filename,
			ContentType: row.ContentType, SizeBytes: int64(row.SizeBytes), CreatedAt: row.CreatedAt,
			ProjectTitle: row.ProjectTitle, UserEmail: row.UserEmail,
		})
	}
	return out, total, nil
}

func (r *adminRepo) ListJobs(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminJob, int64, error) {
	params := db.ListAllJobsParams{
		Status: f.Status, Search: f.Search, RowLimit: int32(limit), RowOffset: int32(offset),
	}
	rows, err := r.s.q.ListAllJobs(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountAllJobs(ctx, db.CountAllJobsParams{Status: f.Status, Search: f.Search})
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.AdminJob, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.AdminJob{
			ID: row.ID, UserID: row.UserID, Kind: row.Kind, Status: row.Status,
			Cost: int64(row.Cost), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			UserEmail: row.UserEmail, UserName: row.UserName,
		})
	}
	return out, total, nil
}

func (r *adminRepo) ListCreditTransactions(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminCreditTransaction, int64, error) {
	params := db.ListAllCreditTransactionsParams{
		Type: f.Status, Search: f.Search, RowLimit: int32(limit), RowOffset: int32(offset),
	}
	rows, err := r.s.q.ListAllCreditTransactions(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountAllCreditTransactions(ctx, db.CountAllCreditTransactionsParams{Type: f.Status, Search: f.Search})
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.AdminCreditTransaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.AdminCreditTransaction{
			ID: row.ID, Type: row.Type, Amount: int64(row.Amount), Operation: row.Operation,
			Status: row.Status, CreatedAt: row.CreatedAt, UserEmail: row.UserEmail,
		})
	}
	return out, total, nil
}

func (r *adminRepo) ListAuditLogs(ctx context.Context, f port.AuditFilter, limit, offset int) ([]domain.AuditLog, int64, error) {
	params := db.ListAuditLogsParams{
		UserID:    optionalUUID(f.UserID),
		Action:    f.Action,
		Entity:    f.Entity,
		RowLimit:  int32(limit),
		RowOffset: int32(offset),
	}
	rows, err := r.s.q.ListAuditLogs(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountAuditLogs(ctx, db.CountAuditLogsParams{UserID: optionalUUID(f.UserID), Action: f.Action, Entity: f.Entity})
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.AuditLog, 0, len(rows))
	for _, row := range rows {
		metadata := map[string]any{}
		if len(row.Metadata) > 0 {
			_ = json.Unmarshal(row.Metadata, &metadata)
		}
		entityID := ""
		if row.EntityID != nil {
			entityID = *row.EntityID
		}
		out = append(out, domain.AuditLog{
			ID: row.ID, UserID: uuidString(row.UserID), Action: row.Action,
			Entity: row.Entity, EntityID: entityID, Metadata: metadata, CreatedAt: row.CreatedAt,
		})
	}
	return out, total, nil
}
