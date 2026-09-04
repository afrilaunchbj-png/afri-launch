// Package admin expose le suivi global réservé au superadmin : statistiques
// plateforme, utilisateurs, modération des tickets de support et journal
// d'activités.
package admin

import (
	"context"

	"afrilaunch/backend/internal/application/audit"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// pageSize par défaut et borne haute pour les listes paginées.
const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Service orchestre les opérations superadmin.
type Service struct {
	repo    port.AdminRepository
	tickets port.SupportRepository
	audit   *audit.Recorder
}

// NewService construit le service admin.
func NewService(repo port.AdminRepository, tickets port.SupportRepository, auditRec *audit.Recorder) *Service {
	return &Service{repo: repo, tickets: tickets, audit: auditRec}
}

func normalize(limit int) int {
	if limit <= 0 || limit > maxPageSize {
		return defaultPageSize
	}
	return limit
}

// Stats renvoie les indicateurs globaux de la plateforme.
func (s *Service) Stats(ctx context.Context) (domain.AdminStats, error) {
	return s.repo.Stats(ctx)
}

// ListUsers renvoie les comptes filtrés (paginés).
func (s *Service) ListUsers(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.User, int64, error) {
	return s.repo.ListUsers(ctx, f, normalize(limit), offset)
}

// ListTickets renvoie tous les tickets filtrés (paginés), avec leur auteur.
func (s *Service) ListTickets(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminTicket, int64, error) {
	return s.tickets.ListAll(ctx, f, normalize(limit), offset)
}

// ListProjects renvoie les projets filtrés (paginés).
func (s *Service) ListProjects(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminProject, int64, error) {
	return s.repo.ListProjects(ctx, f, normalize(limit), offset)
}

// ListConversations renvoie les conversations filtrées (paginées).
func (s *Service) ListConversations(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminConversation, int64, error) {
	return s.repo.ListConversations(ctx, f, normalize(limit), offset)
}

// ListAssets renvoie les assets filtrés (paginés).
func (s *Service) ListAssets(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminAsset, int64, error) {
	return s.repo.ListAssets(ctx, f, normalize(limit), offset)
}

// ListJobs renvoie les jobs de génération filtrés (paginés).
func (s *Service) ListJobs(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminJob, int64, error) {
	return s.repo.ListJobs(ctx, f, normalize(limit), offset)
}

// ListCreditTransactions renvoie les transactions de crédits filtrées (paginées).
func (s *Service) ListCreditTransactions(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminCreditTransaction, int64, error) {
	return s.repo.ListCreditTransactions(ctx, f, normalize(limit), offset)
}

// TicketDetail renvoie un ticket (avec son auteur) et son fil de discussion.
func (s *Service) TicketDetail(ctx context.Context, ticketID string) (domain.AdminTicket, []domain.TicketMessageView, error) {
	ticket, err := s.tickets.GetWithUser(ctx, ticketID)
	if err != nil {
		return domain.AdminTicket{}, nil, err
	}
	messages, err := s.tickets.ListMessages(ctx, ticketID)
	if err != nil {
		return domain.AdminTicket{}, nil, err
	}
	return ticket, messages, nil
}

// ReplyTicket ajoute la réponse du support au fil d'un ticket.
func (s *Service) ReplyTicket(ctx context.Context, adminID, ticketID, content string) (domain.TicketMessageView, error) {
	if content == "" {
		return domain.TicketMessageView{}, domain.ErrInvalidInput
	}
	if _, err := s.tickets.Get(ctx, ticketID); err != nil {
		return domain.TicketMessageView{}, err
	}
	msg, err := s.tickets.AddMessage(ctx, ticketID, domain.TicketMessage{AuthorID: adminID, Content: content, IsAdmin: true})
	if err != nil {
		return domain.TicketMessageView{}, err
	}
	s.audit.Log(ctx, adminID, domain.AuditTicketAdminReply, "support_ticket", ticketID, nil)
	return msg, nil
}

// ResolveTicket marque un ticket comme résolu.
func (s *Service) ResolveTicket(ctx context.Context, adminID, ticketID string) (domain.SupportTicket, error) {
	if _, err := s.tickets.Get(ctx, ticketID); err != nil {
		return domain.SupportTicket{}, err
	}
	ticket, err := s.tickets.SetStatus(ctx, ticketID, domain.TicketResolved)
	if err == nil {
		s.audit.Log(ctx, adminID, domain.AuditTicketResolve, "support_ticket", ticketID, nil)
	}
	return ticket, err
}

// ListAuditLogs renvoie le journal d'activités (paginé, filtrable).
func (s *Service) ListAuditLogs(ctx context.Context, f port.AuditFilter, limit, offset int) ([]domain.AuditLog, int64, error) {
	return s.repo.ListAuditLogs(ctx, f, normalize(limit), offset)
}
