// Package payments orchestre l'achat de crédits (recharges) via un
// provider de paiement choisi par configuration (ADR-008/ADR-018).
// Le code métier ne dépend jamais de PawaPay/FedaPay/PayDunya directement.
package payments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"

	"afrilaunch/backend/internal/application/audit"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// ProviderRegistry résout un provider par clé (pawapay, fedapay, paydunya).
type ProviderRegistry map[string]port.PaymentProvider

// Service orchestre les paiements.
type Service struct {
	providers  ProviderRegistry
	plans      port.PlanRepository
	payments   port.PaymentRepository
	credits    port.CreditRepository
	audit      *audit.Recorder
	storeName  string
	returnURL  string // base FE (ex. https://app.afrilaunch.com/credits)
	webhookURL string // callback serveur-à-serveur (optionnel)
}

// NewService construit le service de paiements.
func NewService(
	providers ProviderRegistry,
	plans port.PlanRepository,
	payments port.PaymentRepository,
	credits port.CreditRepository,
	auditRec *audit.Recorder,
	storeName, returnURL, webhookURL string,
) *Service {
	return &Service{
		providers: providers, plans: plans, payments: payments,
		credits: credits, audit: auditRec,
		storeName: storeName, returnURL: returnURL, webhookURL: webhookURL,
	}
}

// Enabled indique si un provider est configuré (affichage FE).
func (s *Service) Enabled() bool { return len(s.providers) > 0 }

// Capabilities : nom du provider actif ("" si aucun).
func (s *Service) ActiveProvider() string {
	for name := range s.providers {
		return name
	}
	return ""
}

// ListPlans renvoie les packs de crédits actifs.
func (s *Service) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	return s.plans.ListActive(ctx)
}

// GetPayment renvoie un paiement de l'utilisateur (suivi après redirection).
func (s *Service) GetPayment(ctx context.Context, userID, id string) (domain.Payment, error) {
	return s.payments.Get(ctx, userID, id)
}

// ListMyPayments renvoie l'historique des paiements de l'utilisateur.
func (s *Service) ListMyPayments(ctx context.Context, userID string) ([]domain.Payment, error) {
	return s.payments.ListByUser(ctx, userID, 20)
}

// CheckoutInput porte la demande d'achat.
type CheckoutInput struct {
	PlanID string `json:"plan_id"`
}

// Checkout crée le paiement (pending), initie le checkout chez le provider
// et renvoie l'URL de redirection. Les crédits ne sont accordés qu'à la
// confirmation du provider (webhook vérifié + statut reconfirmé par API).
func (s *Service) Checkout(ctx context.Context, userID string, planID string) (domain.Payment, string, error) {
	provider, ok := s.activeProvider()
	if !ok {
		return domain.Payment{}, "", domain.ErrPaymentProviderDisabled
	}
	plan, err := s.plans.Get(ctx, planID)
	if err != nil {
		return domain.Payment{}, "", err
	}
	if !plan.IsActive {
		return domain.Payment{}, "", domain.ErrInvalidInput
	}

	idempotencyKey := "checkout:" + randomHex()
	payment, err := s.payments.Create(ctx, domain.Payment{
		UserID:         userID,
		PlanID:         plan.ID,
		AmountMinor:    plan.PriceMinor,
		Currency:       plan.Currency,
		Provider:       provider.Provider(),
		Status:         domain.PaymentPending,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return domain.Payment{}, "", err
	}

	result, err := provider.CreateCheckout(ctx, port.PaymentCheckoutInput{
		PaymentID:   payment.ID,
		AmountMinor: payment.AmountMinor,
		Currency:    payment.Currency,
		Description: fmt.Sprintf("%s — %d crédits", plan.Name, plan.Credits),
		ReturnURL:   s.returnURL + "?payment=" + payment.ID,
		WebhookURL:  s.webhookURL,
	})
	if err != nil {
		_, _ = s.payments.MarkStatus(ctx, payment.ID, domain.PaymentFailed)
		if s.audit != nil {
			s.audit.Log(ctx, userID, domain.AuditPaymentFailed, "payment", payment.ID,
				map[string]any{"stage": "checkout", "error": err.Error()})
		}
		return domain.Payment{}, "", err
	}
	if err := s.payments.UpdateCheckout(ctx, payment.ID, result.ProviderReference, result.RedirectURL); err != nil {
		return domain.Payment{}, "", err
	}
	if s.audit != nil {
		s.audit.Log(ctx, userID, domain.AuditPaymentCheckout, "payment", payment.ID,
			map[string]any{"provider": payment.Provider, "plan_id": plan.ID, "amount_minor": payment.AmountMinor})
	}
	payment.CheckoutURL = result.RedirectURL
	payment.ProviderReference = result.ProviderReference
	return payment, result.RedirectURL, nil
}

// HandleWebhook traite un webhook entrant : extraction de la référence,
// reconfirmation du statut par API (le corps du webhook n'est jamais une
// preuve), transition de statut et octroi des crédits (idempotent par
// référence "payment:<id>" dans le ledger).
func (s *Service) HandleWebhook(ctx context.Context, providerName string, in port.PaymentWebhookInput) (domain.Payment, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return domain.Payment{}, domain.ErrPaymentProviderDisabled
	}
	result, err := provider.HandleWebhook(ctx, in)
	if err != nil {
		return domain.Payment{}, err
	}
	if !result.Accepted {
		return domain.Payment{}, nil // webhook non final : rien à faire
	}
	payment, err := s.payments.GetByProviderReference(ctx, result.ProviderReference)
	if err != nil {
		return domain.Payment{}, err
	}

	status, err := provider.VerifyStatus(ctx, result.ProviderReference)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("%w: %v", domain.ErrPaymentInvalidState, err)
	}
	return s.applyStatus(ctx, payment, status.Status)
}

// SyncStatus permet au FE de déclencher la reconfirmation (retour navigateur
// avant réception du webhook).
func (s *Service) SyncStatus(ctx context.Context, userID, paymentID string) (domain.Payment, error) {
	payment, err := s.payments.Get(ctx, userID, paymentID)
	if err != nil {
		return domain.Payment{}, err
	}
	if payment.Status != domain.PaymentPending {
		return payment, nil
	}
	provider, ok := s.providers[payment.Provider]
	if !ok {
		return domain.Payment{}, domain.ErrPaymentProviderDisabled
	}
	if payment.ProviderReference == "" {
		return payment, nil
	}
	status, err := provider.VerifyStatus(ctx, payment.ProviderReference)
	if err != nil {
		return payment, nil // indisponible provider : on reste pending
	}
	return s.applyStatus(ctx, payment, status.Status)
}

// applyStatus applique le statut reconfirmé et accorde les crédits si succès.
func (s *Service) applyStatus(ctx context.Context, payment domain.Payment, status string) (domain.Payment, error) {
	if status == domain.PaymentPending && payment.Status != domain.PaymentPending {
		return payment, nil // jamais de retour arrière vers pending
	}
	if status == payment.Status {
		return payment, nil
	}
	updated, err := s.payments.MarkStatus(ctx, payment.ID, status)
	if err != nil {
		return domain.Payment{}, err
	}
	if status == domain.PaymentSucceeded {
		plan, err := s.plans.Get(ctx, payment.PlanID)
		if err != nil {
			return domain.Payment{}, fmt.Errorf("plan introuvable pour le paiement: %w", err)
		}
		if _, err := s.credits.Grant(ctx, payment.UserID, plan.Credits, domain.OperationPurchase, "payment:"+payment.ID); err != nil {
			return domain.Payment{}, fmt.Errorf("octroi des crédits: %w", err)
		}
		slog.Info("payment succeeded", "payment", payment.ID, "user", payment.UserID, "credits", plan.Credits)
	}
	if s.audit != nil {
		action := domain.AuditPaymentFailed
		if status == domain.PaymentSucceeded {
			action = domain.AuditPaymentSucceeded
		}
		s.audit.Log(ctx, payment.UserID, action, "payment", payment.ID,
			map[string]any{"provider": payment.Provider, "amount_minor": payment.AmountMinor, "credits": planCreditsOrZero(s, ctx, payment)})
	}
	return updated, nil
}

func planCreditsOrZero(s *Service, ctx context.Context, payment domain.Payment) int64 {
	plan, err := s.plans.Get(ctx, payment.PlanID)
	if err != nil {
		return 0
	}
	return plan.Credits
}

func (s *Service) activeProvider() (port.PaymentProvider, bool) {
	for _, p := range s.providers {
		return p, true
	}
	return nil, false
}

// randomHex génère une clé d'idempotence interne.
func randomHex() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
