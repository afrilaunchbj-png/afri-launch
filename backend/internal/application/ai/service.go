package ai

import (
	"context"

	"afrilaunch/backend/internal/application/port"
)

// Service expose les providers IA aux use cases/workers, avec le routage de
// modèle par tâche.
type Service struct {
	llm    port.LLMProvider
	images port.ImageProvider
	video  port.VideoProvider
	router ModelRouter
}

// NewService construit le service IA.
func NewService(llm port.LLMProvider, images port.ImageProvider, video port.VideoProvider, router ModelRouter) *Service {
	return &Service{llm: llm, images: images, video: video, router: router}
}

// Complete génère du texte pour une tâche donnée (modèle choisi par le routeur).
func (s *Service) Complete(ctx context.Context, task Task, system, prompt string) (port.LLMResponse, error) {
	return s.llm.Complete(ctx, port.LLMRequest{
		Model:    s.router.ModelFor(task),
		System:   system,
		Messages: []port.LLMMessage{{Role: "user", Content: prompt}},
	})
}

// GenerateImage génère une image avec le modèle image configuré.
func (s *Service) GenerateImage(ctx context.Context, prompt string) (port.Image, error) {
	return s.images.Generate(ctx, port.ImageRequest{
		Model:  s.router.ModelFor(TaskImage),
		Prompt: prompt,
	})
}

// SubmitVideo soumet une génération de vidéo avatar (async).
func (s *Service) SubmitVideo(ctx context.Context, req port.VideoRequest) (string, error) {
	return s.video.Submit(ctx, req)
}

// VideoStatus interroge l'état d'un job vidéo.
func (s *Service) VideoStatus(ctx context.Context, id string) (port.VideoResult, error) {
	return s.video.Status(ctx, id)
}
