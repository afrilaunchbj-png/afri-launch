package jobs

import (
	"fmt"

	"afrilaunch/backend/internal/domain"
)

// ideaInput est la structure JSON attendue du LLM pour les idées.
type ideaInput struct {
	Title            string `json:"title"`
	Subtitle         string `json:"subtitle"`
	Audience         string `json:"audience"`
	Problem          string `json:"problem"`
	Promise          string `json:"promise"`
	Format           string `json:"format"`
	EstimatedPrice   string `json:"estimated_price"`
	Difficulty       string `json:"difficulty"`
	MarketEvidence   string `json:"market_evidence"`
	WhyNow           string `json:"why_now"`
	CompetitiveAngle string `json:"competitive_angle"`
}

const ideaSystem = `You are a product strategist specialized in African digital products.
Generate product concepts for a market opportunity. Return ONLY a JSON object
(no markdown fences) with this exact shape:
{"ideas":[{"title","subtitle","audience","problem","promise","format","estimated_price","difficulty","market_evidence","why_now","competitive_angle"}]}
Generate 5 ideas. Do not invent statistics: qualitative evidence only. Titles must be attractive but honest.
Write every field in the language requested by the user (never in English unless the user asks for English).`

func ideaPrompt(opp domain.Opportunity) string {
	return fmt.Sprintf(
		"Market opportunity (country: %s, sector: %s): %s\nSummary: %s\nDifficulty: %s.\nLanguage: %s — write every idea field entirely in this language.",
		opp.Country, opp.Sector, opp.Title, opp.Summary, opp.Difficulty, opp.Language,
	)
}

// coverPrompt construit le prompt de génération de couverture.
func coverPrompt(topic, audience string) string {
	return fmt.Sprintf(
		"Book cover design, deep emerald (#003527) and warm amber (#FEA619) palette, clean modern African professional aesthetic, title \"%s\", target audience %s, no text besides the title, high quality, print-ready.",
		topic, audience,
	)
}

// posterPrompt construit le prompt d'une affiche publicitaire (3 variantes).
// Tous les textes sont dans la langue du marché (c.language).
func posterPrompt(c genContext, variant int) string {
	base := fmt.Sprintf(
		"Advertising poster, deep emerald (#003527) and warm amber (#FEA619) palette, clean modern African professional aesthetic, all text in %s, print-ready, no invented statistics, generous whitespace, high contrast.",
		c.language,
	)
	switch variant {
	case 1:
		return fmt.Sprintf(
			"%s\nHero poster: big bold headline \"%s\" with a short supporting line \"%s\".",
			base, c.topic, c.promise,
		)
	case 2:
		return fmt.Sprintf(
			"%s\nAudience poster: headline speaks directly to \"%s\", body line \"%s\", with an abstract visual metaphor.",
			base, c.audience, c.promise,
		)
	default:
		return fmt.Sprintf(
			"%s\nOffer poster: headline \"%s\", one short call-to-action line, keep text minimal (10 words max), focal button.",
			base, c.promise,
		)
	}
}
