package document

import (
	"strings"
	"testing"
)

func TestBuildEbookPrompt(t *testing.T) {
	p := BuildEbookPrompt(EbookRequest{
		Topic:    "comptabilité pour PME",
		Language: "fr",
		Country:  "Sénégal",
		Audience: "commerçants",
		Product:  "guide",
	})
	if p.System == "" || p.User == "" {
		t.Fatal("empty prompt")
	}
	for _, want := range []string{"Impeccable", "#003527", "#855300", "Lexend", "@page"} {
		if !strings.Contains(p.System, want) {
			t.Errorf("system prompt missing %q", want)
		}
	}
	if !strings.Contains(p.User, "comptabilité pour PME") || !strings.Contains(p.User, "Sénégal") {
		t.Errorf("user prompt missing request data: %q", p.User)
	}
}

func TestBuildDeckPrompt(t *testing.T) {
	p := BuildDeckPrompt(DeckRequest{Topic: "marketplace", Language: "en", Country: "Nigeria", Audience: "investors"})
	if !strings.Contains(p.System, `class="slide"`) || !strings.Contains(p.System, "1280x720") {
		t.Errorf("deck system prompt missing slide format")
	}
	if !strings.Contains(p.User, "marketplace") {
		t.Errorf("deck user prompt missing topic")
	}
}

func TestStripCodeFences(t *testing.T) {
	cases := []struct{ in, want string }{
		{"<html></html>", "<html></html>"},
		{"```html\n<html></html>\n```", "<html></html>"},
		{"```\n<html></html>\n```", "<html></html>"},
		{"  \n<html></html>\n  ", "<html></html>"},
	}
	for _, c := range cases {
		if got := stripCodeFences(c.in); got != c.want {
			t.Errorf("stripCodeFences(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
