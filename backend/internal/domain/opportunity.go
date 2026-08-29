package domain

import "time"

// Niveaux de difficulté d'entrée sur une opportunité.
const (
	DifficultyLow    = "low"
	DifficultyMedium = "medium"
	DifficultyHigh   = "high"
)

// Classification de la fiabilité d'un signal (jamais de stat inventée).
const (
	SignalVerified   = "verified"
	SignalEstimated  = "estimated"
	SignalInferred   = "inferred"
	SignalHypothesis = "hypothesis"
)

// OpportunityScores décompose l'Opportunity Score (0-100 par critère).
type OpportunityScores struct {
	Demand           int `json:"demand"`
	Pain             int `json:"pain"`
	Competition      int `json:"competition"`
	PurchasingPower  int `json:"purchasing_power"`
	DigitalFit       int `json:"digital_fit"`
	EvidenceStrength int `json:"evidence_strength"`
}

// Evidence décrit une source d'un signal. Pour une donnée vérifiée, on
// conserve source, titre, URL, date de publication, pays, métrique et valeur.
type Evidence struct {
	Source          string `json:"source,omitempty"`
	Title           string `json:"title,omitempty"`
	URL             string `json:"url,omitempty"`
	PublicationDate string `json:"publication_date,omitempty"`
	Country         string `json:"country,omitempty"`
	Metric          string `json:"metric,omitempty"`
	Value           string `json:"value,omitempty"`
	RetrievedAt     string `json:"retrieved_at,omitempty"`
}

// Opportunity est une niche de marché scorée.
type Opportunity struct {
	ID         string
	Title      string
	Summary    string
	Country    string
	Sector     string
	Language   string
	Difficulty string
	Signal     string
	Score      int
	Scores     OpportunityScores
	Evidence   []Evidence
	IsSaved    bool
	CreatedAt  time.Time
}

// Market est un référentiel géo-économique (pays, langue, devise).
type Market struct {
	ID       string
	Code     string
	Name     string
	Currency string
	Language string
}
