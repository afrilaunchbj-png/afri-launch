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

	appai "afrilaunch/backend/internal/application/ai"
	authapp "afrilaunch/backend/internal/application/auth"
	assetsapp "afrilaunch/backend/internal/application/assets"
	creditsapp "afrilaunch/backend/internal/application/credits"
	documentapp "afrilaunch/backend/internal/application/document"
	ideasapp "afrilaunch/backend/internal/application/ideas"
	"afrilaunch/backend/internal/application/jobs"
	opportunitiesapp "afrilaunch/backend/internal/application/opportunities"
	projectsapp "afrilaunch/backend/internal/application/projects"
	"afrilaunch/backend/internal/config"
	authinfra "afrilaunch/backend/internal/infra/auth"
	aiinfra "afrilaunch/backend/internal/infra/ai"
	renderinfra "afrilaunch/backend/internal/infra/render"
	"afrilaunch/backend/internal/infra/postgres"
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
	projectRepo := postgres.NewProjectRepository(store)
	assetRepo := postgres.NewAssetRepository(store)
	jobRepo := postgres.NewJobRepository(store)
	objStorage := storage.NewLocalStorage(cfg.StorageDir)

	verifier, err := authinfra.NewNeonVerifier(cfg.NeonAuthBaseURL, cfg.NeonAuthJWKSURL)
	if err != nil {
		slog.Error("cannot configure auth verifier", "err", err)
		os.Exit(1)
	}

	// Services (application).
	authSvc := authapp.NewService(users, creditRepo)
	creditSvc := creditsapp.NewService(creditRepo)
	oppSvc := opportunitiesapp.NewService(oppRepo)

	// Providers IA (OpenAI = LLM + image, HeyGen = vidéo) + routage de modèle.
	openaiClient := aiinfra.NewOpenAI(cfg.OpenAIAPIKey, "")
	heyGenClient := aiinfra.NewHeyGen(cfg.HeyGenAPIKey, cfg.HeyGenAPIURL)
	modelRouter := appai.NewModelRouter(cfg.OpenAIResearchModel, cfg.OpenAIIdeationModel, cfg.OpenAIImageModel)
	aiSvc := appai.NewService(openaiClient, openaiClient, heyGenClient, modelRouter)

	// Rendu documents (HTML → PDF/PPTX via Chrome headless).
	renderer := renderinfra.NewChromedpRenderer(cfg.ChromePath)
	docSvc := documentapp.NewService(aiSvc, renderer)

	// Worker asynchrone de génération (idées, ebook, assets).
	worker := jobs.NewWorker(jobRepo, creditRepo, ideaRepo, projectRepo, assetRepo, oppRepo, objStorage, aiSvc, docSvc)

	// Services applicatifs.
	ideaSvc := ideasapp.NewService(worker, ideaRepo, oppRepo)
	projectSvc := projectsapp.NewService(worker, projectRepo)
	assetSvc := assetsapp.NewService(assetRepo, objStorage)

	// Handlers HTTP.
	authH := handler.NewAuthHandler(authSvc, int64(cfg.WelcomeCredits))
	creditH := handler.NewCreditHandler(creditSvc)
	oppH := handler.NewOpportunityHandler(oppSvc)
	marketH := handler.NewMarketHandler(marketRepo)
	ideaH := handler.NewIdeaHandler(ideaSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	assetH := handler.NewAssetHandler(assetSvc)
	jobH := handler.NewJobHandler(jobRepo)
	healthH := handler.NewHealth(store)

	router := server.NewRouter(server.Deps{
		Config:        cfg,
		Tokens:        verifier,
		Health:        healthH,
		Auth:          authH,
		Credits:       creditH,
		Opportunities: oppH,
		Markets:       marketH,
		Ideas:         ideaH,
		Projects:      projectH,
		Assets:        assetH,
		Jobs:          jobH,
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
