// Package advertising orchestre les intégrations publicitaires multi-tenant
// (connexions OAuth, comptes, campagnes, creatives, insights) derrière le
// port AdPlatformProvider — aucune logique Meta/Google/TikTok ici
// (ADR-017, prompts/marketing-flow.md).
package advertising

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// ProviderRegistry résout un provider par clé (meta, google_ads, tiktok_ads).
type ProviderRegistry map[string]port.AdPlatformProvider

// SafetyPolicy porte les garde-fous budgétaires (marketing-flow §32) —
// aucune dépense ne doit pouvoir être déclenchée sans limite.
type SafetyPolicy struct {
	MaxDailySpendMinor int64 // budget journalier max par campagne (unités mineures)
	MaxCampaigns       int   // campagnes actives max par utilisateur
}

// DefaultSafetyPolicy : garde-fous prudentiels par défaut.
func DefaultSafetyPolicy() SafetyPolicy {
	return SafetyPolicy{MaxDailySpendMinor: 1_000_000, MaxCampaigns: 10} // ex. 10 000 XOF/jour, 10 campagnes
}

// Service orchestre les intégrations publicitaires.
type Service struct {
	providers   ProviderRegistry
	encryptor   port.SecretEncryptor
	states      port.OAuthStateStore
	connections port.AdConnectionRepository
	campaigns   port.AdCampaignRepository
	creatives   port.AdCreativeRepository
	insights    port.AdInsightRepository
	operations  port.ProviderOperationRepository
	assets      port.AssetRepository
	storage     port.Storage
	signer      port.StorageSigner // optionnel (S3) : URLs signées pour les creatives
	policy      SafetyPolicy
	stateTTL    time.Duration
}

// NewService construit le service advertising.
func NewService(
	providers ProviderRegistry,
	encryptor port.SecretEncryptor,
	states port.OAuthStateStore,
	connections port.AdConnectionRepository,
	campaigns port.AdCampaignRepository,
	creatives port.AdCreativeRepository,
	insights port.AdInsightRepository,
	operations port.ProviderOperationRepository,
	assets port.AssetRepository,
	storage port.Storage,
	signer port.StorageSigner,
	policy SafetyPolicy,
) *Service {
	return &Service{
		providers: providers, encryptor: encryptor, states: states,
		connections: connections, campaigns: campaigns, creatives: creatives,
		insights: insights, operations: operations, assets: assets,
		storage: storage, signer: signer, policy: policy,
		stateTTL: 10 * time.Minute,
	}
}

// Capabilities expose les capacités par provider (affichage FE).
func (s *Service) Capabilities() map[string]port.AdPlatformCapabilities {
	out := make(map[string]port.AdPlatformCapabilities, len(s.providers))
	for name, p := range s.providers {
		out[name] = p.Capabilities()
	}
	return out
}

// ---------- Connexion OAuth ----------

// StartConnect génère un état CSRF lié à l'utilisateur et renvoie l'URL
// d'autorisation du provider.
func (s *Service) StartConnect(ctx context.Context, userID, provider, redirectURI string) (string, error) {
	p, err := s.provider(provider)
	if err != nil {
		return "", err
	}
	state := randomState()
	if err := s.states.Create(ctx, domain.OAuthState{
		State: state, UserID: userID, Provider: provider,
		ExpiresAt: time.Now().Add(s.stateTTL),
	}); err != nil {
		return "", err
	}
	return p.AuthorizationURL(state, redirectURI), nil
}

// HandleCallback traite le retour OAuth : consomme l'état (anti-CSRF/replay),
// échange le code, chiffre et persiste les tokens, puis crée/rafraîchit la
// connexion. Le callback n'est pas authentifié par JWT — l'utilisateur est
// identifié par l'état consommé.
func (s *Service) HandleCallback(ctx context.Context, provider, code, state, redirectURI string) (domain.AdPlatformConnection, error) {
	p, err := s.provider(provider)
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	st, err := s.states.Consume(ctx, state, provider)
	if err != nil {
		return domain.AdPlatformConnection{}, domain.ErrOAuthStateInvalid
	}

	tokens, err := p.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return domain.AdPlatformConnection{}, fmt.Errorf("%w: %v", domain.ErrProviderUnavailable, err)
	}

	// Upsert de la connexion (1 par user+provider) : tokens régénérés.
	existing, getErr := s.connections.GetByProvider(ctx, st.UserID, provider)
	if getErr != nil && !errors.Is(getErr, domain.ErrConnectionNotFound) {
		return domain.AdPlatformConnection{}, getErr
	}
	var conn domain.AdPlatformConnection
	if errors.Is(getErr, domain.ErrConnectionNotFound) {
		conn, err = s.connections.Create(ctx, domain.AdPlatformConnection{
			UserID: st.UserID, Provider: provider, Status: domain.ConnPending,
		})
		if err != nil {
			return domain.AdPlatformConnection{}, err
		}
	} else {
		conn = existing
	}

	accessEnc, err := s.encryptor.Encrypt(tokens.AccessToken)
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	refreshEnc, err := s.encryptor.Encrypt(tokens.RefreshToken)
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	status := domain.ConnPending
	if conn.ExternalAccountID != "" {
		status = domain.ConnActive // compte déjà sélectionné (reconnexion)
	}
	if err := s.connections.UpdateTokens(ctx, conn.ID, accessEnc, refreshEnc, tokens.ExpiresAt, tokens.ExternalUserID, tokens.Scopes, status); err != nil {
		return domain.AdPlatformConnection{}, err
	}
	return s.connections.Get(ctx, st.UserID, conn.ID)
}

// ListAccounts renvoie les comptes publicitaires accessibles via la connexion.
func (s *Service) ListAccounts(ctx context.Context, userID, provider string) ([]domain.AdAccount, error) {
	p, conn, err := s.activeConnection(ctx, userID, provider)
	if err != nil {
		return nil, err
	}
	return p.ListAdAccounts(ctx, conn.AccessToken)
}

// SelectAccount associe un compte publicitaire (l'ID du frontend est
// re-vérifié auprès du provider — jamais considéré fiable).
func (s *Service) SelectAccount(ctx context.Context, userID, provider, externalAccountID string) (domain.AdPlatformConnection, error) {
	p, err := s.provider(provider)
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	conn, err := s.connections.GetByProvider(ctx, userID, provider)
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	account, err := p.VerifyAdAccount(ctx, conn.AccessToken, externalAccountID)
	if err != nil {
		return domain.AdPlatformConnection{}, domain.ErrAccountNotAccessible
	}
	metadata, _ := json.Marshal(map[string]string{"currency": account.Currency})
	if err := s.connections.SelectAccount(ctx, userID, conn.ID, account.ExternalID, account.Name, metadata); err != nil {
		return domain.AdPlatformConnection{}, err
	}
	return s.connections.Get(ctx, userID, conn.ID)
}

// Disconnect déconnecte (révoque si possible) sans supprimer l'historique.
func (s *Service) Disconnect(ctx context.Context, userID, provider string) error {
	p, err := s.provider(provider)
	if err != nil {
		return err
	}
	conn, err := s.connections.GetByProvider(ctx, userID, provider)
	if err != nil {
		return err
	}
	if conn.AccessToken != "" {
		if err := p.Revoke(ctx, conn.AccessToken); err != nil {
			slog.Warn("advertising: revoke failed", "provider", provider, "err", err)
		}
	}
	return s.connections.SetStatus(ctx, userID, conn.ID, domain.ConnDisconnected, "")
}

// ListConnections expose les connexions de l'utilisateur (sans tokens).
func (s *Service) ListConnections(ctx context.Context, userID string) ([]domain.AdPlatformConnection, error) {
	return s.connections.List(ctx, userID)
}

// ---------- Campagnes ----------

// SyncCampaigns synchronise les campagnes du compte (mapping interne/externe).
func (s *Service) SyncCampaigns(ctx context.Context, userID, provider string) ([]domain.Campaign, error) {
	p, conn, err := s.activeConnection(ctx, userID, provider)
	if err != nil {
		return nil, err
	}
	remote, err := p.ListCampaigns(ctx, conn, port.CampaignQuery{})
	if err != nil {
		_ = s.connections.SetStatus(ctx, userID, conn.ID, domain.ConnError, err.Error())
		return nil, err
	}
	out := make([]domain.Campaign, 0, len(remote))
	for _, c := range remote {
		if c.Status == domain.CampaignDeleted {
			continue
		}
		saved, err := s.campaigns.Upsert(ctx, c)
		if err != nil {
			return nil, err
		}
		out = append(out, saved)
	}
	if err := s.connections.SetSynced(ctx, conn.ID, time.Now()); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateCampaignInput décrit la création d'une campagne + creative.
type CreateCampaignInput struct {
	Provider    string `json:"provider"`
	Name        string `json:"name"`
	Objective   string `json:"objective"`
	BudgetMinor int64  `json:"budget_minor"`
	AssetID     string `json:"asset_id"` // creative : asset interne (ex. vidéo)
	Headline    string `json:"headline"`
	PrimaryText string `json:"primary_text"`
	CTA         string `json:"cta"`
}

// CreateCampaign crée la campagne (en pause) puis publie la creative si un
// asset est fourni. Chaque opération est tracée dans provider_operations
// (idempotence, réconciliation) et auditable.
func (s *Service) CreateCampaign(ctx context.Context, userID string, in CreateCampaignInput) (domain.Campaign, error) {
	p, conn, err := s.activeConnection(ctx, userID, in.Provider)
	if err != nil {
		return domain.Campaign{}, err
	}
	if in.Name == "" || in.Objective == "" {
		return domain.Campaign{}, domain.ErrInvalidInput
	}
	if in.BudgetMinor > s.policy.MaxDailySpendMinor {
		return domain.Campaign{}, domain.ErrBudgetLimitExceeded
	}

	// Trace d'opération (créée avant l'appel pour réconciliation).
	op, err := s.operations.Create(ctx, domain.ProviderOperation{
		UserID: userID, ConnectionID: conn.ID, Provider: in.Provider,
		OperationType: "campaign.create", Status: domain.OpPending,
	})
	if err != nil {
		return domain.Campaign{}, err
	}
	if err := s.operations.MarkProcessing(ctx, op.ID); err != nil {
		return domain.Campaign{}, err
	}

	campaign, err := p.CreateCampaign(ctx, conn, port.CreateCampaignInput{
		Name: in.Name, Objective: in.Objective, BudgetMinor: in.BudgetMinor, Currency: currencyOf(conn),
	})
	if err != nil {
		_ = s.operations.Fail(ctx, op.ID, "provider_error", err.Error())
		_ = s.connections.SetStatus(ctx, userID, conn.ID, domain.ConnError, err.Error())
		return domain.Campaign{}, err
	}
	campaign.UserID = userID
	campaign.ConnectionID = conn.ID
	saved, err := s.campaigns.Upsert(ctx, campaign)
	if err != nil {
		_ = s.operations.Fail(ctx, op.ID, "persist_error", err.Error())
		return domain.Campaign{}, err
	}
	if err := s.operations.Complete(ctx, op.ID, saved.ExternalCampaignID); err != nil {
		return domain.Campaign{}, err
	}

	// Creative optionnelle (vidéo/image de la bibliothèque d'assets).
	if in.AssetID != "" {
		if err := s.publishCreative(ctx, userID, p, conn, saved.ID, in); err != nil {
			// La campagne existe : l'échec créative n'invalide pas la création.
			slog.Warn("advertising: creative upload failed", "campaign", saved.ID, "err", err)
		}
	}
	return saved, nil
}

// publishCreative résout l'asset (URL signée) et l'envoie au provider.
func (s *Service) publishCreative(ctx context.Context, userID string, p port.AdPlatformProvider, conn domain.AdPlatformConnection, campaignID string, in CreateCampaignInput) error {
	asset, err := s.assets.Get(ctx, userID, in.AssetID)
	if err != nil {
		return fmt.Errorf("asset introuvable: %w", err)
	}
	if s.signer == nil {
		return errors.New("stockage sans URLs signées (S3 requis pour les creatives)")
	}
	assetURL, err := s.signer.SignedURL(ctx, asset.StorageKey, 24*time.Hour)
	if err != nil {
		return err
	}

	cre, err := s.creatives.Create(ctx, domain.Creative{
		UserID: userID, ConnectionID: conn.ID, CampaignID: &campaignID,
		Type: creativeType(asset), AssetID: &asset.ID,
		Headline: in.Headline, PrimaryText: in.PrimaryText, CTA: in.CTA,
		Status: "uploading",
	})
	if err != nil {
		return err
	}

	op, err := s.operations.Create(ctx, domain.ProviderOperation{
		UserID: userID, ConnectionID: conn.ID, Provider: conn.Provider,
		OperationType: "creative.upload", Status: domain.OpPending, InternalResourceID: cre.ID,
	})
	if err != nil {
		return err
	}
	if err := s.operations.MarkProcessing(ctx, op.ID); err != nil {
		return err
	}

	externalID, err := p.UploadCreative(ctx, conn, port.CreativeInput{
		Type: creativeType(asset), URL: assetURL, MimeType: asset.ContentType,
		Headline: in.Headline, PrimaryText: in.PrimaryText, CTA: in.CTA,
	})
	if err != nil {
		_ = s.operations.Fail(ctx, op.ID, "provider_error", err.Error())
		_ = s.creatives.UpdateExternal(ctx, userID, cre.ID, "", "error")
		return err
	}
	if err := s.creatives.UpdateExternal(ctx, userID, cre.ID, externalID, "active"); err != nil {
		return err
	}
	return s.operations.Complete(ctx, op.ID, externalID)
}

// ListCampaigns renvoie les campagnes internes de l'utilisateur.
func (s *Service) ListCampaigns(ctx context.Context, userID string) ([]domain.Campaign, error) {
	return s.campaigns.List(ctx, userID)
}

// SetCampaignStatus pause/reprend une campagne (interne → externe).
func (s *Service) SetCampaignStatus(ctx context.Context, userID, campaignID, status string) (domain.Campaign, error) {
	campaign, err := s.campaigns.Get(ctx, userID, campaignID)
	if err != nil {
		return domain.Campaign{}, err
	}
	if status != domain.CampaignActive && status != domain.CampaignPaused {
		return domain.Campaign{}, domain.ErrInvalidInput
	}
	conn, err := s.activeConnectionByID(ctx, userID, campaign.ConnectionID)
	if err != nil {
		return domain.Campaign{}, err
	}
	p, err := s.provider(conn.Provider)
	if err != nil {
		return domain.Campaign{}, err
	}
	var opErr error
	if status == domain.CampaignPaused {
		opErr = p.PauseCampaign(ctx, conn, campaign.ExternalCampaignID)
	} else {
		opErr = p.ResumeCampaign(ctx, conn, campaign.ExternalCampaignID)
	}
	if opErr != nil {
		return domain.Campaign{}, opErr
	}
	campaign.Status = status
	return s.campaigns.Update(ctx, userID, campaignID, domain.Campaign{Status: status})
}

// ---------- Insights ----------

// GetInsights synchronise puis renvoie les métriques d'une campagne.
func (s *Service) GetInsights(ctx context.Context, userID, campaignID, since, until string) ([]domain.Insight, error) {
	campaign, err := s.campaigns.Get(ctx, userID, campaignID)
	if err != nil {
		return nil, err
	}
	conn, err := s.activeConnectionByID(ctx, userID, campaign.ConnectionID)
	if err != nil {
		return nil, err
	}
	p, err := s.provider(conn.Provider)
	if err != nil {
		return nil, err
	}
	raw, err := p.GetInsights(ctx, conn, campaign.ExternalCampaignID, port.InsightsQuery{Since: since, Until: until})
	if err != nil {
		return nil, err
	}
	for i := range raw {
		raw[i].Currency = currencyOf(conn)
	}
	if err := s.insights.Upsert(ctx, campaign.ID, userID, raw); err != nil {
		return nil, err
	}
	return s.insights.ListByCampaign(ctx, userID, campaign.ID, since, until)
}

// ---------- Internes ----------

// provider résout le provider par clé.
func (s *Service) provider(name string) (port.AdPlatformProvider, error) {
	p, ok := s.providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", domain.ErrInvalidInput, name)
	}
	return p, nil
}

// activeConnection renvoie la connexion active avec un token valide.
func (s *Service) activeConnection(ctx context.Context, userID, provider string) (port.AdPlatformProvider, domain.AdPlatformConnection, error) {
	p, err := s.provider(provider)
	if err != nil {
		return nil, domain.AdPlatformConnection{}, err
	}
	conn, err := s.connections.GetByProvider(ctx, userID, provider)
	if err != nil {
		return nil, domain.AdPlatformConnection{}, err
	}
	conn, err = s.ensureValidToken(ctx, p, userID, conn)
	if err != nil {
		return nil, domain.AdPlatformConnection{}, err
	}
	if conn.Status != domain.ConnActive {
		return nil, domain.AdPlatformConnection{}, domain.ErrConnectionNotActive
	}
	return p, conn, nil
}

func (s *Service) activeConnectionByID(ctx context.Context, userID, connectionID string) (domain.AdPlatformConnection, error) {
	conn, err := s.connections.Get(ctx, userID, connectionID)
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	p, err := s.provider(conn.Provider)
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	return s.ensureValidToken(ctx, p, userID, conn)
}

// ensureValidToken rafraîchit le token si proche de l'expiration ; en cas
// d'échec la connexion passe en expired (reconnexion requise).
func (s *Service) ensureValidToken(ctx context.Context, p port.AdPlatformProvider, userID string, conn domain.AdPlatformConnection) (domain.AdPlatformConnection, error) {
	const refreshWindow = 24 * time.Hour
	if conn.AccessToken == "" {
		return conn, domain.ErrConnectionNotActive
	}
	if conn.AccessTokenExpiresAt == nil || time.Until(*conn.AccessTokenExpiresAt) > refreshWindow {
		return conn, nil
	}
	res, supported, err := p.Refresh(ctx, conn)
	if err != nil || !supported || res.AccessToken == "" {
		_ = s.connections.SetStatus(ctx, userID, conn.ID, domain.ConnExpired, "token expiré, reconnexion requise")
		return conn, domain.ErrConnectionNotActive
	}
	accessEnc, err := s.encryptor.Encrypt(res.AccessToken)
	if err != nil {
		return conn, err
	}
	refreshEnc := conn.RefreshToken
	if res.RefreshToken != "" {
		if refreshEnc, err = s.encryptor.Encrypt(res.RefreshToken); err != nil {
			return conn, err
		}
	} else {
		refreshEnc, err = s.encryptor.Encrypt(conn.RefreshToken)
		if err != nil {
			return conn, err
		}
	}
	if err := s.connections.UpdateTokens(ctx, conn.ID, accessEnc, refreshEnc, res.ExpiresAt, conn.ExternalUserID, conn.Scopes, conn.Status); err != nil {
		return conn, err
	}
	return s.connections.Get(ctx, userID, conn.ID)
}

func currencyOf(conn domain.AdPlatformConnection) string {
	var md map[string]any
	if json.Unmarshal(conn.Metadata, &md) == nil {
		if c, ok := md["currency"].(string); ok && c != "" {
			return c
		}
	}
	return "XOF"
}

func creativeType(asset domain.Asset) string {
	if strings.HasPrefix(asset.ContentType, "video/") {
		return domain.CreativeVideo
	}
	if strings.HasPrefix(asset.ContentType, "image/") {
		return domain.CreativeImage
	}
	return domain.CreativeText
}

// randomState génère un état CSRF cryptographiquement sûr.
func randomState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) // source système défaillante : impossible à traiter proprement
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
