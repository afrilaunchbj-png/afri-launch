package videoad

import (
	"strings"
	"testing"

	"afrilaunch/backend/internal/domain"
)

func splitWords(s string) []string { return strings.Fields(s) }

func contains(s, sub string) bool { return strings.Contains(s, sub) }

func TestBuildSubtitleCues(t *testing.T) {
	sb := domain.Storyboard{
		Duration: 30,
		Scenes: []domain.Scene{
			{ID: "scene_1", Type: domain.SceneProduct, Duration: 3, TextOverlay: "Hook"},
			{ID: "scene_2", Type: domain.SceneAvatar, Duration: 15, Script: "Un deux trois quatre cinq six sept huit neuf dix onze douze treize quatorze quinze"},
			{ID: "scene_3", Type: domain.SceneAvatar, Duration: 12, Script: "Achète maintenant, c'est simple."},
		},
	}

	cues := BuildSubtitleCues(sb, 28.0)
	if len(cues) == 0 {
		t.Fatal("expected cues, got none")
	}
	for _, c := range cues {
		if c.End <= c.Start {
			t.Fatalf("invalid cue timing: %+v", c)
		}
		if c.Text == "" {
			t.Fatal("empty cue text")
		}
		if wc := len(splitWords(c.Text)); wc > 7 {
			t.Fatalf("cue too long (%d words): %q", wc, c.Text)
		}
	}
	// Le dernier cue se termine approximativement à la durée parlée.
	last := cues[len(cues)-1]
	if last.End < 27 || last.End > 29 {
		t.Fatalf("last cue end = %.2f, want ≈28", last.End)
	}
}

func TestBuildSubtitleCuesFallbackDuration(t *testing.T) {
	sb := domain.Storyboard{
		Duration: 15,
		Scenes: []domain.Scene{
			{ID: "scene_1", Type: domain.SceneAvatar, Duration: 15, Script: "Hello world this is a test"},
		},
	}
	cues := BuildSubtitleCues(sb, 0) // fallback sur sb.Duration
	last := cues[len(cues)-1]
	if last.End < 14.9 || last.End > 15.1 {
		t.Fatalf("last cue end = %.2f, want ≈15", last.End)
	}
}

func TestResolveVideoRequest(t *testing.T) {
	sb := domain.Storyboard{
		AspectRatio: "9:16",
		Scenes: []domain.Scene{
			{Type: domain.SceneAvatar, Script: "Ligne une."},
			{Type: domain.SceneProduct},
			{Type: domain.SceneAvatar, Script: "Ligne deux."},
		},
	}
	params := domain.VideoAdParams{AvatarID: "override-avatar"}
	req := ResolveVideoRequest(sb, params, "Guide business", ProviderDefaults{AvatarID: "default-avatar", VoiceID: "default-voice"})

	if req.AvatarID != "override-avatar" {
		t.Fatalf("AvatarID = %q, want override", req.AvatarID)
	}
	if req.VoiceID != "default-voice" {
		t.Fatalf("VoiceID = %q, want default", req.VoiceID)
	}
	if req.AspectRatio != "9:16" {
		t.Fatalf("AspectRatio = %q, want storyboard ratio", req.AspectRatio)
	}
	if req.Script != "Ligne une.\n\nLigne deux." {
		t.Fatalf("Script = %q", req.Script)
	}
	if req.Resolution != "1080p" || req.Title != "Guide business" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestNormalizeStoryboardDuration(t *testing.T) {
	sb := domain.Storyboard{
		Scenes: []domain.Scene{
			{Type: domain.SceneAvatar, Duration: 4, Script: "hook"},
			{Type: domain.SceneAvatar, Duration: 0, Script: "problem"}, // clamp à 1
			{Type: domain.SceneProduct, Duration: 9},
		},
	}
	normalizeScenes(&sb, 30)
	sum := 0
	for i, sc := range sb.Scenes {
		if sc.ID != "scene_"+string(rune('1'+i)) {
			t.Fatalf("scene ID = %q", sc.ID)
		}
		if sc.Duration < 1 {
			t.Fatalf("scene duration < 1: %v", sc.Duration)
		}
		sum += int(sc.Duration)
	}
	if sum != 30 {
		t.Fatalf("sum of durations = %d, want 30", sum)
	}
}

func TestAnalysisPromptIncludesParams(t *testing.T) {
	c := AdContext{Topic: "Guide business", Language: "fr", Country: "Bénin"}
	params := domain.VideoAdParams{Angle: "urgency", CTA: "Achète maintenant", Duration: 15, AspectRatio: "9:16"}
	p := AnalysisPrompt(c, params)
	for _, want := range []string{"urgency", "Achète maintenant", "15-second", "Guide business", "fr"} {
		if !contains(p, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}
