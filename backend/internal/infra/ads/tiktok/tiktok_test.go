package tiktok

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

func newTestProvider(t *testing.T, handler http.HandlerFunc) (*TikTok, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	p := New("app-1", "secret-1", "https://app.test/cb")
	p.apiBase = srv.URL
	p.oauthBase = srv.URL
	return p, srv
}

func TestAuthorizationURL(t *testing.T) {
	p := New("app-1", "secret-1", "https://app.test/cb")
	raw := p.AuthorizationURL("state-1", "https://app.test/cb")
	if !strings.Contains(raw, "/portal/auth") || !strings.Contains(raw, "app_id=app-1") || !strings.Contains(raw, "state=state-1") {
		t.Fatalf("url = %q", raw)
	}
}

func TestExchangeCode(t *testing.T) {
	p, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/oauth2/access_token/") {
			http.Error(w, "unexpected", 400)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["auth_code"] != "code-1" || body["grant_type"] != "authorization_code" {
			http.Error(w, "bad body", 400)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"access_token": "at-1", "refresh_token": "rt-1", "expires_in": 86400, "open_id": "open-7"},
		})
	})
	res, err := p.ExchangeCode(context.Background(), "code-1", "https://app.test/cb")
	if err != nil {
		t.Fatal(err)
	}
	if res.AccessToken != "at-1" || res.RefreshToken != "rt-1" || res.ExternalUserID != "open-7" {
		t.Fatalf("res = %+v", res)
	}
}

func TestRefresh(t *testing.T) {
	p, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["grant_type"] != "refresh_token" || body["refresh_token"] != "rt-keep" {
			http.Error(w, "bad body", 400)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": map[string]any{"access_token": "at-new", "expires_in": 86400}})
	})
	res, ok, err := p.Refresh(context.Background(), domain.AdPlatformConnection{RefreshToken: "rt-keep"})
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if res.AccessToken != "at-new" || res.RefreshToken != "rt-keep" {
		t.Fatalf("res = %+v", res)
	}
}

func TestListAdAccounts(t *testing.T) {
	p, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/oauth2/advertiser/get/") {
			http.Error(w, "unexpected", 400)
			return
		}
		if r.Header.Get("Access-Token") != "tok" {
			http.Error(w, "missing token", 401)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"list": []map[string]any{
				{"advertiser_id": "7100", "advertiser_name": "Boutique BJ", "currency": "XOF", "timezone": "Africa/Porto-Novo"},
			}},
		})
	})
	accounts, err := p.ListAdAccounts(context.Background(), "tok")
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].ExternalID != "7100" || accounts[0].Currency != "XOF" {
		t.Fatalf("accounts = %+v", accounts)
	}
}

func TestVerifyAdAccountChecksMembership(t *testing.T) {
	p, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "data": map[string]any{"list": []map[string]any{{"advertiser_id": "7100", "advertiser_name": "Boutique BJ"}}},
		})
	})
	if _, err := p.VerifyAdAccount(context.Background(), "tok", "9999"); err != domain.ErrAccountNotAccessible {
		t.Fatalf("compte étranger accepté: %v", err)
	}
	acc, err := p.VerifyAdAccount(context.Background(), "tok", "7100")
	if err != nil || acc.ExternalID != "7100" {
		t.Fatalf("acc = %+v, %v", acc, err)
	}
}

func TestCreateCampaignDisabledByDefault(t *testing.T) {
	p, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/campaign/create/") {
			http.Error(w, "unexpected", 400)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["status"] != "CAMPAIGN_STATUS_DISABLE" {
			t.Errorf("status = %v", body["status"])
		}
		if body["budget"] != 1000.0 { // 100 000 minor / 100
			t.Errorf("budget = %v", body["budget"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": map[string]any{"campaign_id": "camp-9"}})
	})
	conn := domain.AdPlatformConnection{ID: "c1", UserID: "u1", ExternalAccountID: "7100", AccessToken: "tok"}
	c, err := p.CreateCampaign(context.Background(), conn, port.CreateCampaignInput{
		Name: "Pub vidéo", Objective: "OUTCOME_TRAFFIC", BudgetMinor: 100_000, Currency: "XOF",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.ExternalCampaignID != "camp-9" || c.Status != domain.CampaignPaused || c.BudgetMinor != 100_000 {
		t.Fatalf("campaign = %+v", c)
	}
}

func TestPauseResumeCampaign(t *testing.T) {
	var operations []string
	p, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/campaign/status/update/") {
			http.Error(w, "unexpected", 400)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		operations = append(operations, body["operation_type"].(string))
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0})
	})
	conn := domain.AdPlatformConnection{ID: "c1", UserID: "u1", ExternalAccountID: "7100", AccessToken: "tok"}
	if err := p.PauseCampaign(context.Background(), conn, "camp-1"); err != nil {
		t.Fatal(err)
	}
	if err := p.ResumeCampaign(context.Background(), conn, "camp-1"); err != nil {
		t.Fatal(err)
	}
	if len(operations) != 2 || operations[0] != "DISABLE" || operations[1] != "ENABLE" {
		t.Fatalf("operations = %v", operations)
	}
}

func TestListCampaignsMapping(t *testing.T) {
	p, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"list": []map[string]any{
				{"campaign_id": "1", "campaign_name": "Active", "objective_type": "TRAFFIC", "status": "CAMPAIGN_STATUS_ENABLE", "budget": 5000},
				{"campaign_id": "2", "campaign_name": "Paused", "objective_type": "CONVERSIONS", "status": "CAMPAIGN_STATUS_DISABLE", "budget": 6000},
				{"campaign_id": "3", "campaign_name": "Deleted", "status": "DELETE"},
			}},
		})
	})
	conn := domain.AdPlatformConnection{ID: "c1", UserID: "u1", ExternalAccountID: "7100", AccessToken: "tok", Metadata: []byte(`{"currency":"XOF"}`)}
	campaigns, err := p.ListCampaigns(context.Background(), conn, port.CampaignQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(campaigns) != 2 { // supprimée filtrée
		t.Fatalf("len = %d", len(campaigns))
	}
	if campaigns[0].Status != domain.CampaignActive || campaigns[0].BudgetMinor != 500000 {
		t.Fatalf("c0 = %+v", campaigns[0])
	}
	if campaigns[1].Status != domain.CampaignPaused || campaigns[1].Currency != "XOF" {
		t.Fatalf("c1 = %+v", campaigns[1])
	}
}

func TestGetInsights(t *testing.T) {
	p, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/report/integrated/get/") {
			http.Error(w, "unexpected", 400)
			return
		}
		q := r.URL.Query()
		if q.Get("start_date") != "2026-09-01" || !strings.Contains(q.Get("filters"), "camp-1") {
			t.Errorf("params = %v", q)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"list": []map[string]any{
				{"dimensions": map[string]any{"campaign_id": "camp-1"},
					"metrics": map[string]any{"impressions": "900", "clicks": "31", "spend": "1234.5", "conversion": "2"}},
			}},
		})
	})
	conn := domain.AdPlatformConnection{ID: "c1", UserID: "u1", ExternalAccountID: "7100", AccessToken: "tok"}
	insights, err := p.GetInsights(context.Background(), conn, "camp-1", port.InsightsQuery{Since: "2026-09-01", Until: "2026-09-05"})
	if err != nil {
		t.Fatal(err)
	}
	if len(insights) != 1 {
		t.Fatalf("len = %d", len(insights))
	}
	if insights[0].Impressions != 900 || insights[0].SpendMinor != 123450 || insights[0].Conversions != 2 {
		t.Fatalf("insight = %+v", insights[0])
	}
}

func TestUploadCreativeUnsupported(t *testing.T) {
	p := New("app-1", "secret-1", "")
	if p.Capabilities().Creatives {
		t.Fatal("creatives devraient être désactivées")
	}
	if _, err := p.UploadCreative(context.Background(), domain.AdPlatformConnection{}, port.CreativeInput{Type: domain.CreativeVideo}); err == nil {
		t.Fatal("upload accepté alors que non supporté")
	}
}
