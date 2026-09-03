package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/chat"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// ConversationHandler expose les conversations du copilote.
type ConversationHandler struct {
	svc *chat.Service
}

// NewConversationHandler construit le handler de conversations.
func NewConversationHandler(svc *chat.Service) *ConversationHandler {
	return &ConversationHandler{svc: svc}
}

type conversationDTO struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	OpportunityID *string   `json:"opportunity_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toConversationDTO(c domain.Conversation) conversationDTO {
	return conversationDTO{
		ID: c.ID, Title: c.Title, Status: c.Status,
		OpportunityID: c.OpportunityID, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

type conversationMessageDTO struct {
	ID        string          `json:"id"`
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

func toConversationMessageDTO(m domain.ConversationMessage) conversationMessageDTO {
	return conversationMessageDTO{
		ID: m.ID, Role: m.Role, Content: m.Content,
		Payload: json.RawMessage(m.Payload), CreatedAt: m.CreatedAt,
	}
}

// conversationDetailDTO hydrate la page chat en un seul appel.
type conversationDetailDTO struct {
	conversationDTO
	Opportunity *opportunityDTO          `json:"opportunity,omitempty"`
	Messages    []conversationMessageDTO `json:"messages"`
	Ideas       []ideaDTO                `json:"ideas"`
}

// Create gère POST /conversations.
func (h *ConversationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	conv, err := h.svc.Create(r.Context(), userID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toConversationDTO(conv))
}

// List gère GET /conversations.
func (h *ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	items, err := h.svc.List(r.Context(), userID, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]conversationDTO, 0, len(items))
	for _, c := range items {
		out = append(out, toConversationDTO(c))
	}
	writeData(w, http.StatusOK, out)
}

// Get gère GET /conversations/{id} (détail : conversation + opportunité + messages + idées).
func (h *ConversationHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	d, err := h.svc.Detail(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}

	out := conversationDetailDTO{
		conversationDTO: toConversationDTO(d.Conversation),
		Messages:        make([]conversationMessageDTO, 0, len(d.Messages)),
		Ideas:           make([]ideaDTO, 0, len(d.Ideas)),
	}
	if d.Opportunity != nil {
		opp := toOpportunityDTO(*d.Opportunity)
		out.Opportunity = &opp
	}
	for _, m := range d.Messages {
		out.Messages = append(out.Messages, toConversationMessageDTO(m))
	}
	for _, i := range d.Ideas {
		out.Ideas = append(out.Ideas, toIdeaDTO(i))
	}
	writeData(w, http.StatusOK, out)
}

type sendChatMessageRequest struct {
	Content string `json:"content"`
}

type sendMessageAcceptedDTO struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
}

// SendMessage gère POST /conversations/{id}/messages → 202.
// Le streaming du tour arrive sur GET /events (événements chat.*).
func (h *ConversationHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	convID := chi.URLParam(r, "id")

	var in sendChatMessageRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}

	messageID, err := h.svc.SendMessage(r.Context(), userID, convID, in.Content)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusAccepted, sendMessageAcceptedDTO{ConversationID: convID, MessageID: messageID})
}
