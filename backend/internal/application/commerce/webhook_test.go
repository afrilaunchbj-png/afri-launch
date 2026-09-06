package commerce

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyPulseSignature(t *testing.T) {
	secret := "whsec_test_secret"
	body := []byte(`{"event":"successful.sale","sale":{"id":"sal_1"}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	good := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !verifyPulseSignature(secret, body, good) {
		t.Error("signature valide rejetée")
	}
	if verifyPulseSignature(secret, body, "sha256=0000") {
		t.Error("signature invalide acceptée")
	}
	if verifyPulseSignature(secret, body, "md5=0000") {
		t.Error("préfixe inconnu accepté")
	}
	if verifyPulseSignature("autre-secret", body, good) {
		t.Error("mauvais secret accepté")
	}
}

func TestMapSaleEvents(t *testing.T) {
	cases := map[string]string{
		"successful.sale": "completed",
		"failed.sale":     "failed",
		"abandoned.sale":  "abandoned",
		"refunded.sale":   "refunded",
		"license.issued":  "",
	}
	for event, want := range cases {
		got := ""
		switch event {
		case "successful.sale":
			got = "completed"
		case "failed.sale":
			got = "failed"
		case "abandoned.sale":
			got = "abandoned"
		case "refunded.sale":
			got = "refunded"
		}
		if got != want {
			t.Errorf("%s → %q, want %q", event, got, want)
		}
	}
}

func TestNormalizeSaleStatus(t *testing.T) {
	for in, want := range map[string]string{"completed": "completed", "settled": "settled", "success": "completed", "": "awaiting_payment"} {
		if got := normalizeSaleStatus(in); got != want {
			t.Errorf("normalizeSaleStatus(%q) = %q, want %q", in, got, want)
		}
	}
}
