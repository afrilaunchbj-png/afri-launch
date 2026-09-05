package advertising_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
	"time"

	advapp "afrilaunch/backend/internal/application/advertising"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/crypto"
)

func timeNow() time.Time { return time.Now() }

// fakeProvider implémente port.AdPlatformProvider en mémoire.
type fakeProvider struct {
	accounts []domain.AdAccount
	campaign []domain.Campaign
	created  []port.CreateCampaignInput
}

func (f *fakeProvider) Provider() string { return "fake" }
func (f *fakeProvider) Capabilities() port.AdPlatformCapabilities {
	return port.AdPlatformCapabilities{Campaigns: true, Creatives: true}
}
func (f *fakeProvider) AuthorizationURL(state, redirectURI string) string {
	return "https://provider.test/oauth?state=" + url.QueryEscape(state)
}
func (f *fakeProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (port.OAuthTokenResult, error) {
	if code == "bad" {
		return port.OAuthTokenResult{}, errors.New("invalid code")
	}
	return port.OAuthTokenResult{AccessToken: "tok-" + code, ExternalUserID: "ext-1"}, nil
}
func (f *fakeProvider) Refresh(ctx context.Context, conn domain.AdPlatformConnection) (port.OAuthTokenResult, bool, error) {
	return port.OAuthTokenResult{AccessToken: "tok-refreshed"}, true, nil
}
func (f *fakeProvider) Revoke(ctx context.Context, accessToken string) error { return nil }
func (f *fakeProvider) ListAdAccounts(ctx context.Context, accessToken string) ([]domain.AdAccount, error) {
	return f.accounts, nil
}
func (f *fakeProvider) VerifyAdAccount(ctx context.Context, accessToken, externalAccountID string) (domain.AdAccount, error) {
	for _, a := range f.accounts {
		if a.ExternalID == externalAccountID {
			return a, nil
		}
	}
	return domain.AdAccount{}, domain.ErrAccountNotAccessible
}
func (f *fakeProvider) ListCampaigns(ctx context.Context, conn domain.AdPlatformConnection, q port.CampaignQuery) ([]domain.Campaign, error) {
	return f.campaign, nil
}
func (f *fakeProvider) CreateCampaign(ctx context.Context, conn domain.AdPlatformConnection, in port.CreateCampaignInput) (domain.Campaign, error) {
	f.created = append(f.created, in)
	return domain.Campaign{
		UserID: conn.UserID, ConnectionID: conn.ID,
		ExternalCampaignID: "ext-" + in.Name, Name: in.Name,
		Objective: in.Objective, Status: domain.CampaignPaused,
		BudgetMinor: in.BudgetMinor, Currency: in.Currency,
	}, nil
}
func (f *fakeProvider) UpdateCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string, in port.UpdateCampaignInput) (domain.Campaign, error) {
	return domain.Campaign{ExternalCampaignID: externalID}, nil
}
func (f *fakeProvider) PauseCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error {
	return nil
}
func (f *fakeProvider) ResumeCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error {
	return nil
}
func (f *fakeProvider) UploadCreative(ctx context.Context, conn domain.AdPlatformConnection, in port.CreativeInput) (string, error) {
	return "cr-1", nil
}
func (f *fakeProvider) GetInsights(ctx context.Context, conn domain.AdPlatformConnection, externalCampaignID string, q port.InsightsQuery) ([]domain.Insight, error) {
	return nil, nil
}

// fakeStore : OAuth states en mémoire.
type fakeStore struct {
	states map[string]domain.OAuthState
}

func (s *fakeStore) Create(ctx context.Context, st domain.OAuthState) error {
	if s.states == nil {
		s.states = map[string]domain.OAuthState{}
	}
	s.states[st.State] = st
	return nil
}
func (s *fakeStore) Consume(ctx context.Context, state, provider string) (domain.OAuthState, error) {
	st, ok := s.states[state]
	if !ok || st.Provider != provider || st.UsedAt != nil || st.ExpiresAt.Before(timeNow()) {
		return domain.OAuthState{}, domain.ErrOAuthStateInvalid
	}
	now := timeNow()
	st.UsedAt = &now
	s.states[state] = st
	return st, nil
}
func (s *fakeStore) Prune(ctx context.Context) error { return nil }

// fakeConnections : connexions en mémoire (tokens en clair, comme chiffrés).
type fakeConnections struct {
	conns map[string]*domain.AdPlatformConnection
	seq   int
}

func (r *fakeConnections) Create(ctx context.Context, c domain.AdPlatformConnection) (domain.AdPlatformConnection, error) {
	r.seq++
	c.ID = string(rune('a' + r.seq))
	if r.conns == nil {
		r.conns = map[string]*domain.AdPlatformConnection{}
	}
	r.conns[c.ID] = &c
	return c, nil
}
func (r *fakeConnections) UpdateTokens(ctx context.Context, id string, accessEnc, refreshEnc string, expiresAt *time.Time, externalUserID string, scopes []string, status string) error {
	c := r.conns[id]
	c.AccessToken, c.RefreshToken, c.AccessTokenExpiresAt = accessEnc, refreshEnc, expiresAt
	c.ExternalUserID, c.Scopes, c.Status = externalUserID, scopes, status
	return nil
}
func (r *fakeConnections) SelectAccount(ctx context.Context, userID, id, externalAccountID, accountName string, metadata []byte) error {
	c := r.conns[id]
	if c.UserID != userID {
		return domain.ErrConnectionNotFound
	}
	c.ExternalAccountID, c.ExternalAccountName, c.Status = externalAccountID, accountName, domain.ConnActive
	c.Metadata = metadata
	return nil
}
func (r *fakeConnections) SetStatus(ctx context.Context, userID, id, status, lastError string) error {
	r.conns[id].Status, r.conns[id].LastError = status, lastError
	return nil
}
func (r *fakeConnections) SetSynced(ctx context.Context, id string, at time.Time) error { return nil }
func (r *fakeConnections) Get(ctx context.Context, userID, id string) (domain.AdPlatformConnection, error) {
	c, ok := r.conns[id]
	if !ok || c.UserID != userID {
		return domain.AdPlatformConnection{}, domain.ErrConnectionNotFound
	}
	return *c, nil
}
func (r *fakeConnections) GetByProvider(ctx context.Context, userID, provider string) (domain.AdPlatformConnection, error) {
	for _, c := range r.conns {
		if c.UserID == userID && c.Provider == provider {
			return *c, nil
		}
	}
	return domain.AdPlatformConnection{}, domain.ErrConnectionNotFound
}
func (r *fakeConnections) List(ctx context.Context, userID string) ([]domain.AdPlatformConnection, error) {
	return nil, nil
}
func (r *fakeConnections) ListByProvider(ctx context.Context, provider string) ([]domain.AdPlatformConnection, error) {
	return nil, nil
}

// Repos factices minimaux.
type fakeCampaigns struct{ items []domain.Campaign }

func (r *fakeCampaigns) Upsert(ctx context.Context, c domain.Campaign) (domain.Campaign, error) {
	for i := range r.items {
		if r.items[i].ConnectionID == c.ConnectionID && r.items[i].ExternalCampaignID == c.ExternalCampaignID {
			r.items[i].Name, r.items[i].Status = c.Name, c.Status
			return r.items[i], nil
		}
	}
	r.items = append(r.items, c)
	return c, nil
}
func (r *fakeCampaigns) Update(ctx context.Context, userID, id string, c domain.Campaign) (domain.Campaign, error) {
	for i := range r.items {
		if r.items[i].ID == id && r.items[i].UserID == userID {
			r.items[i].Status = c.Status
			return r.items[i], nil
		}
	}
	return domain.Campaign{}, domain.ErrNotFound
}
func (r *fakeCampaigns) Get(ctx context.Context, userID, id string) (domain.Campaign, error) {
	for _, c := range r.items {
		if c.ID == id && c.UserID == userID {
			return c, nil
		}
	}
	return domain.Campaign{}, domain.ErrNotFound
}
func (r *fakeCampaigns) List(ctx context.Context, userID string) ([]domain.Campaign, error) {
	return r.items, nil
}
func (r *fakeCampaigns) ListByConnection(ctx context.Context, connectionID string) ([]domain.Campaign, error) {
	return r.items, nil
}

type noopRepo struct{}

func (noopRepo) Create(ctx context.Context, c domain.Creative) (domain.Creative, error) {
	return c, nil
}
func (noopRepo) UpdateExternal(ctx context.Context, userID, id, externalCreativeID, status string) error {
	return nil
}
func (noopRepo) Get(ctx context.Context, userID, id string) (domain.Creative, error) {
	return domain.Creative{}, domain.ErrNotFound
}
func (noopRepo) List(ctx context.Context, userID string) ([]domain.Creative, error) {
	return nil, nil
}

type noopInsights struct{}

func (noopInsights) Upsert(ctx context.Context, campaignID, userID string, insights []domain.Insight) error {
	return nil
}
func (noopInsights) ListByCampaign(ctx context.Context, userID, campaignID string, since, until string) ([]domain.Insight, error) {
	return nil, nil
}

type noopOps struct{}

func (noopOps) Create(ctx context.Context, op domain.ProviderOperation) (domain.ProviderOperation, error) {
	return op, nil
}
func (noopOps) MarkProcessing(ctx context.Context, id string) error       { return nil }
func (noopOps) Complete(ctx context.Context, id, externalID string) error { return nil }
func (noopOps) Fail(ctx context.Context, id, code, msg string) error      { return nil }
func (noopOps) IncrementAttempts(ctx context.Context, id string) error    { return nil }
func (noopOps) Get(ctx context.Context, userID, id string) (domain.ProviderOperation, error) {
	return domain.ProviderOperation{}, domain.ErrNotFound
}
func (noopOps) List(ctx context.Context, userID string, limit int) ([]domain.ProviderOperation, error) {
	return nil, nil
}

func newTestService(t *testing.T) (*advapp.Service, *fakeProvider, *fakeStore, *fakeConnections, *fakeCampaigns) {
	t.Helper()
	enc, err := crypto.NewEncryptor("test-encryption-key-32-caracteres-xx!", "v1")
	if err != nil {
		t.Fatal(err)
	}
	provider := &fakeProvider{accounts: []domain.AdAccount{{ExternalID: "act_1", Name: "Biz", Currency: "XOF"}}}
	states := &fakeStore{}
	conns := &fakeConnections{}
	campaigns := &fakeCampaigns{}
	svc := advapp.NewService(
		advapp.ProviderRegistry{"fake": provider},
		enc, states, conns,
		campaigns, noopRepo{}, noopInsights{}, noopOps{},
		nil, nil, nil,
		advapp.SafetyPolicy{MaxDailySpendMinor: 100_000, MaxCampaigns: 5},
	)
	return svc, provider, states, conns, campaigns
}

func TestStartConnectAndCallback(t *testing.T) {
	svc, _, states, conns, _ := newTestService(t)
	ctx := context.Background()

	authURL, err := svc.StartConnect(ctx, "user-1", "fake", "https://app.test/callback")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(authURL)
	state := u.Query().Get("state")
	if state == "" {
		t.Fatal("état CSRF manquant dans l'URL")
	}
	if _, ok := states.states[state]; !ok {
		t.Fatal("état non persisté")
	}

	// Callback avec état invalide → refus (anti-CSRF).
	if _, err := svc.HandleCallback(ctx, "fake", "code-1", "mauvais-etat", "https://app.test/callback"); !errors.Is(err, domain.ErrOAuthStateInvalid) {
		t.Fatalf("état invalide accepté: %v", err)
	}

	conn, err := svc.HandleCallback(ctx, "fake", "code-1", state, "https://app.test/callback")
	if err != nil {
		t.Fatal(err)
	}
	if conn.UserID != "user-1" || conn.Status != domain.ConnPending {
		t.Fatalf("conn = %+v", conn)
	}

	// Replay : le même état ne passe plus.
	if _, err := svc.HandleCallback(ctx, "fake", "code-2", state, "https://app.test/callback"); !errors.Is(err, domain.ErrOAuthStateInvalid) {
		t.Fatal("replay accepté")
	}

	// Le token échangé est bien présent (chiffré au repos côté vrai repo).
	stored, _ := conns.Get(ctx, "user-1", conn.ID)
	if stored.AccessToken == "" {
		t.Fatal("token non persisté")
	}
}

func TestSelectAccountVerifiesOwnership(t *testing.T) {
	svc, _, _, conns, _ := newTestService(t)
	ctx := context.Background()

	// Connexion pré-existante avec token (comme après callback).
	created, _ := conns.Create(ctx, domain.AdPlatformConnection{UserID: "user-1", Provider: "fake", Status: domain.ConnPending})
	_ = conns.UpdateTokens(ctx, created.ID, "tok-x", "", nil, "ext-1", nil, domain.ConnPending)

	// Compte non accessible → refus.
	if _, err := svc.SelectAccount(ctx, "user-1", "fake", "act_inconnu"); !errors.Is(err, domain.ErrAccountNotAccessible) {
		t.Fatalf("compte inconnu accepté: %v", err)
	}
	// Compte accessible → active + currency en metadata.
	conn, err := svc.SelectAccount(ctx, "user-1", "fake", "act_1")
	if err != nil {
		t.Fatal(err)
	}
	if conn.Status != domain.ConnActive || conn.ExternalAccountID != "act_1" {
		t.Fatalf("conn = %+v", conn)
	}
	var md map[string]string
	_ = json.Unmarshal(conn.Metadata, &md)
	if md["currency"] != "XOF" {
		t.Fatalf("metadata = %v", md)
	}
}

func TestCreateCampaignBudgetSafety(t *testing.T) {
	svc, provider, _, conns, campaigns := newTestService(t)
	ctx := context.Background()
	created, _ := conns.Create(ctx, domain.AdPlatformConnection{UserID: "user-1", Provider: "fake", Status: domain.ConnActive, ExternalAccountID: "act_1", AccessToken: "tok"})
	_ = created

	// Budget au-dessus de la politique → refus AVANT appel provider.
	_, err := svc.CreateCampaign(ctx, "user-1", advapp.CreateCampaignInput{
		Provider: "fake", Name: "Trop cher", Objective: "OUTCOME_TRAFFIC", BudgetMinor: 500_000,
	})
	if !errors.Is(err, domain.ErrBudgetLimitExceeded) {
		t.Fatalf("garde-fou budget non déclenché: %v", err)
	}
	if len(provider.created) != 0 {
		t.Fatal("le provider a été appelé malgré le garde-fou")
	}

	// Campagne valide → créée en pause.
	c, err := svc.CreateCampaign(ctx, "user-1", advapp.CreateCampaignInput{
		Provider: "fake", Name: "Pub guide", Objective: "OUTCOME_TRAFFIC", BudgetMinor: 50_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != domain.CampaignPaused || len(campaigns.items) != 1 {
		t.Fatalf("campaign = %+v", c)
	}
}
