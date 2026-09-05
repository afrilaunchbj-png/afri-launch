// Package auth fournit le vérificateur de tokens Neon Auth (Managed Better
// Auth) : signature EdDSA (Ed25519), vérification via JWKS.
package auth

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"afrilaunch/backend/internal/application/port"
)

var (
	errNoBaseURL    = errors.New("NEON_AUTH_BASE_URL is not configured")
	errUnknownKeyID = errors.New("no matching JWK for token kid")
)

// NeonVerifier vérifie les JWT émis par Managed Better Auth.
type NeonVerifier struct {
	issuer  string
	jwksURL string
	http    *http.Client

	mu        sync.RWMutex
	keys      map[string]ed25519.PublicKey // kid -> clé publique
	fetchedAt time.Time
}

// NewNeonVerifier construit un vérificateur Neon Auth.
// jwksURL peut être vide : il est alors dérivé de baseURL.
func NewNeonVerifier(baseURL, jwksURL string) (*NeonVerifier, error) {
	if baseURL == "" {
		return nil, errNoBaseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	issuer := u.Scheme + "://" + u.Host
	if jwksURL == "" {
		jwksURL = strings.TrimRight(baseURL, "/") + "/.well-known/jwks.json"
	}
	return &NeonVerifier{
		issuer:  issuer,
		jwksURL: jwksURL,
		http:    &http.Client{Timeout: 10 * time.Second},
		keys:    make(map[string]ed25519.PublicKey),
	}, nil
}

// Verify valide un token JWT et renvoie l'identité utilisateur.
func (v *NeonVerifier) Verify(ctx context.Context, tokenString string) (port.AuthUser, error) {
	kid, err := kidFromToken(tokenString)
	if err != nil {
		return port.AuthUser{}, err
	}
	key, err := v.key(ctx, kid)
	if err != nil {
		return port.AuthUser{}, err
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{"EdDSA"}), jwt.WithIssuer(v.issuer))
	if err != nil {
		return port.AuthUser{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return port.AuthUser{}, errors.New("invalid token claims")
	}

	return port.AuthUser{
		ID:      strClaim(claims["sub"]),
		Email:   strClaim(claims["email"]),
		Name:    strClaim(claims["name"]),
		Picture: strClaim(claims["image"]),
	}, nil
}

// key récupère (et met en cache) la clé publique correspondant au kid.
func (v *NeonVerifier) key(ctx context.Context, kid string) (ed25519.PublicKey, error) {
	v.mu.RLock()
	if k, ok := v.keys[kid]; ok && time.Since(v.fetchedAt) < time.Hour {
		v.mu.RUnlock()
		return k, nil
	}
	v.mu.RUnlock()

	if err := v.refresh(ctx); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()
	if k, ok := v.keys[kid]; ok {
		return k, nil
	}
	return nil, errUnknownKeyID
}

func (v *NeonVerifier) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var set struct {
		Keys []struct {
			Kty string `json:"kty"`
			Crv string `json:"crv"`
			Kid string `json:"kid"`
			X   string `json:"x"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return err
	}

	keys := make(map[string]ed25519.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "OKP" || k.Crv != "Ed25519" {
			continue
		}
		raw, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			continue
		}
		keys[k.Kid] = ed25519.PublicKey(raw)
	}

	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = time.Now()
	v.mu.Unlock()
	return nil
}

// kidFromToken extrait le `kid` du header sans vérifier la signature.
func kidFromToken(tokenString string) (string, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}
	kid, _ := token.Header["kid"].(string)
	if kid == "" {
		return "", errors.New("token header has no kid")
	}
	return kid, nil
}

func strClaim(v any) string {
	s, _ := v.(string)
	return s
}

// DenyVerifier implémente port.TokenVerifier en refusant systématiquement
// les tokens. Utilisé quand Neon Auth n'est pas configuré : l'application
// démarre (healthz OK), mais toutes les routes protégées renvoient 401 —
// une variable d'environnement manquante ne doit jamais rendre le backend
// inaccessible.
type DenyVerifier struct{}

// NewDenyVerifier construit le vérificateur qui refuse tout.
func NewDenyVerifier() *DenyVerifier { return &DenyVerifier{} }

// Verify rejette tout token avec une erreur explicite.
func (d *DenyVerifier) Verify(_ context.Context, _ string) (port.AuthUser, error) {
	return port.AuthUser{}, errNoBaseURL
}
