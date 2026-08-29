package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config regroupe la configuration de l'application, chargée depuis
// l'environnement (et un éventuel fichier .env en développement).
type Config struct {
	Env            string
	Port           string
	DatabaseURL    string
	RedisURL       string
	AllowedOrigins string

	// Auth (Managed Better Auth / Neon Auth).
	// Le backend vérifie les JWT émis par Neon Auth (EdDSA) via JWKS.
	NeonAuthBaseURL string
	NeonAuthJWKSURL string

	// Métier
	WelcomeCredits int
}

// Load lit la configuration depuis les variables d'environnement.
func Load() Config {
	_ = godotenv.Load()

	baseURL := get("NEON_AUTH_BASE_URL", "")

	return Config{
		Env:            get("APP_ENV", "development"),
		Port:           get("PORT", "8080"),
		DatabaseURL:    get("DATABASE_URL", "postgres://afrilaunch:afrilaunch@localhost:5432/afrilaunch?sslmode=disable"),
		RedisURL:       get("REDIS_URL", "redis://localhost:6379/0"),
		AllowedOrigins: get("ALLOWED_ORIGINS", "http://localhost:5173"),

		NeonAuthBaseURL: baseURL,
		NeonAuthJWKSURL: get("NEON_AUTH_JWKS_URL", jwksURL(baseURL)),

		WelcomeCredits: getInt("WELCOME_CREDITS", 100),
	}
}

// jwksURL dérive l'URL JWKS de l'URL de base Neon Auth.
func jwksURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	return strings.TrimRight(baseURL, "/") + "/.well-known/jwks.json"
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
