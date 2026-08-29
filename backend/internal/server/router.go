package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/config"
	"afrilaunch/backend/internal/server/handler"
)

// Deps regroupe les dépendances nécessaires au routeur (injection).
type Deps struct {
	Config        config.Config
	Tokens        port.TokenVerifier
	Health        *handler.Health
	Auth          *handler.AuthHandler
	Credits       *handler.CreditHandler
	Opportunities *handler.OpportunityHandler
	Markets       *handler.MarketHandler
}

// NewRouter construit le routeur HTTP de l'API.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(requestLogger)
	r.Use(cors(d.Config.AllowedOrigins))

	r.NotFound(handler.NotFound)
	r.MethodNotAllowed(handler.MethodNotAllowed)

	r.Get("/healthz", d.Health.Liveness)
	r.Get("/readyz", d.Health.Readiness)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", d.Health.Health)

		// Référentiel des marchés (public).
		r.Get("/markets", d.Markets.List)

		// Routes protégées (JWT Neon Auth).
		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(d.Tokens))

			r.Get("/auth/me", d.Auth.Me)

			r.Get("/credits", d.Credits.Summary)
			r.Get("/credits/transactions", d.Credits.Transactions)
			r.Post("/credits/reserve", d.Credits.Reserve)

			r.Get("/opportunities", d.Opportunities.List)
			r.Get("/opportunities/filters", d.Opportunities.Filters)
			r.Post("/opportunities/{id}/save", d.Opportunities.Save)
			r.Delete("/opportunities/{id}/save", d.Opportunities.Unsave)
		})
	})

	return r
}

// cors applique une politique CORS avec allowlist d'origines explicite.
func cors(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	allow := make(map[string]bool, len(origins))
	for _, o := range origins {
		if o = strings.TrimSpace(o); o != "" {
			allow[o] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allow[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key, X-Correlation-ID")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
