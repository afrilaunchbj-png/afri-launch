package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminapp "afrilaunch/backend/internal/application/admin"
	advapp "afrilaunch/backend/internal/application/advertising"
	appai "afrilaunch/backend/internal/application/ai"
	assetsapp "afrilaunch/backend/internal/application/assets"
	auditapp "afrilaunch/backend/internal/application/audit"
	authapp "afrilaunch/backend/internal/application/auth"
	chatapp "afrilaunch/backend/internal/application/chat"
	creditsapp "afrilaunch/backend/internal/application/credits"
	dashboardapp "afrilaunch/backend/internal/application/dashboard"
	documentapp "afrilaunch/backend/internal/application/document"
	ideasapp "afrilaunch/backend/internal/application/ideas"
	"afrilaunch/backend/internal/application/jobs"
	opportunitiesapp "afrilaunch/backend/internal/application/opportunities"
	"afrilaunch/backend/internal/application/port"
	preferencesapp "afrilaunch/backend/internal/application/preferences"
	projectsapp "afrilaunch/backend/internal/application/projects"
	researchapp "afrilaunch/backend/internal/application/research"
	supportapp "afrilaunch/backend/internal/application/support"
	videoadapp "afrilaunch/backend/internal/application/videoad"
	"afrilaunch/backend/internal/config"
	"afrilaunch/backend/internal/domain"
	adsgoogle "afrilaunch/backend/internal/infra/ads/googleads"
	adsmeta "afrilaunch/backend/internal/infra/ads/meta"
	adstiktok "afrilaunch/backend/internal/infra/ads/tiktok"
	aiinfra "afrilaunch/backend/internal/infra/ai"
	authinfra "afrilaunch/backend/internal/infra/auth"
	cryptoinfra "afrilaunch/backend/internal/infra/crypto"
	eventsinfra "afrilaunch/backend/internal/infra/events"
	"afrilaunch/backend/internal/infra/postgres"
	renderinfra "afrilaunch/backend/internal/infra/render"
	"afrilaunch/backend/internal/infra/storage"
	"afrilaunch/backend/internal/server"
	"afrilaunch/backend/internal/server/handler"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("cannot open postgres pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)

	// Adapters (infra).
	users := postgres.NewUserRepository(store)
	creditRepo := postgres.NewCreditRepository(store)
	oppRepo := postgres.NewOpportunityRepository(store)
	marketRepo := postgres.NewMarketRepository(store)
	ideaRepo := postgres.NewIdeaRepository(store)
	ideaMessageRepo := postgres.NewIdeaMessageRepository(store)
	projectRepo := postgres.NewProjectRepository(store)
	assetRepo := postgres.NewAssetRepository(store)
	jobRepo := postgres.NewJobRepository(store)
	researchRepo := postgres.NewResearchRepository(store)

	// Stockage objet : S3-compatible (Neon) si configuré, sinon disque local (dev).
	var objStorage port.Storage = storage.NewLocalStorage(cfg.StorageDir)
	if cfg.S3Bucket != "" {
		s3Store, err := storage.NewS3(ctx, cfg.S3Endpoint, cfg.S3Region, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3Bucket, cfg.S3PathStyle)
		if err != nil {
			slog.Error("cannot configure s3 storage", "err", err)
			os.Exit(1)
		}
		objStorage = s3Store
		slog.Info("object storage", "provider", "s3", "bucket", cfg.S3Bucket)
	}

	verifier, err := authinfra.NewNeonVerifier(cfg.NeonAuthBaseURL, cfg.NeonAuthJWKSURL)
	if err != nil {
		slog.Error("cannot configure auth verifier", "err", err)
		os.Exit(1)
	}

	// Recorder d'audit (journal des opérations sensibles).
	auditRec := auditapp.NewRecorder(postgres.NewAuditRepository(store))

	// Services (application).
	authSvc := authapp.NewService(users, creditRepo, cfg.SuperadminEmails, auditRec)
	creditSvc := creditsapp.NewService(creditRepo)
	oppSvc := opportunitiesapp.NewService(oppRepo)

	// Providers IA (OpenAI = LLM + image + recherche, HeyGen = vidéo) + routage.
	openaiClient := aiinfra.NewOpenAI(cfg.OpenAIAPIKey, "")
	heyGenClient := aiinfra.NewHeyGen(cfg.HeyGenAPIKey, cfg.HeyGenAPIURL)
	modelRouter := appai.NewModelRouter(cfg.OpenAIResearchModel, cfg.OpenAIIdeationModel, cfg.OpenAIImageModel)
	aiSvc := appai.NewService(openaiClient, openaiClient, heyGenClient, openaiClient, modelRouter)

	// Rendu documents (HTML → PDF/PPTX via Chrome headless).
	renderer := renderinfra.NewChromedpRenderer(cfg.ChromePath)
	docSvc := documentapp.NewService(aiSvc, renderer)

	// Vidéos publicitaires : créatif (LLM) + montage (FFmpeg).
	videoadSvc := videoadapp.NewService(aiSvc)
	videoRenderer := renderinfra.NewFFmpegRenderer(renderer, cfg.FFmpegPath, "")

	// Canal temps réel in-process (SSE unique) : chat + jobs.
	// Multi-instance : à migrer sur Redis pub/sub avec asynq.
	eventBus := eventsinfra.NewBroker()

	// Worker asynchrone de génération (idées, ebook, assets, recherche, vidéos).
	worker := jobs.NewWorker(jobRepo, creditRepo, ideaRepo, projectRepo, assetRepo, oppRepo, researchRepo, objStorage, aiSvc, docSvc, eventBus,
		videoadSvc, videoRenderer, videoadapp.ProviderDefaults{AvatarID: cfg.HeyGenDefaultAvatarID, VoiceID: cfg.HeyGenDefaultVoiceID})

	// Services applicatifs.
	ideaSvc := ideasapp.NewService(worker, ideaRepo, ideaMessageRepo, oppRepo, creditRepo, aiSvc)
	projectSvc := projectsapp.NewService(worker, projectRepo, ideaRepo, assetRepo)
	assetSvc := assetsapp.NewService(assetRepo, objStorage)
	researchSvc := researchapp.NewService(worker, researchRepo)
	prefRepo := postgres.NewPreferenceRepository(store)
	prefSvc := preferencesapp.NewService(prefRepo)
	supportRepo := postgres.NewSupportRepository(store)
	supportSvc := supportapp.NewService(supportRepo, auditRec)
	adminRepo := postgres.NewAdminRepository(store)
	adminSvc := adminapp.NewService(adminRepo, supportRepo, auditRec)
	dashSvc := dashboardapp.NewService(postgres.NewDashboardRepository(store))
	chatRepo := postgres.NewConversationRepository(store)
	chatSvc := chatapp.NewService(chatRepo, ideaRepo, oppRepo, creditRepo, prefRepo, aiSvc, eventBus)

	// Intégrations publicitaires (ADR-017) : Meta au MVP, providers additionnels
	// par simple enregistrement dans la registry.
	encryptor, err := cryptoinfra.NewEncryptor(cfg.EncryptionKey, cfg.EncryptionKeyVersion)
	if err != nil {
		slog.Error("cannot configure encryption", "err", err)
		os.Exit(1)
	}
	providers := advapp.ProviderRegistry{}
	if cfg.MetaAppID != "" && cfg.MetaAppSecret != "" {
		providers[domain.AdPlatformMeta] = adsmeta.New(
			cfg.MetaAppID, cfg.MetaAppSecret, cfg.MetaGraphVersion,
			cfg.MetaOAuthRedirectURI, cfg.MetaOAuthScopes,
		)
	}
	if cfg.GoogleAdsClientID != "" && cfg.GoogleAdsClientSecret != "" && cfg.GoogleAdsDevToken != "" {
		providers[domain.AdPlatformGoogleAds] = adsgoogle.New(
			cfg.GoogleAdsClientID, cfg.GoogleAdsClientSecret, cfg.GoogleAdsDevToken,
			cfg.GoogleAdsLoginCustID, cfg.GoogleAdsAPIVersion,
		)
	}
	if cfg.TikTokAppID != "" && cfg.TikTokAppSecret != "" {
		providers[domain.AdPlatformTikTokAds] = adstiktok.New(
			cfg.TikTokAppID, cfg.TikTokAppSecret, cfg.TikTokRedirectURI,
		)
	}
	var adSigner port.StorageSigner
	if s3Store, ok := objStorage.(*storage.S3); ok {
		adSigner = s3Store
	}
	advSvc := advapp.NewService(
		providers,
		encryptor,
		postgres.NewOAuthStateStore(store),
		postgres.NewAdConnectionRepository(store, encryptor),
		postgres.NewAdCampaignRepository(store),
		postgres.NewAdCreativeRepository(store),
		postgres.NewAdInsightRepository(store),
		postgres.NewProviderOperationRepository(store),
		assetRepo,
		objStorage,
		adSigner,
		advapp.DefaultSafetyPolicy(),
	)

	// Handlers HTTP.
	authH := handler.NewAuthHandler(authSvc, int64(cfg.WelcomeCredits))
	creditH := handler.NewCreditHandler(creditSvc)
	oppH := handler.NewOpportunityHandler(oppSvc)
	marketH := handler.NewMarketHandler(marketRepo)
	ideaH := handler.NewIdeaHandler(ideaSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	assetH := handler.NewAssetHandler(assetSvc)
	jobH := handler.NewJobHandler(jobRepo)
	researchH := handler.NewResearchHandler(researchSvc)
	chatH := handler.NewConversationHandler(chatSvc)
	eventsH := handler.NewEventHandler(eventBus)
	prefH := handler.NewPreferenceHandler(prefSvc)
	supportH := handler.NewSupportHandler(supportSvc)
	integrationsH := handler.NewIntegrationHandler(advSvc, cfg.AppURL, map[string]string{
		domain.AdPlatformMeta:      cfg.MetaOAuthRedirectURI,
		domain.AdPlatformGoogleAds: cfg.GoogleAdsRedirectURI,
		domain.AdPlatformTikTokAds: cfg.TikTokRedirectURI,
	})
	adminH := handler.NewAdminHandler(adminSvc)
	dashboardH := handler.NewDashboardHandler(dashSvc)
	healthH := handler.NewHealth(store)

	router := server.NewRouter(server.Deps{
		Config:        cfg,
		Tokens:        verifier,
		Users:         users,
		Health:        healthH,
		Auth:          authH,
		Credits:       creditH,
		Opportunities: oppH,
		Markets:       marketH,
		Ideas:         ideaH,
		Projects:      projectH,
		Assets:        assetH,
		Jobs:          jobH,
		Research:      researchH,
		Conversations: chatH,
		Events:        eventsH,
		Preferences:   prefH,
		Support:       supportH,
		Integrations:  integrationsH,
		Admin:         adminH,
		Dashboard:     dashboardH,
		AI:            aiSvc,
		Documents:     docSvc,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("api listening", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	slog.Info("api stopped")
}
