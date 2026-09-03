// Package support implémente les demandes d'assistance utilisateur.
package support

import (
	"context"
	"strings"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre les tickets de support.
type Service struct {
	repo port.SupportRepository
}

// NewService construit le service de support.
func NewService(repo port.SupportRepository) *Service { return &Service{repo: repo} }

// Create ouvre une demande d'assistance au nom de l'utilisateur.
func (s *Service) Create(ctx context.Context, userID, subject, message string) (domain.SupportTicket, error) {
	subject = strings.TrimSpace(subject)
	message = strings.TrimSpace(message)
	if subject == "" || message == "" {
		return domain.SupportTicket{}, domain.ErrInvalidInput
	}
	if len(subject) > 150 {
		subject = subject[:150]
	}
	return s.repo.Create(ctx, domain.SupportTicket{UserID: userID, Subject: subject, Message: message})
}

// ListMine renvoie les demandes de l'utilisateur.
func (s *Service) ListMine(ctx context.Context, userID string) ([]domain.SupportTicket, error) {
	return s.repo.ListByUser(ctx, userID)
}
