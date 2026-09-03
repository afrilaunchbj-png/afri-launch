package postgres

import (
	"context"

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

func (r *adminRepo) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int64, error) {
	rows, err := r.s.q.ListUsers(ctx, db.ListUsersParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		out = append(out, toUser(row))
	}
	return out, total, nil
}
