package ai

import (
	"context"
	"encoding/json"
	"strings"

	"afrilaunch/backend/internal/application/port"
)

// Service expose les providers IA aux use cases/workers, avec le routage de
// modèle par tâche.
type Service struct {
	llm      port.LLMProvider
	images   port.ImageProvider
	video    port.VideoProvider
	research port.ResearchProvider
	router   ModelRouter
}

// NewService construit le service IA.
func NewService(llm port.LLMProvider, images port.ImageProvider, video port.VideoProvider, research port.ResearchProvider, router ModelRouter) *Service {
	return &Service{llm: llm, images: images, video: video, research: research, router: router}
}

// Complete génère du texte pour une tâche donnée (modèle choisi par le routeur).
func (s *Service) Complete(ctx context.Context, task Task, system, prompt string) (port.LLMResponse, error) {
	return s.llm.Complete(ctx, port.LLMRequest{
		Model:    s.router.ModelFor(task),
		System:   system,
		Messages: []port.LLMMessage{{Role: "user", Content: prompt}},
	})
}

// StreamComplete génère du texte en continu pour une tâche donnée.
func (s *Service) StreamComplete(ctx context.Context, task Task, system, prompt string, emit func(delta string) error) error {
	return s.StreamMessages(ctx, task, system, []port.LLMMessage{{Role: "user", Content: prompt}}, emit)
}

// StreamMessages génère du texte en continu avec une conversation
// multi-tours complète (historique user/assistant).
func (s *Service) StreamMessages(ctx context.Context, task Task, system string, messages []port.LLMMessage, emit func(delta string) error) error {
	return s.llm.StreamComplete(ctx, port.LLMRequest{
		Model:    s.router.ModelFor(task),
		System:   system,
		Messages: messages,
	}, emit)
}

// CompleteJSON génère une sortie structurée (JSON) et la décode dans out.
func (s *Service) CompleteJSON(ctx context.Context, task Task, system, prompt string, out any) error {
	resp, err := s.llm.Complete(ctx, port.LLMRequest{
		Model:    s.router.ModelFor(task),
		System:   system,
		Messages: []port.LLMMessage{{Role: "user", Content: prompt}},
		JSONMode: true,
	})
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(stripFences(resp.Content)), out)
}

// GenerateImage génère une image avec le modèle image configuré.
func (s *Service) GenerateImage(ctx context.Context, prompt string) (port.Image, error) {
	return s.images.Generate(ctx, port.ImageRequest{
		Model:  s.router.ModelFor(TaskImage),
		Prompt: prompt,
	})
}

// Research recherche en ligne (web search) avec le provider configuré.
func (s *Service) Research(ctx context.Context, system, query string) (port.ResearchResult, error) {
	return s.research.Research(ctx, port.ResearchRequest{
		Model:  s.router.ModelFor(TaskResearch),
		System: system,
		Query:  query,
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

// stripFences retire un éventuel bloc ```json ... ``` autour de la sortie.
func stripFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
