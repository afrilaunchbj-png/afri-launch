package fedapay

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

func newTestProvider(handler http.HandlerFunc) (*FedaPay, *httptest.Server) {
	srv := httptest.NewServer(handler)
	return New("sk_demo", srv.URL, ""), srv
}

func TestCreateCheckout(t *testing.T) {
	p, srv := newTestProvider(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk_demo" {
			http.Error(w, "auth", 401)
			return
		}
		switch {
		case r.URL.Path == "/transactions":
			var body struct {
				Description string `json:"description"`
				Amount      int64  `json:"amount"`
				Currency    struct {
					ISO string `json:"iso"`
				} `json:"currency"`
				MerchantReference string `json:"merchant_reference"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Amount != 5000 || body.Currency.ISO != "XOF" || body.MerchantReference != "pay-1" {
				t.Errorf("body = %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"transaction": map[string]any{"id": 77}})
		case strings.HasSuffix(r.URL.Path, "/token"):
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "tok-9", "url": "https://checkout.fedapay.com/tok-9"})
		default:
			http.Error(w, "unexpected "+r.URL.Path, 400)
		}
	})
	defer srv.Close()

	res, err := p.CreateCheckout(context.Background(), port.PaymentCheckoutInput{
		PaymentID: "pay-1", AmountMinor: 5000, Currency: "XOF",
		Description: "Pack", ReturnURL: "https://app.test/credits",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RedirectURL != "https://checkout.fedapay.com/tok-9" || res.ProviderReference != "77" {
		t.Fatalf("res = %+v", res)
	}
}

func TestVerifyStatusMapping(t *testing.T) {
	cases := map[string]string{
		"approved":    domain.PaymentSucceeded,
		"transferred": domain.PaymentSucceeded,
		"declined":    domain.PaymentFailed,
		"canceled":    domain.PaymentFailed,
		"expired":     domain.PaymentFailed,
		"pending":     domain.PaymentPending,
	}
	for remote, want := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"transaction": map[string]any{"status": remote}})
		}))
		p := New("sk", srv.URL, "")
		got, err := p.VerifyStatus(context.Background(), "77")
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != want {
			t.Fatalf("%q → %q, want %q", remote, got.Status, want)
		}
		srv.Close()
	}
}

func TestHandleWebhook(t *testing.T) {
	p := New("sk", "", "")
	res, err := p.HandleWebhook(context.Background(), port.PaymentWebhookInput{
		Body: []byte(`{"name":"transaction.approved","object":{"id":77,"status":"approved"}}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProviderReference != "77" || !res.Accepted {
		t.Fatalf("res = %+v", res)
	}
	// Événement non transaction : non final.
	res, err = p.HandleWebhook(context.Background(), port.PaymentWebhookInput{
		Body: []byte(`{"name":"customer.created","object":{"id":5}}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Accepted {
		t.Fatal("événement non transaction accepté comme final")
	}
}

func TestStatusNormalization(t *testing.T) {
	if normalizeStatus("approved") != domain.PaymentSucceeded || normalizeStatus("pending") != domain.PaymentPending {
		t.Fatal("mapping de base invalide")
	}
	if majorAmount(5000, "XOF") != 5000 || majorAmount(12345, "GHS") != 123 {
		t.Fatal("conversion de montant invalide")
	}
}
