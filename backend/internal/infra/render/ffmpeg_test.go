package render

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"afrilaunch/backend/internal/application/port"
)

func osOpen(path string) (*os.File, error) { return os.Open(path) }

func TestCanvasFor(t *testing.T) {
	cases := map[string][2]int{
		"9:16":  {1080, 1920},
		"1:1":   {1080, 1080},
		"16:9":  {1920, 1080},
		"":      {1080, 1920}, // défaut portrait
		"20:99": {1080, 1920},
	}
	for ratio, want := range cases {
		w, h := canvasFor(ratio)
		if w != want[0] || h != want[1] {
			t.Fatalf("canvasFor(%q) = %dx%d, want %dx%d", ratio, w, h, want[0], want[1])
		}
	}
}

func TestFormatSRT(t *testing.T) {
	got := string(formatSRT([]port.SubtitleCue{
		{Start: 0, End: 1.5, Text: "Première ligne"},
		{Start: 1.5, End: 62.25, Text: "Deuxième ligne"},
	}))
	want := "1\n00:00:00,000 --> 00:00:01,500\nPremière ligne\n\n" +
		"2\n00:00:01,500 --> 00:01:02,250\nDeuxième ligne\n\n"
	if got != want {
		t.Fatalf("formatSRT =\n%q\nwant\n%q", got, want)
	}
}

func TestFormatSRTClampsNegative(t *testing.T) {
	got := string(formatSRT([]port.SubtitleCue{{Start: -1, End: 0.5, Text: "x"}}))
	if !strings.Contains(got, "00:00:00,000") {
		t.Fatalf("negative start not clamped: %q", got)
	}
}

func TestCardHTMLEscapesText(t *testing.T) {
	in := port.AvatarAdRenderInput{BrandTitle: "Guide & Business", HookText: "Arrête de *perdre* <ton> temps"}
	doc := cardHTML(in, 1080, 1920, in.HookText, true)
	if strings.Contains(doc, "<ton>") {
		t.Fatal("raw HTML injected into card text")
	}
	if !strings.Contains(doc, "<em>perdre</em>") {
		t.Fatal("emphasis not rendered")
	}
	if !strings.Contains(doc, "Guide &amp; Business") {
		t.Fatal("brand title not escaped")
	}
}

func TestCardHTMLWithoutCover(t *testing.T) {
	doc := cardHTML(port.AvatarAdRenderInput{}, 1080, 1920, "CTA final", false)
	if strings.Contains(doc, "<img") {
		t.Fatal("cover img should be absent without CoverPNG")
	}
	if !strings.Contains(doc, "CTA final") {
		t.Fatal("CTA text missing")
	}
}

// TestRenderAvatarAdIntegration valide le montage complet si ffmpeg et
// Chrome sont disponibles (installés dans l'image Docker backend).
func TestRenderAvatarAdIntegration(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg non installé")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe non installé")
	}

	dir := t.TempDir()
	// Vidéo avatar synthétique de 2 s (30 fps, 720x1280) + audio silencieux.
	avatarPath := dir + "/avatar.mp4"
	if err := exec.Command("ffmpeg", "-y",
		"-f", "lavfi", "-i", "testsrc=size=720x1280:rate=30:duration=2",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=2",
		"-c:v", "libx264", "-preset", "ultrafast", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-shortest", avatarPath).Run(); err != nil {
		t.Skipf("génération vidéo test impossible: %v", err)
	}
	avatar, err := readFileBytes(avatarPath)
	if err != nil {
		t.Fatal(err)
	}

	r := NewFFmpegRenderer(nil, "ffmpeg", "ffprobe")
	res, err := r.RenderAvatarAd(context.Background(), port.AvatarAdRenderInput{
		AvatarVideo: avatar,
		HookText:    "Carte intro",
		CTAText:     "Carte outro",
		SubtitleCues: []port.SubtitleCue{
			{Start: 0, End: 1, Text: "Sous-titre un"},
			{Start: 1, End: 2, Text: "Sous-titre deux"},
		},
		AspectRatio: "9:16",
	})
	if err != nil {
		t.Fatalf("RenderAvatarAd: %v", err)
	}
	if len(res.Video) < 10_000 {
		t.Fatalf("vidéo trop petite: %d bytes", len(res.Video))
	}
	if len(res.Thumb) == 0 {
		t.Fatal("vignette manquante")
	}
	if res.Duration < 6.5 || res.Duration > 8 {
		t.Fatalf("durée = %.2f, want ~7", res.Duration)
	}
}

func readFileBytes(path string) ([]byte, error) {
	f, err := osOpen(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
