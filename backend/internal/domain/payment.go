package domain

import (
	"errors"
	"time"
)

// Statuts d'un paiement (alignés sur payments.status, migration 00005).
const (
	PaymentPending   = "pending"
	PaymentSucceeded = "succeeded"
	PaymentFailed    = "failed"
	PaymentRefunded  = "refunded"
)

// Providers de paiement activables via PAYMENT_PROVIDER (ADR-018).
const (
	PaymentProviderPawaPay  = "pawapay"
	PaymentProviderFedaPay  = "fedapay"
	PaymentProviderPayDunya = "paydunya"
)

// Plan est un pack de crédits achetable.
type Plan struct {
	ID         string
	Name       string
	Credits    int64
	PriceMinor int64 // en unités mineures (XOF = zéro décimale)
	Currency   string
	SortOrder  int
	IsActive   bool
	CreatedAt  time.Time
}

// Payment est un achat de crédits (une recharge) via un provider.
type Payment struct {
	ID                string
	UserID            string
	PlanID            string
	AmountMinor       int64
	Currency          string
	Provider          string
	Status            string
	IdempotencyKey    string
	ProviderReference string
	CheckoutURL       string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Erreurs métier paiement.
var (
	ErrPaymentNotFound         = ErrNotFound
	ErrPaymentProviderDisabled = errors.New("payments: aucun provider de paiement configuré")
	ErrPaymentInvalidState     = errors.New("payments: transition de statut invalide")
)
