package port

import (
	"context"

	"afrilaunch/backend/internal/domain"
)

// ResearchRequest est une demande de recherche en ligne auprès du LLM.
type ResearchRequest struct {
	Model    string
	System   string // consigne système (format de sortie, contraintes)
	Query    string // la question/niche à explorer en ligne
}

// ResearchResult est le résultat structuré d'une recherche en ligne.
// Content est le texte JSON à décoder côté use case (opportunités + sources).
type ResearchResult struct {
	Content string
	Sources []Source
}

// Source est une source/citation web remontée par la recherche.
type Source struct {
	Title string
	URL   string
}

// ResearchProvider recherche des informations en ligne (web search) et
// renvoie un contenu structuré + ses sources.
type ResearchProvider interface {
	Research(ctx context.Context, req ResearchRequest) (ResearchResult, error)
}

// ResearchRepository accède aux demandes de recherche.
type ResearchRepository interface {
	Create(ctx context.Context, r domain.ResearchRequest) (domain.ResearchRequest, error)
	Get(ctx context.Context, userID, id string) (domain.ResearchRequest, error)
	UpdateStatus(ctx context.Context, id, status string) (domain.ResearchRequest, error)
}
