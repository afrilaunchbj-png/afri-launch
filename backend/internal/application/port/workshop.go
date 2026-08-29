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
	Select(ctx context.Context, userID, id string) (domain.ProductIdea, error)
	Unselect(ctx context.Context, userID, id string) error
}

// ProjectRepository accède aux projets.
type ProjectRepository interface {
	Create(ctx context.Context, p domain.Project) (domain.Project, error)
	Get(ctx context.Context, userID, id string) (domain.Project, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Project, error)
	UpdateStatus(ctx context.Context, userID, id, status string) (domain.Project, error)
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
	Create(ctx context.Context, userID string, projectID, opportunityID *string, kind string, cost int64) (domain.GenerationJob, error)
	Get(ctx context.Context, userID, id string) (domain.GenerationJob, error)
	UpdateStatus(ctx context.Context, id, status string) (domain.GenerationJob, error)
	Complete(ctx context.Context, id string, result []byte, cost int64) (domain.GenerationJob, error)
	Fail(ctx context.Context, id, errMsg string) (domain.GenerationJob, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.GenerationJob, error)
}
