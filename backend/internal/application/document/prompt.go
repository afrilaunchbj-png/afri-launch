package document

import "fmt"

// EbookRequest décrit la génération d'un ebook/guide.
type EbookRequest struct {
	Topic    string
	Audience string
	Language string
	Country  string
	Product  string
}

// DeckRequest décrit la génération d'un deck (pitch/présentation).
type DeckRequest struct {
	Topic    string
	Audience string
	Language string
	Country  string
}

// SalesPageRequest décrit la génération d'une page de vente.
type SalesPageRequest struct {
	Product  string
	Promise  string
	Audience string
	Language string
	Country  string
	Price    string
}

// Prompt est la consigne envoyée au LLM.
type Prompt struct {
	System string
	User   string
}

const baseSystem = `You are an expert digital-product designer and writer. ` +
	`Generate a single, self-contained HTML document (no markdown fences, no ` +
	`explanations, no code blocks). Inline all CSS in a <style> tag. Do not load ` +
	`any external resource (no external fonts, images, or scripts). ` +
	`Output ONLY the HTML. ` +
	`Write all content in the language requested by the user (never in English unless requested).` + impeccableGuidelines

// BuildEbookPrompt construit la consigne de génération d'un ebook (HTML → PDF).
func BuildEbookPrompt(req EbookRequest) Prompt {
	system := baseSystem + `
## Output format: ebook
- One continuous HTML page styled for print (PDF, A4).
- Use @page { size: A4; margin: 2cm; } and print-friendly typography.
- Structure: cover (title + subtitle), introduction, chapters (with headings),
  short sections, a practical checklist, and a conclusion with a call to action.
- Wrap each chapter in a <section class="chapter"> element.
- In the CSS, add: section.chapter { break-before: page; } so that each chapter
  starts on a new page. Do NOT apply a page break to the cover or the introduction.`

	user := fmt.Sprintf(
		"Write an ebook in %s (target market: %s, audience: %s) about: %s. Format: %s. Aim for about 3000 words, clear and actionable.",
		req.Language, req.Country, req.Audience, req.Topic, req.Product,
	)
	return Prompt{System: system, User: user}
}

// BuildDeckPrompt construit la consigne de génération d'un deck (HTML slides → PPTX).
func BuildDeckPrompt(req DeckRequest) Prompt {
	system := baseSystem + `
## Output format: slide deck (16:9)
- Each slide is a <section class="slide"> element, fixed at 1280x720 CSS px
  (width:1280px; height:720px; overflow:hidden; position:relative).
- Slides are direct children of <body>, no other top-level elements.
- 6 to 10 slides: title slide, problem, solution, market/opportunity,
  how it works, business model, call to action.
- One idea per slide, short bullets, strong titles, generous whitespace.`

	user := fmt.Sprintf(
		"Build a pitch deck in %s (target market: %s, audience: %s) about: %s.",
		req.Language, req.Country, req.Audience, req.Topic,
	)
	return Prompt{System: system, User: user}
}

// BuildEbookDeckPrompt construit la consigne de génération de la version
// paysage de l'ebook (HTML slides → PPTX), destinée à être exportée en PPT.
func BuildEbookDeckPrompt(req EbookRequest) Prompt {
	system := baseSystem + `
## Output format: landscape ebook deck (16:9, for PPT export)
- Each slide is a <section class="slide"> element, fixed at 1280x720 CSS px
  (width:1280px; height:720px; overflow:hidden; position:relative).
- Slides are direct children of <body>, no other top-level elements.
- Turn the ebook into 8 to 14 slides: cover, introduction, the key chapters
  (one idea per slide), practical steps, a checklist, and a conclusion with
  a clear call to action.
- One idea per slide, short bullets, strong titles, generous whitespace.`

	user := fmt.Sprintf(
		"Build a landscape slide-deck version of an ebook in %s (target market: %s, audience: %s) about: %s. Format: %s.",
		req.Language, req.Country, req.Audience, req.Topic, req.Product,
	)
	return Prompt{System: system, User: user}
}

// BuildSalesPagePrompt construit la consigne de génération d'une page de vente (HTML).
func BuildSalesPagePrompt(req SalesPageRequest) Prompt {
	system := baseSystem + `
## Output format: sales landing page (Persuade mode)
- One self-contained HTML page (no markdown fences, inline CSS only).
- Hero with a sharp, benefit-driven headline (no vague claim, no invented statistic).
- Sections: problem, promise/solution, what's included, social proof placeholder, price, one clear CTA.
- Mobile-first, high contrast, generous whitespace.`

	user := fmt.Sprintf(
		"Write a sales page in %s (target market: %s, audience: %s) for: %s. Promise: %s. Price: %s.",
		req.Language, req.Country, req.Audience, req.Product, req.Promise, req.Price,
	)
	return Prompt{System: system, User: user}
}
