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
Generate 5 ideas. Do not invent statistics: qualitative evidence only. Titles must be attractive but honest.`

func ideaPrompt(opp domain.Opportunity) string {
	return fmt.Sprintf(
		"Market opportunity (country: %s, sector: %s, language: %s): %s\nSummary: %s\nDifficulty: %s.",
		opp.Country, opp.Sector, opp.Language, opp.Title, opp.Summary, opp.Difficulty,
	)
}

// coverPrompt construit le prompt de génération de couverture.
func coverPrompt(topic, audience string) string {
	return fmt.Sprintf(
		"Book cover design, deep emerald (#003527) and warm amber (#FEA619) palette, clean modern African professional aesthetic, title \"%s\", target audience %s, no text besides the title, high quality, print-ready.",
		topic, audience,
	)
}
