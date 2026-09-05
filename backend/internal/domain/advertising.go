package domain

import (
	"errors"
	"time"
)

// Plateformes publicitaires supportées (extensible, ADR-017).
const (
	AdPlatformMeta      = "meta"
	AdPlatformGoogleAds = "google_ads"
	AdPlatformTikTokAds = "tiktok_ads"
)

// Statuts d'une connexion de plateforme publicitaire.
const (
	ConnPending      = "pending"      // OAuth en cours (compte non sélectionné)
	ConnActive       = "active"       // compte sélectionné, opérationnelle
	ConnExpired      = "expired"      // token expiré, reconnexion requise
	ConnRevoked      = "revoked"      // accès révoqué côté plateforme
	ConnError        = "error"        // erreur transitoire lors d'un appel
	ConnDisconnected = "disconnected" // déconnectée par l'utilisateur (historique conservé)
)

// Statuts de campagne publicitaire (normalisés inter-plateformes).
const (
	CampaignDraft   = "draft"
	CampaignActive  = "active"
	CampaignPaused  = "paused"
	CampaignDeleted = "deleted"
)

// Statuts d'une opération provider asynchrone.
const (
	OpPending    = "pending"
	OpProcessing = "processing"
	OpCompleted  = "completed"
	OpFailed     = "failed"
)

// AdPlatformConnection est une connexion OAuth d'un utilisateur à une
// plateforme publicitaire. Les tokens sont stockés chiffrés ; les champs
// AccessToken/RefreshToken de cette entité ne sont peuplés qu'en mémoire,
// juste avant un appel provider (jamais sérialisés vers le frontend).
type AdPlatformConnection struct {
	ID                   string
	UserID               string
	Provider             string
	Status               string
	ExternalUserID       string
	ExternalAccountID    string // compte publicitaire sélectionné (indexable)
	ExternalAccountName  string
	AccessToken          string // déchiffré en mémoire uniquement
	RefreshToken         string // déchiffré en mémoire uniquement
	AccessTokenExpiresAt *time.Time
	Scopes               []string
	Metadata             []byte // JSONB : business_id, customer_id, ad_account_currency…
	LastError            string
	LastErrorAt          *time.Time
	LastSyncAt           *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// AdAccount est un compte publicitaire accessible via une connexion
// (account discovery, normalisé inter-plateformes).
type AdAccount struct {
	ExternalID string `json:"id"`
	Name       string `json:"name"`
	Currency   string `json:"currency,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
	Status     string `json:"status,omitempty"`
}

// Campaign est une campagne publicitaire gérée par le SaaS. L'ID interne est
// un UUID ; l'ID externe (Meta/Google/TikTok) est conservé à part.
type Campaign struct {
	ID                 string
	UserID             string
	ConnectionID       string
	ExternalCampaignID string
	Name               string
	Objective          string
	Status             string
	BudgetMinor        int64 // budget journalier en unités mineures (jamais de float)
	Currency           string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Kinds de creative.
const (
	CreativeVideo    = "video"
	CreativeImage    = "image"
	CreativeCarousel = "carousel"
	CreativeText     = "text"
)

// Creative est un visuel publicitaire lié à un asset interne (ex. vidéo du
// pipeline ADR-016) et, après publication, à son ID externe.
type Creative struct {
	ID                 string
	UserID             string
	CampaignID         *string
	ConnectionID       string
	Type               string
	AssetID            *string
	ExternalCreativeID string
	Headline           string
	PrimaryText        string
	CTA                string
	Status             string
	Metadata           []byte // JSONB
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Insight est une métrique de performance normalisée pour une journée.
// Les définitions exactes restent dans Metadata (brut provider).
type Insight struct {
	CampaignID  string  `json:"campaign_id,omitempty"`
	Date        string  `json:"date"`
	Impressions int64   `json:"impressions"`
	Reach       int64   `json:"reach"`
	Clicks      int64   `json:"clicks"`
	SpendMinor  int64   `json:"spend_minor"`
	CTR         float64 `json:"ctr,omitempty"`
	CPCMinor    int64   `json:"cpc_minor,omitempty"`
	CPMMinor    int64   `json:"cpm_minor,omitempty"`
	Conversions float64 `json:"conversions,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Metadata    []byte  `json:"-"` // brut provider, jamais exposé au FE
}

// OAuthState est un jeton CSRF éphémère lié à un utilisateur et un provider.
type OAuthState struct {
	State     string
	UserID    string
	Provider  string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

// ProviderOperation trace chaque appel mutatif vers un provider (idempotence,
// retry, reconciliation). Ne stocke jamais de token.
type ProviderOperation struct {
	ID                 string
	UserID             string
	ConnectionID       string
	Provider           string
	OperationType      string
	Status             string
	Attempts           int
	InternalResourceID string
	ExternalResourceID string
	ErrorCode          string
	ErrorMessage       string
	CreatedAt          time.Time
	StartedAt          *time.Time
	CompletedAt        *time.Time
}

// Erreurs métier du module advertising (sentinelles, cf. errors.go).
var (
	ErrConnectionNotFound   = ErrNotFound
	ErrConnectionNotActive  = errors.New("advertising: connexion non active")
	ErrOAuthStateInvalid    = errors.New("advertising: état OAuth invalide ou expiré")
	ErrAccountNotAccessible = errors.New("advertising: compte non accessible via cette connexion")
	ErrProviderUnavailable  = errors.New("advertising: plateforme indisponible")
	ErrBudgetLimitExceeded  = errors.New("advertising: limite de budget dépassée")
)
