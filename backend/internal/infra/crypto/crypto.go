// Package crypto fournit le chiffrement au repos des secrets (tokens OAuth).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Encryptor implémente port.SecretEncryptor (AES-256-GCM).
// Format stocké : "<version>:<nonce b64url>:<ciphertext b64url>" — le préfixe
// de version permet la rotation de clé (déchiffrement par version).
type Encryptor struct {
	aead    cipher.AEAD
	version string
}

// NewEncryptor construit un chiffreur AES-256-GCM. La clé est dérivée en
// 32 octets via SHA-256 (accepte toute passphrase ≥ 32 caractères).
func NewEncryptor(secret, version string) (*Encryptor, error) {
	if len(secret) < 32 {
		return nil, errors.New("crypto: ENCRYPTION_KEY doit faire au moins 32 caractères")
	}
	if version == "" {
		version = "v1"
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("crypto: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: %w", err)
	}
	return &Encryptor{aead: aead, version: version}, nil
}

// Encrypt chiffre plaintext avec un nonce aléatoire.
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: nonce: %w", err)
	}
	sealed := e.aead.Seal(nil, nonce, []byte(plaintext), nil)
	return e.version + ":" +
		base64.RawURLEncoding.EncodeToString(nonce) + ":" +
		base64.RawURLEncoding.EncodeToString(sealed), nil
}

// Decrypt déchiffre une valeur produite par Encrypt.
func (e *Encryptor) Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	parts := strings.SplitN(value, ":", 3)
	if len(parts) != 3 {
		return "", errors.New("crypto: format invalide")
	}
	// parts[0] = version : une seule clé active pour l'instant ; la rotation
	// consistera à conserver un map version→clé et à résoudre parts[0].
	nonce, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("crypto: nonce: %w", err)
	}
	if len(nonce) != e.aead.NonceSize() {
		return "", errors.New("crypto: format invalide")
	}
	sealed, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("crypto: ciphertext: %w", err)
	}
	plaintext, err := e.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: déchiffrement: %w", err)
	}
	return string(plaintext), nil
}
