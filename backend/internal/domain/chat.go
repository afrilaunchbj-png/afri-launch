package domain

import "time"

// Conversation est un chat multi-tours entre l'utilisateur et le copilote
// (recherche d'opportunité → idées → idée validée).
type Conversation struct {
	ID            string
	UserID        string
	Title         string
	Status        string
	OpportunityID *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Statuts de conversation.
const (
	ConversationActive   = "active"
	ConversationArchived = "archived"
)

// ConversationMessage est un message du chat. Payload contient les
// métadonnées JSON (ids d'idées proposées, sources de recherche…).
type ConversationMessage struct {
	ID             string
	ConversationID string
	UserID         string
	Role           string
	Content        string
	Payload        []byte
	CreatedAt      time.Time
}

// Rôles des messages de conversation.
const (
	ConversationMessageUser      = "user"
	ConversationMessageAssistant = "assistant"
)
