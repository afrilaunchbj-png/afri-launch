package domain

import (
	"encoding/json"
	"strconv"
	"time"
)

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

// evidenceWire est la forme wire d'Evidence : `value` est décodé en brut pour
// accepter une chaîne OU un nombre renvoyé par le LLM (ex. 45 vs "45%").
type evidenceWire struct {
	Source          string          `json:"source"`
	Title           string          `json:"title"`
	URL             string          `json:"url"`
	PublicationDate string          `json:"publication_date"`
	Country         string          `json:"country"`
	Metric          string          `json:"metric"`
	Value           json.RawMessage `json:"value"`
	RetrievedAt     string          `json:"retrieved_at"`
}

// UnmarshalJSON tolère un `value` numérique : il est normalisé en chaîne.
func (e *Evidence) UnmarshalJSON(b []byte) error {
	var raw evidenceWire
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	e.Source, e.Title, e.URL = raw.Source, raw.Title, raw.URL
	e.PublicationDate, e.Country, e.Metric, e.RetrievedAt = raw.PublicationDate, raw.Country, raw.Metric, raw.RetrievedAt
	e.Value = scalarString(raw.Value)
	return nil
}

// scalarString convertit un json.RawMessage (string, nombre, bool) en chaîne.
func scalarString(v json.RawMessage) string {
	if len(v) == 0 || string(v) == "null" {
		return ""
	}
	var s string
	if json.Unmarshal(v, &s) == nil {
		return s
	}
	var f float64
	if json.Unmarshal(v, &f) == nil {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	var b bool
	if json.Unmarshal(v, &b) == nil {
		return strconv.FormatBool(b)
	}
	return ""
}

// Opportunity est une niche de marché scorée.
type Opportunity struct {
	ID         string
	UserID     *string // nil = catalogue global ; sinon propre à l'utilisateur
	ResearchID *string
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
