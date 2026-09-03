package port

import (
	"context"

	"afrilaunch/backend/internal/domain"
)

// SupportRepository accède aux demandes d'assistance.
type SupportRepository interface {
	Create(ctx context.Context, t domain.SupportTicket) (domain.SupportTicket, error)
	ListByUser(ctx context.Context, userID string) ([]domain.SupportTicket, error)
	ListAll(ctx context.Context, limit, offset int) ([]domain.AdminTicket, int64, error)
	Get(ctx context.Context, id string) (domain.SupportTicket, error)
	SetStatus(ctx context.Context, id, status string) (domain.SupportTicket, error)
	CountOpen(ctx context.Context) (int64, error)
}

// AdminRepository expose les agrégats du suivi global superadmin.
type AdminRepository interface {
	Stats(ctx context.Context) (domain.AdminStats, error)
	ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int64, error)
}
