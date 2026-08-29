// Package jobs orchestre les générations asynchrones (worker in-process).
// À terme, remplacé par asynq (Redis) — la structure (jobs + statuts) est conservée.
package jobs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/document"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

type task struct {
	userID string
	jobID  string
}

// Worker traite les jobs de génération en arrière-plan.
type Worker struct {
	tasks         chan task
	jobs          port.JobRepository
	credits       port.CreditRepository
	ideas         port.IdeaRepository
	projects      port.ProjectRepository
	assets        port.AssetRepository
	opportunities port.OpportunityRepository
	storage       port.Storage
	ai            *ai.Service
	docs          *document.Service
}

// NewWorker construit le worker et démarre un pool de goroutines.
func NewWorker(
	jobs port.JobRepository,
	credits port.CreditRepository,
	ideas port.IdeaRepository,
	projects port.ProjectRepository,
	assets port.AssetRepository,
	opportunities port.OpportunityRepository,
	storage port.Storage,
	ai *ai.Service,
	docs *document.Service,
) *Worker {
	w := &Worker{
		tasks:         make(chan task, 64),
		jobs:          jobs,
		credits:       credits,
		ideas:         ideas,
		projects:      projects,
		assets:        assets,
		opportunities: opportunities,
		storage:       storage,
		ai:            ai,
		docs:          docs,
	}
	for i := 0; i < 3; i++ {
		go w.run()
	}
	return w
}

// Dispatch crée un job, réserve les crédits et le met en file.
func (w *Worker) Dispatch(ctx context.Context, userID string, projectID, opportunityID *string, kind string) (domain.GenerationJob, error) {
	op := operationFor(kind)
	if op == "" {
		return domain.GenerationJob{}, fmt.Errorf("unknown job kind %q", kind)
	}
	cost, err := w.credits.GetGenerationCost(ctx, op)
	if err != nil {
		return domain.GenerationJob{}, err
	}

	job, err := w.jobs.Create(ctx, userID, projectID, opportunityID, kind, cost.Credits)
	if err != nil {
		return domain.GenerationJob{}, err
	}

	if _, err := w.credits.Reserve(ctx, userID, cost.Credits, op, "job:"+job.ID, 24*time.Hour); err != nil {
		_, _ = w.jobs.Fail(ctx, job.ID, "crédits insuffisants")
		return domain.GenerationJob{}, err
	}

	w.tasks <- task{userID: userID, jobID: job.ID}
	return job, nil
}

func (w *Worker) run() {
	for t := range w.tasks {
		w.process(t)
	}
}

func (w *Worker) process(t task) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	job, err := w.jobs.Get(ctx, t.userID, t.jobID)
	if err != nil {
		return
	}
	_, _ = w.jobs.UpdateStatus(ctx, job.ID, domain.JobProcessing)

	result, err := w.runKind(ctx, job)
	if err != nil {
		_ = w.credits.Release(ctx, job.UserID, "job:"+job.ID)
		_, _ = w.jobs.Fail(ctx, job.ID, err.Error())
		if job.ProjectID != nil {
			_, _ = w.projects.UpdateStatus(ctx, job.UserID, *job.ProjectID, domain.ProjectFailed)
		}
		return
	}

	_, _ = w.credits.Consume(ctx, job.UserID, "job:"+job.ID)
	_, _ = w.jobs.Complete(ctx, job.ID, result, job.Cost)

	if job.ProjectID != nil {
		switch job.Kind {
		case domain.JobEbook:
			_, _ = w.projects.UpdateStatus(ctx, job.UserID, *job.ProjectID, domain.ProjectContentReady)
		case domain.JobCover, domain.JobPosters, domain.JobSalesPage:
			_, _ = w.projects.UpdateStatus(ctx, job.UserID, *job.ProjectID, domain.ProjectCompleted)
		}
	}
}

func (w *Worker) runKind(ctx context.Context, job domain.GenerationJob) ([]byte, error) {
	switch job.Kind {
	case domain.JobIdeas:
		return w.runIdeas(ctx, job)
	case domain.JobEbook:
		return w.runEbook(ctx, job)
	case domain.JobCover:
		return w.runCover(ctx, job)
	case domain.JobPosters:
		return w.runPosters(ctx, job)
	case domain.JobSalesPage:
		return w.runSalesPage(ctx, job)
	default:
		return nil, fmt.Errorf("unknown job kind %q", job.Kind)
	}
}

func (w *Worker) runIdeas(ctx context.Context, job domain.GenerationJob) ([]byte, error) {
	if job.OpportunityID == nil {
		return nil, errors.New("opportunité manquante")
	}
	opp, err := w.opportunities.Get(ctx, *job.OpportunityID)
	if err != nil {
		return nil, err
	}

	var out struct {
		Ideas []ideaInput `json:"ideas"`
	}
	if err := w.ai.CompleteJSON(ctx, ai.TaskIdeation, ideaSystem, ideaPrompt(opp), &out); err != nil {
		return nil, err
	}
	if len(out.Ideas) == 0 {
		return nil, errors.New("aucune idée générée")
	}

	ids := make([]string, 0, len(out.Ideas))
	for _, in := range out.Ideas {
		idea, err := w.ideas.Create(ctx, domain.ProductIdea{
			UserID:           job.UserID,
			OpportunityID:    job.OpportunityID,
			Title:            in.Title,
			Subtitle:         in.Subtitle,
			Audience:         in.Audience,
			Problem:          in.Problem,
			Promise:          in.Promise,
			Format:           in.Format,
			EstimatedPrice:   in.EstimatedPrice,
			Difficulty:       in.Difficulty,
			MarketEvidence:   in.MarketEvidence,
			WhyNow:           in.WhyNow,
			CompetitiveAngle: in.CompetitiveAngle,
		})
		if err != nil {
			return nil, err
		}
		ids = append(ids, idea.ID)
	}
	return json.Marshal(map[string]any{"idea_ids": ids})
}

func (w *Worker) runEbook(ctx context.Context, job domain.GenerationJob) ([]byte, error) {
	c, err := w.context(ctx, job)
	if err != nil {
		return nil, err
	}
	ebookReq := document.EbookRequest{
		Topic:    c.topic,
		Audience: c.audience,
		Language: c.language,
		Country:  c.country,
		Product:  c.format,
	}

	// Version portrait (PDF).
	pdf, err := w.docs.GenerateEbook(ctx, ebookReq)
	if err != nil {
		return nil, err
	}
	if _, err := w.storeAsset(ctx, job, domain.AssetEbookPDF, c.topic+".pdf", "application/pdf", pdf); err != nil {
		return nil, err
	}

	// Version paysage (PPTX, destinée à l'export PPT).
	deck, err := w.docs.GenerateEbookDeck(ctx, ebookReq)
	if err != nil {
		return nil, err
	}
	deckAsset, err := w.storeAsset(ctx, job, domain.AssetEbookDeck, c.topic+".pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation", deck)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]any{"asset_ids": []string{deckAsset.ID}})
}

func (w *Worker) runCover(ctx context.Context, job domain.GenerationJob) ([]byte, error) {
	c, err := w.context(ctx, job)
	if err != nil {
		return nil, err
	}
	img, err := w.ai.GenerateImage(ctx, coverPrompt(c.topic, c.audience))
	if err != nil {
		return nil, err
	}
	png, err := base64.StdEncoding.DecodeString(img.B64JSON)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	asset, err := w.storeAsset(ctx, job, domain.AssetCover, "cover.png", "image/png", png)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"asset_id": asset.ID})
}

// runPosters génère 3 affiches publicitaires.
func (w *Worker) runPosters(ctx context.Context, job domain.GenerationJob) ([]byte, error) {
	c, err := w.context(ctx, job)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, 3)
	for variant := 1; variant <= 3; variant++ {
		img, err := w.ai.GenerateImage(ctx, posterPrompt(c, variant))
		if err != nil {
			return nil, err
		}
		png, err := base64.StdEncoding.DecodeString(img.B64JSON)
		if err != nil {
			return nil, fmt.Errorf("decode image: %w", err)
		}
		asset, err := w.storeAsset(ctx, job, domain.AssetPoster, fmt.Sprintf("poster-%d.png", variant), "image/png", png)
		if err != nil {
			return nil, err
		}
		ids = append(ids, asset.ID)
	}
	return json.Marshal(map[string]any{"asset_ids": ids})
}

func (w *Worker) runSalesPage(ctx context.Context, job domain.GenerationJob) ([]byte, error) {
	c, err := w.context(ctx, job)
	if err != nil {
		return nil, err
	}
	html, err := w.docs.GenerateSalesPage(ctx, document.SalesPageRequest{
		Product:  c.topic,
		Promise:  c.promise,
		Audience: c.audience,
		Language: c.language,
		Country:  c.country,
		Price:    c.price,
	})
	if err != nil {
		return nil, err
	}
	asset, err := w.storeAsset(ctx, job, domain.AssetSalesPage, "sales-page.html", "text/html", html)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"asset_id": asset.ID})
}

// genContext regroupe les infos du projet/idée/opportunité pour les prompts.
type genContext struct {
	topic    string
	audience string
	language string
	country  string
	format   string
	promise  string
	price    string
}

func (w *Worker) context(ctx context.Context, job domain.GenerationJob) (genContext, error) {
	if job.ProjectID == nil {
		return genContext{}, errors.New("projet manquant")
	}
	project, err := w.projects.Get(ctx, job.UserID, *job.ProjectID)
	if err != nil {
		return genContext{}, err
	}
	c := genContext{topic: project.Title, language: "fr"}

	if project.IdeaID != nil {
		if idea, err := w.ideas.Get(ctx, job.UserID, *project.IdeaID); err == nil {
			c.audience = idea.Audience
			c.format = idea.Format
			c.promise = idea.Promise
			if idea.Title != "" {
				c.topic = idea.Title
			}
		}
	}
	if project.OpportunityID != nil {
		if opp, err := w.opportunities.Get(ctx, *project.OpportunityID); err == nil {
			c.country = opp.Country
			if opp.Language != "" {
				c.language = opp.Language
			}
		}
	}
	if c.format == "" {
		c.format = "ebook"
	}
	return c, nil
}

func (w *Worker) storeAsset(ctx context.Context, job domain.GenerationJob, kind, filename, contentType string, data []byte) (domain.Asset, error) {
	if job.ProjectID == nil {
		return domain.Asset{}, errors.New("projet manquant")
	}
	key := fmt.Sprintf("projects/%s/%s", *job.ProjectID, filename)
	if err := w.storage.Put(ctx, key, data, contentType); err != nil {
		return domain.Asset{}, err
	}
	return w.assets.Create(ctx, domain.Asset{
		UserID:      job.UserID,
		ProjectID:   *job.ProjectID,
		Kind:        kind,
		StorageKey:  key,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   int64(len(data)),
	})
}

func operationFor(kind string) string {
	switch kind {
	case domain.JobIdeas:
		return domain.OperationIdeaGeneration
	case domain.JobEbook:
		return domain.OperationEbookGen
	case domain.JobCover:
		return domain.OperationImageGen
	case domain.JobPosters:
		return domain.OperationPosterGen
	case domain.JobSalesPage:
		return domain.OperationSalesPage
	default:
		return ""
	}
}
