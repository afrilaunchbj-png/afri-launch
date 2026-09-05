package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	e, err := NewEncryptor("une-passphrase-secrete-de-32-caracteres-minimum", "v1")
	if err != nil {
		t.Fatal(err)
	}
	for _, plain := range []string{"EAAW_access_token_123", "mot avec accents éàç", "x"} {
		enc, err := e.Encrypt(plain)
		if err != nil {
			t.Fatalf("Encrypt(%q): %v", plain, err)
		}
		if !strings.HasPrefix(enc, "v1:") {
			t.Fatalf("format attendu v1:..., got %q", enc[:6])
		}
		if strings.Contains(enc, plain) {
			t.Fatal("ciphertext contient le plaintext")
		}
		got, err := e.Decrypt(enc)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}
		if got != plain {
			t.Fatalf("roundtrip: got %q, want %q", got, plain)
		}
	}
}

func TestEncryptEmpty(t *testing.T) {
	e, _ := NewEncryptor("une-passphrase-secrete-de-32-caracteres-minimum", "")
	enc, err := e.Encrypt("")
	if err != nil || enc != "" {
		t.Fatalf("Encrypt(\"\") = %q, %v", enc, err)
	}
	got, err := e.Decrypt("")
	if err != nil || got != "" {
		t.Fatalf("Decrypt(\"\") = %q, %v", got, err)
	}
}

func TestEncryptUniqueNonces(t *testing.T) {
	e, _ := NewEncryptor("une-passphrase-secrete-de-32-caracteres-minimum", "v1")
	a, _ := e.Encrypt("token")
	b, _ := e.Encrypt("token")
	if a == b {
		t.Fatal("deux chiffrements identiques produisent le même ciphertext (nonce fixe ?)")
	}
}

func TestDecryptInvalidFormat(t *testing.T) {
	e, _ := NewEncryptor("une-passphrase-secrete-de-32-caracteres-minimum", "v1")
	for _, bad := range []string{"pas-un-chiffre", "v1:nope:XXX", "v1:abc:def"} {
		if _, err := e.Decrypt(bad); err == nil {
			t.Fatalf("Decrypt(%q) devrait échouer", bad)
		}
	}
}

func TestShortKeyRejected(t *testing.T) {
	if _, err := NewEncryptor("trop-court", "v1"); err == nil {
		t.Fatal("clé courte acceptée")
	}
}
