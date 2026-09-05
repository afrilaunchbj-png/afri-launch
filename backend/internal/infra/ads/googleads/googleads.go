// Package googleads encapsule TOUTE la logique Google Ads (OAuth offline,
// accessibleCustomers, campagnes GAQL, insights) derrière le port
// AdPlatformProvider — via l'API REST officielle (pas de gRPC au MVP).
// Les creatives vidéo (ad_group_ad) ne sont pas supportées à ce stade
// (Capabilities.Creatives=false) : elles nécessitent ad groups + ads
// complets (voir docs/integrations/advertising.md).
package googleads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

const (
	defaultAPIVersion = "v19"
	scopeAdWords      = "https://www.googleapis.com/auth/adwords"
)

// GoogleAds implémente port.AdPlatformProvider pour Google Ads.
type GoogleAds struct {
	clientID        string
	clientSecret    string
	developerToken  string
	loginCustomerID string // MCC contextuel si nécessaire (env, jamais l'ID client)
	apiVersion      string
	oauthBase       string
	tokenBase       string
	apiBase         string
	userInfoBase    string
	http            *http.Client
}

// New construit le provider Google Ads.
func New(clientID, clientSecret, developerToken, loginCustomerID, apiVersion string) *GoogleAds {
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}
	return &GoogleAds{
		clientID:        clientID,
		clientSecret:    clientSecret,
		developerToken:  developerToken,
		loginCustomerID: loginCustomerID,
		apiVersion:      apiVersion,
		oauthBase:       "https://accounts.google.com",
		tokenBase:       "https://oauth2.googleapis.com",
		apiBase:         "https://googleads.googleapis.com",
		userInfoBase:    "https://www.googleapis.com",
		http:            &http.Client{Timeout: 60 * time.Second},
	}
}

func (g *GoogleAds) Provider() string { return domain.AdPlatformGoogleAds }

func (g *GoogleAds) Capabilities() port.AdPlatformCapabilities {
	return port.AdPlatformCapabilities{
		Campaigns:        true,
		Creatives:        false, // ad_group_ad + ad groups : phase suivante
		VideoAds:         false,
		ImageAds:         false,
		Reporting:        true,
		BudgetManagement: true,
	}
}

// ---------- OAuth (offline : refresh_token requis pour les workers) ----------

func (g *GoogleAds) AuthorizationURL(state, redirectURI string) string {
	q := url.Values{}
	q.Set("client_id", g.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", scopeAdWords)
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	q.Set("state", state)
	return g.oauthBase + "/o/oauth2/v2/auth?" + q.Encode()
}

func (g *GoogleAds) ExchangeCode(ctx context.Context, code, redirectURI string) (port.OAuthTokenResult, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", g.clientID)
	form.Set("client_secret", g.clientSecret)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")

	res, err := g.token(ctx, form)
	if err != nil {
		return port.OAuthTokenResult{}, fmt.Errorf("googleads: exchange: %w", err)
	}
	out := port.OAuthTokenResult{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		Scopes:       []string{scopeAdWords},
	}
	if res.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
		out.ExpiresAt = &t
	}
	// Identité Google (email).
	var user struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := g.getUserInfo(ctx, res.AccessToken, &user); err == nil && user.ID != "" {
		out.ExternalUserID = user.ID
	}
	return out, nil
}

// Refresh renouvelle l'access token via le refresh token (offline).
func (g *GoogleAds) Refresh(ctx context.Context, conn domain.AdPlatformConnection) (port.OAuthTokenResult, bool, error) {
	if conn.RefreshToken == "" {
		return port.OAuthTokenResult{}, false, fmt.Errorf("googleads: refresh token manquant (reconnexion requise)")
	}
	form := url.Values{}
	form.Set("refresh_token", conn.RefreshToken)
	form.Set("client_id", g.clientID)
	form.Set("client_secret", g.clientSecret)
	form.Set("grant_type", "refresh_token")
	res, err := g.token(ctx, form)
	if err != nil {
		return port.OAuthTokenResult{}, true, fmt.Errorf("googleads: refresh: %w", err)
	}
	out := port.OAuthTokenResult{AccessToken: res.AccessToken, RefreshToken: conn.RefreshToken}
	if res.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
		out.ExpiresAt = &t
	}
	return out, true, nil
}

// Revoke révoque le token (endpoint OAuth révocable).
func (g *GoogleAds) Revoke(ctx context.Context, accessToken string) error {
	q := url.Values{}
	q.Set("token", accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/revoke?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	return g.do(req, nil, "")
}

// ---------- Comptes (accessibleCustomers + customer info) ----------

func (g *GoogleAds) ListAdAccounts(ctx context.Context, accessToken string) ([]domain.AdAccount, error) {
	var out struct {
		ResourceNames []string `json:"resourceNames"`
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		g.apiBase+"/"+g.apiVersion+"/customers:listAccessibleCustomers", nil)
	if err != nil {
		return nil, err
	}
	if err := g.do(req, &out, accessToken); err != nil {
		return nil, err
	}

	accounts := make([]domain.AdAccount, 0, len(out.ResourceNames))
	for _, rn := range out.ResourceNames {
		id := customerID(rn)
		info, err := g.customerInfo(ctx, accessToken, id)
		if err != nil {
			// Compte inaccessible : on le liste quand même avec son ID.
			accounts = append(accounts, domain.AdAccount{ExternalID: id, Status: "unknown"})
			continue
		}
		accounts = append(accounts, info)
	}
	return accounts, nil
}

// VerifyAdAccount vérifie l'accès réel au compte via une requête GAQL.
func (g *GoogleAds) VerifyAdAccount(ctx context.Context, accessToken, externalAccountID string) (domain.AdAccount, error) {
	return g.customerInfo(ctx, accessToken, customerID(externalAccountID))
}

// customerInfo lit descriptive_name/currency via searchStream.
func (g *GoogleAds) customerInfo(ctx context.Context, accessToken, id string) (domain.AdAccount, error) {
	query := "SELECT customer.descriptive_name, customer.currency_code, customer.time_zone FROM customer LIMIT 1"
	rows, err := g.searchStream(ctx, accessToken, id, query)
	if err != nil {
		return domain.AdAccount{}, domain.ErrAccountNotAccessible
	}
	if len(rows) == 0 {
		return domain.AdAccount{}, domain.ErrAccountNotAccessible
	}
	var customer struct {
		Customer struct {
			DescriptiveName string `json:"descriptiveName"`
			CurrencyCode    string `json:"currencyCode"`
			TimeZone        string `json:"timeZone"`
		} `json:"customer"`
	}
	if err := json.Unmarshal(rows[0], &customer); err != nil {
		return domain.AdAccount{}, domain.ErrAccountNotAccessible
	}
	return domain.AdAccount{
		ExternalID: id,
		Name:       customer.Customer.DescriptiveName,
		Currency:   customer.Customer.CurrencyCode,
		Timezone:   customer.Customer.TimeZone,
		Status:     "active",
	}, nil
}

// ---------- Campagnes (GAQL + mutate) ----------

func (g *GoogleAds) ListCampaigns(ctx context.Context, conn domain.AdPlatformConnection, q port.CampaignQuery) ([]domain.Campaign, error) {
	_ = q
	id := customerID(conn.ExternalAccountID)
	query := "SELECT campaign.id, campaign.name, campaign.status, campaign_budget.amount_micros, customer.currency_code " +
		"FROM campaign WHERE campaign.status != 'REMOVED'"
	rows, err := g.searchStream(ctx, conn.AccessToken, id, query)
	if err != nil {
		return nil, err
	}
	campaigns := make([]domain.Campaign, 0)
	for _, raw := range rows {
		var row struct {
			Campaign struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"campaign"`
			CampaignBudget struct {
				AmountMicros json.Number `json:"amountMicros"`
			} `json:"campaignBudget"`
			Customer struct {
				CurrencyCode string `json:"currencyCode"`
			} `json:"customer"`
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			continue
		}
		campaigns = append(campaigns, domain.Campaign{
			UserID:             conn.UserID,
			ConnectionID:       conn.ID,
			ExternalCampaignID: row.Campaign.ID,
			Name:               row.Campaign.Name,
			Objective:          "SEARCH",
			Status:             fromGoogleStatus(row.Campaign.Status),
			BudgetMinor:        microsToMinor(row.CampaignBudget.AmountMicros),
			Currency:           row.Customer.CurrencyCode,
		})
	}
	return campaigns, nil
}

// CreateCampaign crée budget + campagne (canal SEARCH, statut PAUSED —
// garde-fou §32 ; Google exige un CampaignBudget séparé).
func (g *GoogleAds) CreateCampaign(ctx context.Context, conn domain.AdPlatformConnection, in port.CreateCampaignInput) (domain.Campaign, error) {
	id := customerID(conn.ExternalAccountID)
	currency := in.Currency
	if currency == "" {
		currency = "XOF"
	}

	// 1. Budget journalier explicite (micros = unités mineures × 10 000).
	budgetBody := map[string]any{
		"operations": []map[string]any{{
			"create": map[string]any{
				"name":             fmt.Sprintf("Budget %s — %s", in.Name, time.Now().Format("2006-01-02")),
				"amountMicros":     fmt.Sprintf("%d", in.BudgetMinor*10_000),
				"deliveryMethod":   "STANDARD",
				"explicitlyShared": false,
			},
		}},
	}
	var budgetOut googleMutateResponse
	if err := g.mutate(ctx, conn.AccessToken, id, "/campaignBudgets:mutate", budgetBody, &budgetOut); err != nil {
		return domain.Campaign{}, err
	}
	if len(budgetOut.Results) == 0 {
		return domain.Campaign{}, fmt.Errorf("googleads: budget non créé")
	}

	// 2. Campagne (SEARCH + manual CPC, créée en pause).
	campaign := map[string]any{
		"name":                   in.Name,
		"advertisingChannelType": "SEARCH",
		"status":                 "PAUSED",
		"campaignBudget":         budgetOut.Results[0].ResourceName,
		"manualCpc":              map[string]any{"enhancedCpcEnabled": false},
		"networkSettings": map[string]any{
			"targetGoogleSearch":   true,
			"targetSearchNetwork":  true,
			"targetContentNetwork": false,
		},
	}
	campaignBody := map[string]any{
		"operations": []map[string]any{{"create": campaign}},
	}
	var campaignOut googleMutateResponse
	if err := g.mutate(ctx, conn.AccessToken, id, "/campaigns:mutate", campaignBody, &campaignOut); err != nil {
		return domain.Campaign{}, err
	}
	if len(campaignOut.Results) == 0 {
		return domain.Campaign{}, fmt.Errorf("googleads: campagne non créée")
	}

	return domain.Campaign{
		UserID:             conn.UserID,
		ConnectionID:       conn.ID,
		ExternalCampaignID: campaignIDFromResource(campaignOut.Results[0].ResourceName),
		Name:               in.Name,
		Objective:          in.Objective,
		Status:             domain.CampaignPaused,
		BudgetMinor:        in.BudgetMinor,
		Currency:           currency,
	}, nil
}

// UpdateCampaign modifie nom/budget (status: PAUSED→ENABLED via Pause/Resume).
func (g *GoogleAds) UpdateCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string, in port.UpdateCampaignInput) (domain.Campaign, error) {
	id := customerID(conn.ExternalAccountID)
	if in.BudgetMinor <= 0 && in.Name == "" {
		return domain.Campaign{ExternalCampaignID: externalID}, nil
	}
	update := map[string]any{"resourceName": resourceCampaign(id, externalID)}
	mask := []string{}
	if in.Name != "" {
		update["name"] = in.Name
		mask = append(mask, "name")
	}
	if in.BudgetMinor > 0 {
		return domain.Campaign{}, fmt.Errorf("googleads: la mise à jour du budget requiert le budget partagé (phase suivante)")
	}
	body := map[string]any{
		"operations": []map[string]any{{"update": update}},
		"updateMask": strings.Join(mask, ","),
	}
	if len(mask) == 0 {
		return domain.Campaign{ExternalCampaignID: externalID}, nil
	}
	if err := g.mutate(ctx, conn.AccessToken, id, "/campaigns:mutate", body, nil); err != nil {
		return domain.Campaign{}, err
	}
	return domain.Campaign{ExternalCampaignID: externalID}, nil
}

func (g *GoogleAds) PauseCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error {
	return g.setCampaignStatus(ctx, conn, externalID, "PAUSED")
}

func (g *GoogleAds) ResumeCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error {
	return g.setCampaignStatus(ctx, conn, externalID, "ENABLED")
}

func (g *GoogleAds) setCampaignStatus(ctx context.Context, conn domain.AdPlatformConnection, externalID, status string) error {
	id := customerID(conn.ExternalAccountID)
	body := map[string]any{
		"operations": []map[string]any{{
			"update":     map[string]any{"resourceName": resourceCampaign(id, externalID), "status": status},
			"updateMask": "status",
		}},
	}
	return g.mutate(ctx, conn.AccessToken, id, "/campaigns:mutate", body, nil)
}

// UploadCreative non supporté au MVP (voir Capabilities).
func (g *GoogleAds) UploadCreative(ctx context.Context, conn domain.AdPlatformConnection, in port.CreativeInput) (string, error) {
	return "", fmt.Errorf("googleads: upload de créatives non supporté au MVP")
}

// ---------- Insights ----------

func (g *GoogleAds) GetInsights(ctx context.Context, conn domain.AdPlatformConnection, externalCampaignID string, q port.InsightsQuery) ([]domain.Insight, error) {
	id := customerID(conn.ExternalAccountID)
	query := fmt.Sprintf(
		"SELECT campaign.id, segments.date, metrics.impressions, metrics.clicks, metrics.cost_micros, metrics.conversions "+
			"FROM campaign WHERE campaign.id = %s AND segments.date BETWEEN '%s' AND '%s'",
		externalCampaignID, q.Since, q.Until)
	rows, err := g.searchStream(ctx, conn.AccessToken, id, query)
	if err != nil {
		return nil, err
	}
	insights := make([]domain.Insight, 0)
	for _, raw := range rows {
		var row struct {
			Campaign struct {
				ID string `json:"id"`
			} `json:"campaign"`
			Segments struct {
				Date string `json:"date"`
			} `json:"segments"`
			Metrics struct {
				Impressions json.Number `json:"impressions"`
				Clicks      json.Number `json:"clicks"`
				CostMicros  json.Number `json:"costMicros"`
				Conversions json.Number `json:"conversions"`
			} `json:"metrics"`
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			continue
		}
		costMicros, _ := row.Metrics.CostMicros.Float64()
		conversions, _ := row.Metrics.Conversions.Float64()
		impressions, _ := row.Metrics.Impressions.Int64()
		clicks, _ := row.Metrics.Clicks.Int64()
		insights = append(insights, domain.Insight{
			Date:        row.Segments.Date,
			Impressions: impressions,
			Clicks:      clicks,
			// Micros → unités mineures (2 décimales) : micros / 10 000.
			SpendMinor:  int64(costMicros / 10_000),
			Conversions: conversions,
		})
	}
	return insights, nil
}

// ---------- HTTP interne ----------

type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type googleMutateResponse struct {
	Results []struct {
		ResourceName string `json:"resourceName"`
	} `json:"results"`
}

type googleErrorResponse struct {
	Err struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func (e googleErrorResponse) Error() string {
	return fmt.Sprintf("googleads: %s (code %d)", e.Err.Message, e.Err.Code)
}

// Transient : erreurs serveur / quota / indisponibilité.
func (e googleErrorResponse) Transient() bool {
	return e.Err.Code == 429 || e.Err.Code >= 500 || e.Err.Code == 408
}

func (g *GoogleAds) token(ctx context.Context, form url.Values) (googleTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.tokenBase+"/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return googleTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var out googleTokenResponse
	if err := g.do(req, &out, ""); err != nil {
		return googleTokenResponse{}, err
	}
	return out, nil
}

func (g *GoogleAds) getUserInfo(ctx context.Context, accessToken string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.userInfoBase+"/oauth2/v2/userinfo", nil)
	if err != nil {
		return err
	}
	return g.do(req, out, accessToken)
}

// searchStream exécute une requête GAQL et renvoie les lignes JSON brutes.
func (g *GoogleAds) searchStream(ctx context.Context, accessToken, customerID, query string) ([]json.RawMessage, error) {
	payload, _ := json.Marshal(map[string]string{"query": query})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.apiBase+"/"+g.apiVersion+"/customers/"+customerID+"/googleAds:searchStream",
		bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if g.loginCustomerID != "" {
		req.Header.Set("login-customer-id", g.loginCustomerID)
	}
	var stream []struct {
		Results []json.RawMessage `json:"results"`
	}
	if err := g.do(req, &stream, accessToken); err != nil {
		return nil, err
	}
	rows := make([]json.RawMessage, 0)
	for _, chunk := range stream {
		rows = append(rows, chunk.Results...)
	}
	return rows, nil
}

func (g *GoogleAds) mutate(ctx context.Context, accessToken, customerID, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.apiBase+"/"+g.apiVersion+"/customers/"+customerID+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if g.loginCustomerID != "" {
		req.Header.Set("login-customer-id", g.loginCustomerID)
	}
	return g.do(req, out, accessToken)
}

func (g *GoogleAds) do(req *http.Request, out any, accessToken string) error {
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("developer-token", g.developerToken)
	resp, err := g.http.Do(req)
	if err != nil {
		return fmt.Errorf("googleads: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return fmt.Errorf("googleads: lecture: %w", err)
	}
	if resp.StatusCode >= 400 {
		var apiErr googleErrorResponse
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Err.Message != "" {
			return apiErr
		}
		return fmt.Errorf("googleads: status %d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

// ---------- helpers ----------

func customerID(resourceOrID string) string {
	s := strings.TrimPrefix(resourceOrID, "customers/")
	return strings.TrimPrefix(s, "/")
}

func resourceCampaign(customerID, campaignID string) string {
	return fmt.Sprintf("customers/%s/campaigns/%s", customerID, campaignID)
}

func campaignIDFromResource(rn string) string {
	parts := strings.Split(rn, "/")
	return parts[len(parts)-1]
}

// microsToMinor convertit des micros (1e-6) en unités mineures (1e-2).
func microsToMinor(m json.Number) int64 {
	f, err := m.Float64()
	if err != nil {
		return 0
	}
	return int64(f / 10_000)
}

func fromGoogleStatus(status string) string {
	switch strings.ToUpper(status) {
	case "ENABLED":
		return domain.CampaignActive
	case "PAUSED":
		return domain.CampaignPaused
	case "REMOVED":
		return domain.CampaignDeleted
	default:
		return domain.CampaignDraft
	}
}
