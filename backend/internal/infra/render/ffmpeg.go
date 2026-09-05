package render

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"afrilaunch/backend/internal/application/port"
)

// Dimensions d'export par ratio d'aspect (presets TikTok/Reels, carré, paysage).
func canvasFor(aspectRatio string) (width, height int) {
	switch aspectRatio {
	case "1:1":
		return 1080, 1080
	case "16:9":
		return 1920, 1080
	default: // 9:16
		return 1080, 1920
	}
}

// Durée (s) des cartes d'intro/outro du montage.
const (
	cardDuration = 2.5
	introColor   = "#003527" // primary Emerald & Amber
	accentColor  = "#fea619" // accent
	textColor    = "#f7f4ec"
)

// FFmpegRenderer implémente port.VideoRenderer : assemble la vidéo du
// provider (ex. HeyGen) avec une carte d'intro (cover + hook), des
// sous-titres burnés et une carte d'outro (cover + CTA). Les cartes sont
// rendues en HTML via chromedp pour respecter l'identité visuelle.
type FFmpegRenderer struct {
	cards *ChromedpRenderer
	bin   string
	probe string
}

// NewFFmpegRenderer construit le renderer vidéo. bin/probe vides = "ffmpeg"/"ffprobe".
func NewFFmpegRenderer(cards *ChromedpRenderer, bin, probe string) *FFmpegRenderer {
	if bin == "" {
		bin = "ffmpeg"
	}
	if probe == "" {
		probe = "ffprobe"
	}
	return &FFmpegRenderer{cards: cards, bin: bin, probe: probe}
}

// RenderAvatarAd assemble la vidéo publicitaire finale (MP4 H.264 + vignette JPEG).
func (r *FFmpegRenderer) RenderAvatarAd(ctx context.Context, in port.AvatarAdRenderInput) (port.AvatarAdRenderResult, error) {
	w, h := canvasFor(in.AspectRatio)

	dir, err := os.MkdirTemp("", "afri-ad-*")
	if err != nil {
		return port.AvatarAdRenderResult{}, err
	}
	defer os.RemoveAll(dir)

	if len(in.AvatarVideo) == 0 {
		return port.AvatarAdRenderResult{}, fmt.Errorf("render: vidéo avatar manquante")
	}
	avatarPath := filepath.Join(dir, "avatar.mp4")
	if err := os.WriteFile(avatarPath, in.AvatarVideo, 0o600); err != nil {
		return port.AvatarAdRenderResult{}, err
	}

	spoken, err := r.probeDuration(ctx, avatarPath)
	if err != nil {
		return port.AvatarAdRenderResult{}, err
	}

	// Les cues sont estimés sur la durée cible : on les redimensionne sur la
	// durée réellement parlée (mesurée via ffprobe) pour éviter la dérive.
	cues := rescaleCues(in.SubtitleCues, spoken)

	// Cartes intro/outro (HTML → PNG, identité visuelle).
	introPNG, err := r.card(ctx, in, w, h, in.HookText, true)
	if err != nil {
		return port.AvatarAdRenderResult{}, err
	}
	outroPNG, err := r.card(ctx, in, w, h, in.CTAText, false)
	if err != nil {
		return port.AvatarAdRenderResult{}, err
	}
	introPath := filepath.Join(dir, "intro.png")
	outroPath := filepath.Join(dir, "outro.png")
	if err := os.WriteFile(introPath, introPNG, 0o600); err != nil {
		return port.AvatarAdRenderResult{}, err
	}
	if err := os.WriteFile(outroPath, outroPNG, 0o600); err != nil {
		return port.AvatarAdRenderResult{}, err
	}

	// Sous-titres (SRT) : timing aligné sur le segment avatar.
	subsPath := ""
	if len(cues) > 0 {
		subsPath = filepath.Join(dir, "subs.srt")
		if err := os.WriteFile(subsPath, []byte(formatSRT(cues)), 0o600); err != nil {
			return port.AvatarAdRenderResult{}, err
		}
	}

	outPath := filepath.Join(dir, "final.mp4")
	if err := r.assemble(ctx, dir, w, h, introPath, outroPath, avatarPath, subsPath, outPath); err != nil {
		return port.AvatarAdRenderResult{}, err
	}
	video, err := os.ReadFile(outPath)
	if err != nil {
		return port.AvatarAdRenderResult{}, err
	}

	// Vignette : un instant de la carte d'intro (mockup du produit).
	thumbPath := filepath.Join(dir, "thumb.jpg")
	if err := r.run(ctx, dir, r.bin, "-y", "-ss", "1.2", "-i", outPath,
		"-frames:v", "1", "-q:v", "4", thumbPath); err != nil {
		return port.AvatarAdRenderResult{}, fmt.Errorf("render: vignette: %w", err)
	}
	thumb, err := os.ReadFile(thumbPath)
	if err != nil {
		return port.AvatarAdRenderResult{}, err
	}

	return port.AvatarAdRenderResult{
		Video:    video,
		Thumb:    thumb,
		Duration: spoken + 2*cardDuration,
	}, nil
}

// assemble concatène intro + avatar (sous-titres burnés) + outro.
func (r *FFmpegRenderer) assemble(ctx context.Context, dir string, w, h int, introPath, outroPath, avatarPath, subsPath, outPath string) error {
	style := fmt.Sprintf("FontSize=%d,PrimaryColour=&H00FFFFFF,OutlineColour=&H90000000,BorderStyle=1,Outline=2,Shadow=0,Alignment=2,MarginV=%d,Bold=1",
		subtitleFontSize(h), h/24)

	main := fmt.Sprintf(
		"[0:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=0x003527,setsar=1,fps=30,format=yuv420p[intro];",
		w, h, w, h)
	if subsPath != "" {
		// Chemin relatif (cmd.Dir = répertoire temporaire) : pas d'échappement de chemin.
		main += fmt.Sprintf(
			"[2:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=0x003527,setsar=1,fps=30,subtitles=%s:force_style='%s'[main];",
			w, h, w, h, filepath.Base(subsPath), style)
	} else {
		main += fmt.Sprintf(
			"[2:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=0x003527,setsar=1,fps=30,format=yuv420p[main];", w, h, w, h)
	}
	filter := main + fmt.Sprintf(
		"[1:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=0x003527,setsar=1,fps=30,format=yuv420p[outro];"+
			"[intro][3:a][main][2:a][outro][4:a]concat=n=3:v=1:a=1[v][a]",
		w, h, w, h)

	args := []string{
		"-y",
		"-loop", "1", "-t", fmt.Sprintf("%.2f", cardDuration), "-i", introPath,
		"-loop", "1", "-t", fmt.Sprintf("%.2f", cardDuration), "-i", outroPath,
		"-i", avatarPath,
		"-f", "lavfi", "-t", fmt.Sprintf("%.2f", cardDuration), "-i", "anullsrc=r=44100:cl=stereo",
		"-f", "lavfi", "-t", fmt.Sprintf("%.2f", cardDuration), "-i", "anullsrc=r=44100:cl=stereo",
		"-filter_complex", filter,
		"-map", "[v]", "-map", "[a]",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "22", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "128k",
		"-movflags", "+faststart",
		outPath,
	}
	return r.run(ctx, dir, r.bin, args...)
}

// card rend une carte texte (hook ou CTA) aux dimensions du canvas, avec le
// mockup du produit si disponible.
func (r *FFmpegRenderer) card(ctx context.Context, in port.AvatarAdRenderInput, w, h int, text string, intro bool) ([]byte, error) {
	htmlDoc := cardHTML(in, w, h, text, intro)
	return r.cards.HTMLToPNG(ctx, []byte(htmlDoc), w, h)
}

// cardHTML produit le HTML d'une carte brandée (Emerald & Amber Ledger).
func cardHTML(in port.AvatarAdRenderInput, w, h int, text string, intro bool) string {
	cover := ""
	if len(in.CoverPNG) > 0 {
		cover = fmt.Sprintf(
			`<img class="cover" src="data:image/png;base64,%s" alt=""/>`,
			base64Std(in.CoverPNG))
	}
	badge := ""
	if intro && in.BrandTitle != "" {
		badge = `<div class="brand">` + html.EscapeString(in.BrandTitle) + `</div>`
	}
	accent := "#a5bda1"
	if !intro {
		accent = accentColor
	}
	return `<!doctype html><html><head><meta charset="utf-8"><style>
* { box-sizing: border-box; margin: 0; }
html, body { width: ` + strconv.Itoa(w) + `px; height: ` + strconv.Itoa(h) + `px; overflow: hidden;
  font-family: Lexend, Inter, system-ui, sans-serif; }
body { background: ` + introColor + `; color: ` + textColor + `; display: flex;
  flex-direction: column; align-items: center; justify-content: center; padding: ` + fmt.Sprint(h/12) + `px; gap: ` + fmt.Sprint(h/28) + `px; }
.cover { max-width: 72%; max-height: 52%; border-radius: 24px; box-shadow: 0 24px 64px rgba(0,0,0,.45); object-fit: contain; }
.text { font-weight: 700; font-size: ` + fmt.Sprint(h/16) + `px; line-height: 1.25; text-align: center; text-wrap: balance; }
.text em { color: ` + accentColor + `; font-style: normal; }
.brand { position: absolute; bottom: 5%; font-size: ` + fmt.Sprint(h/45) + `px; letter-spacing: .14em;
  text-transform: uppercase; opacity: .72; }
.rule { width: 96px; height: 6px; border-radius: 3px; background: ` + accent + `; }
</style></head><body>` + cover + `<div class="rule"></div><div class="text">` +
		escapeEmphasis(text) + `</div>` + badge + `</body></html>`
}

// escapeEmphasis échappe le texte en convertissant *mots* en <em>.
func escapeEmphasis(text string) string {
	parts := strings.Split(html.EscapeString(text), "*")
	var b strings.Builder
	for i, p := range parts {
		if i%2 == 1 {
			b.WriteString("<em>" + p + "</em>")
		} else {
			b.WriteString(p)
		}
	}
	return b.String()
}

// subtitleFontSize renvoie la taille ASS (repère PlayRes 288 px de haut)
// pour un rendu lisible sur la hauteur réelle.
func subtitleFontSize(h int) int {
	switch {
	case h >= 1920:
		return 13
	case h >= 1080:
		return 12
	default:
		return 10
	}
}

// probeDuration mesure la durée (s) d'un média via ffprobe.
func (r *FFmpegRenderer) probeDuration(ctx context.Context, path string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, r.probe, "-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", path).Output()
	if err != nil {
		return 0, fmt.Errorf("render: ffprobe: %w", err)
	}
	d, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("render: durée invalide %q", strings.TrimSpace(string(out)))
	}
	return d, nil
}

func (r *FFmpegRenderer) run(ctx context.Context, dir, bin string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := stderr.String()
		if len(msg) > 600 {
			msg = msg[len(msg)-600:]
		}
		return fmt.Errorf("render: %s: %w: %s", filepath.Base(bin), err, msg)
	}
	return nil
}

// rescaleCues remet à l'échelle les cues (estimés sur la durée cible) sur la
// durée parlée réelle de la vidéo provider.
func rescaleCues(cues []port.SubtitleCue, spoken float64) []port.SubtitleCue {
	if len(cues) == 0 || spoken <= 0 {
		return nil
	}
	last := cues[len(cues)-1].End
	if last <= 0 {
		return nil
	}
	f := spoken / last
	out := make([]port.SubtitleCue, len(cues))
	for i, c := range cues {
		out[i] = port.SubtitleCue{Start: c.Start * f, End: c.End * f, Text: c.Text}
	}
	return out
}

// formatSRT encode les cues en SubRip.
func formatSRT(cues []port.SubtitleCue) []byte {
	var b strings.Builder
	for i, c := range cues {
		b.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n", i+1, srtTime(c.Start), srtTime(c.End), strings.TrimSpace(c.Text)))
	}
	return []byte(b.String())
}

func srtTime(sec float64) string {
	if sec < 0 {
		sec = 0
	}
	ms := int(sec * 1000)
	h, m, s := ms/3600000, (ms%3600000)/60000, (ms%60000)/1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms%1000)
}

func base64Std(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
