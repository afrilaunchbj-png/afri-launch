package paydunya

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

const (
	testMaster  = "master-key-test"
	testPrivate = "test_private_abc"
	testToken   = "token-1"
)

func TestCreateCheckoutAndConfirm(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/checkout-invoice/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PAYDUNYA-MASTER-KEY") != testMaster ||
			r.Header.Get("PAYDUNYA-PRIVATE-KEY") != testPrivate ||
			r.Header.Get("PAYDUNYA-TOKEN") != testToken {
			http.Error(w, "auth", 401)
			return
		}
		var body struct {
			Invoice struct {
				TotalAmount int64  `json:"total_amount"`
				Description string `json:"description"`
			} `json:"invoice"`
			Actions struct {
				ReturnURL   string `json:"return_url"`
				CallbackURL string `json:"callback_url"`
			} `json:"actions"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Invoice.TotalAmount != 5000 {
			t.Errorf("total_amount = %d", body.Invoice.TotalAmount)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"response_code": "00",
			"response_text": "https://app.paydunya.com/checkout/invoice/tok-1",
			"token":         "tok-1",
		})
	})
	mux.HandleFunc("/checkout-invoice/confirm/tok-1", func(w http.ResponseWriter, r *http.Request) {
		sum := sha512.Sum512([]byte(testMaster))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"response_code": "00",
			"hash":          hex.EncodeToString(sum[:]),
			"status":        "completed",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := New(testMaster, testPrivate, testToken, "live", "AfriLaunch")
	p.apiBaseOverride = srv.URL

	res, err := p.CreateCheckout(context.Background(), port.PaymentCheckoutInput{
		PaymentID: "pay-1", AmountMinor: 5000, Currency: "XOF",
		Description: "Pack", ReturnURL: "https://app.test/credits",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RedirectURL != "https://app.paydunya.com/checkout/invoice/tok-1" || res.ProviderReference != "tok-1" {
		t.Fatalf("res = %+v", res)
	}

	status, err := p.VerifyStatus(context.Background(), "tok-1")
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != domain.PaymentSucceeded {
		t.Fatalf("status = %q", status.Status)
	}
}

func TestVerifyStatusRejectsBadHash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"response_code": "00", "hash": "bogus", "status": "completed",
		})
	}))
	defer srv.Close()
	p := New(testMaster, testPrivate, testToken, "live", "")
	p.apiBaseOverride = srv.URL
	if _, err := p.VerifyStatus(context.Background(), "tok-1"); err == nil {
		t.Fatal("hash invalide accepté")
	}
}

func TestHandleWebhookFormAndJSON(t *testing.T) {
	p := New(testMaster, testPrivate, testToken, "test", "")

	// IPN form-urlencoded.
	res, err := p.HandleWebhook(context.Background(), port.PaymentWebhookInput{
		Headers: map[string]string{"content-type": "application/x-www-form-urlencoded"},
		Body:    []byte("data[token]=tok-1&data[status]=completed"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProviderReference != "tok-1" || !res.Accepted {
		t.Fatalf("res = %+v", res)
	}

	// JSON.
	res, err = p.HandleWebhook(context.Background(), port.PaymentWebhookInput{
		Body: []byte(`{"data":{"token":"tok-2","status":"cancelled"}}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProviderReference != "tok-2" {
		t.Fatalf("res = %+v", res)
	}

	if _, err := p.HandleWebhook(context.Background(), port.PaymentWebhookInput{Body: []byte("rien")}); err == nil {
		t.Fatal("webhook sans token accepté")
	}
}

func TestModeURLs(t *testing.T) {
	if strings.Contains(New("m", "p", "t", "test", "").apiURL(), "live") {
		t.Fatal("mode test doit pointer vers sandbox-api")
	}
}
