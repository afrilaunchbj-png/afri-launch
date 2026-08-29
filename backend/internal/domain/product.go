package domain

import "time"

// ProductIdea est un concept de produit généré pour une opportunité.
type ProductIdea struct {
	ID              string
	UserID          string
	OpportunityID   *string
	Title           string
	Subtitle        string
	Audience        string
	Problem         string
	Promise         string
	Format          string
	EstimatedPrice  string
	Difficulty      string
	MarketEvidence  string
	WhyNow          string
	CompetitiveAngle string
	IsSelected      bool
	CreatedAt       time.Time
}

// Project est un projet de produit issu d'une idée.
type Project struct {
	ID              string
	UserID          string
	OpportunityID   *string
	IdeaID          *string
	Title           string
	Status          string
	CreditsConsumed int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Statuts de projet.
const (
	ProjectDraft        = "draft"
	ProjectIdeaSelected = "idea_selected"
	ProjectGenerating   = "generating"
	ProjectContentReady = "content_ready"
	ProjectCompleted    = "completed"
	ProjectFailed       = "failed"
)

// Asset est un fichier généré (ebook PDF, couverture, page de vente…).
type Asset struct {
	ID          string
	UserID      string
	ProjectID   string
	Kind        string
	StorageKey  string
	Filename    string
	ContentType string
	SizeBytes   int64
	CreatedAt   time.Time
}

// Kinds d'asset.
const (
	AssetEbookPDF  = "ebook_pdf"
	AssetEbookDeck = "ebook_deck"
	AssetCover     = "cover"
	AssetPoster    = "poster"
	AssetSalesPage = "sales_page"
)

// GenerationJob est un job de génération asynchrone.
type GenerationJob struct {
	ID            string
	UserID        string
	ProjectID     *string
	OpportunityID *string
	Kind          string
	Status        string
	Error         string
	Cost          int64
	Result        []byte // JSONB
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
}

// Kinds et statuts de job.
const (
	JobIdeas     = "ideas"
	JobEbook     = "ebook"
	JobCover     = "cover"
	JobPosters   = "posters"
	JobSalesPage = "sales_page"
)

const (
	JobPending    = "pending"
	JobProcessing = "processing"
	JobCompleted  = "completed"
	JobFailed     = "failed"
)
