package postgres

import (
	"context"

	"github.com/google/uuid"

	"afrilaunch/backend/internal/application/port"
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

func (r *supportRepo) ListAll(ctx context.Context, f port.AdminListFilter, limit, offset int) ([]domain.AdminTicket, int64, error) {
	rows, err := r.s.q.ListAllTickets(ctx, db.ListAllTicketsParams{
		Status: f.Status, Search: f.Search, RowLimit: int32(limit), RowOffset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.s.q.CountAllTickets(ctx, db.CountAllTicketsParams{Status: f.Status, Search: f.Search})
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

func (r *supportRepo) GetWithUser(ctx context.Context, id string) (domain.AdminTicket, error) {
	row, err := r.s.q.GetTicketWithUser(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return domain.AdminTicket{}, domain.ErrNotFound
		}
		return domain.AdminTicket{}, err
	}
	ticket := domain.SupportTicket{
		ID: row.ID, UserID: row.UserID, Subject: row.Subject, Message: row.Message,
		Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	return domain.AdminTicket{SupportTicket: ticket, UserEmail: row.UserEmail, UserName: row.UserName}, nil
}

func (r *supportRepo) ListMessages(ctx context.Context, ticketID string) ([]domain.TicketMessageView, error) {
	rows, err := r.s.q.ListTicketMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.TicketMessageView, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.TicketMessageView{
			TicketMessage: domain.TicketMessage{
				ID: row.ID, TicketID: row.TicketID, AuthorID: row.AuthorID,
				Content: row.Content, IsAdmin: row.IsAdmin, CreatedAt: row.CreatedAt,
			},
			AuthorEmail: row.AuthorEmail,
			AuthorName:  row.AuthorName,
		})
	}
	return out, nil
}

func (r *supportRepo) AddMessage(ctx context.Context, ticketID string, msg domain.TicketMessage) (domain.TicketMessageView, error) {
	row, err := r.s.q.InsertTicketMessage(ctx, db.InsertTicketMessageParams{
		TicketID: ticketID, AuthorID: msg.AuthorID, Content: msg.Content, IsAdmin: msg.IsAdmin,
	})
	if err != nil {
		return domain.TicketMessageView{}, err
	}
	return domain.TicketMessageView{
		TicketMessage: domain.TicketMessage{
			ID: row.ID, TicketID: row.TicketID, AuthorID: row.AuthorID,
			Content: row.Content, IsAdmin: row.IsAdmin, CreatedAt: row.CreatedAt,
		},
		AuthorEmail: row.AuthorEmail,
		AuthorName:  row.AuthorName,
	}, nil
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

func (r *supportRepo) CreateAttachment(ctx context.Context, a domain.SupportAttachment) (domain.SupportAttachment, error) {
	row, err := r.s.q.CreateSupportAttachment(ctx, db.CreateSupportAttachmentParams{
		UserID:      a.UserID,
		Filename:    a.Filename,
		StorageKey:  a.StorageKey,
		ContentType: a.ContentType,
		SizeBytes:   a.SizeBytes,
	})
	if err != nil {
		return domain.SupportAttachment{}, err
	}
	return toAttachment(row), nil
}

func (r *supportRepo) BindAttachments(ctx context.Context, userID string, attachmentIDs []string, ticketID, messageID string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}
	return r.s.q.BindSupportAttachments(ctx, db.BindSupportAttachmentsParams{
		UserID:   userID,
		Column2:  attachmentIDs,
		TicketID: toUUIDPtr(&ticketID),
		Column4:  messageID,
	})
}

func (r *supportRepo) GetAttachment(ctx context.Context, userID, id string) (domain.SupportAttachment, error) {
	row, err := r.s.q.GetSupportAttachment(ctx, db.GetSupportAttachmentParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.SupportAttachment{}, domain.ErrNotFound
		}
		return domain.SupportAttachment{}, err
	}
	return toAttachment(row), nil
}

func (r *supportRepo) ListAttachmentsByTicket(ctx context.Context, ticketID string) ([]domain.SupportAttachment, error) {
	rows, err := r.s.q.ListSupportAttachmentsByTicket(ctx, toUUIDPtr(&ticketID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.SupportAttachment, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAttachment(row))
	}
	return out, nil
}

func (r *supportRepo) ListAttachmentsByMessages(ctx context.Context, messageIDs []string) ([]domain.SupportAttachment, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}
	rows, err := r.s.q.ListSupportAttachmentsByMessages(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	out := make([]domain.SupportAttachment, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAttachment(row))
	}
	return out, nil
}

func toAttachment(row db.SupportAttachment) domain.SupportAttachment {
	ticketID, messageID := "", ""
	if row.TicketID.Valid {
		ticketID = uuid.UUID(row.TicketID.Bytes).String()
	}
	if row.MessageID.Valid {
		messageID = uuid.UUID(row.MessageID.Bytes).String()
	}
	return domain.SupportAttachment{
		ID: row.ID, UserID: row.UserID, TicketID: ticketID, MessageID: messageID,
		Filename: row.Filename, StorageKey: row.StorageKey,
		ContentType: row.ContentType, SizeBytes: row.SizeBytes, CreatedAt: row.CreatedAt,
	}
}

func (r *supportRepo) GetAttachmentByID(ctx context.Context, id string) (domain.SupportAttachment, error) {
	row, err := r.s.q.GetSupportAttachmentByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return domain.SupportAttachment{}, domain.ErrNotFound
		}
		return domain.SupportAttachment{}, err
	}
	return toAttachment(row), nil
}
