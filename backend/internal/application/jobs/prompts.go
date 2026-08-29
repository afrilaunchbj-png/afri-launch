package jobs

import (
	"fmt"
	"strings"

	"afrilaunch/backend/internal/domain"
)

// ideaInput est la structure JSON attendue du LLM pour les idées.
type ideaInput struct {
	Title            string `json:"title"`
	Hook             string `json:"hook"`
	Explanation      string `json:"explanation"`
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
{"ideas":[{"title","hook","explanation","subtitle","audience","problem","promise","format","estimated_price","difficulty","market_evidence","why_now","competitive_angle"}]}
Generate 5 ideas. For each idea:
- "title": a short, honest product title.
- "hook": a punchy one-line pitch with a clear, honest number (only numbers supported by the opportunity's evidence — never invent a statistic).
- "explanation": 2-3 sentences explaining the product and why it fits the market.
Do not invent statistics: qualitative evidence only. Titles must be attractive but honest.
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

// researchSystem est la consigne système de la recherche en ligne.
const researchSystem = `You are a market researcher specialized in African digital-product opportunities.
Use the web_search tool to gather real, current information before answering.
Return ONLY a JSON object (no markdown fences) with this exact shape:
{"opportunities":[{"country","title","summary","difficulty","signal","score","scores":{"demand","pain","competition","purchasing_power","digital_fit","evidence_strength"},"evidence":[{"source","title","url","publication_date","country","metric","value"}]}]}
Rules:
- One opportunity per target market (country).
- NEVER invent statistics: only report figures you actually found, and put each one in "evidence" with its source URL. If you cannot verify a figure, omit it.
- "signal" is "verified" only when backed by solid evidence; otherwise "estimated" or "hypothesis".
- "score" and each sub-score are your qualitative assessment (0-100).
- Write "title" and "summary" in the requested language.`

// researchQuery construit la requête de recherche en ligne.
func researchQuery(query, sector string, markets []string, language string) string {
	return fmt.Sprintf(
		"Research this niche: %s.\nSector: %s.\nTarget markets: %s.\nLanguage for titles/summaries: %s.",
		query, sector, strings.Join(markets, ", "), language,
	)
}

// researchInput est la structure JSON attendue du LLM pour la recherche.
type researchInput struct {
	Opportunities []researchOpportunityInput `json:"opportunities"`
}

type researchOpportunityInput struct {
	Country    string                   `json:"country"`
	Title      string                   `json:"title"`
	Summary    string                   `json:"summary"`
	Difficulty string                   `json:"difficulty"`
	Signal     string                   `json:"signal"`
	Score      int                      `json:"score"`
	Scores     domain.OpportunityScores `json:"scores"`
	Evidence   []domain.Evidence        `json:"evidence"`
}

// ideaReviseSystem est la consigne système de l'itération sur une idée.
const ideaReviseSystem = `You are a product-idea coach. Given a product idea and the user's feedback,
revise the title, hook and explanation accordingly. Return ONLY a JSON object
(no markdown fences) with this exact shape: {"title","hook","explanation"}.
- "hook" is a punchy one-line pitch with a clear, honest number (only numbers already present in the idea/evidence — never invent).
- Keep the revised text in the same language as the current idea.`

// ideaRevisePrompt construit la consigne d'itération avec l'historique.
func ideaRevisePrompt(idea domain.ProductIdea, history []domain.IdeaMessage, feedback string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Product idea to refine:\nTitle: %s\nHook: %s\nExplanation: %s\n\n", idea.Title, idea.Hook, idea.Explanation)
	if len(history) > 0 {
		b.WriteString("Conversation history:\n")
		for _, m := range history {
			fmt.Fprintf(&b, "%s: %s\n", m.Role, m.Content)
		}
		b.WriteString("\n")
	}
	b.WriteString("Latest user feedback: " + feedback + "\n\n")
	b.WriteString("Revise the title, hook and explanation to address this feedback.")
	return b.String()
}
