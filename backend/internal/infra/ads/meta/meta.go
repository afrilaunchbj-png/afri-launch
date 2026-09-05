// Package meta encapsule TOUTE la logique Meta Marketing API (OAuth Graph,
// ad accounts, campagnes, creatives vidéo, insights) derrière le port
// AdPlatformProvider. Aucune logique Meta ne doit fuiter dans le métier
// (prompts/marketing-flow.md §1, §12).
package meta

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
	defaultGraphVersion = "v23.0"
	defaultOAuthBase    = "https://www.facebook.com"
	defaultGraphBase    = "https://graph.facebook.com"
)

// Meta implémente port.AdPlatformProvider pour Facebook/Instagram Ads.
type Meta struct {
	appID        string
	appSecret    string
	graphVersion string
	oauthBase    string
	graphBase    string
	redirectURI  string
	scopes       string
	http         *http.Client
}

// Scopes par défaut (configurables via META_OAUTH_SCOPES).
const defaultScopes = "ads_management,ads_read,business_management,pages_show_list,pages_read_engagement"

// New construit le provider Meta.
func New(appID, appSecret, graphVersion, redirectURI, scopes string) *Meta {
	if graphVersion == "" {
		graphVersion = defaultGraphVersion
	}
	if scopes == "" {
		scopes = defaultScopes
	}
	return &Meta{
		appID:        appID,
		appSecret:    appSecret,
		graphVersion: graphVersion,
		oauthBase:    defaultOAuthBase,
		graphBase:    defaultGraphBase,
		redirectURI:  redirectURI,
		scopes:       scopes,
		http:         &http.Client{Timeout: 60 * time.Second},
	}
}

// Provider identifie la plateforme.
func (m *Meta) Provider() string { return domain.AdPlatformMeta }

// Capabilities décrit les capacités Meta.
func (m *Meta) Capabilities() port.AdPlatformCapabilities {
	return port.AdPlatformCapabilities{
		Campaigns:          true,
		Creatives:          true,
		VideoAds:           true,
		ImageAds:           true,
		ConversionTracking: true,
		Reporting:          true,
		BudgetManagement:   true,
	}
}

// AuthorizationURL construit l'URL du dialogue OAuth Facebook.
func (m *Meta) AuthorizationURL(state, redirectURI string) string {
	q := url.Values{}
	q.Set("client_id", m.appID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("response_type", "code")
	q.Set("scope", m.scopes)
	return m.oauthBase + "/" + m.graphVersion + "/dialog/oauth?" + q.Encode()
}

// ExchangeCode échange le code OAuth contre un token longue durée
// (fb_exchange_token, ~60 jours ; pas de refresh token chez Meta).
func (m *Meta) ExchangeCode(ctx context.Context, code, redirectURI string) (port.OAuthTokenResult, error) {
	var short struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	q := url.Values{}
	q.Set("client_id", m.appID)
	q.Set("client_secret", m.appSecret)
	q.Set("redirect_uri", redirectURI)
	q.Set("code", code)
	if err := m.get(ctx, "/oauth/access_token", q, &short); err != nil {
		return port.OAuthTokenResult{}, fmt.Errorf("meta: exchange: %w", err)
	}

	// Prolonge en token longue durée.
	long := struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}{}
	q2 := url.Values{}
	q2.Set("grant_type", "fb_exchange_token")
	q2.Set("client_id", m.appID)
	q2.Set("client_secret", m.appSecret)
	q2.Set("fb_exchange_token", short.AccessToken)
	if err := m.get(ctx, "/oauth/access_token", q2, &long); err != nil {
		// On garde le token court si la prolongation échoue.
		long.AccessToken = short.AccessToken
	}

	res := port.OAuthTokenResult{AccessToken: long.AccessToken}
	if long.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(long.ExpiresIn) * time.Second)
		res.ExpiresAt = &t
	}
	// Identité de l'utilisateur Meta.
	var me struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := m.getWithToken(ctx, "/me", nil, long.AccessToken, &me); err == nil {
		res.ExternalUserID = me.ID
	}
	return res, nil
}

// Refresh prolonge le token courant (Meta n'a pas de refresh token dédié).
func (m *Meta) Refresh(ctx context.Context, conn domain.AdPlatformConnection) (port.OAuthTokenResult, bool, error) {
	res, err := m.ExchangeCodeLongLived(ctx, conn.AccessToken)
	return res, true, err
}

// ExchangeCodeLongLived prolonge un token existant.
func (m *Meta) ExchangeCodeLongLived(ctx context.Context, token string) (port.OAuthTokenResult, error) {
	long := struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}{}
	q := url.Values{}
	q.Set("grant_type", "fb_exchange_token")
	q.Set("client_id", m.appID)
	q.Set("client_secret", m.appSecret)
	q.Set("fb_exchange_token", token)
	if err := m.get(ctx, "/oauth/access_token", q, &long); err != nil {
		return port.OAuthTokenResult{}, fmt.Errorf("meta: refresh: %w", err)
	}
	res := port.OAuthTokenResult{AccessToken: long.AccessToken}
	if long.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(long.ExpiresIn) * time.Second)
		res.ExpiresAt = &t
	}
	return res, nil
}

// Revoke révoque toutes les permissions accordées.
func (m *Meta) Revoke(ctx context.Context, accessToken string) error {
	return m.delete(ctx, "/me/permissions", accessToken, nil)
}

// ListAdAccounts découvre les comptes publicitaires accessibles.
func (m *Meta) ListAdAccounts(ctx context.Context, accessToken string) ([]domain.AdAccount, error) {
	var out struct {
		Data []struct {
			ID             string  `json:"id"`
			Name           string  `json:"name"`
			AccountStatus  int     `json:"account_status"`
			Currency       string  `json:"currency"`
			TimezoneOffset float64 `json:"timezone_offset_hours"`
		} `json:"data"`
	}
	q := url.Values{}
	q.Set("fields", "id,name,account_status,currency,timezone_offset_hours")
	q.Set("limit", "100")
	if err := m.getWithToken(ctx, "/me/adaccounts", q, accessToken, &out); err != nil {
		return nil, err
	}
	accounts := make([]domain.AdAccount, 0, len(out.Data))
	for _, a := range out.Data {
		accounts = append(accounts, domain.AdAccount{
			ExternalID: a.ID,
			Name:       a.Name,
			Currency:   a.Currency,
			Status:     adAccountStatus(a.AccountStatus),
		})
	}
	return accounts, nil
}

// VerifyAdAccount vérifie que le compte est réellement accessible (l'ID
// reçu du frontend n'est jamais fiable).
func (m *Meta) VerifyAdAccount(ctx context.Context, accessToken, externalAccountID string) (domain.AdAccount, error) {
	if !strings.HasPrefix(externalAccountID, "act_") {
		return domain.AdAccount{}, domain.ErrAccountNotAccessible
	}
	var out struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Currency string `json:"currency"`
	}
	q := url.Values{}
	q.Set("fields", "id,name,currency")
	if err := m.getWithToken(ctx, "/"+externalAccountID, q, accessToken, &out); err != nil {
		return domain.AdAccount{}, domain.ErrAccountNotAccessible
	}
	return domain.AdAccount{ExternalID: out.ID, Name: out.Name, Currency: out.Currency}, nil
}

// ListCampaigns liste les campagnes d'un compte.
func (m *Meta) ListCampaigns(ctx context.Context, conn domain.AdPlatformConnection, q port.CampaignQuery) ([]domain.Campaign, error) {
	_ = q // filtrage côté application (repo/service) au MVP
	account := actID(conn.ExternalAccountID)
	var out struct {
		Data []struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Objective       string `json:"objective"`
			Status          string `json:"status"`
			EffectiveStatus string `json:"effective_status"`
			DailyBudget     int64  `json:"daily_budget"`
		} `json:"data"`
	}
	query := url.Values{}
	query.Set("fields", "id,name,objective,status,effective_status,daily_budget")
	query.Set("limit", "100")
	if err := m.getWithToken(ctx, "/"+account+"/campaigns", query, conn.AccessToken, &out); err != nil {
		return nil, err
	}
	campaigns := make([]domain.Campaign, 0, len(out.Data))
	for _, c := range out.Data {
		campaigns = append(campaigns, domain.Campaign{
			UserID:             conn.UserID,
			ConnectionID:       conn.ID,
			ExternalCampaignID: c.ID,
			Name:               c.Name,
			Objective:          c.Objective,
			Status:             fromMetaStatus(c.Status, c.EffectiveStatus),
			BudgetMinor:        c.DailyBudget,
			Currency:           currencyFromMetadata(conn.Metadata),
		})
	}
	return campaigns, nil
}

// CreateCampaign crée la campagne (toujours en PAUSED — garde-fou §32 :
// la publication reste un geste utilisateur explicite).
func (m *Meta) CreateCampaign(ctx context.Context, conn domain.AdPlatformConnection, in port.CreateCampaignInput) (domain.Campaign, error) {
	account := actID(conn.ExternalAccountID)
	body := map[string]any{
		"name":                  in.Name,
		"objective":             in.Objective,
		"status":                "PAUSED",
		"special_ad_categories": []string{},
	}
	if in.BudgetMinor > 0 {
		body["daily_budget"] = in.BudgetMinor
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := m.post(ctx, "/"+account+"/campaigns", conn.AccessToken, body, &out); err != nil {
		return domain.Campaign{}, err
	}
	return domain.Campaign{
		UserID:             conn.UserID,
		ConnectionID:       conn.ID,
		ExternalCampaignID: out.ID,
		Name:               in.Name,
		Objective:          in.Objective,
		Status:             domain.CampaignPaused,
		BudgetMinor:        in.BudgetMinor,
		Currency:           in.Currency,
	}, nil
}

// UpdateCampaign modifie nom/budget.
func (m *Meta) UpdateCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string, in port.UpdateCampaignInput) (domain.Campaign, error) {
	body := map[string]any{}
	if in.Name != "" {
		body["name"] = in.Name
	}
	if in.BudgetMinor > 0 {
		body["daily_budget"] = in.BudgetMinor
	}
	if len(body) > 0 {
		if err := m.post(ctx, "/"+externalID, conn.AccessToken, body, nil); err != nil {
			return domain.Campaign{}, err
		}
	}
	return domain.Campaign{ExternalCampaignID: externalID}, nil
}

// PauseCampaign met la campagne en pause.
func (m *Meta) PauseCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error {
	return m.post(ctx, "/"+externalID, conn.AccessToken, map[string]any{"status": "PAUSED"}, nil)
}

// ResumeCampaign réactive la campagne.
func (m *Meta) ResumeCampaign(ctx context.Context, conn domain.AdPlatformConnection, externalID string) error {
	return m.post(ctx, "/"+externalID, conn.AccessToken, map[string]any{"status": "ACTIVE"}, nil)
}

// UploadCreative publie une vidéo (advideos puis adcreatives). La page est
// résolue depuis les métadonnées de la connexion (requis par object_story_spec).
func (m *Meta) UploadCreative(ctx context.Context, conn domain.AdPlatformConnection, in port.CreativeInput) (string, error) {
	if in.Type != domain.CreativeVideo {
		return "", fmt.Errorf("meta: type de creative non supporté au MVP : %q", in.Type)
	}
	pageID := pageIDFromMetadata(conn.Metadata)
	if pageID == "" {
		return "", fmt.Errorf("meta: page Facebook requise (page_id manquant dans la connexion)")
	}
	account := actID(conn.ExternalAccountID)

	// 1. Upload de la vidéo.
	var videoID string
	if strings.HasPrefix(in.URL, "http://") || strings.HasPrefix(in.URL, "https://") {
		var up struct {
			ID string `json:"id"`
		}
		if err := m.postForm(ctx, "/"+account+"/advideos", conn.AccessToken,
			url.Values{"file_url": {in.URL}}, &up); err != nil {
			return "", err
		}
		videoID = up.ID
	} else {
		return "", fmt.Errorf("meta: source vidéo invalide")
	}

	// 2. Creative (object_story_spec vidéo).
	story := map[string]any{
		"page_id": pageID,
		"video_data": map[string]any{
			"video_id": videoID,
			"message":  in.PrimaryText,
			"title":    in.Headline,
			"call_to_action": map[string]any{
				"type":  ctaType(in.CTA),
				"value": map[string]any{"link": "https://afrilaunch.com"},
			},
		},
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := m.post(ctx, "/"+account+"/adcreatives", conn.AccessToken,
		map[string]any{"object_story_spec": story}, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// GetInsights récupère les métriques normalisées d'une campagne.
func (m *Meta) GetInsights(ctx context.Context, conn domain.AdPlatformConnection, externalCampaignID string, q port.InsightsQuery) ([]domain.Insight, error) {
	var out struct {
		Data []struct {
			DateStart   string `json:"date_start"`
			DateStop    string `json:"date_stop"`
			Impressions string `json:"impressions"`
			Reach       string `json:"reach"`
			Clicks      string `json:"clicks"`
			Spend       string `json:"spend"`
			CPC         string `json:"cpc"`
			CPM         string `json:"cpm"`
			CTR         string `json:"ctr"`
			Actions     []struct {
				ActionType string `json:"action_type"`
				Value      string `json:"value"`
			} `json:"actions"`
		} `json:"data"`
	}
	query := url.Values{}
	query.Set("fields", "impressions,reach,clicks,spend,cpc,cpm,ctr,actions")
	query.Set("level", "campaign")
	query.Set("time_increment", "1")
	tr := map[string]string{"since": q.Since, "until": q.Until}
	raw, _ := json.Marshal(tr)
	query.Set("time_range", string(raw))
	if err := m.getWithToken(ctx, "/"+externalCampaignID+"/insights", query, conn.AccessToken, &out); err != nil {
		return nil, err
	}
	insights := make([]domain.Insight, 0, len(out.Data))
	for _, d := range out.Data {
		conv := 0.0
		for _, a := range d.Actions {
			if strings.Contains(a.ActionType, "purchase") || a.ActionType == "lead" {
				conv += parseFloat(a.Value)
			}
		}
		spendMinor := int64(parseFloat(d.Spend) * 100)
		insights = append(insights, domain.Insight{
			Date:        d.DateStart,
			Impressions: int64(parseFloat(d.Impressions)),
			Reach:       int64(parseFloat(d.Reach)),
			Clicks:      int64(parseFloat(d.Clicks)),
			SpendMinor:  spendMinor,
			CTR:         parseFloat(d.CTR),
			Conversions: conv,
			Metadata:    rawInsights(d),
		})
	}
	return insights, nil
}

// ---------- HTTP interne ----------

type graphError struct {
	Code    int    `json:"code"`
	Subcode int    `json:"error_subcode"`
	Type    string `json:"type"`
	Message string `json:"message"`
	FBTrace string `json:"fbtrace_id"`
}

func (e graphError) Error() string {
	return fmt.Sprintf("meta graph: %s (code %d)", e.Message, e.Code)
}

// Transient indique si l'erreur mérite un retry (rate limit / indispo).
func (e graphError) Transient() bool {
	return e.Code == 1 || e.Code == 2 || e.Code == 4 || e.Code == 17 || e.Code == 32 || e.Code == 613
}

func (m *Meta) get(ctx context.Context, path string, query url.Values, out any) error {
	return m.getWithToken(ctx, path, query, "", out)
}

func (m *Meta) getWithToken(ctx context.Context, path string, query url.Values, token string, out any) error {
	if query == nil {
		query = url.Values{}
	}
	if token != "" {
		query.Set("access_token", token)
	} else {
		query.Set("access_token", m.appID+"|"+m.appSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.graphBase+"/"+m.graphVersion+path+"?"+query.Encode(), nil)
	if err != nil {
		return err
	}
	return m.do(req, out)
}

func (m *Meta) post(ctx context.Context, path, token string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		m.graphBase+"/"+m.graphVersion+path+"?access_token="+url.QueryEscape(token), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return m.do(req, out)
}

// postForm envoie des champs form-urlencoded (ex. file_url pour advideos).
func (m *Meta) postForm(ctx context.Context, path, token string, form url.Values, out any) error {
	form.Set("access_token", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		m.graphBase+"/"+m.graphVersion+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return m.do(req, out)
}

func (m *Meta) delete(ctx context.Context, path, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		m.graphBase+"/"+m.graphVersion+path+"?access_token="+url.QueryEscape(token), nil)
	if err != nil {
		return err
	}
	return m.do(req, out)
}

func (m *Meta) do(req *http.Request, out any) error {
	resp, err := m.http.Do(req)
	if err != nil {
		return fmt.Errorf("meta: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return fmt.Errorf("meta: lecture: %w", err)
	}
	if resp.StatusCode >= 400 {
		var wrapper struct {
			Error *graphError `json:"error"`
		}
		if json.Unmarshal(data, &wrapper) == nil && wrapper.Error != nil {
			return *wrapper.Error
		}
		return fmt.Errorf("meta: status %d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

// ---------- helpers ----------

func actID(account string) string { return strings.TrimPrefix(account, "/") }
func parseFloat(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(strings.TrimSpace(s), "%g", &f)
	return f
}

func adAccountStatus(code int) string {
	switch code {
	case 1:
		return "active"
	case 100, 101, 201, 202:
		return "suspended"
	default:
		return "unknown"
	}
}

func fromMetaStatus(status, effective string) string {
	effectiveUpper := strings.ToUpper(effective)
	statusUpper := strings.ToUpper(status)
	switch statusUpper {
	case "DELETED", "ARCHIVED":
		return domain.CampaignDeleted
	case "PAUSED":
		return domain.CampaignPaused
	case "ACTIVE":
		if strings.Contains(effectiveUpper, "PAUSED") {
			return domain.CampaignPaused
		}
		return domain.CampaignActive
	default:
		switch {
		case strings.Contains(effectiveUpper, "PAUSED"):
			return domain.CampaignPaused
		case effectiveUpper == "ACTIVE" || effectiveUpper == "SCHEDULED":
			return domain.CampaignActive
		case effectiveUpper == "DELETED" || effectiveUpper == "ARCHIVED":
			return domain.CampaignDeleted
		default:
			return domain.CampaignDraft
		}
	}
}

func ctaType(cta string) string {
	switch strings.ToUpper(strings.TrimSpace(cta)) {
	case "SHOP_NOW", "BUY":
		return "SHOP_NOW"
	case "LEARN_MORE":
		return "LEARN_MORE"
	case "SIGN_UP":
		return "SIGN_UP"
	case "DOWNLOAD":
		return "DOWNLOAD"
	default:
		return "LEARN_MORE"
	}
}

func currencyFromMetadata(metadata []byte) string {
	var md map[string]any
	if json.Unmarshal(metadata, &md) == nil {
		if c, ok := md["currency"].(string); ok {
			return c
		}
	}
	return "XOF"
}

func pageIDFromMetadata(metadata []byte) string {
	var md map[string]any
	if json.Unmarshal(metadata, &md) == nil {
		if p, ok := md["page_id"].(string); ok {
			return p
		}
	}
	return ""
}

func rawInsights(d any) []byte {
	raw, _ := json.Marshal(map[string]any{"actions": d})
	return raw
}
