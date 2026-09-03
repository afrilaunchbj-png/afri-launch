package jobs

import (
	"encoding/json"
	"errors"
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

// coverPrompt construit le prompt de génération de couverture. La palette
// est celle du projet (proposée par l'IA ou fixée par l'utilisateur) ;
// fallback = identité de l'application.
func coverPrompt(p *domain.ProjectPalette, style, topic, audience, instructions string) string {
	colors := "deep emerald (#003527) and warm amber (#FEA619)"
	if p != nil && !p.Empty() {
		colors = fmt.Sprintf("primary %s, secondary %s, accent %s, background %s, text %s",
			p.Primary, p.Secondary, p.Accent, p.Background, p.Text)
	}
	base := fmt.Sprintf(
		"Book cover design, %s palette, %s, clean modern African professional aesthetic, title \"%s\", target audience %s, no text besides the title, high quality, print-ready.",
		colors, style, topic, audience,
	)
	if instructions != "" {
		base += "\nAdditional directions from the user: " + instructions
	}
	return base
}

// paletteSystem est la consigne de proposition d'identité visuelle par l'IA.
const paletteSystem = `You are a brand designer. Given a digital product project and its market context,
propose a visual identity. Return ONLY a JSON object (no markdown fences) with this exact shape:
{"primary":"#RRGGBB","secondary":"#RRGGBB","accent":"#RRGGBB","background":"#RRGGBB","text":"#RRGGBB","style":"3-6 style keywords"}
Rules:
- All colors are hex codes starting with # (e.g. "#0F766E").
- "background" must be light/neutral and "text" dark (or a consistent dark-brand variant) — printed documents must stay highly readable.
- High contrast between text and background; harmonious, culturally appropriate for the target market.
- "style": 3-6 short keywords describing the visual mood, in the project language.`

// palettePrompt construit la demande d'identité visuelle.
func palettePrompt(c genContext, instructions string) string {
	b := fmt.Sprintf(
		"Product: %s\nFormat: %s\nAudience: %s\nCountry: %s\nLanguage: %s",
		c.topic, c.format, c.audience, c.country, c.language,
	)
	if c.promise != "" {
		b += "\nPromise: " + c.promise
	}
	if instructions != "" {
		b += "\nUser directions: " + instructions
	}
	b += "\nPropose the visual identity (palette + style keywords in the language above)."
	return b
}

// posterPrompt construit le prompt d'une affiche publicitaire (3 variantes).
// Tous les textes sont dans la langue du marché (c.language).
func posterPrompt(c genContext, variant int) string {
	colors := "deep emerald (#003527) and warm amber (#FEA619)"
	if c.palette != nil && !c.palette.Empty() {
		colors = fmt.Sprintf("primary %s, secondary %s, accent %s, background %s, text %s",
			c.palette.Primary, c.palette.Secondary, c.palette.Accent, c.palette.Background, c.palette.Text)
	}
	base := fmt.Sprintf(
		"Advertising poster, %s palette, %s, clean modern African professional aesthetic, all text in %s, print-ready, no invented statistics, generous whitespace, high contrast.",
		colors, c.config.Style, c.language,
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

// ResearchSystem est la consigne système de la recherche en ligne
// (réutilisée par le copilote conversationnel).
const ResearchSystem = `You are a market researcher specialized in African digital-product opportunities.
Use the web_search tool to gather real, current information before answering.
Return ONLY a JSON object (no markdown fences) with this exact shape:
{"opportunities":[{"country","title","summary","difficulty","signal","score","scores":{"demand","pain","competition","purchasing_power","digital_fit","evidence_strength"},"evidence":[{"source","title","url","publication_date","country","metric","value"}]}]}
Rules:
- One opportunity per target market (country).
- NEVER invent statistics: only report figures you actually found, and put each one in "evidence" with its source URL. If you cannot verify a figure, omit it.
- "signal" is "verified" only when backed by solid evidence; otherwise "estimated" or "hypothesis".
- "score" and each sub-score are your qualitative assessment (0-100).
- Write "title" and "summary" in the requested language.`

// ResearchQuery construit la requête de recherche en ligne.
func ResearchQuery(query, sector string, markets []string, language string) string {
	return fmt.Sprintf(
		"Research this niche: %s.\nSector: %s.\nTarget markets: %s.\nLanguage for titles/summaries: %s.",
		query, sector, strings.Join(markets, ", "), language,
	)
}

// researchInput est la structure JSON attendue du LLM pour la recherche.
type researchInput struct {
	Opportunities []ResearchOpportunityInput `json:"opportunities"`
}

// ResearchOpportunityInput est une opportunité renvoyée par la recherche.
type ResearchOpportunityInput struct {
	Country    string                   `json:"country"`
	Title      string                   `json:"title"`
	Summary    string                   `json:"summary"`
	Difficulty string                   `json:"difficulty"`
	Signal     string                   `json:"signal"`
	Score      int                      `json:"score"`
	Scores     domain.OpportunityScores `json:"scores"`
	Evidence   []domain.Evidence        `json:"evidence"`
}

// ParseResearchResult décode et normalise le JSON renvoyé par le LLM.
func ParseResearchResult(content string) ([]ResearchOpportunityInput, error) {
	var out researchInput
	if err := json.Unmarshal([]byte(stripFences(content)), &out); err != nil {
		return nil, fmt.Errorf("decode research: %w", err)
	}
	if len(out.Opportunities) == 0 {
		return nil, errors.New("aucune opportunité trouvée")
	}
	return out.Opportunities, nil
}

// OpportunityFromResearch convertit un résultat de recherche en entité domain.
func OpportunityFromResearch(in ResearchOpportunityInput, sector, language, userID string, researchID *string) domain.Opportunity {
	return domain.Opportunity{
		UserID:     &userID,
		ResearchID: researchID,
		Title:      in.Title,
		Summary:    in.Summary,
		Country:    in.Country,
		Sector:     sector,
		Language:   language,
		Difficulty: normalizeDifficulty(in.Difficulty),
		Signal:     normalizeSignal(in.Signal),
		Score:      clamp(in.Score, 0, 100),
		Scores:     in.Scores,
		Evidence:   in.Evidence,
	}
}
