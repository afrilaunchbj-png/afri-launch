package videoad

import (
	"context"
	"fmt"
	"strings"

	"afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service génère le plan créatif d'une vidéo publicitaire (analyse
// marketing → script → storyboard) via le LLM. La vidéo elle-même est
// produite par le worker via les ports VideoProvider/VideoRenderer.
type Service struct {
	ai *ai.Service
}

// NewService construit le service créatif vidéo.
func NewService(ai *ai.Service) *Service { return &Service{ai: ai} }

// Plan produit le storyboard validé de la publicité (deux appels LLM en
// mode JSON : analyse marketing puis storyboard).
func (s *Service) Plan(ctx context.Context, c AdContext, params domain.VideoAdParams) (domain.Storyboard, error) {
	params = params.Normalized()
	if c.Topic == "" {
		return domain.Storyboard{}, domain.ErrInvalidInput
	}
	analysis, err := s.Analyze(ctx, c, params)
	if err != nil {
		return domain.Storyboard{}, err
	}
	return s.Story(ctx, c, params, analysis)
}

// Analyze exécute l'étape 1 : analyse marketing (angle, pain points, hook, CTA).
func (s *Service) Analyze(ctx context.Context, c AdContext, params domain.VideoAdParams) (Analysis, error) {
	params = params.Normalized()
	if c.Topic == "" {
		return Analysis{}, domain.ErrInvalidInput
	}
	var analysis Analysis
	if err := s.ai.CompleteJSON(ctx, ai.TaskContent, analysisSystem, AnalysisPrompt(c, params), &analysis); err != nil {
		return Analysis{}, fmt.Errorf("analyse marketing: %w", err)
	}
	if analysis.Hook == "" || analysis.CTA == "" {
		return Analysis{}, fmt.Errorf("analyse marketing incomplète (hook/cta vides)")
	}
	return analysis, nil
}

// Story exécute l'étape 2 : storyboard des scènes à partir de l'analyse.
func (s *Service) Story(ctx context.Context, c AdContext, params domain.VideoAdParams, analysis Analysis) (domain.Storyboard, error) {
	params = params.Normalized()
	var sb domain.Storyboard
	if err := s.ai.CompleteJSON(ctx, ai.TaskContent, storyboardSystem, StoryboardPrompt(c, params, analysis), &sb); err != nil {
		return domain.Storyboard{}, fmt.Errorf("storyboard: %w", err)
	}

	sb.Angle = firstNonEmpty(sb.Angle, analysis.Angle)
	sb.Hook = firstNonEmpty(sb.Hook, analysis.Hook)
	sb.CTA = firstNonEmpty(sb.CTA, analysis.CTA)
	sb.Duration = params.Duration
	sb.AspectRatio = params.AspectRatio
	normalizeScenes(&sb, params.Duration)

	if err := sb.Validate(); err != nil {
		return domain.Storyboard{}, err
	}
	return sb, nil
}

// ProviderDefaults porte la configuration HeyGen par défaut (env).
type ProviderDefaults struct {
	AvatarID string
	VoiceID  string
}

// ResolveVideoRequest construit la requête provider (HeyGen) depuis le
// storyboard : override params > défauts configurés. Aucun identifiant
// provider n'est hard-codé.
func ResolveVideoRequest(sb domain.Storyboard, params domain.VideoAdParams, title string, d ProviderDefaults) port.VideoRequest {
	return port.VideoRequest{
		AvatarID:    firstNonEmpty(params.AvatarID, d.AvatarID),
		VoiceID:     firstNonEmpty(params.VoiceID, d.VoiceID),
		Script:      sb.AvatarScript(),
		AspectRatio: sb.AspectRatio,
		Resolution:  "1080p",
		Title:       title,
	}
}

// normalizeScenes réattribue des IDs stables et ramène la somme des durées
// sur la durée demandée (proportionnel, en secondes entières).
func normalizeScenes(sb *domain.Storyboard, duration int) {
	total := 0
	for _, sc := range sb.Scenes {
		d := int(sc.Duration)
		if d < 1 {
			d = 1
		}
		total += d
	}
	if total < 1 {
		total = 1
	}
	sum := 0
	for i := range sb.Scenes {
		sb.Scenes[i].ID = fmt.Sprintf("scene_%d", i+1)
		d := int(float64(int(sb.Scenes[i].Duration)) * float64(duration) / float64(total))
		if d < 1 {
			d = 1
		}
		sb.Scenes[i].Duration = float64(d)
		sum += d
	}
	// Ajuste l'écart résiduel sur les scènes avatar (durée parlée réelle).
	for i := len(sb.Scenes) - 1; i >= 0 && sum != duration; i-- {
		step := 1
		if sum > duration {
			step = -1
		}
		if next := sb.Scenes[i].Duration + float64(step); next >= 1 {
			sb.Scenes[i].Duration = next
			sum += step
		}
	}
}

// BuildSubtitleCues découpe le script des scènes avatar en cues de
// sous-titres, répartis proportionnellement au nombre de mots sur la durée
// parlée réelle (mesurée sur la vidéo provider ; fallback = durée cible).
func BuildSubtitleCues(sb domain.Storyboard, spoken float64) []port.SubtitleCue {
	if spoken <= 0 {
		spoken = float64(sb.Duration)
	}

	scenes := make([]domain.Scene, 0, len(sb.Scenes))
	totalWords := 0
	for _, sc := range sb.Scenes {
		if sc.Type == domain.SceneAvatar && sc.Script != "" {
			scenes = append(scenes, sc)
			totalWords += wordCount(sc.Script)
		}
	}
	if totalWords == 0 {
		return nil
	}

	cues := make([]port.SubtitleCue, 0, len(scenes)*3)
	t := 0.0
	for _, sc := range scenes {
		sceneDur := spoken * float64(wordCount(sc.Script)) / float64(totalWords)
		cues = append(cues, sceneCues(sc.Script, t, sceneDur)...)
		t += sceneDur
	}
	return cues
}

// sceneCues découpe un script de scène en cues séquentiels (max 7 mots),
// durées proportionnelles au nombre de mots.
func sceneCues(script string, start, sceneDur float64) []port.SubtitleCue {
	words := strings.Fields(script)
	if len(words) == 0 {
		return nil
	}
	const maxWords = 7
	var chunks []string
	for start := 0; start < len(words); start += maxWords {
		end := start + maxWords
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[start:end], " "))
	}
	weights := make([]float64, len(chunks))
	totalW := 0
	for i, ch := range chunks {
		weights[i] = float64(wordCount(ch))
		totalW += int(weights[i])
	}
	if totalW == 0 {
		totalW = len(chunks)
		for i := range weights {
			weights[i] = 1
		}
	}
	cues := make([]port.SubtitleCue, 0, len(chunks))
	t := start
	for i, ch := range chunks {
		d := sceneDur * weights[i] / float64(totalW)
		cues = append(cues, port.SubtitleCue{Start: t, End: t + d, Text: ch})
		t += d
	}
	return cues
}

func wordCount(s string) int {
	return len(strings.Fields(s))
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
