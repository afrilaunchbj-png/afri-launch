package handler

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	advapp "afrilaunch/backend/internal/application/advertising"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// IntegrationHandler expose les intégrations publicitaires (connexions
// OAuth, comptes, campagnes, insights) — ADR-017.
type IntegrationHandler struct {
	svc          *advapp.Service
	appURL       string
	redirectURIs map[string]string // provider → URI de callback (identique plateforme)
}

// NewIntegrationHandler construit le handler d'intégrations.
func NewIntegrationHandler(svc *advapp.Service, appURL string, redirectURIs map[string]string) *IntegrationHandler {
	return &IntegrationHandler{svc: svc, appURL: appURL, redirectURIs: redirectURIs}
}

// connDTO sérialise une connexion — jamais de token ici.
type connDTO struct {
	ID                  string  `json:"id"`
	Provider            string  `json:"provider"`
	Status              string  `json:"status"`
	ExternalAccountID   string  `json:"external_account_id,omitempty"`
	ExternalAccountName string  `json:"external_account_name,omitempty"`
	LastError           string  `json:"last_error,omitempty"`
	LastSyncAt          *string `json:"last_sync_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
}

func toConnDTO(c domain.AdPlatformConnection) connDTO {
	var lastSync *string
	if c.LastSyncAt != nil {
		s := c.LastSyncAt.Format("2006-01-02 15:04")
		lastSync = &s
	}
	return connDTO{
		ID: c.ID, Provider: c.Provider, Status: c.Status,
		ExternalAccountID: c.ExternalAccountID, ExternalAccountName: c.ExternalAccountName,
		LastError: c.LastError, LastSyncAt: lastSync, CreatedAt: c.CreatedAt.Format("2006-01-02"),
	}
}

// List gère GET /integrations.
func (h *IntegrationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	conns, err := h.svc.ListConnections(r.Context(), userID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]connDTO, 0, len(conns))
	for _, c := range conns {
		out = append(out, toConnDTO(c))
	}
	writeData(w, http.StatusOK, map[string]any{"connections": out, "capabilities": h.svc.Capabilities()})
}

// Connect gère GET /integrations/{provider}/connect → URL d'autorisation.
func (h *IntegrationHandler) Connect(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	provider := chi.URLParam(r, "provider")
	authURL, err := h.svc.StartConnect(r.Context(), userID, provider, h.redirectURI(provider))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"authorization_url": authURL})
}

// Callback gère GET /integrations/{provider}/callback (redirect OAuth —
// non authentifié par JWT ; l'utilisateur est résolu via l'état consommé).
// Redirige vers le frontend sans jamais exposer de token.
func (h *IntegrationHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	q := r.URL.Query()
	target := h.appURL + "/integrations?connect=" + url.QueryEscape(provider) + "&status="

	if errMsg := q.Get("error"); errMsg != "" {
		http.Redirect(w, r, target+"error&reason="+url.QueryEscape(errMsg), http.StatusFound)
		return
	}

	_, err := h.svc.HandleCallback(r.Context(), provider, q.Get("code"), q.Get("state"), h.redirectURI(provider))
	if err != nil {
		target += "error"
	} else {
		target += "success"
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// ListAccounts gère GET /integrations/{provider}/accounts.
func (h *IntegrationHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	provider := chi.URLParam(r, "provider")
	accounts, err := h.svc.ListAccounts(r.Context(), userID, provider)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, accounts)
}

// SelectAccount gère POST /integrations/{provider}/accounts/select.
func (h *IntegrationHandler) SelectAccount(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	provider := chi.URLParam(r, "provider")
	var in struct {
		ExternalAccountID string `json:"external_account_id"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	conn, err := h.svc.SelectAccount(r.Context(), userID, provider, in.ExternalAccountID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toConnDTO(conn))
}

// Disconnect gère DELETE /integrations/{provider}.
func (h *IntegrationHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	provider := chi.URLParam(r, "provider")
	if err := h.svc.Disconnect(r.Context(), userID, provider); err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"ok": true})
}

// SyncCampaigns gère POST /integrations/{provider}/campaigns/sync.
func (h *IntegrationHandler) SyncCampaigns(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	provider := chi.URLParam(r, "provider")
	campaigns, err := h.svc.SyncCampaigns(r.Context(), userID, provider)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toCampaignDTOs(campaigns))
}

// campaignDTO sérialise une campagne interne.
type campaignDTO struct {
	ID                 string `json:"id"`
	ConnectionID       string `json:"connection_id"`
	ExternalCampaignID string `json:"external_campaign_id,omitempty"`
	Name               string `json:"name"`
	Objective          string `json:"objective,omitempty"`
	Status             string `json:"status"`
	BudgetMinor        int64  `json:"budget_minor"`
	Currency           string `json:"currency"`
}

func toCampaignDTO(c domain.Campaign) campaignDTO {
	return campaignDTO{
		ID: c.ID, ConnectionID: c.ConnectionID, ExternalCampaignID: c.ExternalCampaignID,
		Name: c.Name, Objective: c.Objective, Status: c.Status,
		BudgetMinor: c.BudgetMinor, Currency: c.Currency,
	}
}

func toCampaignDTOs(campaigns []domain.Campaign) []campaignDTO {
	out := make([]campaignDTO, 0, len(campaigns))
	for _, c := range campaigns {
		out = append(out, toCampaignDTO(c))
	}
	return out
}

// ListCampaigns gère GET /ad-campaigns.
func (h *IntegrationHandler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	campaigns, err := h.svc.ListCampaigns(r.Context(), userID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toCampaignDTOs(campaigns))
}

// CreateCampaign gère POST /ad-campaigns.
func (h *IntegrationHandler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	var in advapp.CreateCampaignInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	campaign, err := h.svc.CreateCampaign(r.Context(), userID, in)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toCampaignDTO(campaign))
}

// PauseCampaign gère POST /ad-campaigns/{id}/pause.
func (h *IntegrationHandler) PauseCampaign(w http.ResponseWriter, r *http.Request) {
	h.setCampaignStatus(w, r, domain.CampaignPaused)
}

// ResumeCampaign gère POST /ad-campaigns/{id}/resume.
func (h *IntegrationHandler) ResumeCampaign(w http.ResponseWriter, r *http.Request) {
	h.setCampaignStatus(w, r, domain.CampaignActive)
}

func (h *IntegrationHandler) setCampaignStatus(w http.ResponseWriter, r *http.Request, status string) {
	userID := authctx.UserID(r.Context())
	id := chi.URLParam(r, "id")
	campaign, err := h.svc.SetCampaignStatus(r.Context(), userID, id, status)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toCampaignDTO(campaign))
}

// CampaignInsights gère GET /ad-campaigns/{id}/insights?since=&until=.
func (h *IntegrationHandler) CampaignInsights(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	until := q.Get("until")
	if until == "" {
		until = q.Get("since") // journée unique par défaut
	}
	insights, err := h.svc.GetInsights(r.Context(), userID, id, q.Get("since"), until)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, insights)
}

func (h *IntegrationHandler) redirectURI(provider string) string {
	return h.redirectURIs[provider]
}
