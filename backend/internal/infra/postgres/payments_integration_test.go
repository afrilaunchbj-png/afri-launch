package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	creditsapp "afrilaunch/backend/internal/application/credits"
	"afrilaunch/backend/internal/application/payments"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres"
)

// fakePaymentProvider implémente port.PaymentProvider pour tester le service.
type fakePaymentProvider struct {
	remoteStatus string
	lastRef      string // référence du dernier checkout créé
}

func (f *fakePaymentProvider) Provider() string { return "fakepay" }

func (f *fakePaymentProvider) CreateCheckout(ctx context.Context, in port.PaymentCheckoutInput) (port.PaymentCheckoutResult, error) {
	f.lastRef = "ref-" + in.PaymentID
	return port.PaymentCheckoutResult{RedirectURL: "https://pay.test/" + in.PaymentID, ProviderReference: f.lastRef}, nil
}

func (f *fakePaymentProvider) VerifyStatus(ctx context.Context, providerReference string) (port.PaymentStatusResult, error) {
	return port.PaymentStatusResult{Status: f.remoteStatus}, nil
}

func (f *fakePaymentProvider) HandleWebhook(ctx context.Context, in port.PaymentWebhookInput) (port.PaymentWebhookResult, error) {
	return port.PaymentWebhookResult{ProviderReference: f.lastRef, Accepted: true}, nil
}

// TestPaymentsCheckoutFlow valide le cycle complet contre une vraie base :
// checkout → webhook (statut reconfirmé) → crédits accordés une seule fois.
func TestPaymentsCheckoutFlow(t *testing.T) {
	url := os.Getenv("AFRILAUNCH_TEST_DB")
	if url == "" {
		t.Skip("AFRILAUNCH_TEST_DB non défini — test d'intégration ignoré")
	}

	ctx := context.Background()
	pool, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	users := postgres.NewUserRepository(store)
	creditRepo := postgres.NewCreditRepository(store)
	planRepo := postgres.NewPlanRepository(store)
	paymentRepo := postgres.NewPaymentRepository(store)
	creditSvc := creditsapp.NewService(creditRepo)

	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: "pay-" + uuid.NewString() + "@test.local", FullName: "Payer"})
	if err != nil {
		t.Fatal(err)
	}

	planID := "22222222-2222-4222-8222-222222222222" // Pack Business (seed 00019)
	provider := &fakePaymentProvider{remoteStatus: domain.PaymentPending}
	svc := payments.NewService(
		payments.ProviderRegistry{"fakepay": provider},
		planRepo, paymentRepo, creditRepo, nil,
		"AfriLaunch", "https://app.test/credits", "",
	)

	// 1. Checkout : paiement pending + URL.
	payment, redirect, err := svc.Checkout(ctx, user.ID, planID)
	if err != nil {
		t.Fatal(err)
	}
	if payment.Status != domain.PaymentPending || redirect == "" || payment.ProviderReference != "ref-"+payment.ID {
		t.Fatalf("checkout = %+v, redirect %q", payment, redirect)
	}

	// 2. Le provider confirme (statut distant = succeeded) → webhook.
	provider.remoteStatus = domain.PaymentSucceeded
	final, err := svc.HandleWebhook(ctx, "fakepay", port.PaymentWebhookInput{
		Body: []byte(`{"reference":"ref-pay-1"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if final.Status != domain.PaymentSucceeded {
		t.Fatalf("status final = %q", final.Status)
	}

	// 3. Crédits accordés (Pack Business = 120).
	summary, err := creditSvc.Summary(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Balance != 120 {
		t.Fatalf("balance = %d, want 120", summary.Balance)
	}

	// 4. Idempotence : un second webhook ne re-crédite pas.
	if _, err := svc.HandleWebhook(ctx, "fakepay", port.PaymentWebhookInput{
		Body: []byte(`{"reference":"ref-pay-1"}`),
	}); err != nil {
		t.Fatal(err)
	}
	summary, _ = creditSvc.Summary(ctx, user.ID)
	if summary.Balance != 120 {
		t.Fatalf("double octroi ! balance = %d", summary.Balance)
	}

	// 5. Sync depuis le FE : paiement déjà final → inchangé.
	if _, err := svc.SyncStatus(ctx, user.ID, payment.ID); err != nil {
		t.Fatal(err)
	}
	summary, _ = creditSvc.Summary(ctx, user.ID)
	if summary.Balance != 120 {
		t.Fatalf("balance modifiée par sync : %d", summary.Balance)
	}

	// 6. Provider inconnu → erreur explicite, pas de panique.
	if _, err := svc.HandleWebhook(ctx, "inconnu", port.PaymentWebhookInput{Body: []byte(`{}`)}); err == nil {
		t.Fatal("webhook pour un provider inconnu accepté")
	}
}
