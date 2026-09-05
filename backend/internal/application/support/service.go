// Package support implémente les demandes d'assistance utilisateur :
// création de tickets, fil de discussion (réponses utilisateur et support),
// pièces jointes (captures d'écran, PDF).
package support

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"afrilaunch/backend/internal/application/audit"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre les tickets de support.
type Service struct {
	repo    port.SupportRepository
	audit   *audit.Recorder
	storage port.Storage
}

// NewService construit le service de support.
func NewService(repo port.SupportRepository, auditRec *audit.Recorder, storage port.Storage) *Service {
	return &Service{repo: repo, audit: auditRec, storage: storage}
}

// Create ouvre une demande d'assistance au nom de l'utilisateur. Les IDs de
// pièces jointes (préalablement uploadées) sont rattachés au ticket.
func (s *Service) Create(ctx context.Context, userID, subject, message string, attachmentIDs []string) (domain.SupportTicket, error) {
	subject = strings.TrimSpace(subject)
	message = strings.TrimSpace(message)
	if subject == "" || message == "" {
		return domain.SupportTicket{}, domain.ErrInvalidInput
	}
	if len(subject) > 150 {
		subject = subject[:150]
	}
	if len(attachmentIDs) > domain.AttachmentMaxPerSubmit {
		return domain.SupportTicket{}, domain.ErrInvalidInput
	}
	ticket, err := s.repo.Create(ctx, domain.SupportTicket{UserID: userID, Subject: subject, Message: message})
	if err == nil {
		if err := s.repo.BindAttachments(ctx, userID, attachmentIDs, ticket.ID, ""); err != nil {
			return domain.SupportTicket{}, err
		}
		s.audit.Log(ctx, userID, domain.AuditTicketCreate, "support_ticket", ticket.ID,
			map[string]any{"attachments": len(attachmentIDs)})
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
func (s *Service) Reply(ctx context.Context, userID, ticketID, content string, attachmentIDs []string) (domain.SupportTicket, domain.TicketMessageView, error) {
	content = strings.TrimSpace(content)
	if content == "" && len(attachmentIDs) == 0 {
		return domain.SupportTicket{}, domain.TicketMessageView{}, domain.ErrInvalidInput
	}
	if content == "" {
		content = "(pièces jointes)"
	}
	if len(attachmentIDs) > domain.AttachmentMaxPerSubmit {
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
	if len(attachmentIDs) > 0 {
		if err := s.repo.BindAttachments(ctx, userID, attachmentIDs, "", msg.ID); err != nil {
			return domain.SupportTicket{}, domain.TicketMessageView{}, err
		}
	}
	if ticket.Status == domain.TicketResolved {
		if ticket, err = s.repo.SetStatus(ctx, ticketID, domain.TicketOpen); err != nil {
			return domain.SupportTicket{}, domain.TicketMessageView{}, err
		}
	}
	s.audit.Log(ctx, userID, domain.AuditTicketReply, "support_ticket", ticketID,
		map[string]any{"attachments": len(attachmentIDs)})
	return ticket, msg, nil
}

// UploadAttachment enregistre un fichier joint (capture d'écran, PDF) avant
// sa liaison au ticket/message. La pièce jointe est possédée par l'utilisateur.
func (s *Service) UploadAttachment(ctx context.Context, userID, filename, contentType string, data []byte) (domain.SupportAttachment, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" || len(data) == 0 {
		return domain.SupportAttachment{}, domain.ErrInvalidInput
	}
	if !domain.AttachmentAllowedContentType(contentType) {
		return domain.SupportAttachment{}, domain.ErrInvalidInput
	}
	if len(data) > domain.AttachmentMaxSize {
		return domain.SupportAttachment{}, domain.ErrInvalidInput
	}
	attachment, err := s.repo.CreateAttachment(ctx, domain.SupportAttachment{
		UserID:      userID,
		Filename:    filename,
		StorageKey:  attachmentKey(userID, filename),
		ContentType: contentType,
		SizeBytes:   int64(len(data)),
	})
	if err != nil {
		return domain.SupportAttachment{}, err
	}
	if err := s.storage.Put(ctx, attachment.StorageKey, data, contentType); err != nil {
		return domain.SupportAttachment{}, err
	}
	return attachment, nil
}

// DetailAttachments renvoie les pièces jointes d'un ticket (message initial)
// et celles indexées par message, pour affichage dans le fil.
func (s *Service) DetailAttachments(ctx context.Context, userID, ticketID string) (ticketFiles []domain.SupportAttachment, messageFiles map[string][]domain.SupportAttachment, err error) {
	ticket, err := s.Get(ctx, userID, ticketID)
	if err != nil {
		return nil, nil, err
	}
	ticketFiles, err = s.repo.ListAttachmentsByTicket(ctx, ticket.ID)
	if err != nil {
		return nil, nil, err
	}
	messages, err := s.repo.ListMessages(ctx, ticket.ID)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, 0, len(messages))
	for _, m := range messages {
		ids = append(ids, m.ID)
	}
	all, err := s.repo.ListAttachmentsByMessages(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	messageFiles = make(map[string][]domain.SupportAttachment)
	for _, a := range all {
		messageFiles[a.MessageID] = append(messageFiles[a.MessageID], a)
	}
	return ticketFiles, messageFiles, nil
}

// AdminAttachments renvoie les pièces jointes d'un ticket (côté support,
// sans vérification d'appartenance — route superadmin uniquement).
func (s *Service) AdminAttachments(ctx context.Context, ticketID string) (ticketFiles []domain.SupportAttachment, messageFiles map[string][]domain.SupportAttachment, err error) {
	ticketFiles, err = s.repo.ListAttachmentsByTicket(ctx, ticketID)
	if err != nil {
		return nil, nil, err
	}
	messages, err := s.repo.ListMessages(ctx, ticketID)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, 0, len(messages))
	for _, m := range messages {
		ids = append(ids, m.ID)
	}
	all, err := s.repo.ListAttachmentsByMessages(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	messageFiles = make(map[string][]domain.SupportAttachment)
	for _, a := range all {
		messageFiles[a.MessageID] = append(messageFiles[a.MessageID], a)
	}
	return ticketFiles, messageFiles, nil
}

// DownloadAttachment renvoie le contenu d'une pièce jointe si l'utilisateur
// en est le propriétaire.
func (s *Service) DownloadAttachment(ctx context.Context, userID, attachmentID string) (domain.SupportAttachment, []byte, error) {
	attachment, err := s.repo.GetAttachment(ctx, userID, attachmentID)
	if err != nil {
		return domain.SupportAttachment{}, nil, err
	}
	data, err := s.storage.Get(ctx, attachment.StorageKey)
	if err != nil {
		return domain.SupportAttachment{}, nil, err
	}
	return attachment, data, nil
}

// DownloadAttachmentForAdmin télécharge une pièce jointe sans contrainte
// d'appartenance (route superadmin uniquement).
func (s *Service) DownloadAttachmentForAdmin(ctx context.Context, attachmentID string) (domain.SupportAttachment, []byte, error) {
	attachment, err := s.repo.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return domain.SupportAttachment{}, nil, err
	}
	data, err := s.storage.Get(ctx, attachment.StorageKey)
	if err != nil {
		return domain.SupportAttachment{}, nil, err
	}
	return attachment, data, nil
}

// attachmentKey construit la clé de stockage : support/<user>/<random>/<fichier>.
func attachmentKey(userID, filename string) string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return "support/" + userID + "/" + hex.EncodeToString(buf) + "/" + filename
}
