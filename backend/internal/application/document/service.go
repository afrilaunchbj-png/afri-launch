package document

import (
	"context"
	"encoding/base64"
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

// GenerateEbookDeckWithCover génère la version paysage avec la cover générée
// en première slide (workflow cover-first).
func (s *Service) GenerateEbookDeckWithCover(ctx context.Context, req EbookRequest, coverPNG []byte) ([]byte, error) {
	html, err := s.generateHTML(ctx, BuildEbookDeckPrompt(req))
	if err != nil {
		return nil, err
	}
	return s.render.SlidesToPPTXWithCover(ctx, html, coverPNG)
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

// PrependCoverPage injecte la cover générée (PNG) en première page pleine
// du PDF (workflow cover-first) : marge 0 sur la 1re page, saut de page après.
func PrependCoverPage(html []byte, coverPNG []byte) []byte {
	if len(coverPNG) == 0 {
		return html
	}
	s := string(html)

	const css = `<style>@page :first { margin: 0; }` +
		`section.cover-page { break-after: page; page-break-after: always; margin: 0; padding: 0; }` +
		`section.cover-page img { display: block; width: 210mm; height: 296mm; object-fit: cover; }</style>`
	if i := strings.Index(s, "</head>"); i >= 0 {
		s = s[:i] + css + s[i:]
	} else if i := strings.Index(s, "<body"); i >= 0 {
		if j := strings.Index(s[i:], ">"); j >= 0 {
			pos := i + j + 1
			s = s[:pos] + css + s[pos:]
		} else {
			s = css + s
		}
	} else {
		s = css + s
	}

	cover := `<section class="cover-page"><img src="data:image/png;base64,` +
		base64.StdEncoding.EncodeToString(coverPNG) + `" alt="Cover"/></section>`
	if i := strings.Index(s, "<body"); i >= 0 {
		if j := strings.Index(s[i:], ">"); j >= 0 {
			pos := i + j + 1
			return []byte(s[:pos] + cover + s[pos:])
		}
	}
	return []byte(cover + s)
}
