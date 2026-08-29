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
	html = ensureChapterPageBreaks(html)
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

// GenerateEbookDeck génère la version paysage d'un ebook (HTML slides → PPTX).
func (s *Service) GenerateEbookDeck(ctx context.Context, req EbookRequest) ([]byte, error) {
	html, err := s.generateHTML(ctx, BuildEbookDeckPrompt(req))
	if err != nil {
		return nil, err
	}
	return s.render.SlidesToPPTX(ctx, html)
}

// GenerateSalesPage génère une page de vente (HTML auto-porteur).
func (s *Service) GenerateSalesPage(ctx context.Context, req SalesPageRequest) ([]byte, error) {
	return s.generateHTML(ctx, BuildSalesPagePrompt(req))
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

// ensureChapterPageBreaks garantit que les chapitres (section.chapter)
// démarrent sur une nouvelle page. Si le LLM n'a pas inclus la règle CSS,
// on l'injecte avant le rendu.
func ensureChapterPageBreaks(html []byte) []byte {
	s := string(html)
	if strings.Contains(s, "break-before: page") || strings.Contains(s, "page-break-before") {
		return html
	}
	const css = `<style>section.chapter{break-before:page;page-break-before:always;}</style>`
	if i := strings.Index(s, "</head>"); i >= 0 {
		return []byte(s[:i] + css + s[i:])
	}
	if i := strings.Index(s, "<body"); i >= 0 {
		return []byte(s[:i] + css + s[i:])
	}
	return []byte(css + s)
}
