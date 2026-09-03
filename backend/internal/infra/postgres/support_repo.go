package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// supportRepo implémente port.SupportRepository.
type supportRepo struct {
	s *Store
}

// NewSupportRepository construit le repository de support.
func NewSupportRepository(s *Store) *supportRepo { return &supportRepo{s: s} }

func (r *supportRepo) Create(ctx context.Context, t domain.SupportTicket) (domain.SupportTicket, error) {
	row, err := r.s.q.CreateTicket(ctx, db.CreateTicketParams{UserID: t.UserID, Subject: t.Subject, Message: t.Message})
	if err != nil {
		return domain.SupportTicket{}, err
	}
	return toTicket(row), nil
}

func (r *supportRepo) ListByUser(ctx context.Context, userID string) ([]domain.SupportTicket, error) {
	rows, err := r.s.q.ListTicketsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.SupportTicket, 0, len(rows))
	for _, row := range rows {
		out = append(out, toTicket(row))
	}
	return out, nil
}

func (r *supportRepo) ListAll(ctx context.Context, limit, offset int) ([]domain.AdminTicket, int64, error) {
	rows, err := r.s.q.ListAllTickets(ctx, db.ListAllTicketsParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountAllTickets(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.AdminTicket, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.AdminTicket{
			SupportTicket: domain.SupportTicket{
				ID: row.ID, UserID: row.UserID, Subject: row.Subject, Message: row.Message,
				Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			},
			UserEmail: row.UserEmail,
			UserName:  row.UserName,
		})
	}
	return out, total, nil
}

func (r *supportRepo) Get(ctx context.Context, id string) (domain.SupportTicket, error) {
	row, err := r.s.q.GetTicket(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return domain.SupportTicket{}, domain.ErrNotFound
		}
		return domain.SupportTicket{}, err
	}
	return toTicket(row), nil
}

func (r *supportRepo) SetStatus(ctx context.Context, id, status string) (domain.SupportTicket, error) {
	row, err := r.s.q.SetTicketStatus(ctx, db.SetTicketStatusParams{ID: id, Status: status})
	if err != nil {
		if isNoRows(err) {
			return domain.SupportTicket{}, domain.ErrNotFound
		}
		return domain.SupportTicket{}, err
	}
	return toTicket(row), nil
}

func (r *supportRepo) CountOpen(ctx context.Context) (int64, error) {
	return r.s.q.CountOpenTickets(ctx)
}

func toTicket(t db.SupportTicket) domain.SupportTicket {
	return domain.SupportTicket{
		ID: t.ID, UserID: t.UserID, Subject: t.Subject, Message: t.Message,
		Status: t.Status, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}
