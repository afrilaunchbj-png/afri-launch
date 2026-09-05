// Package tiktok encapsule TOUTE la logique TikTok Ads (OAuth auth_code,
// advertiser discovery, campagnes, status, reporting) derrière le port
// AdPlatformProvider. Les créatives vidéo (file/video/ad/upload + adgroup/ad)
// ne sont pas supportées au MVP (Capabilities.Creatives=false).
package tiktok

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
	defaultAPIVersion = "v1.3"
)

// TikTok implémente port.AdPlatformProvider pour TikTok Ads (Business API).
type TikTok struct {
	appID       string
	appSecret   string
	apiVersion  string
	oauthBase   string
	apiBase     string
	redirectURI string
	http        *http.Client
}

// New construit le provider TikTok Ads.
func New(appID, appSecret, redirectURI string) *TikTok {
	return &TikTok{
		appID:       appID,
		appSecret:   appSecret,
		apiVersion:  defaultAPIVersion,
		oauthBase:   "https://business-api.tiktok.com",
		apiBase:     "https://business-api.tiktok.com",
		redirectURI: redirectURI,
		http:        &http.Client{Timeout: 60 * time.Second},
	}
}

func (t *TikTok) Provider() string { return domain.AdPlatformTikTokAds }

func (t *TikTok) Capabilities() port.AdPlatformCapabilities {
	return port.AdPlatformCapabilities{
		Campaigns:        true,
		Creatives:        false, // upload vidéo + adgroup/ad : phase suivante
		VideoAds:         false,
		ImageAds:         false,
		Reporting:        true,
		BudgetManagement: true,
	}
}

// ---------- OAuth (auth_code échangé contre access_token + refresh_token) ----------

func (t *TikTok) AuthorizationURL(state, redirectURI string) string {
	q := url.Values{}
	q.Set("app_id", t.appID)
	q.Set("state", state)
	q.Set("redirect_uri", redirectURI)
	return t.oauthBase + "/portal/auth?" + q.Encode()
}

func (t *TikTok) ExchangeCode(ctx context.Context, code, redirectURI string) (port.OAuthTokenResult, error) {
	body := map[string]any{
		"app_id":     t.appID,
		"secret":     t.appSecret,
		"auth_code":  code,
		"grant_type": "authorization_code",
	}
	var out struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
			OpenID       string `json:"open_id"`
		} `json:"data"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := t.post(ctx, "/oauth2/access_token/", body, "", &out); err != nil {
		return port.OAuthTokenResult{}, err
	}
	if out.Code != 0 {
		return port.OAuthTokenResult{}, fmt.Errorf("tiktok: exchange: %s (code %d)", out.Message, out.Code)
	}
	res := port.OAuthTokenResult{
		AccessToken:    out.Data.AccessToken,
		RefreshToken:   out.Data.RefreshToken,
		ExternalUserID: out.Data.OpenID,
	}
	if out.Data.ExpiresIn > 0 {
		at := time.Now().Add(time.Duration(out.Data.ExpiresIn) * time.Second)
		res.ExpiresAt = &at
	}
	return res, nil
}

// Refresh renouvelle l'access token (TikTok fournit un refresh_token).
func (t *TikTok) Refresh(ctx context.Context, conn domain.AdPlatformConnection) (port.OAuthTokenResult, bool, error) {
	if conn.RefreshToken == "" {
		return port.OAuthTokenResult{}, false, fmt.Errorf("tiktok: refresh token manquant (reconnexion requise)")
	}
	body := map[string]any{
		"app_id":        t.appID,
		"secret":        t.appSecret,
		"grant_type":    "refresh_token",
		"refresh_token": conn.RefreshToken,
	}
	var out struct {
		Data struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int64  `json:"expires_in"`
		} `json:"data"`
		Code int `json:"code"`
	}
	if err := t.post(ctx, "/oauth2/refresh_token/", body, "", &out); err != nil {
		return port.OAuthTokenResult{}, true, err
	}
	if out.Code != 0 {
		return port.OAuthTokenResult{}, true, fmt.Errorf("tiktok: refresh: code %d", out.Code)
	}
	res := port.OAuthTokenResult{AccessToken: out.Data.AccessToken, RefreshToken: conn.RefreshToken}
	if out.Data.ExpiresIn > 0 {
		at := time.Now().Add(time.Duration(out.Data.ExpiresIn) * time.Second)
		res.ExpiresAt = &at
	}
	return res, true, nil
}

// Revoke : la Business API n'expose pas de révocation programmatique
// documentée au MVP — la déconnexion locale suffit (statut disconnected).
func (t *TikTok) Revoke(ctx context.Context, accessToken string) error { return nil }

// ---------- Comptes (advertisers liés au token) ----------

func (t *TikTok) ListAdAccounts(ctx context.Context, accessToken string) ([]domain.AdAccount, error) {
	var out struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				AdvertiserID   string `json:"advertiser_id"`
				AdvertiserName string `json:"advertiser_name"`
				Currency       string `json:"currency"`
				Timezone       string `json:"timezone"`
			} `json:"list"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := t.get(ctx, "/oauth2/advertiser/get/", url.Values{}, accessToken, &out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("tiktok: advertiser/get: %s (code %d)", out.Message, out.Code)
	}
	accounts := make([]domain.AdAccount, 0, len(out.Data.List))
	for _, a := range out.Data.List {
		accounts = append(accounts, domain.AdAccount{
			ExternalID: a.AdvertiserID,
			Name:       a.AdvertiserName,
			Currency:   a.Currency,
			Timezone:   a.Timezone,
			Status:     "active",
		})
	}
	return accounts, nil
}

// VerifyAdAccount vérifie que le compte appartient bien à la connexion.
func (t *TikTok) VerifyAdAccount(ctx context.Context, accessToken, externalAccountID string) (domain.AdAccount, error) {
	accounts, err := t.ListAdAccounts(ctx, accessToken)
	if err != nil {
		return domain.AdAccount{}, domain.ErrAccountNotAccessible
	}
	for _, a := range accounts {
		if a.ExternalID == externalAccountID {
			return a, nil
		}
	}
	return domain.AdAccount{}, domain.ErrAccountNotAccessible
}

// ---------- Campagnes ----------

func (t *TikTok) ListCampaigns(ctx context.Context, conn domain.AdPlatformConnection, q port.CampaignQuery) ([]domain.Campaign, error) {
	qv := url.Values{}
	qv.Set("advertiser_id", conn.ExternalAccountID)
	qv.Set("page", "1")
	qv.Set("page_size", "100")
	var out struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				CampaignID    string  `json:"campaign_id"`
				CampaignName  string  `json:"campaign_name"`
				ObjectiveType string  `json:"objective_type"`
				Status        string  `json:"status"`
				Budget        float64 `json:"budget"`
				BudgetMode    string  `json:"budget_mode"`
			} `json:"list"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := t.get(ctx, "/campaign/get/", qv, conn.AccessToken, &out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("tiktok: campaign/get: %s (code %d)", out.Message, out.Code)
	}
	campaigns := make([]domain.Campaign, 0, len(out.Data.List))
	for _, c := range out.Data.List {
		if fromTikTokStatus(c.Status) == domain.CampaignDeleted {
			continue
		}
		campaigns = append(campaigns, domain.Campaign{
			UserID:             conn.UserID,
			ConnectionID:       conn.ID,
			ExternalCampaignID: c.CampaignID,
			Name:               c.CampaignName,
			Objective:          c.ObjectiveType,
			Status:             fromTikTokStatus(c.Status),
			// Budget en unités majeures chez TikTok → mineures (×100).
			BudgetMinor: int64(c.Budget * 100),
			Currency:    currencyFromMetadata(conn.Metadata),
		})
	}
	return campaigns, nil
}

// CreateCampaign crée la campagne (statut DISABLE = en pause — garde-fou §32).
func (t *TikTok) CreateCampaign(ctx context.Context, conn domain.AdPlatformConnection, in port.CreateCampaignInput) (domain.Campaign, error) {
	body := map[string]any{
		"advertiser_id":  conn.ExternalAccountID,
		"campaign_name":  in.Name,
		"objective_type": objectiveType(in.Objective),
		"status":         "CAMPAIGN_STATUS_DISABLE",
	}
	if in.BudgetMinor > 0 {
		body["budget_mode"] = "BUDGET_MODE_DAY"
		body["budget"] = float64(in.BudgetMinor) / 100 // mineures → majeures
	}
	var out struct {
		Code int `json:"code"`
		Data struct {
			CampaignID string `json:"campaign_id"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := t.post(ctx, "/campaign/create/", body, conn.AccessToken, &out); err != nil {
		return domain.Campaign{}, err
	}
	if out.Code != 0 {
		return domain.Campaign{}, fmt.Errorf("tiktok: campaign/create: %s (code %d)", out.Message, out.Code)
	}
	return domain.Campaign{
		UserID:             conn.UserID,
		ConnectionID:       conn.ID,
		ExternalCampaignID: out.Data.CampaignID,
		Name:               in.Name,
		Objective:          in.Objective,
		Status:             domain.CampaignPaused,
		BudgetMinor:        in.BudgetMinor,
		Currency:           in.Currency,
	}, nil
}

func (t *TikTok) UpdateCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string, in port.UpdateCampaignInput) (domain.Campaign, error) {
	if in.Name == "" && in.BudgetMinor <= 0 {
		return domain.Campaign{ExternalCampaignID: externalID}, nil
	}
	body := map[string]any{
		"advertiser_id": conn.ExternalAccountID,
		"campaign_id":   externalID,
	}
	if in.Name != "" {
		body["campaign_name"] = in.Name
	}
	if in.BudgetMinor > 0 {
		body["budget"] = float64(in.BudgetMinor) / 100
		body["budget_mode"] = "BUDGET_MODE_DAY"
	}
	var out struct {
		Code int `json:"code"`
	}
	if err := t.post(ctx, "/campaign/update/", body, conn.AccessToken, &out); err != nil {
		return domain.Campaign{}, err
	}
	if out.Code != 0 {
		return domain.Campaign{}, fmt.Errorf("tiktok: campaign/update: code %d", out.Code)
	}
	return domain.Campaign{ExternalCampaignID: externalID}, nil
}

func (t *TikTok) PauseCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error {
	return t.updateStatus(ctx, conn, externalID, "DISABLE")
}

func (t *TikTok) ResumeCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error {
	return t.updateStatus(ctx, conn, externalID, "ENABLE")
}

func (t *TikTok) updateStatus(ctx context.Context, conn domain.AdPlatformConnection, externalID, operation string) error {
	body := map[string]any{
		"advertiser_id":  conn.ExternalAccountID,
		"campaign_ids":   []string{externalID},
		"operation_type": operation,
	}
	var out struct {
		Code int `json:"code"`
	}
	if err := t.post(ctx, "/campaign/status/update/", body, conn.AccessToken, &out); err != nil {
		return err
	}
	if out.Code != 0 {
		return fmt.Errorf("tiktok: status/update: code %d", out.Code)
	}
	return nil
}

// UploadCreative non supporté au MVP (voir Capabilities).
func (t *TikTok) UploadCreative(ctx context.Context, conn domain.AdPlatformConnection, in port.CreativeInput) (string, error) {
	return "", fmt.Errorf("tiktok: upload de créatives non supporté au MVP")
}

// ---------- Insights ----------

func (t *TikTok) GetInsights(ctx context.Context, conn domain.AdPlatformConnection, externalCampaignID string, q port.InsightsQuery) ([]domain.Insight, error) {
	qv := url.Values{}
	qv.Set("report_type", "BASIC")
	qv.Set("data_level", "AUCTION_CAMPAIGN")
	qv.Set("start_date", q.Since)
	qv.Set("end_date", q.Until)
	qv.Set("dimensions", `["campaign_id"]`)
	qv.Set("metrics", `["impressions","clicks","spend","conversion"]`)
	qv.Set("filters", fmt.Sprintf(`[{"field":"campaign_id","operator":"EQUAL","value":[%q]}]`, externalCampaignID))

	var out struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				Dimensions struct {
					CampaignID string `json:"campaign_id"`
				} `json:"dimensions"`
				Metrics struct {
					Impressions string `json:"impressions"`
					Clicks      string `json:"clicks"`
					Spend       string `json:"spend"`
					Conversion  string `json:"conversion"`
				} `json:"metrics"`
			} `json:"list"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := t.get(ctx, "/report/integrated/get/", qv, conn.AccessToken, &out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("tiktok: report: %s (code %d)", out.Message, out.Code)
	}
	insights := make([]domain.Insight, 0, len(out.Data.List))
	for _, row := range out.Data.List {
		// Spend en unités majeures (string) → mineures.
		spend, _ := parseFloat(row.Metrics.Spend)
		conv, _ := parseFloat(row.Metrics.Conversion)
		impressions, _ := parseFloat(row.Metrics.Impressions)
		clicks, _ := parseFloat(row.Metrics.Clicks)
		date := ""
		// La dimension date est absente si data_level campagne agrégé :
		// TikTok renvoie une ligne par jour avec segments via data_level.
		insights = append(insights, domain.Insight{
			Date:        date,
			Impressions: int64(impressions),
			Clicks:      int64(clicks),
			SpendMinor:  int64(spend * 100),
			Conversions: conv,
		})
	}
	return insights, nil
}

// ---------- HTTP interne ----------

func (t *TikTok) get(ctx context.Context, path string, query url.Values, accessToken string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		t.apiBase+"/open_api/"+t.apiVersion+path+"?"+query.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Access-Token", accessToken)
	return t.do(req, out)
}

func (t *TikTok) post(ctx context.Context, path string, body any, accessToken string, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		t.apiBase+"/open_api/"+t.apiVersion+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Access-Token", accessToken)
	}
	return t.do(req, out)
}

func (t *TikTok) do(req *http.Request, out any) error {
	resp, err := t.http.Do(req)
	if err != nil {
		return fmt.Errorf("tiktok: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return fmt.Errorf("tiktok: lecture: %w", err)
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("tiktok: décodage: %w", err)
		}
	}
	return nil
}

// ---------- helpers ----------

func parseFloat(s string) (float64, error) {
	return strconvParseFloat(strings.TrimSpace(s))
}

func strconvParseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	var f float64
	_, err := fmt.Sscanf(s, "%g", &f)
	return f, err
}

func fromTikTokStatus(status string) string {
	switch status {
	case "CAMPAIGN_STATUS_ENABLE", "STATUS_ENABLE":
		return domain.CampaignActive
	case "CAMPAIGN_STATUS_DISABLE", "STATUS_DISABLE":
		return domain.CampaignPaused
	case "DELETE", "STATUS_DELETE", "ALL_STATUS_TOTAL_DELETE":
		return domain.CampaignDeleted
	default:
		return domain.CampaignDraft
	}
}

func objectiveType(objective string) string {
	switch strings.ToUpper(objective) {
	case "OUTCOME_TRAFFIC", "TRAFFIC":
		return "TRAFFIC"
	case "OUTCOME_SALES", "CONVERSIONS":
		return "CONVERSIONS"
	case "OUTCOME_LEADS", "LEAD_GENERATION":
		return "LEAD_GENERATION"
	case "OUTCOME_ENGAGEMENT", "ENGAGEMENT":
		return "ENGAGEMENT"
	case "OUTCOME_AWARENESS", "REACH":
		return "REACH"
	default:
		return "TRAFFIC"
	}
}

func currencyFromMetadata(metadata []byte) string {
	var md map[string]any
	if json.Unmarshal(metadata, &md) == nil {
		if c, ok := md["currency"].(string); ok {
			return c
		}
	}
	return "USD"
}
