// Package admin expose le suivi global réservé au superadmin : statistiques
// plateforme, utilisateurs et modération des tickets de support.
package admin

import (
	"context"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre les opérations superadmin.
type Service struct {
	repo    port.AdminRepository
	tickets port.SupportRepository
}

// NewService construit le service admin.
func NewService(repo port.AdminRepository, tickets port.SupportRepository) *Service {
	return &Service{repo: repo, tickets: tickets}
}

// Stats renvoie les indicateurs globaux de la plateforme.
func (s *Service) Stats(ctx context.Context) (domain.AdminStats, error) {
	return s.repo.Stats(ctx)
}

// ListUsers renvoie les comptes (paginés).
func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListUsers(ctx, limit, offset)
}

// ListTickets renvoie tous les tickets (paginés), avec leur auteur.
func (s *Service) ListTickets(ctx context.Context, limit, offset int) ([]domain.AdminTicket, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.tickets.ListAll(ctx, limit, offset)
}

// ResolveTicket marque un ticket comme résolu.
func (s *Service) ResolveTicket(ctx context.Context, ticketID string) (domain.SupportTicket, error) {
	if _, err := s.tickets.Get(ctx, ticketID); err != nil {
		return domain.SupportTicket{}, err
	}
	return s.tickets.SetStatus(ctx, ticketID, domain.TicketResolved)
}
