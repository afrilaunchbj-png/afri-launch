package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/document"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/config"
	"afrilaunch/backend/internal/server/handler"
)

// Deps regroupe les dépendances nécessaires au routeur (injection).
type Deps struct {
	Config        config.Config
	Tokens        port.TokenVerifier
	Users         port.UserRepository
	Health        *handler.Health
	Auth          *handler.AuthHandler
	Credits       *handler.CreditHandler
	Opportunities *handler.OpportunityHandler
	Markets       *handler.MarketHandler
	Ideas         *handler.IdeaHandler
	Projects      *handler.ProjectHandler
	Assets        *handler.AssetHandler
	Jobs          *handler.JobHandler
	Research      *handler.ResearchHandler
	Conversations *handler.ConversationHandler
	Events        *handler.EventHandler
	Preferences   *handler.PreferenceHandler
	Support       *handler.SupportHandler
	Integrations  *handler.IntegrationHandler
	Admin         *handler.AdminHandler
	Dashboard     *handler.DashboardHandler
	// AI : providers IA, consommés par les workers (générations asynchrones).
	AI *ai.Service
	// Documents : génération ebook/deck (LLM → HTML → chromedp → PDF/PPTX).
	Documents *document.Service
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

		// Callbacks OAuth des plateformes publicitaires : arrivent sans JWT
		// (navigation navigateur) — l'utilisateur est résolu via l'état CSRF.
		r.Get("/integrations/{provider}/callback", d.Integrations.Callback)

		// Routes protégées (JWT Neon Auth).
		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(d.Tokens))

			r.Get("/auth/me", d.Auth.Me)

			// Canal temps réel unique (SSE) : chat, jobs, notifications.
			r.Get("/events", d.Events.Stream)

			// Copilote conversationnel.
			r.Get("/conversations", d.Conversations.List)
			r.Post("/conversations", d.Conversations.Create)
			r.Get("/conversations/{id}", d.Conversations.Get)
			r.Post("/conversations/{id}/messages", d.Conversations.SendMessage)

			// Préférences utilisateur (langue, thème).
			r.Get("/preferences", d.Preferences.Get)
			r.Put("/preferences", d.Preferences.Update)

			// Support utilisateur.
			r.Post("/support/tickets", d.Support.Create)
			r.Get("/support/tickets", d.Support.ListMine)
			r.Get("/support/tickets/{id}", d.Support.GetTicket)
			r.Post("/support/tickets/{id}/messages", d.Support.Reply)
			r.Post("/support/attachments", d.Support.UploadAttachment)
			r.Get("/support/attachments/{id}/download", d.Support.DownloadAttachment)

			// Intégrations publicitaires (ADR-017).
			r.Get("/integrations", d.Integrations.List)
			r.Get("/integrations/{provider}/connect", d.Integrations.Connect)
			r.Get("/integrations/{provider}/accounts", d.Integrations.ListAccounts)
			r.Post("/integrations/{provider}/accounts/select", d.Integrations.SelectAccount)
			r.Post("/integrations/{provider}/campaigns/sync", d.Integrations.SyncCampaigns)
			r.Delete("/integrations/{provider}", d.Integrations.Disconnect)
			r.Get("/ad-campaigns", d.Integrations.ListCampaigns)
			r.Post("/ad-campaigns", d.Integrations.CreateCampaign)
			r.Post("/ad-campaigns/{id}/pause", d.Integrations.PauseCampaign)
			r.Post("/ad-campaigns/{id}/resume", d.Integrations.ResumeCampaign)
			r.Post("/ad-campaigns/{id}/creatives", d.Integrations.PublishCreative)
			r.Get("/ad-campaigns/{id}/insights", d.Integrations.CampaignInsights)
			r.Get("/ad-creatives", d.Integrations.ListCreatives)

			// Tableau de bord personnel : statistiques et courbes.
			r.Get("/dashboard/stats", d.Dashboard.Stats)

			// Suivi global superadmin.
			r.Route("/admin", func(r chi.Router) {
				r.Use(RequireSuperadmin(d.Users))
				r.Get("/stats", d.Admin.Stats)
				r.Get("/users", d.Admin.Users)
				r.Get("/tickets", d.Admin.Tickets)
				r.Get("/tickets/{id}", d.Admin.TicketDetail)
				r.Post("/tickets/{id}/resolve", d.Admin.ResolveTicket)
				r.Post("/tickets/{id}/messages", d.Admin.ReplyTicket)
				r.Get("/support/attachments/{id}/download", d.Support.AdminDownloadAttachment)
				r.Get("/projects", d.Admin.Projects)
				r.Get("/conversations", d.Admin.Conversations)
				r.Get("/assets", d.Admin.Assets)
				r.Get("/jobs", d.Admin.Jobs)
				r.Get("/credit-transactions", d.Admin.CreditTransactions)
				r.Get("/audit-logs", d.Admin.AuditLogs)
			})

			r.Get("/credits", d.Credits.Summary)
			r.Get("/credits/transactions", d.Credits.Transactions)
			r.Post("/credits/reserve", d.Credits.Reserve)

			r.Get("/opportunities", d.Opportunities.List)
			r.Get("/opportunities/filters", d.Opportunities.Filters)
			r.Post("/opportunities/{id}/save", d.Opportunities.Save)
			r.Delete("/opportunities/{id}/save", d.Opportunities.Unsave)
			r.Post("/opportunities/{id}/ideas", d.Ideas.Generate)
			r.Get("/opportunities/{id}/ideas", d.Ideas.List)

			r.Get("/ideas", d.Ideas.List)
			r.Get("/ideas/{id}/messages", d.Ideas.ListMessages)
			r.Post("/ideas/{id}/messages", d.Ideas.StreamMessage)
			r.Post("/ideas/{id}/confirm", d.Ideas.Confirm)

			r.Post("/research", d.Research.Start)

			r.Get("/projects", d.Projects.List)
			r.Post("/projects", d.Projects.Create)
			r.Get("/projects/{id}", d.Projects.Get)
			r.Put("/projects/{id}/config", d.Projects.UpdateConfig)
			r.Post("/projects/{id}/ebook", d.Projects.GenerateEbook)
			r.Post("/projects/{id}/cover", d.Projects.GenerateCover)
			r.Post("/projects/{id}/posters", d.Projects.GeneratePosters)
			r.Post("/projects/{id}/sales-page", d.Projects.GenerateSalesPage)
			r.Post("/projects/{id}/video-ads", d.Projects.GenerateVideoAd)
			r.Get("/projects/{id}/assets", d.Assets.List)

			r.Get("/assets/{id}/download", d.Assets.Download)
			r.Get("/jobs/{id}", d.Jobs.Get)
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
