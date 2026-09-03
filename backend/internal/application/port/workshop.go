package port

import (
	"context"

	"afrilaunch/backend/internal/domain"
)

// Storage stocke et relit des objets (PDF, images…).
type Storage interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
}

// IdeaRepository accède aux idées de produit.
type IdeaRepository interface {
	Create(ctx context.Context, idea domain.ProductIdea) (domain.ProductIdea, error)
	Get(ctx context.Context, userID, id string) (domain.ProductIdea, error)
	ListByUser(ctx context.Context, userID string) ([]domain.ProductIdea, error)
	ListByOpportunity(ctx context.Context, userID, opportunityID string) ([]domain.ProductIdea, error)
	ListByConversation(ctx context.Context, userID, conversationID string) ([]domain.ProductIdea, error)
	Select(ctx context.Context, userID, id string) (domain.ProductIdea, error)
	Unselect(ctx context.Context, userID, id string) error
	UpdateContent(ctx context.Context, idea domain.ProductIdea) (domain.ProductIdea, error)
	SetStatus(ctx context.Context, userID, id, status string) (domain.ProductIdea, error)
}

// ConversationRepository accède aux conversations du chat.
type ConversationRepository interface {
	Create(ctx context.Context, c domain.Conversation) (domain.Conversation, error)
	Get(ctx context.Context, userID, id string) (domain.Conversation, error)
	List(ctx context.Context, userID string, limit, offset int) ([]domain.Conversation, error)
	Touch(ctx context.Context, id string) (domain.Conversation, error)
	SetOpportunity(ctx context.Context, id string, opportunityID *string) (domain.Conversation, error)
	SetTitle(ctx context.Context, id, title string) (domain.Conversation, error)
	CreateMessage(ctx context.Context, m domain.ConversationMessage) (domain.ConversationMessage, error)
	ListMessages(ctx context.Context, conversationID string) ([]domain.ConversationMessage, error)
}

// PreferenceRepository accède aux préférences utilisateur (langue, thème).
type PreferenceRepository interface {
	// GetOrCreate renvoie les préférences, en créant les valeurs par défaut
	// (langue fr, thème system) si absentes.
	GetOrCreate(ctx context.Context, userID string) (domain.UserPreference, error)
	Upsert(ctx context.Context, p domain.UserPreference) (domain.UserPreference, error)
}

// IdeaMessageRepository accède à l'historique de conversation d'une idée.
type IdeaMessageRepository interface {
	Create(ctx context.Context, m domain.IdeaMessage) (domain.IdeaMessage, error)
	ListByIdea(ctx context.Context, ideaID string) ([]domain.IdeaMessage, error)
}

// ProjectRepository accède aux projets.
type ProjectRepository interface {
	Create(ctx context.Context, p domain.Project) (domain.Project, error)
	Get(ctx context.Context, userID, id string) (domain.Project, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Project, error)
	UpdateStatus(ctx context.Context, userID, id, status string) (domain.Project, error)
	UpdateConfig(ctx context.Context, userID, id string, config []byte) (domain.Project, error)
	AddCredits(ctx context.Context, id string, amount int64) (domain.Project, error)
}

// AssetRepository accède aux assets générés.
type AssetRepository interface {
	Create(ctx context.Context, a domain.Asset) (domain.Asset, error)
	Get(ctx context.Context, userID, id string) (domain.Asset, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.Asset, error)
}

// JobRepository accède aux jobs de génération.
type JobRepository interface {
	Create(ctx context.Context, userID string, projectID, opportunityID, researchID, ideaID *string, kind string, cost int64, params []byte) (domain.GenerationJob, error)
	Get(ctx context.Context, userID, id string) (domain.GenerationJob, error)
	UpdateStatus(ctx context.Context, id, status string) (domain.GenerationJob, error)
	Complete(ctx context.Context, id string, result []byte, cost int64) (domain.GenerationJob, error)
	Fail(ctx context.Context, id, errMsg string) (domain.GenerationJob, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.GenerationJob, error)
}
