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

// Prompt est la consigne envoyée au LLM.
type Prompt struct {
	System string
	User   string
}

const baseSystem = `You are an expert digital-product designer and writer. ` +
	`Generate a single, self-contained HTML document (no markdown fences, no ` +
	`explanations, no code blocks). Inline all CSS in a <style> tag. Do not load ` +
	`any external resource (no external fonts, images, or scripts). ` +
	`Output ONLY the HTML.` + impeccableGuidelines

// BuildEbookPrompt construit la consigne de génération d'un ebook (HTML → PDF).
func BuildEbookPrompt(req EbookRequest) Prompt {
	system := baseSystem + `
## Output format: ebook
- One continuous HTML page styled for print (PDF, A4).
- Use @page { size: A4; margin: 2cm; } and print-friendly typography.
- Structure: cover (title + subtitle), introduction, chapters (with headings),
  short sections, a practical checklist, and a conclusion with a call to action.`

	user := fmt.Sprintf(
		"Write an ebook in %s (target market: %s, audience: %s) about: %s. Format: %s. Aim for 4000-6000 words equivalent, clear and actionable.",
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
