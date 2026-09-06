// Package port contient les interfaces (ports) du module commerce.
package port

import (
	"context"
	"time"

	"afrilaunch/backend/internal/domain"
)

// ---------- Provider de commerce (Chariow) ----------

// CommerceStore décrit une boutique renvoyée par le provider.
type CommerceStore struct {
	ID       string
	Name     string
	URL      string
	Currency string
}

// CommerceProduct décrit un produit publié (API Chariow : lecture seule).
type CommerceProduct struct {
	ID         string
	Slug       string
	Name       string
	Type       string
	IsFree     bool
	Currency   string
	PriceMinor int64
}

// CommerceCheckoutInput porte les données d'un checkout hébergé.
type CommerceCheckoutInput struct {
	ProductID         string
	BuyerEmail        string
	BuyerFirstName    string
	BuyerLastName     string
	BuyerPhoneNumber  string
	BuyerPhoneCountry string
	RedirectURL       string
	CustomerIP        string
	Metadata          map[string]string
}

// CommerceCheckoutResult est le résultat d'une initiation de checkout.
type CommerceCheckoutResult struct {
	Step        string // payment | completed | already_purchased
	SaleID      string
	Status      string
	CheckoutURL string
	AmountMinor int64
	Currency    string
}

// CommerceSaleInfo est une vente renvoyée par l'API du provider.
type CommerceSaleInfo struct {
	ID          string
	Status      string
	AmountMinor int64
	Currency    string
	BuyerEmail  string
	BuyerName   string
	Metadata    map[string]string
	Raw         []byte
}

// CommerceProvider est l'abstraction d'une plateforme de commerce.
type CommerceProvider interface {
	Provider() string
	// GetStore valide une clé API et renvoie la boutique associée.
	GetStore(ctx context.Context, apiKey string) (CommerceStore, error)
	// ListProducts liste les produits publiés (paginated par curseur).
	ListProducts(ctx context.Context, apiKey, cursor string, perPage int) ([]CommerceProduct, string, error)
	// CreateCheckout initie un checkout hébergé.
	CreateCheckout(ctx context.Context, apiKey string, in CommerceCheckoutInput) (CommerceCheckoutResult, error)
	// ListSales liste les ventes (paginated par curseur).
	ListSales(ctx context.Context, apiKey, cursor string, perPage int) ([]CommerceSaleInfo, string, error)
}

// ---------- Repository (persistance) ----------

// CommerceRepository persiste connexions, liens produit, ventes et events.
type CommerceRepository interface {
	UpsertConnection(ctx context.Context, c domain.CommerceConnection, apiKeyEnc, webhookSecretEnc string) error
	GetConnectionByProvider(ctx context.Context, userID, provider string) (conn domain.CommerceConnection, apiKeyEnc, webhookSecretEnc string, err error)
	GetConnectionByStoreID(ctx context.Context, storeID string) (conn domain.CommerceConnection, apiKeyEnc, webhookSecretEnc string, err error)
	ListConnections(ctx context.Context, userID string) ([]domain.CommerceConnection, error)
	UpdateConnectionState(ctx context.Context, id, userID, status, lastError string, lastSyncAt, connectedAt *time.Time) error
	DeleteConnection(ctx context.Context, id, userID string) error

	UpsertProductLink(ctx context.Context, link domain.CommerceProductLink) (domain.CommerceProductLink, error)
	GetProductLinkByProject(ctx context.Context, userID, projectID string) (domain.CommerceProductLink, error)
	GetProductLinkByExternal(ctx context.Context, connectionID, externalProductID string) (domain.CommerceProductLink, error)
	GetProductLink(ctx context.Context, id, userID string) (domain.CommerceProductLink, error)
	GetProductLinkByToken(ctx context.Context, token string) (domain.CommerceProductLink, error)
	UpdateProductLinkPublic(ctx context.Context, id, userID string, isPublic bool, publicToken string) error
	DeleteProductLink(ctx context.Context, id, userID string) error
	DeleteProductLinksByProject(ctx context.Context, userID, projectID string) error
	ListProductLinks(ctx context.Context, userID string) ([]domain.CommerceProductLink, error)

	UpsertSale(ctx context.Context, sale domain.CommerceSale) (domain.CommerceSale, error)
	GetSaleByExternalID(ctx context.Context, connectionID, externalID string) (domain.CommerceSale, error)
	ListSalesByProductLink(ctx context.Context, linkID, userID string, limit int) ([]domain.CommerceSale, error)
	ListSalesByUser(ctx context.Context, userID string, limit int) ([]domain.CommerceSale, error)

	// CreateWebhookEvent insère un événement si sa livraison est inédite.
	// Renvoie false si l'event (provider+delivery_id) existe déjà.
	CreateWebhookEvent(ctx context.Context, userID, provider, pulseID, deliveryID, eventType string, payload []byte, status string) (bool, error)
}
