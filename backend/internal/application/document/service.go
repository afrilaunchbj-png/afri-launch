package document

import (
	"context"
	"strings"

	"afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/port"
)

// Service génère des documents : LLM → HTML → PDF (ebook) ou PPTX (deck).
type Service struct {
	ai     *ai.Service
	render port.Renderer
}

// NewService construit le service de génération de documents.
func NewService(ai *ai.Service, render port.Renderer) *Service {
	return &Service{ai: ai, render: render}
}

// GenerateEbook génère un ebook (HTML → PDF).
func (s *Service) GenerateEbook(ctx context.Context, req EbookRequest) ([]byte, error) {
	html, err := s.generateHTML(ctx, BuildEbookPrompt(req))
	if err != nil {
		return nil, err
	}
	return s.render.HTMLToPDF(ctx, html)
}

// GenerateDeck génère un deck (HTML slides → PPTX image-par-slide).
func (s *Service) GenerateDeck(ctx context.Context, req DeckRequest) ([]byte, error) {
	html, err := s.generateHTML(ctx, BuildDeckPrompt(req))
	if err != nil {
		return nil, err
	}
	return s.render.SlidesToPPTX(ctx, html)
}

func (s *Service) generateHTML(ctx context.Context, prompt Prompt) ([]byte, error) {
	resp, err := s.ai.Complete(ctx, ai.TaskContent, prompt.System, prompt.User)
	if err != nil {
		return nil, err
	}
	return []byte(stripCodeFences(resp.Content)), nil
}

// stripCodeFences retire un éventuel bloc ```html ... ``` autour du HTML.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```html")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
