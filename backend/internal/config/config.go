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

	// Providers IA (voir docs/ai.md).
	OpenAIAPIKey        string
	OpenAIResearchModel string
	OpenAIIdeationModel string
	OpenAIImageModel    string
	HeyGenAPIKey        string
	HeyGenAPIURL        string

	// Rendu HTML → PDF/PPTX (chromedp / Chrome headless).
	ChromePath string

	// Stockage des fichiers générés (local au MVP ; S3 à venir).
	StorageDir string

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

		OpenAIAPIKey:        get("OPENAI_API_KEY", ""),
		OpenAIResearchModel: get("OPENAI_MODEL_RESEARCH", "gpt-5.6-terra"),
		OpenAIIdeationModel: get("OPENAI_MODEL_IDEATION", "gpt-5.6-luna"),
		OpenAIImageModel:    get("OPENAI_MODEL_IMAGE", "gpt-image-2"),
		HeyGenAPIKey:        get("HEYGEN_API_KEY", ""),
		HeyGenAPIURL:        get("HEYGEN_API_URL", "https://api.heygen.com"),

		ChromePath: get("CHROME_PATH", ""),

		StorageDir: get("STORAGE_DIR", "./.storage"),

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
