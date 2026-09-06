package document

import (
	"context"
	"encoding/base64"
	"html"
	"regexp"
	"strconv"
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
	html = prepareEbookHTML(html, req.Language)
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

// ---------- Préparation du rendu ebook (PDF) ----------

// ebookPrintCSS garantit un fond blanc sur toutes les pages du PDF et style
// la page de sommaire. Injecté APRÈS le <style> du LLM (d'où !important).
const ebookPrintCSS = `<style>
html, body { background: #ffffff !important; }
.toc-page { break-after: page; page-break-after: always; }
.toc-title { margin: 0 0 0.6em; }
.toc-list { list-style: none; margin: 0; padding: 0; }
.toc-list li { display: flex; align-items: baseline; gap: 0.6em; margin: 0.45em 0; }
.toc-num { min-width: 1.4em; font-weight: 600; color: #855300; }
.toc-entry { font-weight: 500; }
</style>`

// prepareEbookHTML : ① force le fond blanc d'impression ; ② insère une page
// « Sommaire » (après le <body>, donc après la cover injectée ensuite par
// PrependCoverPage) listant les chapitres détectés dans le HTML généré.
func prepareEbookHTML(html []byte, language string) []byte {
	s := string(html)

	if toc, ok := buildTOC(s, language); ok {
		s = insertAfterBodyOpen(s, toc)
	}
	return []byte(insertBeforeHeadEnd(s, ebookPrintCSS))
}

// tocLabel donne le titre de la page de sommaire dans la langue de l'ebook.
func tocLabel(language string) string {
	switch strings.ToLower(language) {
	case "en", "en-us", "en-gb":
		return "Table of contents"
	case "fr", "fr-fr":
		return "Sommaire"
	default:
		return "Sommaire"
	}
}

var (
	chapterOpenRe = regexp.MustCompile(`(?is)<section\b[^>]*class\s*=\s*["'][^"']*\bchapter\b[^"']*["'][^>]*>`)
	headingOpenRe = regexp.MustCompile(`(?is)<(h[1-4])(?:\s[^>]*)?>`)
	tagRe         = regexp.MustCompile(`(?s)<[^>]*>`)
)

// chapterHeadings extrait le premier titre (h1→h4) de chaque chapitre.
func chapterHeadings(s string) []string {
	locs := chapterOpenRe.FindAllStringIndex(s, -1)
	headings := make([]string, 0, len(locs))
	for i, loc := range locs {
		from := loc[1]
		until := len(s)
		if i+1 < len(locs) {
			until = locs[i+1][0]
		}
		if h := nextHeadingText(s, from, until); h != "" {
			headings = append(headings, h)
		}
	}
	return headings
}

// nextHeadingText renvoie le texte du premier heading (h1→h4) dans [from,until).
func nextHeadingText(s string, from, until int) string {
	if until > len(s) {
		until = len(s)
	}
	region := s[from:until]
	m := headingOpenRe.FindStringSubmatchIndex(region)
	if m == nil {
		return ""
	}
	level := region[m[2]:m[3]]
	contentStart := from + m[1]
	closeTag := "</" + level
	ci := strings.Index(strings.ToLower(s[contentStart:until]), strings.ToLower(closeTag))
	end := until
	if ci >= 0 {
		end = contentStart + ci
	}
	inner := s[contentStart:end]
	inner = tagRe.ReplaceAllString(inner, "")
	return strings.Join(strings.Fields(html.UnescapeString(inner)), " ")
}

// buildTOC génère le HTML de la page de sommaire à partir des chapitres.
func buildTOC(s, language string) (string, bool) {
	headings := chapterHeadings(s)
	if len(headings) == 0 {
		return "", false
	}
	var b strings.Builder
	b.WriteString(`<section class="toc-page">`)
	b.WriteString(`<h1 class="toc-title">` + html.EscapeString(tocLabel(language)) + `</h1>`)
	b.WriteString(`<ul class="toc-list">`)
	for i, h := range headings {
		b.WriteString(`<li><span class="toc-num">` + strconv.Itoa(i+1) + `.</span><span class="toc-entry">` + html.EscapeString(h) + `</span></li>`)
	}
	b.WriteString(`</ul></section>`)
	return b.String(), true
}

// insertBeforeHeadEnd injecte `css` juste avant </head> (ou avant <body>).
func insertBeforeHeadEnd(s, css string) string {
	if i := strings.Index(s, "</head>"); i >= 0 {
		return s[:i] + css + s[i:]
	}
	return insertAfterBodyOpen(s, css)
}

// insertAfterBodyOpen injecte `content` juste après la balise <body …> ;
// en l'absence de <body>, après </head>, sinon au tout début du document.
func insertAfterBodyOpen(s, content string) string {
	if i := strings.Index(s, "<body"); i >= 0 {
		if j := strings.Index(s[i:], ">"); j >= 0 {
			pos := i + j + 1
			return s[:pos] + content + s[pos:]
		}
	}
	if j := strings.Index(s, "</head>"); j >= 0 {
		pos := j + len("</head>")
		return s[:pos] + content + s[pos:]
	}
	return content + s
}
