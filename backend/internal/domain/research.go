package domain

import "time"

// Statuts d'une demande de recherche.
const (
	ResearchPending    = "pending"
	ResearchProcessing = "processing"
	ResearchCompleted  = "completed"
	ResearchFailed     = "failed"
)

// ResearchRequest est une demande de recherche en ligne (une niche, un
// secteur, plusieurs marchés cibles).
type ResearchRequest struct {
	ID        string
	UserID    string
	Query     string
	Sector    string
	Markets   []string
	Language  string
	Status    string
	Error     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
