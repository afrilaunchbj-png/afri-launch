package port

import (
	"context"

	"afrilaunch/backend/internal/domain"
)

// PaymentCheckoutInput décrit un checkout à créer chez le provider.
type PaymentCheckoutInput struct {
	// PaymentID est l'UUID interne du paiement — utilisé comme référence
	// idempotente chez le provider (depositId / merchant_reference…).
	PaymentID string
	// AmountMinor est le montant en unités mineures de la devise.
	AmountMinor int64
	Currency    string
	Description string
	// ReturnURL est la redirection navigateur APRÈS paiement (jamais fiable
	// pour valider un paiement — uniquement pour l'UX).
	ReturnURL string
	// WebhookURL est le callback serveur-à-serveur (optionnel : certains
	// providers l'acceptent inline, d'autres via leur dashboard).
	WebhookURL string
	// Customer est optionnel selon les providers.
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
}

// PaymentCheckoutResult est le résultat d'une création de checkout.
type PaymentCheckoutResult struct {
	// RedirectURL est l'URL hébergée par le provider vers laquelle rediriger
	// l'utilisateur (page de choix de l'opérateur / saisie du PIN).
	RedirectURL string
	// ProviderReference est l'identifiant du paiement chez le provider
	// (token, checkoutId, transaction id…).
	ProviderReference string
}

// PaymentStatusResult est le statut d'un paiement interrogé chez le provider
// (source de vérité — jamais le retour navigateur).
type PaymentStatusResult struct {
	Status string // domain.PaymentPending / PaymentSucceeded / PaymentFailed
}

// PaymentWebhookInput porte le webhook brut (headers + corps) : le parsing
// spécifique au provider reste dans le provider.
type PaymentWebhookInput struct {
	Headers map[string]string
	Body    []byte
}

// PaymentWebhookResult est le résultat de l'analyse d'un webhook.
type PaymentWebhookResult struct {
	// ProviderReference identifie le paiement concerné chez le provider.
	ProviderReference string
	// Accepted indique si le provider considère le webhook valide (signature
	// vérifiée ou structure conforme).
	Accepted bool
}

// PaymentProvider encapsule TOUTE la logique d'un provider de paiement
// (ADR-008/ADR-018) : création de checkout, vérification de statut et
// traitement des webhooks. Aucune logique provider dans le métier.
type PaymentProvider interface {
	Provider() string
	// CreateCheckout initie un paiement hébergé et renvoie l'URL de redirection.
	CreateCheckout(ctx context.Context, in PaymentCheckoutInput) (PaymentCheckoutResult, error)
	// VerifyStatus interroge le provider (source de vérité) pour un paiement.
	VerifyStatus(ctx context.Context, providerReference string) (PaymentStatusResult, error)
	// HandleWebhook analyse un webhook entrant et renvoie la référence
	// provider concernée. La validation définitive passe toujours par
	// VerifyStatus (le corps d'un webhook n'est jamais une preuve).
	HandleWebhook(ctx context.Context, in PaymentWebhookInput) (PaymentWebhookResult, error)
}

// PlanRepository accède aux packs de crédits.
type PlanRepository interface {
	ListActive(ctx context.Context) ([]domain.Plan, error)
	Get(ctx context.Context, id string) (domain.Plan, error)
}

// PaymentRepository accède aux paiements (recharges de crédits).
type PaymentRepository interface {
	Create(ctx context.Context, p domain.Payment) (domain.Payment, error)
	// UpdateCheckout persiste référence provider + URL de checkout.
	UpdateCheckout(ctx context.Context, id, providerReference, checkoutURL string) error
	// MarkStatus applique une transition de statut (idempotent : renvoie
	// l'état courant sans erreur si déjà appliqué).
	MarkStatus(ctx context.Context, id, status string) (domain.Payment, error)
	Get(ctx context.Context, userID, id string) (domain.Payment, error)
	GetByProviderReference(ctx context.Context, providerReference string) (domain.Payment, error)
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.Payment, error)
}
