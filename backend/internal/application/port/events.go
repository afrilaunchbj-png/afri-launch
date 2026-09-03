package port

import "encoding/json"

// Types d'événements temps réel diffusés sur le canal SSE unique
// (GET /api/v1/events). Le client route par type + payload.
const (
	EventChatStarted   = "chat.started"
	EventChatDelta     = "chat.delta"
	EventChatTool      = "chat.tool"
	EventChatCompleted = "chat.completed"
	EventChatError     = "chat.error"
	EventJobUpdated    = "job.updated"
)

// AppEvent est une notification server→client du canal temps réel.
// Data est un JSON déjà sérialisé (évite de re-marshaler par abonné).
type AppEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// EventPublisher diffuse un événement aux connexions d'un utilisateur.
type EventPublisher interface {
	Publish(userID string, evt AppEvent)
}

// EventBus abonne des connexions SSE au flux d'événements d'un utilisateur.
// cancel désabonne (idempotent) ; la fermeture du channel signale la fin.
type EventBus interface {
	EventPublisher
	Subscribe(userID string) (<-chan AppEvent, func())
}
