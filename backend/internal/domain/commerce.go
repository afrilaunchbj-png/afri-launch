package domain

import (
	"errors"
	"time"
)

// Erreurs métier du module commerce (Chariow).
var (
	ErrCommerceNotConnected       = errors.New("commerce non connecté")
	ErrCommerceInvalidCredentials = errors.New("credentials Chariow invalides")
	ErrCommerceUnavailable        = errors.New("Chariow indisponible")
	ErrCommerceWebhookInvalid     = errors.New("signature webhook Chariow invalide")
	ErrCommerceProductUnavailable = errors.New("produit Chariow indisponible")
	ErrCommerceSecretMissing      = errors.New("secret webhook non configuré")
	ErrCommerceConfigMissing      = errors.New("chiffrement non configuré")
)

// CommerceRemoteError est une erreur 4xx renvoyée par Chariow (message sûr).
type CommerceRemoteError struct {
	Status  int
	Message string
}

// Error implémente error.
func (e *CommerceRemoteError) Error() string { return "chariow: " + e.Message }

// Provider de commerce supporté.
const CommerceProviderChariow = "chariow"

// Statuts d'une connexion de commerce.
const (
	CommerceStatusConnected         = "connected"
	CommerceStatusDisconnected      = "disconnected"
	CommerceStatusError             = "error"
	CommerceStatusReconnectRequired = "reconnect_required"
)

// Statuts d'une vente.
const (
	CommerceSaleAwaitingPayment = "awaiting_payment"
	CommerceSaleCompleted       = "completed"
	CommerceSaleFailed          = "failed"
	CommerceSaleAbandoned       = "abandoned"
	CommerceSaleSettled         = "settled"
	CommerceSaleRefunded        = "refunded"
)

// CommerceConnection : une boutique Chariow connectée par un utilisateur.
// La clé API et le secret Pulse sont chiffrés au repos, jamais exposés.
type CommerceConnection struct {
	ID            string
	UserID        string
	Provider      string
	Status        string
	StoreID       string
	StoreName     string
	StoreURL      string
	StoreCurrency string
	ConnectedAt   *time.Time
	LastSyncAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CommerceProductLink associe un projet SaaS à un produit Chariow existant.
type CommerceProductLink struct {
	ID                string
	UserID            string
	ConnectionID      string
	ProjectID         string
	ExternalProductID string
	ExternalSlug      string
	ExternalName      string
	PriceMinor        int64
	Currency          string
	Status            string
	IsPublic          bool
	PublicToken       string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CommerceSale : une vente Chariow (créée à l'init du checkout puis mise à
// jour par webhook/sync). custom_metadata et payload sont du JSON brut.
type CommerceSale struct {
	ID            string
	UserID        string
	ConnectionID  string
	ProductLinkID *string
	ExternalID    string
	Status        string
	AmountMinor   int64
	Currency      string
	BuyerEmail    string
	BuyerName     string
	CheckoutURL   string
	Metadata      []byte
	Payload       []byte
	CompletedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
