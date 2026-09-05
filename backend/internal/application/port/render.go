package port

import "context"

// Renderer transforme du HTML (auto-porteur) en PDF ou en PPTX (image-par-slide).
type Renderer interface {
	HTMLToPDF(ctx context.Context, html []byte) ([]byte, error)
	SlidesToPPTX(ctx context.Context, html []byte) ([]byte, error)
	// SlidesToPPTXWithCover assemble le PPTX en plaçant coverPNG en première
	// slide (workflow cover-first).
	SlidesToPPTXWithCover(ctx context.Context, html []byte, coverPNG []byte) ([]byte, error)
}

// SubtitleCue est une entrée de sous-titres (timing en secondes).
type SubtitleCue struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// AvatarAdRenderInput décrit le montage d'une vidéo publicitaire avatar :
// carte d'intro (cover + hook), vidéo du provider avec sous-titres burnés,
// carte d'outro (cover + CTA).
type AvatarAdRenderInput struct {
	// AvatarVideo est le MP4 du provider (ex. HeyGen) — base du montage.
	AvatarVideo []byte
	// CoverPNG est le mockup du produit (cover du projet) — nil si absent.
	CoverPNG []byte
	// HookText s'affiche sur la carte d'intro, CTAText sur la carte d'outro.
	HookText string
	CTAText  string
	// SubtitleCues sont les sous-titres du script (burnés sur la vidéo).
	SubtitleCues []SubtitleCue
	// AspectRatio pilote les dimensions d'export : 9:16 → 1080x1920,
	// 1:1 → 1080x1080, 16:9 → 1920x1080.
	AspectRatio string
	// BrandTitle est le titre du produit (petit bandeau discret).
	BrandTitle string
}

// AvatarAdRenderResult est le rendu final : MP4 assemblé + vignette JPEG.
type AvatarAdRenderResult struct {
	Video    []byte
	Thumb    []byte
	Duration float64
}

// VideoRenderer assemble les assets d'une publicité en vidéo finale
// (FFmpeg ou équivalent). Abstraction provider-agnostic du montage.
type VideoRenderer interface {
	RenderAvatarAd(ctx context.Context, in AvatarAdRenderInput) (AvatarAdRenderResult, error)
}
