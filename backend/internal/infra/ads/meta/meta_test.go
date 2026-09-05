package meta

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

func timeNow() time.Time { return time.Now() }

func asGraphError(err error, target *graphError) bool {
	var g graphError
	if errors.As(err, &g) {
		*target = g
		return true
	}
	return false
}

// newTestMeta construit un provider pointant vers un serveur Graph simulé.
func newTestMeta(t *testing.T, handler http.HandlerFunc) (*Meta, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	m := New("app-123", "secret-456", "v23.0", "https://app.example.com/callback", "")
	m.graphBase = srv.URL
	m.oauthBase = srv.URL
	return m, srv
}

func TestAuthorizationURL(t *testing.T) {
	m := New("app-123", "secret-456", "v23.0", "https://app.example.com/callback", "ads_management")
	raw := m.AuthorizationURL("state-xyz", "https://app.example.com/callback")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/v23.0/dialog/oauth" {
		t.Fatalf("path = %q", u.Path)
	}
	q := u.Query()
	if q.Get("client_id") != "app-123" || q.Get("state") != "state-xyz" || q.Get("scope") != "ads_management" {
		t.Fatalf("query = %v", q)
	}
	if q.Get("redirect_uri") != "https://app.example.com/callback" {
		t.Fatalf("redirect_uri = %q", q.Get("redirect_uri"))
	}
}

func TestExchangeCodeLongLivedAndMe(t *testing.T) {
	var sawExchange, sawLongLived, sawMe bool
	m, _ := newTestMeta(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Query().Get("code") == "the-code":
			sawExchange = true
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "short-token"})
		case r.URL.Query().Get("fb_exchange_token") == "short-token":
			sawLongLived = true
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "long-token", "expires_in": 5184000})
		case r.URL.Path == "/v23.0/me":
			sawMe = true
			if r.URL.Query().Get("access_token") != "long-token" {
				http.Error(w, "bad token", 401)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "user-9", "name": "Alice"})
		default:
			http.Error(w, "unexpected", 400)
		}
	})

	res, err := m.ExchangeCode(context.Background(), "the-code", "https://app.example.com/callback")
	if err != nil {
		t.Fatal(err)
	}
	if !sawExchange || !sawLongLived || !sawMe {
		t.Fatalf("flux incomplet: exchange=%v long=%v me=%v", sawExchange, sawLongLived, sawMe)
	}
	if res.AccessToken != "long-token" || res.ExternalUserID != "user-9" {
		t.Fatalf("result = %+v", res)
	}
	if res.ExpiresAt == nil || res.ExpiresAt.Before(timeNow().Add(50*24*time.Hour)) {
		t.Fatalf("expires_at = %v", res.ExpiresAt)
	}
}

func TestListAdAccounts(t *testing.T) {
	m, _ := newTestMeta(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/me/adaccounts") {
			http.Error(w, "unexpected", 400)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "act_111", "name": "Business principal", "account_status": 1, "currency": "XOF"},
				{"id": "act_222", "name": "Agence", "account_status": 100, "currency": "NGN"},
			},
		})
	})

	accounts, err := m.ListAdAccounts(context.Background(), "tok")
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 {
		t.Fatalf("len = %d", len(accounts))
	}
	if accounts[0].ExternalID != "act_111" || accounts[0].Status != "active" || accounts[0].Currency != "XOF" {
		t.Fatalf("accounts[0] = %+v", accounts[0])
	}
	if accounts[1].Status != "suspended" {
		t.Fatalf("status suspendu non mappé: %+v", accounts[1])
	}
}

func TestVerifyAdAccount(t *testing.T) {
	m, _ := newTestMeta(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v23.0/act_111" {
			http.Error(w, "unexpected", 404)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "act_111", "name": "Business", "currency": "XOF"})
	})
	if _, err := m.VerifyAdAccount(context.Background(), "tok", "123456"); err != domain.ErrAccountNotAccessible {
		t.Fatalf("ID sans préfixe act_ accepté: %v", err)
	}
	acc, err := m.VerifyAdAccount(context.Background(), "tok", "act_111")
	if err != nil || acc.ExternalID != "act_111" || acc.Currency != "XOF" {
		t.Fatalf("verify = %+v, %v", acc, err)
	}
}

func TestCreateCampaignPausedByDefault(t *testing.T) {
	var gotBody map[string]any
	m, _ := newTestMeta(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/act_111/campaigns") {
			http.Error(w, "unexpected", 400)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "camp_9", "success": true})
	})
	conn := domain.AdPlatformConnection{ID: "conn-1", UserID: "u1", ExternalAccountID: "act_111", AccessToken: "tok", Metadata: []byte(`{"currency":"XOF"}`)}

	c, err := m.CreateCampaign(context.Background(), conn, port.CreateCampaignInput{
		Name: "Pub guide", Objective: "OUTCOME_TRAFFIC", BudgetMinor: 500000, Currency: "XOF",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Garde-fou §32 : toujours créée en pause.
	if gotBody["status"] != "PAUSED" {
		t.Fatalf("status = %v", gotBody["status"])
	}
	if gotBody["daily_budget"] != float64(500000) {
		t.Fatalf("daily_budget = %v", gotBody["daily_budget"])
	}
	if c.Status != domain.CampaignPaused || c.ExternalCampaignID != "camp_9" {
		t.Fatalf("campaign = %+v", c)
	}
}

func TestPauseResumeCampaign(t *testing.T) {
	var statuses []string
	m, _ := newTestMeta(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		statuses = append(statuses, body["status"].(string))
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	conn := domain.AdPlatformConnection{ID: "conn-1", UserID: "u1", ExternalAccountID: "act_111", AccessToken: "tok"}
	if err := m.PauseCampaign(context.Background(), conn, "camp_1"); err != nil {
		t.Fatal(err)
	}
	if err := m.ResumeCampaign(context.Background(), conn, "camp_1"); err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 2 || statuses[0] != "PAUSED" || statuses[1] != "ACTIVE" {
		t.Fatalf("statuses = %v", statuses)
	}
}

func TestListCampaignsMapping(t *testing.T) {
	m, _ := newTestMeta(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "camp_1", "name": "Active", "objective": "OUTCOME_SALES", "status": "ACTIVE", "effective_status": "ACTIVE", "daily_budget": 1000},
				{"id": "camp_2", "name": "Paused", "objective": "OUTCOME_TRAFFIC", "status": "ACTIVE", "effective_status": "CAMPAIGN_PAUSED", "daily_budget": 2000},
				{"id": "camp_3", "name": "Deleted", "status": "DELETED", "effective_status": "DELETED"},
			},
		})
	})
	conn := domain.AdPlatformConnection{ID: "conn-1", UserID: "u1", ExternalAccountID: "act_111", AccessToken: "tok", Metadata: []byte(`{"currency":"XOF"}`)}
	campaigns, err := m.ListCampaigns(context.Background(), conn, port.CampaignQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(campaigns) != 3 {
		t.Fatalf("len = %d", len(campaigns))
	}
	if campaigns[0].Status != domain.CampaignActive {
		t.Fatalf("active: %+v", campaigns[0])
	}
	if campaigns[1].Status != domain.CampaignPaused {
		t.Fatalf("effective PAUSED non mappé: %+v", campaigns[1])
	}
	if campaigns[2].Status != domain.CampaignDeleted || campaigns[2].Currency != "XOF" {
		t.Fatalf("deleted: %+v", campaigns[2])
	}
}

func TestGetInsightsParsing(t *testing.T) {
	m, _ := newTestMeta(t, func(w http.ResponseWriter, r *http.Request) {
		var tr struct {
			Since string `json:"since"`
			Until string `json:"until"`
		}
		_ = json.Unmarshal([]byte(r.URL.Query().Get("time_range")), &tr)
		if tr.Since != "2026-09-01" {
			t.Errorf("time_range.since = %q", tr.Since)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"date_start": "2026-09-01", "impressions": "1200", "reach": "900",
					"clicks": "50", "spend": "1234.5", "cpc": "24.69", "cpm": "1028.75", "ctr": "4.17",
					"actions": []map[string]string{{"action_type": "purchase", "value": "3"}},
				},
			},
		})
	})
	conn := domain.AdPlatformConnection{ID: "conn-1", UserID: "u1", ExternalAccountID: "act_111", AccessToken: "tok"}
	insights, err := m.GetInsights(context.Background(), conn, "camp_1", port.InsightsQuery{Since: "2026-09-01", Until: "2026-09-05"})
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
	if i.SpendMinor != 123450 { // 1234.5 XOF → unités mineures
		t.Fatalf("spend_minor = %d", i.SpendMinor)
	}
	if i.Conversions != 3 {
		t.Fatalf("conversions = %v", i.Conversions)
	}
}

func TestGraphErrorTransient(t *testing.T) {
	var called int
	m, _ := newTestMeta(t, func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "rate limited", "code": 4, "type": "OAuthException"},
		})
	})
	conn := domain.AdPlatformConnection{ID: "conn-1", UserID: "u1", ExternalAccountID: "act_111", AccessToken: "tok"}
	_, err := m.ListCampaigns(context.Background(), conn, port.CampaignQuery{})
	if err == nil {
		t.Fatal("erreur attendue")
	}
	var gErr graphError
	if !asGraphError(err, &gErr) {
		t.Fatalf("erreur non typée: %v", err)
	}
	if !gErr.Transient() {
		t.Fatal("code 4 devrait être transient")
	}
	if called != 1 {
		t.Fatalf("called = %d", called)
	}
}
