package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNeonVerifier(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	const issuer = "https://ep-xxxx.aws.neon.tech"
	const kid = "test-key-1"

	// Serveur JWKS factice (endpoint /.well-known/jwks.json).
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []any{
				map[string]any{
					"kty": "OKP",
					"crv": "Ed25519",
					"kid": kid,
					"x":   base64.RawURLEncoding.EncodeToString(pub),
				},
			},
		})
	}))
	defer jwksServer.Close()

	verifier, err := NewNeonVerifier(issuer+"/neondb/auth", jwksServer.URL)
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	// JWT signé EdDSA avec le kid attendu.
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"sub":   "860dc360-609f-4b7d-9e70-ec93fe6414d3",
		"email": "user@example.com",
		"name":  "Test User",
		"image": "https://example.com/a.png",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
		"iss":   issuer,
	})
	token.Header["kid"] = kid
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	u, err := verifier.Verify(context.Background(), signed)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if u.ID != "860dc360-609f-4b7d-9e70-ec93fe6414d3" || u.Email != "user@example.com" || u.Name != "Test User" {
		t.Fatalf("unexpected claims: %+v", u)
	}

	// Token altéré → rejeté.
	if _, err := verifier.Verify(context.Background(), signed+"tampered"); err == nil {
		t.Fatal("expected verification to fail for tampered token")
	}

	// Token signé par une autre clé (kid inconnu) → rejeté.
	_, priv2, _ := ed25519.GenerateKey(rand.Reader)
	badToken := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{"sub": "x", "iss": issuer})
	badToken.Header["kid"] = "unknown-kid"
	badSigned, _ := badToken.SignedString(priv2)
	if _, err := verifier.Verify(context.Background(), badSigned); err == nil {
		t.Fatal("expected verification to fail for unknown kid")
	}
}
