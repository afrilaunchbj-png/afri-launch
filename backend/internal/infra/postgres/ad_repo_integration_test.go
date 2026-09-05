package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/crypto"
	"afrilaunch/backend/internal/infra/postgres"
)

// testEncryptor construit un chiffreur de test.
func testEncryptor(t *testing.T) *crypto.Encryptor {
	t.Helper()
	enc, err := crypto.NewEncryptor("test-encryption-key-32-caracteres-xx!", "v1")
	if err != nil {
		t.Fatal(err)
	}
	return enc
}

// TestAdConnectionsIsolation valide la persistance des connexions
// publicitaires (tokens chiffrés) et l'isolation par utilisateur
// contre une vraie base PostgreSQL.
// Activé uniquement lorsque AFRILAUNCH_TEST_DB est défini.
func TestAdConnectionsIsolation(t *testing.T) {
	url := os.Getenv("AFRILAUNCH_TEST_DB")
	if url == "" {
		t.Skip("AFRILAUNCH_TEST_DB non défini — test d'intégration ignoré")
	}

	ctx := context.Background()
	pool, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	users := postgres.NewUserRepository(store)
	conns := postgres.NewAdConnectionRepository(store, testEncryptor(t))
	states := postgres.NewOAuthStateStore(store)

	// Deux utilisateurs isolés.
	userA, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: "adv-a-" + uuid.NewString() + "@test.local", FullName: "User A"})
	if err != nil {
		t.Fatalf("create userA: %v", err)
	}
	userB, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: "adv-b-" + uuid.NewString() + "@test.local", FullName: "User B"})
	if err != nil {
		t.Fatalf("create userB: %v", err)
	}

	// Connexion Meta de A.
	created, err := conns.Create(ctx, domain.AdPlatformConnection{
		UserID:   userA.ID,
		Provider: domain.AdPlatformMeta,
		Status:   domain.ConnPending,
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}

	// Tokens chiffrés au repos (le repository reçoit du chiffré, rend du clair).
	enc := testEncryptor(t)
	encAccess, err := enc.Encrypt("secret-access-token")
	if err != nil {
		t.Fatal(err)
	}
	encRefresh, err := enc.Encrypt("secret-refresh-token")
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(60 * time.Hour).UTC()
	if err := conns.UpdateTokens(ctx, created.ID, encAccess, encRefresh, &expires, "ext-user-1", []string{"ads_management"}, domain.ConnPending); err != nil {
		t.Fatalf("update tokens: %v", err)
	}
	got, err := conns.Get(ctx, userA.ID, created.ID)
	if err != nil {
		t.Fatalf("get connection: %v", err)
	}
	if got.AccessToken != "secret-access-token" || got.ExternalUserID != "ext-user-1" {
		t.Fatalf("tokens non persistés: %+v", got)
	}

	// Isolation : B ne peut pas lire la connexion de A.
	if _, err := conns.Get(ctx, userB.ID, created.ID); err != domain.ErrConnectionNotFound {
		t.Fatalf("isolation rompue: err = %v", err)
	}
	if err := conns.SelectAccount(ctx, userB.ID, created.ID, "act_x", "Robbed", nil); err == nil {
		t.Fatal("B a sélectionné un compte sur la connexion de A")
	}

	// Sélection de compte par le propriétaire → status active.
	if err := conns.SelectAccount(ctx, userA.ID, created.ID, "act_123", "Business Test", []byte(`{"currency":"XOF"}`)); err != nil {
		t.Fatalf("select account: %v", err)
	}
	got, _ = conns.Get(ctx, userA.ID, created.ID)
	if got.Status != domain.ConnActive || got.ExternalAccountID != "act_123" {
		t.Fatalf("sélection compte: %+v", got)
	}

	// Statuts : déconnexion conserve la ligne (historique).
	if err := conns.SetStatus(ctx, userA.ID, created.ID, domain.ConnDisconnected, ""); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	got, _ = conns.Get(ctx, userA.ID, created.ID)
	if got.Status != domain.ConnDisconnected {
		t.Fatalf("status après disconnect = %q", got.Status)
	}

	// OAuth states : création, consommation unique.
	state := uuid.NewString()
	if err := states.Create(ctx, domain.OAuthState{State: state, UserID: userA.ID, Provider: domain.AdPlatformMeta, ExpiresAt: time.Now().Add(10 * time.Minute)}); err != nil {
		t.Fatalf("create state: %v", err)
	}
	consumed, err := states.Consume(ctx, state, domain.AdPlatformMeta)
	if err != nil {
		t.Fatalf("consume state: %v", err)
	}
	if consumed.UserID != userA.ID {
		t.Fatalf("state consommé sans user lié: %+v", consumed)
	}
	// Anti-replay + anti-CSRF.
	if _, err := states.Consume(ctx, state, domain.AdPlatformMeta); err != domain.ErrOAuthStateInvalid {
		t.Fatalf("replay accepté: %v", err)
	}
	if _, err := states.Consume(ctx, state, domain.AdPlatformGoogleAds); err != domain.ErrOAuthStateInvalid {
		t.Fatal("state consommé pour un autre provider")
	}
}

// TestAdCampaignsAndOps valide campagnes, creatives, insights et opérations
// provider (mapping interne/externe) contre une vraie base PostgreSQL.
func TestAdCampaignsAndOps(t *testing.T) {
	url := os.Getenv("AFRILAUNCH_TEST_DB")
	if url == "" {
		t.Skip("AFRILAUNCH_TEST_DB non défini — test d'intégration ignoré")
	}

	ctx := context.Background()
	pool, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	users := postgres.NewUserRepository(store)
	conns := postgres.NewAdConnectionRepository(store, testEncryptor(t))
	campaigns := postgres.NewAdCampaignRepository(store)
	creatives := postgres.NewAdCreativeRepository(store)
	insights := postgres.NewAdInsightRepository(store)
	ops := postgres.NewProviderOperationRepository(store)

	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: "adv-c-" + uuid.NewString() + "@test.local", FullName: "User C"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	conn, err := conns.Create(ctx, domain.AdPlatformConnection{
		UserID:   user.ID,
		Provider: domain.AdPlatformMeta,
		Status:   domain.ConnActive,
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}

	// Campagne : upsert idempotent sur (connection, external_id).
	created, err := campaigns.Upsert(ctx, domain.Campaign{
		UserID: user.ID, ConnectionID: conn.ID,
		ExternalCampaignID: "camp_ext_1", Name: "Pub guide", Objective: "OUTCOME_TRAFFIC",
		Status: domain.CampaignActive, BudgetMinor: 500000, Currency: "XOF",
	})
	if err != nil {
		t.Fatalf("upsert campaign: %v", err)
	}
	same, err := campaigns.Upsert(ctx, domain.Campaign{
		UserID: user.ID, ConnectionID: conn.ID,
		ExternalCampaignID: "camp_ext_1", Name: "Pub guide v2", Status: domain.CampaignPaused,
	})
	if err != nil {
		t.Fatalf("upsert 2: %v", err)
	}
	if same.ID != created.ID {
		t.Fatal("upsert a créé une deuxième campagne")
	}
	if same.Name != "Pub guide v2" || same.Status != domain.CampaignPaused {
		t.Fatalf("upsert n'a pas mis à jour: %+v", same)
	}

	// Creative liée à l'asset vidéo + mapping externe.
	cre, err := creatives.Create(ctx, domain.Creative{
		UserID: user.ID, ConnectionID: conn.ID, CampaignID: &created.ID,
		Type: domain.CreativeVideo, Headline: "Lance ton business", PrimaryText: "Guide complet", CTA: "SHOP_NOW",
		Status: "uploading",
	})
	if err != nil {
		t.Fatalf("create creative: %v", err)
	}
	if err := creatives.UpdateExternal(ctx, user.ID, cre.ID, "cr_ext_9", "active"); err != nil {
		t.Fatalf("update external: %v", err)
	}
	gotCre, _ := creatives.Get(ctx, user.ID, cre.ID)
	if gotCre.ExternalCreativeID != "cr_ext_9" || gotCre.Status != "active" {
		t.Fatalf("creative non mise à jour: %+v", gotCre)
	}

	// Insights : upsert par (campagne, jour).
	day := time.Now().UTC().Format("2006-01-02")
	if err := insights.Upsert(ctx, created.ID, user.ID, []domain.Insight{{
		Date: day, Impressions: 1000, Reach: 800, Clicks: 42, SpendMinor: 12345, Conversions: 3, Currency: "XOF",
	}}); err != nil {
		t.Fatalf("upsert insights: %v", err)
	}
	if err := insights.Upsert(ctx, created.ID, user.ID, []domain.Insight{{
		Date: day, Impressions: 1100, Reach: 850, Clicks: 45, SpendMinor: 13000, Conversions: 4, Currency: "XOF",
	}}); err != nil {
		t.Fatalf("upsert insights 2: %v", err)
	}
	list, err := insights.ListByCampaign(ctx, user.ID, created.ID, day, day)
	if err != nil || len(list) != 1 {
		t.Fatalf("insights: %v, len = %d", err, len(list))
	}
	if list[0].Impressions != 1100 || list[0].Clicks != 45 {
		t.Fatalf("insight non mis à jour: %+v", list[0])
	}

	// Opérations provider : cycle complet.
	op, err := ops.Create(ctx, domain.ProviderOperation{
		UserID: user.ID, ConnectionID: conn.ID, Provider: domain.AdPlatformMeta,
		OperationType: "campaign.create", Status: domain.OpPending, InternalResourceID: created.ID,
	})
	if err != nil {
		t.Fatalf("create op: %v", err)
	}
	if err := ops.MarkProcessing(ctx, op.ID); err != nil {
		t.Fatalf("processing: %v", err)
	}
	if err := ops.Complete(ctx, op.ID, "camp_ext_1"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	gotOp, _ := ops.Get(ctx, user.ID, op.ID)
	if gotOp.Status != domain.OpCompleted || gotOp.ExternalResourceID != "camp_ext_1" || gotOp.Attempts != 1 {
		t.Fatalf("opération: %+v", gotOp)
	}
}
