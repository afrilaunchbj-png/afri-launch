package domain

import (
	"encoding/json"
	"strings"
)

// JobVideoAd est le kind de job pour la génération d'une vidéo publicitaire.
const JobVideoAd = "video_ad"

// Kinds d'asset vidéo publicitaire.
const (
	AssetVideoAd      = "video_ad"       // MP4 final (montage complet)
	AssetVideoAdThumb = "video_ad_thumb" // vignette JPEG
)

// Étapes du pipeline vidéo, publiées via SSE (champ "stage" de job.updated)
// pour une progression détaillée côté frontend (cf. prompts/video-flow.md §26).
const (
	VideoStageAnalyzing       = "analyzing"         // analyse marketing + script (LLM)
	VideoStageStoryboarding   = "storyboarding"     // storyboard des scènes (LLM)
	VideoStageGeneratingVideo = "generating_avatar" // génération HeyGen
	VideoStageRendering       = "rendering"         // montage FFmpeg (sous-titres, cartes)
)

// Durées supportées (secondes).
const (
	VideoDurationMin       = 15
	VideoDurationDefault   = 30
	VideoDurationMax       = 60
	VideoDurationIncrement = 15
)

// Ratios d'export supportés.
const (
	VideoRatioPortrait  = "9:16" // TikTok / Reels
	VideoRatioSquare    = "1:1"  // feed Facebook
	VideoRatioLandscape = "16:9" // YouTube
	VideoRatioDefault   = VideoRatioPortrait
)

// VideoAdParams porte les paramètres d'une génération de vidéo pub
// (stockés en JSONB dans generation_jobs.params).
type VideoAdParams struct {
	Angle        string `json:"angle,omitempty"`        // angle marketing souhaité (régénération)
	Duration     int    `json:"duration,omitempty"`     // 15/30/45/60 s
	AspectRatio  string `json:"aspect_ratio,omitempty"` // 9:16 / 1:1 / 16:9
	AvatarID     string `json:"avatar_id,omitempty"`    // override HeyGen
	VoiceID      string `json:"voice_id,omitempty"`     // override HeyGen
	CTA          string `json:"cta,omitempty"`          // CTA explicite (sinon dérivé de la promise)
	Instructions string `json:"instructions,omitempty"` // consignes libres (régénération)
}

// Normalized applique les valeurs par défaut et les bornes.
func (p VideoAdParams) Normalized() VideoAdParams {
	switch p.Duration {
	case VideoDurationMin, 30, 45, VideoDurationMax:
	default:
		p.Duration = VideoDurationDefault
	}
	switch p.AspectRatio {
	case VideoRatioPortrait, VideoRatioSquare, VideoRatioLandscape:
	default:
		p.AspectRatio = VideoRatioDefault
	}
	return p
}

// Marshal sérialise les paramètres pour generation_jobs.params (JSONB).
func (p VideoAdParams) Marshal() []byte {
	raw, _ := json.Marshal(p.Normalized())
	return raw
}

// Types de scène du storyboard (extensible : broll, screen_recording… post-MVP).
const (
	SceneAvatar  = "avatar"  // présentateur qui parle (HeyGen)
	SceneProduct = "product" // mockup / visuel du produit (cover générée)
	SceneText    = "text"    // carte texte (hook, CTA)
)

// Storyboard est le plan de montage de la vidéo publicitaire (JSONB dans
// generation_jobs.result). Les scènes découpent le script short-form.
type Storyboard struct {
	Angle       string  `json:"angle"`
	CTA         string  `json:"cta"`
	Hook        string  `json:"hook"`
	Duration    int     `json:"duration"`
	AspectRatio string  `json:"aspect_ratio"`
	Scenes      []Scene `json:"scenes"`
}

// Scene est une scène du storyboard.
type Scene struct {
	ID                string  `json:"id"`
	Type              string  `json:"type"`
	Duration          float64 `json:"duration"`
	Script            string  `json:"script,omitempty"`             // texte parlé (scène avatar)
	TextOverlay       string  `json:"text_overlay,omitempty"`       // texte affiché à l'écran
	VisualDescription string  `json:"visual_description,omitempty"` // description visuelle (futurs providers B-roll)
}

// AvatarScript concatène le texte parlé des scènes avatar (script complet
// envoyé au provider de vidéo avatar).
func (s Storyboard) AvatarScript() string {
	out := make([]string, 0, len(s.Scenes))
	for _, sc := range s.Scenes {
		if sc.Type == SceneAvatar && sc.Script != "" {
			out = append(out, sc.Script)
		}
	}
	return strings.Join(out, "\n\n")
}

// Validate vérifie la cohérence minimale du storyboard.
func (s Storyboard) Validate() error {
	if len(s.Scenes) == 0 {
		return ErrVideoStoryboardInvalid
	}
	if s.AvatarScript() == "" {
		return ErrVideoStoryboardInvalid
	}
	for _, sc := range s.Scenes {
		switch sc.Type {
		case SceneAvatar, SceneProduct, SceneText:
		default:
			return ErrVideoStoryboardInvalid
		}
	}
	return nil
}

// ResultVideoAd est le résultat JSONB du job video_ad.
type ResultVideoAd struct {
	AssetID         string     `json:"asset_id"`
	ThumbAssetID    string     `json:"thumb_asset_id,omitempty"`
	Storyboard      Storyboard `json:"storyboard"`
	ProviderVideoID string     `json:"provider_video_id,omitempty"`
	ProviderName    string     `json:"provider_name,omitempty"`
	Duration        float64    `json:"duration,omitempty"`
}
