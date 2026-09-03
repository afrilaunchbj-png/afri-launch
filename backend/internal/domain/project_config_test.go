package domain

import "testing"

func TestProjectConfigResolvedPageRange(t *testing.T) {
	// Défauts quand non configuré.
	c := ProjectConfig{}
	minP, maxP := c.ResolvedPageRange()
	if minP != EbookDefaultMinPages || maxP != EbookDefaultMaxPages {
		t.Fatalf("défauts : got (%d,%d), want (%d,%d)", minP, maxP, EbookDefaultMinPages, EbookDefaultMaxPages)
	}

	// Valeurs valides conservées.
	c = ProjectConfig{EbookMinPages: 4, EbookMaxPages: 10}
	minP, maxP = c.ResolvedPageRange()
	if minP != 4 || maxP != 10 {
		t.Fatalf("got (%d,%d), want (4,10)", minP, maxP)
	}

	// Borne haute plafonnée, min > max borné au max.
	c = ProjectConfig{EbookMinPages: 99, EbookMaxPages: 3}
	minP, maxP = c.ResolvedPageRange()
	if maxP > EbookMaxPagesCeiling {
		t.Fatalf("max %d > plafond %d", maxP, EbookMaxPagesCeiling)
	}
	if minP > maxP {
		t.Fatalf("min %d > max %d", minP, maxP)
	}
}

func TestValidHexColor(t *testing.T) {
	valid := []string{"", "#003527", "#0F766E", "#abc"}
	for _, v := range valid {
		if !ValidHexColor(v) {
			t.Errorf("%q devrait être valide", v)
		}
	}
	invalid := []string{"003527", "#1", "#12345", "#003527ff", "bleu", "#GGGGGG"}
	for _, v := range invalid {
		if ValidHexColor(v) {
			t.Errorf("%q devrait être invalide", v)
		}
	}
}

func TestProjectPaletteNormalize(t *testing.T) {
	p := ProjectPalette{Primary: "#003527", Secondary: " #FEA619 ", Text: "#111111"}.Normalize(PaletteSourceUser)
	if p.Secondary != "#fea619" {
		t.Errorf("Secondary = %q", p.Secondary)
	}
	if p.Source != PaletteSourceUser {
		t.Errorf("Source = %q", p.Source)
	}
	if !(ProjectPalette{}).Empty() {
		t.Error("palette vide attendue")
	}
}

func TestParseProjectConfigRoundTrip(t *testing.T) {
	cfg := ProjectConfig{
		Palette:       &ProjectPalette{Primary: "#003527", Source: PaletteSourceAI},
		Style:         "moderne, chaleureux",
		EbookMinPages: 8,
		EbookMaxPages: 20,
	}
	parsed := ParseProjectConfig(cfg.Marshal())
	if parsed.Style != cfg.Style || parsed.EbookMinPages != 8 || parsed.EbookMaxPages != 20 {
		t.Fatalf("round trip : %+v", parsed)
	}
	if parsed.Palette == nil || parsed.Palette.Primary != "#003527" || parsed.Palette.Source != PaletteSourceAI {
		t.Fatalf("palette : %+v", parsed.Palette)
	}

	// JSONB vide/invalide → config par défaut.
	if _, maxP := ParseProjectConfig(nil).ResolvedPageRange(); maxP != EbookDefaultMaxPages {
		t.Errorf("config nil devrait donner les défauts")
	}
	if _, maxP := ParseProjectConfig([]byte("not json")).ResolvedPageRange(); maxP != EbookDefaultMaxPages {
		t.Errorf("json invalide devrait donner les défauts")
	}
}
