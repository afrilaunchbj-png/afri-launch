package port

import (
	"context"

	"afrilaunch/backend/internal/domain"
)

// SupportRepository accède aux demandes d'assistance.
type SupportRepository interface {
	Create(ctx context.Context, t domain.SupportTicket) (domain.SupportTicket, error)
	ListByUser(ctx context.Context, userID string) ([]domain.SupportTicket, error)
	ListAll(ctx context.Context, f AdminListFilter, limit, offset int) ([]domain.AdminTicket, int64, error)
	Get(ctx context.Context, id string) (domain.SupportTicket, error)
	GetWithUser(ctx context.Context, id string) (domain.AdminTicket, error)
	ListMessages(ctx context.Context, ticketID string) ([]domain.TicketMessageView, error)
	AddMessage(ctx context.Context, ticketID string, msg domain.TicketMessage) (domain.TicketMessageView, error)
	SetStatus(ctx context.Context, id, status string) (domain.SupportTicket, error)
	CountOpen(ctx context.Context) (int64, error)
}

// AdminListFilter filtre les listes du suivi global (chaîne vide = pas de filtre).
type AdminListFilter struct {
	Search string // recherche texte (titre, nom, email…)
	Status string // statut ou type selon l'entité
	Role   string // users uniquement
}

// AdminRepository expose les agrégats et listes du suivi global superadmin.
type AdminRepository interface {
	Stats(ctx context.Context) (domain.AdminStats, error)
	ListUsers(ctx context.Context, f AdminListFilter, limit, offset int) ([]domain.User, int64, error)
	ListProjects(ctx context.Context, f AdminListFilter, limit, offset int) ([]domain.AdminProject, int64, error)
	ListConversations(ctx context.Context, f AdminListFilter, limit, offset int) ([]domain.AdminConversation, int64, error)
	ListAssets(ctx context.Context, f AdminListFilter, limit, offset int) ([]domain.AdminAsset, int64, error)
	ListJobs(ctx context.Context, f AdminListFilter, limit, offset int) ([]domain.AdminJob, int64, error)
	ListCreditTransactions(ctx context.Context, f AdminListFilter, limit, offset int) ([]domain.AdminCreditTransaction, int64, error)
	ListAuditLogs(ctx context.Context, f AuditFilter, limit, offset int) ([]domain.AuditLog, int64, error)
}
