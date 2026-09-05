// Package videoad orchestre la génération de vidéos publicitaires
// short-form (TikTok/Reels) : analyse marketing + script + storyboard par
// LLM, puis vidéo avatar (HeyGen) et montage (VideoRenderer). Le code métier
// ne dépend jamais d'un provider concret (cf. ADR-016).
package videoad

import (
	"fmt"
	"strings"

	"afrilaunch/backend/internal/domain"
)

// AdContext regroupe le contexte produit pour les prompts vidéo.
type AdContext struct {
	Topic            string // nom du produit
	Audience         string
	Language         string
	Country          string
	Problem          string
	Promise          string
	Price            string
	Format           string // ebook, guide…
	CompetitiveAngle string
	HasCover         bool // mockup du produit disponible
}

// Analysis est la sortie structurée de l'analyse marketing (étape 1).
type Analysis struct {
	Angle            string   `json:"angle"`
	PainPoints       []string `json:"pain_points"`
	ValueProposition string   `json:"value_proposition"`
	Hook             string   `json:"hook"`
	CTA              string   `json:"cta"`
	MarketingAngles  []string `json:"marketing_angles"`
}

const analysisSystem = `You are a senior direct-response marketing analyst specializing in short-form video ads (TikTok, Instagram Reels) for digital products sold in African markets. ` +
	`Reply with STRICT JSON only (no markdown fences, no commentary). ` +
	`Write all human-readable content in the language requested by the user prompt.`

// AnalysisPrompt construit la consigne d'analyse marketing (étape 1).
func AnalysisPrompt(c AdContext, params domain.VideoAdParams) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		"Analyze this digital product for a %d-second video ad.\n"+
			"Product: %s\nFormat: %s\nAudience: %s\nMain problem it solves: %s\n"+
			"Promise: %s\nPrice: %s\nMarket: %s\n",
		params.Duration, c.Topic, c.Format, c.Audience, c.Problem,
		c.Promise, c.Price, c.Country,
	))
	if c.CompetitiveAngle != "" {
		b.WriteString("Competitive angle: " + c.CompetitiveAngle + "\n")
	}
	if params.Angle != "" {
		b.WriteString("Preferred marketing angle (follow it): " + params.Angle + "\n")
	}
	if params.CTA != "" {
		b.WriteString("Mandatory call to action (use it verbatim): " + params.CTA + "\n")
	}
	if params.Instructions != "" {
		b.WriteString("Additional creator instructions: " + params.Instructions + "\n")
	}
	b.WriteString(`
Return JSON with exactly these fields:
- "angle": the single marketing angle this ad will follow (one of: pain_point, transformation, curiosity, social_proof, urgency, educational, problem_solution)
- "pain_points": 3 short pain points of the audience (each under 12 words)
- "value_proposition": one sentence, concrete and outcome-driven
- "hook": the opening line spoken in the first 3 seconds — bold, native spoken style, no emojis, under 15 words
- "cta": the closing call to action — imperative, under 10 words
- "marketing_angles": 4 alternative angle names from the list above (for future variants)
All string content in ` + c.Language + `.`)
	return b.String()
}

const storyboardSystem = `You are a short-form video scriptwriter and storyboard artist for TikTok/Reels ads. ` +
	`Reply with STRICT JSON only (no markdown fences, no commentary). ` +
	`The ad features ONE talking-head presenter (avatar) speaking the whole script; the visual scenes describe how text overlays and product shots punctuate the video. ` +
	`Write all human-readable content in the language requested by the user prompt.`

// StoryboardPrompt construit la consigne de storyboard (étape 2).
func StoryboardPrompt(c AdContext, params domain.VideoAdParams, a Analysis) string {
	angles := strings.Join(a.MarketingAngles, ", ")
	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		"Write the storyboard of a %d-second video ad in %s (market: %s).\n"+
			"Product: %s (%s)\nAudience: %s\nValue proposition: %s\n"+
			"Chosen angle: %s\nMain pain points: %s\nHook (opening line): %s\nCTA (closing line): %s\n",
		params.Duration, c.Language, c.Country,
		c.Topic, c.Format, c.Audience, a.ValueProposition,
		a.Angle, strings.Join(a.PainPoints, " ; "), a.Hook, a.CTA,
	))
	if angles != "" {
		b.WriteString("Alternative angles for reference: " + angles + "\n")
	}
	if c.HasCover {
		b.WriteString("A product mockup image exists: you may use a \"product\" scene (max 3 s) right after the hook.\n")
	}
	if params.Instructions != "" {
		b.WriteString("Creator instructions: " + params.Instructions + "\n")
	}
	b.WriteString(fmt.Sprintf(`
Return JSON: {"angle":"%s","hook":string,"cta":string,"duration":%d,"aspect_ratio":"%s","scenes":[...]}
Rules:
- 4 to 6 scenes following the short-form structure: HOOK (0-3s) → PROBLEM → SOLUTION → BENEFITS → CTA.
- Scene "type" is one of: "avatar" (presenter speaking — the default), "product" (product mockup shot), "text" (bold text card).
- Each "avatar" scene has a "script" field: the exact words spoken, natural spoken style, no emojis, no stage directions.
- Each scene has "text_overlay" (max 6 words, punchy, optional) and "visual_description" (one short sentence).
- Each scene has "duration" in seconds (integers); the SUM of scene durations must equal exactly %d.
- The first scene must open with the hook; the last scene is the CTA.`, a.Angle, params.Duration, params.AspectRatio, params.Duration))
	return b.String()
}
