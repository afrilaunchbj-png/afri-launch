package document

import (
	"strings"
	"testing"

	"afrilaunch/backend/internal/domain"
)

func TestBuildEbookPromptWithPaletteAndPages(t *testing.T) {
	palette := &domain.ProjectPalette{Primary: "#0f766e", Secondary: "#f59e0b", Accent: "#ef4444", Background: "#fffbeb", Text: "#1c1917"}
	p := BuildEbookPrompt(EbookRequest{
		Topic:    "comptabilité",
		Language: "fr",
		Country:  "Bénin",
		Palette:  palette,
		Style:    "moderne, chaleureux",
		MinPages: 6,
		MaxPages: 14,
		HasCover: true,
	})

	// L'identité visuelle est injectée en dur dans le système.
	for _, want := range []string{"#0f766e", "#f59e0b", "#fffbeb", "--color-primary", "moderne, chaleureux"} {
		if !strings.Contains(p.System, want) {
			t.Errorf("system prompt missing %q", want)
		}
	}
	// Consigne de longueur.
	if !strings.Contains(p.User, "between 6 and 14 A4 pages") {
		t.Errorf("user prompt missing page range: %q", p.User)
	}
	// La cover est fournie : le LLM ne doit PAS créer sa propre cover.
	if !strings.Contains(p.System, "Do NOT create your own cover page") {
		t.Errorf("system prompt should exclude LLM cover")
	}
	if strings.Contains(p.System, "Structure: cover") {
		t.Errorf("l'ancienne structure avec cover LLM devrait avoir disparu")
	}
}

func TestBuildEbookPromptWithoutPalette(t *testing.T) {
	p := BuildEbookPrompt(EbookRequest{Topic: "x", Language: "fr", MinPages: 2, MaxPages: 40})
	if strings.Contains(p.System, "--color-primary") {
		t.Errorf("sans palette, aucune directive de couleurs attendue")
	}
	if !strings.Contains(p.User, "between 2 and 40") {
		t.Errorf("page range manquant")
	}
}

func TestPrependCoverPage(t *testing.T) {
	html := []byte("<html><head><title>x</title></head><body><section class=\"chapter\">Intro</section></body></html>")
	out := string(PrependCoverPage(html, []byte{0x89, 'P', 'N', 'G'}))

	// La cover est la première section du body.
	bodyIdx := strings.Index(out, "<body>")
	coverIdx := strings.Index(out, `<section class="cover-page">`)
	if coverIdx < 0 {
		t.Fatal("cover page manquante")
	}
	if coverIdx < bodyIdx {
		t.Fatalf("cover injectée avant le body")
	}
	chapterIdx := strings.Index(out, `class="chapter"`)
	if chapterIdx < coverIdx {
		t.Error("la cover doit précéder le contenu")
	}

	// Règles print : marge 0 sur la 1re page + saut après la cover.
	for _, want := range []string{"@page :first", "break-after: page", "data:image/png;base64,"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}

	// Pas de cover → inchangé.
	if got := PrependCoverPage(html, nil); string(got) != string(html) {
		t.Error("sans cover, le HTML doit rester inchangé")
	}
}

func TestBuildSalesPagePromptPalette(t *testing.T) {
	palette := &domain.ProjectPalette{Primary: "#0f766e", Background: "#fffbeb", Text: "#1c1917"}
	p := BuildSalesPagePrompt(SalesPageRequest{Product: "guide", Language: "fr", Palette: palette})
	if !strings.Contains(p.System, "#0f766e") {
		t.Errorf("palette manquante dans la page de vente")
	}
}
