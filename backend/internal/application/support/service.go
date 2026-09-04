// Package support implémente les demandes d'assistance utilisateur :
// création de tickets, fil de discussion (réponses utilisateur et support).
package support

import (
	"context"
	"strings"

	"afrilaunch/backend/internal/application/audit"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre les tickets de support.
type Service struct {
	repo  port.SupportRepository
	audit *audit.Recorder
}

// NewService construit le service de support.
func NewService(repo port.SupportRepository, auditRec *audit.Recorder) *Service {
	return &Service{repo: repo, audit: auditRec}
}

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
	ticket, err := s.repo.Create(ctx, domain.SupportTicket{UserID: userID, Subject: subject, Message: message})
	if err == nil {
		s.audit.Log(ctx, userID, domain.AuditTicketCreate, "support_ticket", ticket.ID, nil)
	}
	return ticket, err
}

// ListMine renvoie les demandes de l'utilisateur.
func (s *Service) ListMine(ctx context.Context, userID string) ([]domain.SupportTicket, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Get renvoie un ticket si l'utilisateur en est l'auteur.
func (s *Service) Get(ctx context.Context, userID, ticketID string) (domain.SupportTicket, error) {
	ticket, err := s.repo.Get(ctx, ticketID)
	if err != nil {
		return domain.SupportTicket{}, err
	}
	if ticket.UserID != userID {
		return domain.SupportTicket{}, domain.ErrForbidden
	}
	return ticket, nil
}

// Detail renvoie un ticket et son fil de discussion (accès propriétaire).
func (s *Service) Detail(ctx context.Context, userID, ticketID string) (domain.SupportTicket, []domain.TicketMessageView, error) {
	ticket, err := s.Get(ctx, userID, ticketID)
	if err != nil {
		return domain.SupportTicket{}, nil, err
	}
	messages, err := s.repo.ListMessages(ctx, ticketID)
	if err != nil {
		return domain.SupportTicket{}, nil, err
	}
	return ticket, messages, nil
}

// Reply ajoute la réponse de l'utilisateur au fil ; un ticket résolu
// est rouvert automatiquement (le support doit répondre à nouveau).
func (s *Service) Reply(ctx context.Context, userID, ticketID, content string) (domain.SupportTicket, domain.TicketMessageView, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return domain.SupportTicket{}, domain.TicketMessageView{}, domain.ErrInvalidInput
	}
	ticket, err := s.Get(ctx, userID, ticketID)
	if err != nil {
		return domain.SupportTicket{}, domain.TicketMessageView{}, err
	}
	msg, err := s.repo.AddMessage(ctx, ticketID, domain.TicketMessage{AuthorID: userID, Content: content})
	if err != nil {
		return domain.SupportTicket{}, domain.TicketMessageView{}, err
	}
	if ticket.Status == domain.TicketResolved {
		if ticket, err = s.repo.SetStatus(ctx, ticketID, domain.TicketOpen); err != nil {
			return domain.SupportTicket{}, domain.TicketMessageView{}, err
		}
	}
	s.audit.Log(ctx, userID, domain.AuditTicketReply, "support_ticket", ticketID, nil)
	return ticket, msg, nil
}
