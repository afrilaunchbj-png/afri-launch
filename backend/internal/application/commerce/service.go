// Package commerce orchestre l'intégration de boutiques Chariow par
// utilisateur (ADR-019) : connexion (API key chiffrée), association produit,
// checkout hébergé multi-tenant, webhooks Pulse (HMAC + idempotence) et
// synchronisation des ventes.
package commerce

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"strings"
	"time"

	"afrilaunch/backend/internal/application/audit"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Provider clé du module.
const ProviderChariow = domain.CommerceProviderChariow

// Service orchestre le module commerce.
type Service struct {
	provider port.CommerceProvider
	repo     port.CommerceRepository
	enc      port.SecretEncryptor
	audit    *audit.Recorder
	appURL   string
}

// NewService construit le service commerce.
func NewService(provider port.CommerceProvider, repo port.CommerceRepository, enc port.SecretEncryptor, auditRec *audit.Recorder, appURL string) *Service {
	return &Service{provider: provider, repo: repo, enc: enc, audit: auditRec, appURL: appURL}
}

// WebhookHeaders porte les en-têtes d'une livraison Pulse.
type WebhookHeaders struct {
	Signature  string
	PulseID    string
	DeliveryID string
	Event      string
}

// ---------- Connexion ----------

// ConnectInput contient la clé API Chariow (et, optionnel, le secret du
// Pulse webhook). Jamais persistés en clair.
type ConnectInput struct {
	APIKey        string
	WebhookSecret string
}

// Connect valide la clé auprès de Chariow puis enregistre la connexion.
func (s *Service) Connect(ctx context.Context, userID string, in ConnectInput) (domain.CommerceConnection, error) {
	apiKey := strings.TrimSpace(in.APIKey)
	if apiKey == "" {
		return domain.CommerceConnection{}, domain.ErrInvalidInput
	}
	store, err := s.provider.GetStore(ctx, apiKey)
	if err != nil {
		return domain.CommerceConnection{}, err
	}
	if store.ID == "" {
		return domain.CommerceConnection{}, domain.ErrCommerceInvalidCredentials
	}
	apiKeyEnc, err := s.encrypt(apiKey)
	if err != nil {
		return domain.CommerceConnection{}, err
	}
	secretEnc, err := s.encrypt(strings.TrimSpace(in.WebhookSecret))
	if err != nil {
		return domain.CommerceConnection{}, err
	}
	now := time.Now().UTC()
	conn := domain.CommerceConnection{
		UserID: userID, Provider: ProviderChariow, Status: domain.CommerceStatusConnected,
		StoreID: store.ID, StoreName: store.Name, StoreURL: store.URL, StoreCurrency: store.Currency,
		ConnectedAt: &now,
	}
	if err := s.repo.UpsertConnection(ctx, conn, apiKeyEnc, secretEnc); err != nil {
		return domain.CommerceConnection{}, err
	}
	s.log(ctx, userID, "commerce.connect", map[string]any{"store": store.ID, "store_name": store.Name})
	return conn, nil
}

// Status renvoie la connexion de l'utilisateur, un indicateur « actif »
// (statut connected) et la présence d'un secret de webhook configuré.
func (s *Service) Status(ctx context.Context, userID string) (conn domain.CommerceConnection, ok bool, hasWebhookSecret bool, err error) {
	conn, secretEnc, _, err := s.repo.GetConnectionByProvider(ctx, userID, ProviderChariow)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.CommerceConnection{}, false, false, nil
		}
		return domain.CommerceConnection{}, false, false, err
	}
	return conn, conn.Status == domain.CommerceStatusConnected, secretEnc != "", nil
}

// UpdateWebhookSecret met à jour le secret Pulse (permet d'ajouter le
// webhook après la connexion, sans redonner la clé API).
func (s *Service) UpdateWebhookSecret(ctx context.Context, userID, webhookSecret string) error {
	conn, apiKeyEnc, _, err := s.repo.GetConnectionByProvider(ctx, userID, ProviderChariow)
	if err != nil {
		return err
	}
	secretEnc, err := s.encrypt(strings.TrimSpace(webhookSecret))
	if err != nil {
		return err
	}
	return s.repo.UpsertConnection(ctx, conn, apiKeyEnc, secretEnc)
}

// Disconnect invalide la connexion : secrets effacés, historique conservé.
func (s *Service) Disconnect(ctx context.Context, userID string) error {
	conn, _, _, err := s.repo.GetConnectionByProvider(ctx, userID, ProviderChariow)
	if err != nil {
		return err
	}
	conn.Status = domain.CommerceStatusDisconnected
	if err := s.repo.UpsertConnection(ctx, conn, "", ""); err != nil {
		return err
	}
	s.log(ctx, userID, "commerce.disconnect", map[string]any{"store": conn.StoreID})
	return nil
}

// DeleteConnection supprime entièrement la connexion (liens et ventes liées
// supprimés en cascade) — utilisée pour changer de boutique Chariow.
func (s *Service) DeleteConnection(ctx context.Context, userID string) error {
	conn, _, _, err := s.repo.GetConnectionByProvider(ctx, userID, ProviderChariow)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteConnection(ctx, conn.ID, userID); err != nil {
		return err
	}
	s.log(ctx, userID, "commerce.delete_connection", map[string]any{"store": conn.StoreID})
	return nil
}

// SetReconnectRequired marque la connexion après un 401/403 Chariow.
func (s *Service) SetReconnectRequired(ctx context.Context, userID string, cause string) {
	conn, ok, _, err := s.Status(ctx, userID)
	if err != nil || !ok {
		return
	}
	now := time.Now().UTC()
	_ = s.repo.UpdateConnectionState(ctx, conn.ID, userID, domain.CommerceStatusReconnectRequired, cause, &now, conn.ConnectedAt)
}

// ---------- Produits ----------

// ListRemoteProducts liste les produits publiés de la boutique connectée.
func (s *Service) ListRemoteProducts(ctx context.Context, userID string) ([]port.CommerceProduct, error) {
	apiKey, err := s.decryptKey(ctx, userID)
	if err != nil {
		return nil, err
	}
	var out []port.CommerceProduct
	cursor := ""
	for page := 0; page < 10; page++ {
		items, next, err := s.provider.ListProducts(ctx, apiKey, cursor, 100)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if next == "" {
			break
		}
		cursor = next
	}
	return out, nil
}

// LinkProduct associe un projet SaaS à un produit Chariow existant.
func (s *Service) LinkProduct(ctx context.Context, userID, projectID, externalProductID string) (domain.CommerceProductLink, error) {
	conn, apiKeyEnc, _, err := s.repo.GetConnectionByProvider(ctx, userID, ProviderChariow)
	if err != nil {
		return domain.CommerceProductLink{}, err
	}
	apiKey, err := s.decrypt(apiKeyEnc)
	if err != nil {
		return domain.CommerceProductLink{}, err
	}
	var product *port.CommerceProduct
	cursor := ""
	for page := 0; page < 10 && product == nil; page++ {
		items, next, err := s.provider.ListProducts(ctx, apiKey, cursor, 100)
		if err != nil {
			return domain.CommerceProductLink{}, err
		}
		for i := range items {
			if items[i].ID == externalProductID {
				p := items[i]
				product = &p
				break
			}
		}
		cursor = next
		if cursor == "" {
			break
		}
	}
	if product == nil {
		return domain.CommerceProductLink{}, domain.ErrCommerceProductUnavailable
	}
	if err := s.repo.DeleteProductLinksByProject(ctx, userID, projectID); err != nil {
		return domain.CommerceProductLink{}, err
	}
	link, err := s.repo.UpsertProductLink(ctx, domain.CommerceProductLink{
		UserID: userID, ConnectionID: conn.ID, ProjectID: projectID,
		ExternalProductID: product.ID, ExternalSlug: product.Slug, ExternalName: product.Name,
		PriceMinor: product.PriceMinor, Currency: product.Currency, Status: "linked",
	})
	if err != nil {
		return domain.CommerceProductLink{}, err
	}
	s.log(ctx, userID, "commerce.link_product", map[string]any{"link": link.ID, "external_product": product.ID, "project": projectID})
	return link, nil
}

// UnlinkProduct supprime l'association.
func (s *Service) UnlinkProduct(ctx context.Context, userID, linkID string) error {
	return s.repo.DeleteProductLink(ctx, linkID, userID)
}

// SetProductPublic publie/dépublie la page d'achat publique du lien.
func (s *Service) SetProductPublic(ctx context.Context, userID, linkID string, isPublic bool) (domain.CommerceProductLink, error) {
	link, err := s.repo.GetProductLink(ctx, linkID, userID)
	if err != nil {
		return domain.CommerceProductLink{}, err
	}
	token := link.PublicToken
	if isPublic && token == "" {
		token = randomToken()
	} else if !isPublic {
		token = ""
	}
	if err := s.repo.UpdateProductLinkPublic(ctx, linkID, userID, isPublic, token); err != nil {
		return domain.CommerceProductLink{}, err
	}
	link.IsPublic = isPublic
	link.PublicToken = token
	s.log(ctx, userID, "commerce.publish", map[string]any{"link": link.ID, "public": isPublic})
	return link, nil
}

// ProjectLink renvoie le lien d'un projet.
func (s *Service) ProjectLink(ctx context.Context, userID, projectID string) (domain.CommerceProductLink, error) {
	return s.repo.GetProductLinkByProject(ctx, userID, projectID)
}

// ListLinks renvoie toutes les associations de l'utilisateur.
func (s *Service) ListLinks(ctx context.Context, userID string) ([]domain.CommerceProductLink, error) {
	return s.repo.ListProductLinks(ctx, userID)
}

// ---------- Checkout public ----------

// PublicProduct est la vue publique d'un produit en vente.
type PublicProduct struct {
	Token      string
	Name       string
	PriceMinor int64
	Currency   string
	IsFree     bool
	StoreName  string
}

// PublicInfo renvoie les infos publiques d'une page d'achat.
func (s *Service) PublicInfo(ctx context.Context, token string) (PublicProduct, error) {
	link, err := s.repo.GetProductLinkByToken(ctx, token)
	if err != nil {
		return PublicProduct{}, err
	}
	conn, _, _, err := s.repo.GetConnectionByProvider(ctx, link.UserID, ProviderChariow)
	if err != nil {
		return PublicProduct{}, err
	}
	return PublicProduct{
		Token: link.PublicToken, Name: link.ExternalName, PriceMinor: link.PriceMinor,
		Currency: link.Currency, IsFree: link.PriceMinor == 0, StoreName: conn.StoreName,
	}, nil
}

// CheckoutBuyer est l'acheteur final (page publique).
type CheckoutBuyer struct {
	Email        string
	FirstName    string
	LastName     string
	PhoneNumber  string
	PhoneCountry string
	CustomerIP   string
}

// PublicCheckout initie un checkout Chariow pour un produit publié.
func (s *Service) PublicCheckout(ctx context.Context, token string, buyer CheckoutBuyer) (port.CommerceCheckoutResult, domain.CommerceSale, error) {
	link, err := s.repo.GetProductLinkByToken(ctx, token)
	if err != nil {
		return port.CommerceCheckoutResult{}, domain.CommerceSale{}, err
	}
	conn, apiKeyEnc, _, err := s.repo.GetConnectionByProvider(ctx, link.UserID, ProviderChariow)
	if err != nil {
		return port.CommerceCheckoutResult{}, domain.CommerceSale{}, err
	}
	apiKey, err := s.decrypt(apiKeyEnc)
	if err != nil {
		return port.CommerceCheckoutResult{}, domain.CommerceSale{}, err
	}
	meta := map[string]string{
		"saas_user_id":    link.UserID,
		"saas_project_id": link.ProjectID,
		"saas_link_id":    link.ID,
	}
	redirect := strings.TrimRight(s.appURL, "/") + "/buy/" + link.PublicToken + "?thanks=1"
	res, err := s.provider.CreateCheckout(ctx, apiKey, port.CommerceCheckoutInput{
		ProductID: link.ExternalProductID, BuyerEmail: buyer.Email,
		BuyerFirstName: buyer.FirstName, BuyerLastName: buyer.LastName,
		BuyerPhoneNumber: buyer.PhoneNumber, BuyerPhoneCountry: buyer.PhoneCountry,
		RedirectURL: redirect, CustomerIP: buyer.CustomerIP, Metadata: meta,
	})
	if err != nil {
		return port.CommerceCheckoutResult{}, domain.CommerceSale{}, err
	}
	metaJSON, _ := json.Marshal(meta)
	completed := completedAtFor(res.Status)
	sale, err := s.repo.UpsertSale(ctx, domain.CommerceSale{
		UserID: link.UserID, ConnectionID: conn.ID,
		ProductLinkID: &link.ID, ExternalID: res.SaleID, Status: res.Status,
		AmountMinor: res.AmountMinor, Currency: res.Currency,
		BuyerEmail: buyer.Email, BuyerName: strings.TrimSpace(buyer.FirstName + " " + buyer.LastName),
		CheckoutURL: res.CheckoutURL, Metadata: metaJSON, CompletedAt: completed,
	})
	if err != nil {
		return port.CommerceCheckoutResult{}, domain.CommerceSale{}, err
	}
	return res, sale, nil
}

// ---------- Ventes & webhooks ----------

// SyncSales synchronise les ventes de la boutique connectée (idempotent).
func (s *Service) SyncSales(ctx context.Context, userID string) error {
	conn, apiKeyEnc, _, err := s.repo.GetConnectionByProvider(ctx, userID, ProviderChariow)
	if err != nil {
		return err
	}
	apiKey, err := s.decrypt(apiKeyEnc)
	if err != nil {
		return err
	}
	cursor := ""
	for page := 0; page < 10; page++ {
		items, next, err := s.provider.ListSales(ctx, apiKey, cursor, 100)
		if err != nil {
			s.SetReconnectRequired(ctx, userID, "sync: "+err.Error())
			return err
		}
		for _, info := range items {
			if info.ID == "" {
				continue
			}
			payload, _ := json.Marshal(map[string]any{"sale_id": info.ID, "status": info.Status})
			_, _ = s.repo.UpsertSale(ctx, domain.CommerceSale{
				UserID: conn.UserID, ConnectionID: conn.ID, ExternalID: info.ID,
				Status: normalizeSaleStatus(info.Status), AmountMinor: info.AmountMinor, Currency: info.Currency,
				BuyerEmail: info.BuyerEmail, BuyerName: info.BuyerName,
				Payload: payload, CompletedAt: completedAtFor(info.Status),
			})
		}
		if next == "" {
			break
		}
		cursor = next
	}
	now := time.Now().UTC()
	if err := s.repo.UpdateConnectionState(ctx, conn.ID, userID, domain.CommerceStatusConnected, "", &now, conn.ConnectedAt); err != nil {
		return err
	}
	s.log(ctx, userID, "commerce.sync", map[string]any{"store": conn.StoreID})
	return nil
}

// Sales liste les dernières ventes de l'utilisateur.
func (s *Service) Sales(ctx context.Context, userID string) ([]domain.CommerceSale, error) {
	return s.repo.ListSalesByUser(ctx, userID, 100)
}

// HandleWebhook valide, déduplique et traite une livraison Pulse.
func (s *Service) HandleWebhook(ctx context.Context, body []byte, h WebhookHeaders) error {
	if len(body) == 0 {
		return domain.ErrInvalidInput
	}
	var p struct {
		Event string `json:"event"`
		Store struct {
			ID string `json:"id"`
		} `json:"store"`
		Sale struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Amount struct {
				Value    float64 `json:"value"`
				Currency string  `json:"currency"`
			} `json:"amount"`
			CustomMetadata map[string]string `json:"custom_metadata"`
		} `json:"sale"`
		Product struct {
			ID string `json:"id"`
		} `json:"product"`
		Customer struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"customer"`
		Checkout struct {
			URL string `json:"url"`
		} `json:"checkout"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return domain.ErrInvalidInput
	}
	if p.Store.ID == "" {
		return nil // livraison sans contexte boutique : on acquitte
	}
	conn, _, secretEnc, err := s.repo.GetConnectionByStoreID(ctx, p.Store.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil // boutique inconnue du SaaS : acquitter (pas de retry)
		}
		return err
	}
	secret, err := s.decrypt(secretEnc)
	if err != nil {
		return err
	}
	if secret == "" {
		return domain.ErrCommerceSecretMissing
	}
	if !verifyPulseSignature(secret, body, h.Signature) {
		return domain.ErrCommerceWebhookInvalid
	}

	event := h.Event
	if event == "" {
		event = p.Event
	}
	saleStatus := ""
	switch event {
	case "successful.sale":
		saleStatus = domain.CommerceSaleCompleted
	case "failed.sale":
		saleStatus = domain.CommerceSaleFailed
	case "abandoned.sale":
		saleStatus = domain.CommerceSaleAbandoned
	case "refunded.sale":
		saleStatus = domain.CommerceSaleRefunded
	}

	if saleStatus != "" && p.Sale.ID != "" {
		var linkID *string
		if link, err := s.repo.GetProductLinkByExternal(ctx, conn.ID, p.Product.ID); err == nil {
			linkID = &link.ID
		}
		meta, _ := json.Marshal(p.Sale.CustomMetadata)
		_, _ = s.repo.UpsertSale(ctx, domain.CommerceSale{
			UserID: conn.UserID, ConnectionID: conn.ID, ProductLinkID: linkID,
			ExternalID: p.Sale.ID, Status: saleStatus,
			AmountMinor: toMinor(p.Sale.Amount.Value, p.Sale.Amount.Currency), Currency: p.Sale.Amount.Currency,
			BuyerEmail: p.Customer.Email, BuyerName: p.Customer.Name,
			CheckoutURL: p.Checkout.URL, Metadata: meta, CompletedAt: completedAtFor(saleStatus),
		})
	}

	eventStatus := "ignored"
	if saleStatus != "" {
		eventStatus = "processed"
	}
	_, _ = s.repo.CreateWebhookEvent(ctx, conn.UserID, ProviderChariow, h.PulseID, h.DeliveryID, event, body, eventStatus)
	return nil
}

// ---------- Helpers ----------

func (s *Service) decryptKey(ctx context.Context, userID string) (string, error) {
	_, apiKeyEnc, _, err := s.repo.GetConnectionByProvider(ctx, userID, ProviderChariow)
	if err != nil {
		return "", err
	}
	return s.decrypt(apiKeyEnc)
}

func (s *Service) encrypt(plain string) (string, error) {
	v, err := s.enc.Encrypt(plain)
	if err != nil {
		return "", domain.ErrCommerceConfigMissing
	}
	return v, nil
}

func (s *Service) decrypt(cipher string) (string, error) {
	v, err := s.enc.Decrypt(cipher)
	if err != nil {
		return "", domain.ErrCommerceConfigMissing
	}
	return v, nil
}

func (s *Service) log(ctx context.Context, userID, action string, meta map[string]any) {
	if s.audit != nil {
		s.audit.Log(ctx, userID, action, "commerce", "", meta)
	}
}

func normalizeSaleStatus(status string) string {
	switch status {
	case domain.CommerceSaleCompleted, domain.CommerceSaleSettled:
		return status
	case "success":
		return domain.CommerceSaleCompleted
	default:
		if status == "" {
			return domain.CommerceSaleAwaitingPayment
		}
		return status
	}
}

func completedAtFor(status string) *time.Time {
	switch status {
	case domain.CommerceSaleCompleted, domain.CommerceSaleSettled, domain.CommerceSaleRefunded, domain.CommerceSaleFailed:
		now := time.Now().UTC()
		return &now
	default:
		return nil
	}
}

func toMinor(value float64, currency string) int64 {
	if currency == "XOF" || currency == "XAF" {
		return int64(math.Round(value))
	}
	return int64(math.Round(value * 100))
}

// verifyPulseSignature vérifie x-chariow-signature (HMAC-SHA256 du corps brut).
func verifyPulseSignature(secret string, rawBody []byte, header string) bool {
	const prefix = "sha256="
	if secret == "" || !strings.HasPrefix(header, prefix) {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(rawBody)
	expected := prefix + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(header), []byte(expected))
}

func randomToken() string {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		slog.Error("commerce: random token", "err", err)
		return hex.EncodeToString([]byte("fallback-token"))
	}
	return hex.EncodeToString(buf)
}
