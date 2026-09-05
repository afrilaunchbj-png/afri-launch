package pawapay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

func TestCreateCheckoutAccepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/checkouts" || r.Method != http.MethodPost {
			http.Error(w, "unexpected", 400)
			return
		}
		if r.Header.Get("Authorization") != "Bearer tok-1" {
			http.Error(w, "bad auth", 401)
			return
		}
		var body struct {
			CheckoutID string `json:"checkoutId"`
			ReturnURL  string `json:"returnUrl"`
			Amounts    []struct {
				Country  string `json:"country"`
				Currency string `json:"currency"`
				Amount   string `json:"amount"`
			} `json:"amounts"`
			ClientReferenceID string `json:"clientReferenceId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.CheckoutID != "pay-1" || body.ClientReferenceID != "pay-1" {
			t.Errorf("ids = %v", body)
		}
		if len(body.Amounts) != 1 || body.Amounts[0].Amount != "5000" || body.Amounts[0].Currency != "XOF" {
			t.Errorf("amounts = %+v", body.Amounts)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"checkoutId": "pay-1", "status": "ACCEPTED", "redirectUrl": "https://checkout.pawapay.io/abc",
		})
	}))
	defer srv.Close()

	p := New("tok-1", srv.URL, "BEN")
	res, err := p.CreateCheckout(context.Background(), port.PaymentCheckoutInput{
		PaymentID: "pay-1", AmountMinor: 5000, Currency: "XOF",
		Description: "Pack Business — 120 crédits", ReturnURL: "https://app.test/credits",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RedirectURL != "https://checkout.pawapay.io/abc" || res.ProviderReference != "pay-1" {
		t.Fatalf("res = %+v", res)
	}
}

func TestVerifyStatusMapping(t *testing.T) {
	cases := map[string]string{
		"COMPLETED":       domain.PaymentSucceeded,
		"FAILED":          domain.PaymentFailed,
		"CANCELLED":       domain.PaymentFailed,
		"EXPIRED":         domain.PaymentFailed,
		"WAITING_PAYMENT": domain.PaymentPending,
		"PROCESSING":      domain.PaymentPending,
	}
	for remote, want := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"status": remote})
		}))
		p := New("tok", srv.URL, "")
		got, err := p.VerifyStatus(context.Background(), "ref-1")
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != want {
			t.Fatalf("status %q → %q, want %q", remote, got.Status, want)
		}
		srv.Close()
	}
}

func TestHandleWebhookExtractsReference(t *testing.T) {
	p := New("tok", "", "")

	res, err := p.HandleWebhook(context.Background(), port.PaymentWebhookInput{
		Body: []byte(`{"depositId":"d-1","clientReferenceId":"pay-1","status":"COMPLETED"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProviderReference != "pay-1" || !res.Accepted {
		t.Fatalf("res = %+v", res)
	}

	res, err = p.HandleWebhook(context.Background(), port.PaymentWebhookInput{
		Body: []byte(`{"checkoutId":"pay-2","status":"FAILED"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProviderReference != "pay-2" || !res.Accepted {
		t.Fatalf("res = %+v", res)
	}

	res, err = p.HandleWebhook(context.Background(), port.PaymentWebhookInput{
		Body: []byte(`{"checkoutId":"pay-3","status":"PROCESSING"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Accepted {
		t.Fatal("un statut transitoire ne doit pas être final")
	}

	if _, err := p.HandleWebhook(context.Background(), port.PaymentWebhookInput{Body: []byte(`{}`)}); err == nil {
		t.Fatal("webhook sans référence accepté")
	}
}

func TestAmountFormatNonDecimalCurrency(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Amounts []struct {
				Amount string `json:"amount"`
			} `json:"amounts"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Amounts[0].Amount != "123.45" {
			t.Errorf("amount = %q, want 123.45", body.Amounts[0].Amount)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ACCEPTED", "redirectUrl": "u"})
	}))
	defer srv.Close()

	p := New("tok", srv.URL, "")
	if _, err := p.CreateCheckout(context.Background(), port.PaymentCheckoutInput{
		PaymentID: "p", AmountMinor: 12345, Currency: "GHS", ReturnURL: "r",
	}); err != nil {
		t.Fatal(err)
	}
}
