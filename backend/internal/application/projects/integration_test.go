package projects_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	appai "afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/document"
	"afrilaunch/backend/internal/application/jobs"
	"afrilaunch/backend/internal/application/port"
	projectsapp "afrilaunch/backend/internal/application/projects"
	videoad "afrilaunch/backend/internal/application/videoad"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres"
	"afrilaunch/backend/internal/infra/storage"
)

// Fakes "hors ligne" : le worker échoue proprement (release des crédits)
// sans jamais paniquer pendant le test.
type errLLM struct{}

func (errLLM) Complete(ctx context.Context, req port.LLMRequest) (port.LLMResponse, error) {
	return port.LLMResponse{}, errors.New("offline")
}
func (errLLM) StreamComplete(ctx context.Context, req port.LLMRequest, emit func(string) error) error {
	return errors.New("offline")
}

type errImage struct{}

func (errImage) Generate(ctx context.Context, req port.ImageRequest) (port.Image, error) {
	return port.Image{}, errors.New("offline")
}

type errVideo struct{}

func (errVideo) Submit(ctx context.Context, req port.VideoRequest) (string, error) {
	return "", errors.New("offline")
}
func (errVideo) Status(ctx context.Context, id string) (port.VideoResult, error) {
	return port.VideoResult{}, errors.New("offline")
}

type errResearch struct{}

func (errResearch) Research(ctx context.Context, req port.ResearchRequest) (port.ResearchResult, error) {
	return port.ResearchResult{}, errors.New("offline")
}

// TestProjectsCoverFirstGate valide la configuration projet (palette, pages)
// et le gate cover-first contre une vraie base PostgreSQL.
// Activé uniquement lorsque AFRILAUNCH_TEST_DB est défini.
func TestProjectsCoverFirstGate(t *testing.T) {
	url := os.Getenv("AFRILAUNCH_TEST_DB")
	if url == "" {
		t.Skip("AFRILAUNCH_TEST_DB non défini — test d'intégration ignoré")
	}

	ctx := context.Background()
	pool, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	users := postgres.NewUserRepository(store)
	credits := postgres.NewCreditRepository(store)
	ideas := postgres.NewIdeaRepository(store)
	opps := postgres.NewOpportunityRepository(store)
	projectRepo := postgres.NewProjectRepository(store)
	assetRepo := postgres.NewAssetRepository(store)
	jobRepo := postgres.NewJobRepository(store)
	researchRepo := postgres.NewResearchRepository(store)
	objStorage := storage.NewLocalStorage(t.TempDir())

	aiSvc := appai.NewService(errLLM{}, errImage{}, errVideo{}, errResearch{}, appai.NewModelRouter("r", "i", "m"))
	docSvc := document.NewService(aiSvc, nil)
	worker := jobs.NewWorker(jobRepo, credits, ideas, projectRepo, assetRepo, opps, researchRepo, objStorage, aiSvc, docSvc, nil, nil, nil, videoad.ProviderDefaults{})

	svc := projectsapp.NewService(worker, projectRepo, ideas, assetRepo)

	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: fmt.Sprintf("proj-test-%d@example.com", time.Now().UnixNano()), FullName: "Proj Test"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := credits.Grant(ctx, user.ID, 100, domain.OperationWelcomeBonus, "welcome:"+user.ID); err != nil {
		t.Fatalf("grant: %v", err)
	}

	idea, err := ideas.Create(ctx, domain.ProductIdea{UserID: user.ID, Title: "Guide couture", Status: domain.IdeaDraft})
	if err != nil {
		t.Fatalf("create idea: %v", err)
	}
	if _, err := ideas.SetStatus(ctx, user.ID, idea.ID, domain.IdeaConfirmed); err != nil {
		t.Fatalf("confirm idea: %v", err)
	}

	project, err := svc.Create(ctx, user.ID, nil, &idea.ID, "Guide couture")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	// Gate cover-first : ebook interdit sans cover.
	if _, err := svc.GenerateEbook(ctx, user.ID, project.ID); !errors.Is(err, domain.ErrCoverRequired) {
		t.Fatalf("expected ErrCoverRequired, got %v", err)
	}

	// Configuration : palette utilisateur + pages (bornées).
	palette := domain.ProjectPalette{Primary: "#0F766E", Secondary: "#F59E0B", Background: "#FFF", Text: "#1c1917"}
	minP, maxP := 99, 3
	if _, err := svc.UpdateConfig(ctx, user.ID, project.ID, projectsapp.ConfigInput{
		Palette:       &palette,
		EbookMinPages: &minP,
		EbookMaxPages: &maxP,
	}); err != nil {
		t.Fatalf("update config: %v", err)
	}

	updated, err := projectRepo.Get(ctx, user.ID, project.ID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	cfg := domain.ParseProjectConfig(updated.Config)
	if cfg.Palette == nil || cfg.Palette.Primary != "#0f766e" || cfg.Palette.Source != domain.PaletteSourceUser {
		t.Fatalf("palette non persistée : %+v", cfg.Palette)
	}
	if cfg.EbookMinPages != domain.EbookMaxPagesCeiling {
		t.Errorf("min pages = %d, want clamp %d", cfg.EbookMinPages, domain.EbookMaxPagesCeiling)
	}

	// Une cover existe → le gate passe et le job est créé (crédits réservés).
	asset, err := assetRepo.Create(ctx, domain.Asset{
		UserID: user.ID, ProjectID: project.ID, Kind: domain.AssetCover,
		StorageKey: "projects/" + project.ID + "/cover.png", Filename: "cover.png",
		ContentType: "image/png", SizeBytes: 4,
	})
	if err != nil {
		t.Fatalf("create cover asset: %v", err)
	}
	if err := objStorage.Put(ctx, asset.StorageKey, []byte("PNG!"), "image/png"); err != nil {
		t.Fatalf("put cover: %v", err)
	}

	job, err := svc.GenerateEbook(ctx, user.ID, project.ID)
	if err != nil {
		t.Fatalf("generate ebook with cover: %v", err)
	}
	if job.Kind != domain.JobEbook || job.Status != domain.JobPending {
		t.Fatalf("job inattendu : %+v", job)
	}

	// Régénération de cover avec instructions (paramètres persistés sur le job).
	coverJob, err := svc.GenerateCover(ctx, user.ID, project.ID, "plus minimaliste")
	if err != nil {
		t.Fatalf("generate cover: %v", err)
	}
	if len(coverJob.Params) == 0 || !strings.Contains(string(coverJob.Params), "minimaliste") {
		t.Errorf("instructions non persistées sur le job : %q", string(coverJob.Params))
	}
}
