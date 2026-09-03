package postgres

import (
	"context"
	"encoding/json"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// conversationRepo implémente port.ConversationRepository.
type conversationRepo struct {
	s *Store
}

// NewConversationRepository construit le repository de conversations.
func NewConversationRepository(s *Store) *conversationRepo { return &conversationRepo{s: s} }

func (r *conversationRepo) Create(ctx context.Context, c domain.Conversation) (domain.Conversation, error) {
	row, err := r.s.q.CreateConversation(ctx, c.UserID)
	if err != nil {
		return domain.Conversation{}, err
	}
	return toConversation(row), nil
}

func (r *conversationRepo) Get(ctx context.Context, userID, id string) (domain.Conversation, error) {
	row, err := r.s.q.GetConversation(ctx, db.GetConversationParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.Conversation{}, domain.ErrNotFound
		}
		return domain.Conversation{}, err
	}
	return toConversation(row), nil
}

func (r *conversationRepo) List(ctx context.Context, userID string, limit, offset int) ([]domain.Conversation, error) {
	rows, err := r.s.q.ListConversations(ctx, db.ListConversationsParams{UserID: userID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Conversation, 0, len(rows))
	for _, row := range rows {
		out = append(out, toConversation(row))
	}
	return out, nil
}

func (r *conversationRepo) Touch(ctx context.Context, id string) (domain.Conversation, error) {
	row, err := r.s.q.TouchConversation(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return domain.Conversation{}, domain.ErrNotFound
		}
		return domain.Conversation{}, err
	}
	return toConversation(row), nil
}

func (r *conversationRepo) SetOpportunity(ctx context.Context, id string, opportunityID *string) (domain.Conversation, error) {
	row, err := r.s.q.SetConversationOpportunity(ctx, db.SetConversationOpportunityParams{ID: id, OpportunityID: strPtrToUUID(opportunityID)})
	if err != nil {
		if isNoRows(err) {
			return domain.Conversation{}, domain.ErrNotFound
		}
		return domain.Conversation{}, err
	}
	return toConversation(row), nil
}

func (r *conversationRepo) SetTitle(ctx context.Context, id, title string) (domain.Conversation, error) {
	row, err := r.s.q.SetConversationTitle(ctx, db.SetConversationTitleParams{ID: id, Title: title})
	if err != nil {
		if isNoRows(err) {
			return domain.Conversation{}, domain.ErrNotFound
		}
		return domain.Conversation{}, err
	}
	return toConversation(row), nil
}

func (r *conversationRepo) CreateMessage(ctx context.Context, m domain.ConversationMessage) (domain.ConversationMessage, error) {
	row, err := r.s.q.CreateConversationMessage(ctx, db.CreateConversationMessageParams{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		UserID:         m.UserID,
		Role:           m.Role,
		Content:        m.Content,
		Payload:        m.Payload,
	})
	if err != nil {
		return domain.ConversationMessage{}, err
	}
	return toConversationMessage(row), nil
}

func (r *conversationRepo) ListMessages(ctx context.Context, conversationID string) ([]domain.ConversationMessage, error) {
	rows, err := r.s.q.ListConversationMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ConversationMessage, 0, len(rows))
	for _, row := range rows {
		out = append(out, toConversationMessage(row))
	}
	return out, nil
}

func toConversation(c db.Conversation) domain.Conversation {
	return domain.Conversation{
		ID:            c.ID,
		UserID:        c.UserID,
		Title:         c.Title,
		Status:        c.Status,
		OpportunityID: uuidPtr(c.OpportunityID),
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func toConversationMessage(m db.ConversationMessage) domain.ConversationMessage {
	var payload []byte
	if m.Payload != nil {
		payload = []byte(m.Payload)
		if !json.Valid(payload) {
			payload = nil
		}
	}
	return domain.ConversationMessage{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		UserID:         m.UserID,
		Role:           m.Role,
		Content:        m.Content,
		Payload:        payload,
		CreatedAt:      m.CreatedAt,
	}
}
