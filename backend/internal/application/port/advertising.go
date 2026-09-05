package port

import (
	"context"
	"time"

	"afrilaunch/backend/internal/domain"
)

// AdPlatformCapabilities décrit ce que supporte une plateforme — les
// providers n'implémentent pas tous les mêmes fonctionnalités.
type AdPlatformCapabilities struct {
	Campaigns          bool
	Creatives          bool
	VideoAds           bool
	ImageAds           bool
	ConversionTracking bool
	Reporting          bool
	BudgetManagement   bool
}

// OAuthTokenResult est le résultat d'un exchange/refresh de token OAuth.
type OAuthTokenResult struct {
	AccessToken    string
	RefreshToken   string // vide si la plateforme n'en fournit pas
	ExpiresAt      *time.Time
	Scopes         []string
	ExternalUserID string
}

// CampaignQuery filtre la liste des campagnes chez le provider.
type CampaignQuery struct {
	// vide = toutes les campagnes non supprimées
	Status string
}

// CreateCampaignInput décrit la création d'une campagne (validation des
// garde-fous budget effectuée par le service applicatif, pas par le provider).
type CreateCampaignInput struct {
	Name        string
	Objective   string
	BudgetMinor int64
	Currency    string
}

// UpdateCampaignInput permet de modifier nom et budget (champs vides = inchangés).
type UpdateCampaignInput struct {
	Name        string
	BudgetMinor int64 // 0 = inchangé
}

// CreativeInput décrit la publication d'un visuel (l'asset est d'abord
// résolu en URL par le service, le provider ne connaît que l'URL publique).
type CreativeInput struct {
	Type        string // domain.CreativeVideo / CreativeImage
	URL         string
	MimeType    string
	Headline    string
	PrimaryText string
	CTA         string
}

// InsightsQuery borne la période de reporting (dates YYYY-MM-DD).
type InsightsQuery struct {
	Since string
	Until string
}

// AdPlatformProvider encapsule TOUTE la logique d'une plateforme
// publicitaire (OAuth + comptes + campagnes + creatives + insights).
// Aucune logique Meta/Google/TikTok ne doit sortir de cette interface
// (prompts/marketing-flow.md §1, §3).
type AdPlatformProvider interface {
	Provider() string
	Capabilities() AdPlatformCapabilities

	// OAuth (l'état CSRF est géré par le service applicatif).
	AuthorizationURL(state, redirectURI string) string
	ExchangeCode(ctx context.Context, code, redirectURI string) (OAuthTokenResult, error)
	// Refresh renvoie un nouveau token ; refreshable=false si la plateforme
	// n'a pas de mécanisme de refresh (Meta = prolongation longue durée).
	Refresh(ctx context.Context, conn domain.AdPlatformConnection) (OAuthTokenResult, bool, error)
	Revoke(ctx context.Context, accessToken string) error

	// Account discovery.
	ListAdAccounts(ctx context.Context, accessToken string) ([]domain.AdAccount, error)
	// VerifyAdAccount vérifie que le compte est réellement accessible (un ID
	// envoyé par le frontend n'est jamais fiable).
	VerifyAdAccount(ctx context.Context, accessToken, externalAccountID string) (domain.AdAccount, error)

	// Campagnes.
	ListCampaigns(ctx context.Context, conn domain.AdPlatformConnection, q CampaignQuery) ([]domain.Campaign, error)
	CreateCampaign(ctx context.Context, conn domain.AdPlatformConnection, in CreateCampaignInput) (domain.Campaign, error)
	UpdateCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string, in UpdateCampaignInput) (domain.Campaign, error)
	PauseCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error
	ResumeCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error

	// Creatives.
	UploadCreative(ctx context.Context, conn domain.AdPlatformConnection, in CreativeInput) (externalCreativeID string, err error)

	// Reporting.
	GetInsights(ctx context.Context, conn domain.AdPlatformConnection, externalCampaignID string, q InsightsQuery) ([]domain.Insight, error)
}

// SecretEncryptor chiffre/déchiffre les tokens au repos (AES-256-GCM).
// Format : "v1:<iv b64>:<ciphertext b64>" — le préfixe permet la rotation.
type SecretEncryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// OAuthStateStore persistance des états CSRF OAuth (anti-replay : consume).
type OAuthStateStore interface {
	Create(ctx context.Context, s domain.OAuthState) error
	// Consume vérifie (provider, expiration, non utilisé) puis marque utilisé
	// et renvoie l'état (dont l'utilisateur lié — le callback OAuth n'est pas
	// authentifié par le JWT). ErrOAuthStateInvalid si absent/expiré/replay.
	Consume(ctx context.Context, state, provider string) (domain.OAuthState, error)
	// Prune supprime les états expirés (maintenance).
	Prune(ctx context.Context) error
}

// StorageSigner produit des URLs signées temporairement publiques pour un
// objet privé (nécessaire pour que la plateforme télécharge la creative).
// Implémenté par le stockage S3 ; absent en local (uploads créatives
// indisponibles en dev sans S3).
type StorageSigner interface {
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// AdConnectionRepository accède aux connexions publicitaires. Le repository
// chiffre/déchiffre les tokens au repos via le SecretEncryptor injecté ;
// l'entité domaine porte les tokens en clair en mémoire uniquement.
type AdConnectionRepository interface {
	Create(ctx context.Context, c domain.AdPlatformConnection) (domain.AdPlatformConnection, error)
	// UpdateTokens met à jour tokens chiffrés + statut (après callback/refresh).
	UpdateTokens(ctx context.Context, id string, accessEnc, refreshEnc string, expiresAt *time.Time, externalUserID string, scopes []string, status string) error
	// SelectAccount associe le compte publicitaire choisi (vérifié côté backend).
	SelectAccount(ctx context.Context, userID, id, externalAccountID, accountName string, metadata []byte) error
	SetStatus(ctx context.Context, userID, id, status, lastError string) error
	SetSynced(ctx context.Context, id string, at time.Time) error
	Get(ctx context.Context, userID, id string) (domain.AdPlatformConnection, error)
	GetByProvider(ctx context.Context, userID, provider string) (domain.AdPlatformConnection, error)
	List(ctx context.Context, userID string) ([]domain.AdPlatformConnection, error)
	ListByProvider(ctx context.Context, provider string) ([]domain.AdPlatformConnection, error)
}

// AdCampaignRepository accède aux campagnes internes (mapping externe).
type AdCampaignRepository interface {
	Upsert(ctx context.Context, c domain.Campaign) (domain.Campaign, error)
	Update(ctx context.Context, userID, id string, c domain.Campaign) (domain.Campaign, error)
	Get(ctx context.Context, userID, id string) (domain.Campaign, error)
	List(ctx context.Context, userID string) ([]domain.Campaign, error)
	ListByConnection(ctx context.Context, connectionID string) ([]domain.Campaign, error)
}

// AdCreativeRepository accède aux creatives.
type AdCreativeRepository interface {
	Create(ctx context.Context, c domain.Creative) (domain.Creative, error)
	UpdateExternal(ctx context.Context, userID, id, externalCreativeID, status string) error
	Get(ctx context.Context, userID, id string) (domain.Creative, error)
	List(ctx context.Context, userID string) ([]domain.Creative, error)
}

// AdInsightRepository persiste les insights normalisés (upsert par campagne+jour).
type AdInsightRepository interface {
	Upsert(ctx context.Context, campaignID, userID string, insights []domain.Insight) error
	ListByCampaign(ctx context.Context, userID, campaignID string, since, until string) ([]domain.Insight, error)
}

// ProviderOperationRepository trace les opérations mutatives provider.
type ProviderOperationRepository interface {
	Create(ctx context.Context, op domain.ProviderOperation) (domain.ProviderOperation, error)
	MarkProcessing(ctx context.Context, id string) error
	Complete(ctx context.Context, id, externalResourceID string) error
	Fail(ctx context.Context, id, code, message string) error
	IncrementAttempts(ctx context.Context, id string) error
	Get(ctx context.Context, userID, id string) (domain.ProviderOperation, error)
	List(ctx context.Context, userID string, limit int) ([]domain.ProviderOperation, error)
}
