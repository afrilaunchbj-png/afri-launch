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

	// Vidéos publicitaires (avatar HeyGen par défaut, configurable par projet).
	HeyGenDefaultAvatarID string
	HeyGenDefaultVoiceID  string

	// Intégrations publicitaires (ADR-017) — tokens clients chiffrés en DB.
	AppURL                string // URL publique du frontend (redirections OAuth)
	EncryptionKey         string
	EncryptionKeyVersion  string
	MetaAppID             string
	MetaAppSecret         string
	MetaGraphVersion      string
	MetaOAuthRedirectURI  string
	MetaOAuthScopes       string
	GoogleAdsClientID     string
	GoogleAdsClientSecret string
	GoogleAdsDevToken     string
	GoogleAdsRedirectURI  string
	GoogleAdsLoginCustID  string
	GoogleAdsAPIVersion   string
	TikTokAppID           string
	TikTokAppSecret       string
	TikTokRedirectURI     string

	// Stockage objet (S3-compatible / Neon en prod, sinon disque local).
	StorageDir        string
	S3Endpoint        string
	S3Region          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3Bucket          string
	S3PathStyle       bool

	// Rendu HTML → PDF/PPTX (chromedp / Chrome headless).
	ChromePath string

	// Montage vidéo (FFmpeg) — binaire ffmpeg (ffprobe dérivé).
	FFmpegPath string

	// Métier
	WelcomeCredits int

	// Superadmins (emails séparés par des virgules) — promus au login,
	// accès au suivi global /admin.
	SuperadminEmails []string
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

		HeyGenDefaultAvatarID: get("HEYGEN_DEFAULT_AVATAR_ID", ""),
		HeyGenDefaultVoiceID:  get("HEYGEN_DEFAULT_VOICE_ID", ""),

		AppURL:               get("APP_URL", "http://localhost:5173"),
		EncryptionKey:        get("ENCRYPTION_KEY", ""),
		EncryptionKeyVersion: get("ENCRYPTION_KEY_VERSION", "v1"),
		MetaAppID:            get("META_APP_ID", ""),
		MetaAppSecret:        get("META_APP_SECRET", ""),
		MetaGraphVersion:     get("META_GRAPH_API_VERSION", "v23.0"),
		MetaOAuthRedirectURI: get("META_OAUTH_REDIRECT_URI", ""),
		MetaOAuthScopes:      get("META_OAUTH_SCOPES", ""),

		GoogleAdsClientID:     get("GOOGLE_ADS_CLIENT_ID", ""),
		GoogleAdsClientSecret: get("GOOGLE_ADS_CLIENT_SECRET", ""),
		GoogleAdsDevToken:     get("GOOGLE_ADS_DEVELOPER_TOKEN", ""),
		GoogleAdsRedirectURI:  get("GOOGLE_ADS_OAUTH_REDIRECT_URI", ""),
		GoogleAdsLoginCustID:  get("GOOGLE_ADS_LOGIN_CUSTOMER_ID", ""),
		GoogleAdsAPIVersion:   get("GOOGLE_ADS_API_VERSION", ""),

		TikTokAppID:       get("TIKTOK_APP_ID", ""),
		TikTokAppSecret:   get("TIKTOK_APP_SECRET", ""),
		TikTokRedirectURI: get("TIKTOK_OAUTH_REDIRECT_URI", ""),

		ChromePath: get("CHROME_PATH", ""),
		FFmpegPath: get("FFMPEG_PATH", "ffmpeg"),

		StorageDir:        get("STORAGE_DIR", "./.storage"),
		S3Endpoint:        get("S3_ENDPOINT", ""),
		S3Region:          get("S3_REGION", "us-east-1"),
		S3AccessKeyID:     get("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey: get("S3_SECRET_ACCESS_KEY", ""),
		S3Bucket:          get("S3_BUCKET", ""),
		S3PathStyle:       get("S3_PATH_STYLE", "true") == "true",

		WelcomeCredits: getInt("WELCOME_CREDITS", 100),

		SuperadminEmails: parseEmails(os.Getenv("SUPERADMIN_EMAILS")),
	}
}

// parseEmails découpe une liste d'emails séparés par des virgules.
func parseEmails(raw string) []string {
	emails := make([]string, 0)
	for _, e := range strings.Split(raw, ",") {
		if e = strings.ToLower(strings.TrimSpace(e)); e != "" {
			emails = append(emails, e)
		}
	}
	return emails
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
