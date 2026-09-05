package googleads

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

func newTestProvider(t *testing.T, handler http.HandlerFunc) (*GoogleAds, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	g := New("cid", "csecret", "dev-token", "", "")
	g.apiBase = srv.URL
	g.tokenBase = srv.URL
	g.oauthBase = srv.URL
	g.userInfoBase = srv.URL
	return g, srv
}

func TestAuthorizationURL(t *testing.T) {
	g := New("cid", "csecret", "dev-token", "", "")
	raw := g.AuthorizationURL("state-1", "https://app.test/cb")
	if !strings.Contains(raw, "accounts.google.com/o/oauth2/v2/auth") {
		t.Fatalf("url = %q", raw)
	}
	if !strings.Contains(raw, "access_type=offline") || !strings.Contains(raw, "prompt=consent") {
		t.Fatal("offline access (refresh_token) requis")
	}
	if !strings.Contains(raw, "scope=") || !strings.Contains(raw, "adwords") {
		t.Fatal("scope adwords manquant")
	}
}

func TestExchangeCodeStoresRefreshToken(t *testing.T) {
	g, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "at-1", "refresh_token": "rt-1", "expires_in": 3600, "token_type": "Bearer",
			})
			return
		}
		if r.URL.Path == "/oauth2/v2/userinfo" {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "g-9", "email": "x@y.z"})
			return
		}
		http.Error(w, "unexpected", 400)
	})
	res, err := g.ExchangeCode(context.Background(), "code-1", "https://app.test/cb")
	if err != nil {
		t.Fatal(err)
	}
	if res.AccessToken != "at-1" || res.RefreshToken != "rt-1" || res.ExternalUserID != "g-9" {
		t.Fatalf("res = %+v", res)
	}
}

func TestRefreshUsesRefreshToken(t *testing.T) {
	g, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "rt-keep" {
			http.Error(w, "bad form", 400)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "at-new", "expires_in": 3600})
	})
	res, ok, err := g.Refresh(context.Background(), domain.AdPlatformConnection{RefreshToken: "rt-keep"})
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if res.AccessToken != "at-new" || res.RefreshToken != "rt-keep" {
		t.Fatalf("res = %+v", res)
	}
}

func TestListAdAccounts(t *testing.T) {
	g, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "customers:listAccessibleCustomers") {
			if r.Header.Get("developer-token") != "dev-token" {
				http.Error(w, "missing developer token", 403)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"resourceNames": []string{"customers/111", "customers/222"}})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/googleAds:searchStream") {
			var body struct {
				Query string `json:"query"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if !strings.Contains(body.Query, "FROM customer") {
				http.Error(w, "bad query", 400)
				return
			}
			name := "Compte principal"
			if strings.Contains(r.URL.Path, "222") {
				name = "Agence"
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"results": []map[string]any{{
					"customer": map[string]any{
						"descriptiveName": name, "currencyCode": "EUR", "timeZone": "Europe/Paris",
					},
				}},
			}})
			return
		}
		http.Error(w, "unexpected "+r.URL.Path, 400)
	})

	accounts, err := g.ListAdAccounts(context.Background(), "tok")
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 {
		t.Fatalf("len = %d", len(accounts))
	}
	if accounts[0].ExternalID != "111" || accounts[0].Name != "Compte principal" || accounts[0].Currency != "EUR" {
		t.Fatalf("accounts[0] = %+v", accounts[0])
	}
}

func TestVerifyAdAccountInaccessible(t *testing.T) {
	g, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": 403, "message": "PERMISSION_DENIED"}})
	})
	if _, err := g.VerifyAdAccount(context.Background(), "tok", "customers/333"); err != domain.ErrAccountNotAccessible {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateCampaignBudgetThenCampaignPaused(t *testing.T) {
	var calls []string
	g, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		switch {
		case strings.HasSuffix(r.URL.Path, "/campaignBudgets:mutate"):
			var body struct {
				Operations []struct {
					Create map[string]any `json:"create"`
				} `json:"operations"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			amount := body.Operations[0].Create["amountMicros"].(string)
			if amount != "5000000000" { // 500 000 minor × 10 000
				t.Errorf("amountMicros = %q", amount)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{"resourceName": "customers/111/campaignBudgets/77"}},
			})
		case strings.HasSuffix(r.URL.Path, "/campaigns:mutate"):
			var body struct {
				Operations []struct {
					Create map[string]any `json:"create"`
				} `json:"operations"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			create := body.Operations[0].Create
			if create["status"] != "PAUSED" {
				t.Errorf("status = %v", create["status"])
			}
			if create["campaignBudget"] != "customers/111/campaignBudgets/77" {
				t.Errorf("budget = %v", create["campaignBudget"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{"resourceName": "customers/111/campaigns/42"}},
			})
		default:
			http.Error(w, "unexpected "+r.URL.Path, 400)
		}
	})
	conn := domain.AdPlatformConnection{ID: "c1", UserID: "u1", ExternalAccountID: "customers/111", AccessToken: "tok"}

	c, err := g.CreateCampaign(context.Background(), conn, port.CreateCampaignInput{
		Name: "Campagne EUR", Objective: "OUTCOME_TRAFFIC", BudgetMinor: 500_000, Currency: "EUR",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("calls = %v", calls)
	}
	if c.ExternalCampaignID != "42" || c.Status != domain.CampaignPaused || c.BudgetMinor != 500_000 {
		t.Fatalf("campaign = %+v", c)
	}
}

func TestPauseResumeCampaign(t *testing.T) {
	var statuses []string
	g, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Operations []struct {
				Update struct {
					Status string `json:"status"`
				} `json:"update"`
			} `json:"operations"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		statuses = append(statuses, body.Operations[0].Update.Status)
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []map[string]any{}})
	})
	conn := domain.AdPlatformConnection{ID: "c1", UserID: "u1", ExternalAccountID: "customers/111", AccessToken: "tok"}
	if err := g.PauseCampaign(context.Background(), conn, "42"); err != nil {
		t.Fatal(err)
	}
	if err := g.ResumeCampaign(context.Background(), conn, "42"); err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 2 || statuses[0] != "PAUSED" || statuses[1] != "ENABLED" {
		t.Fatalf("statuses = %v", statuses)
	}
}

func TestListCampaignsMapping(t *testing.T) {
	g, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"results": []map[string]any{
				{"campaign": map[string]any{"id": "1", "name": "Active", "status": "ENABLED"}, "campaignBudget": map[string]any{"amountMicros": "10000000"}, "customer": map[string]any{"currencyCode": "EUR"}},
				{"campaign": map[string]any{"id": "2", "name": "Paused", "status": "PAUSED"}, "campaignBudget": map[string]any{"amountMicros": "20000000"}, "customer": map[string]any{"currencyCode": "EUR"}},
				{"campaign": map[string]any{"id": "3", "name": "Removed", "status": "REMOVED"}},
			},
		}})
	})
	conn := domain.AdPlatformConnection{ID: "c1", UserID: "u1", ExternalAccountID: "customers/111", AccessToken: "tok"}
	campaigns, err := g.ListCampaigns(context.Background(), conn, port.CampaignQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(campaigns) != 3 {
		t.Fatalf("len = %d", len(campaigns))
	}
	if campaigns[0].Status != domain.CampaignActive || campaigns[0].BudgetMinor != 1000 {
		t.Fatalf("c0 = %+v", campaigns[0])
	}
	if campaigns[1].Status != domain.CampaignPaused || campaigns[1].BudgetMinor != 2000 {
		t.Fatalf("c1 = %+v", campaigns[1])
	}
	if campaigns[2].Status != domain.CampaignDeleted {
		t.Fatalf("c2 = %+v", campaigns[2])
	}
}

func TestGetInsightsMicrosConversion(t *testing.T) {
	g, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !strings.Contains(body.Query, "campaign.id = 42") {
			t.Errorf("query = %q", body.Query)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"results": []map[string]any{
				{"campaign": map[string]any{"id": "42"}, "segments": map[string]any{"date": "2026-09-01"},
					"metrics": map[string]any{"impressions": "1200", "clicks": "50", "costMicros": "1234500000", "conversions": "3.5"}},
			},
		}})
	})
	conn := domain.AdPlatformConnection{ID: "c1", UserID: "u1", ExternalAccountID: "customers/111", AccessToken: "tok"}
	insights, err := g.GetInsights(context.Background(), conn, "42", port.InsightsQuery{Since: "2026-09-01", Until: "2026-09-05"})
	if err != nil {
		t.Fatal(err)
	}
	if len(insights) != 1 {
		t.Fatalf("len = %d", len(insights))
	}
	i := insights[0]
	if i.Date != "2026-09-01" || i.Impressions != 1200 || i.Clicks != 50 {
		t.Fatalf("insight = %+v", i)
	}
	if i.SpendMinor != 123450 { // 1 234 500 000 micros → 123 450 minor
		t.Fatalf("spend_minor = %d", i.SpendMinor)
	}
	if i.Conversions != 3.5 {
		t.Fatalf("conversions = %v", i.Conversions)
	}
}

func TestUploadCreativeUnsupported(t *testing.T) {
	g := New("cid", "csecret", "dev-token", "", "")
	if g.Capabilities().Creatives {
		t.Fatal("creatives devraient être désactivées")
	}
	if _, err := g.UploadCreative(context.Background(), domain.AdPlatformConnection{}, port.CreativeInput{Type: domain.CreativeVideo}); err == nil {
		t.Fatal("upload accepté alors que non supporté")
	}
}
