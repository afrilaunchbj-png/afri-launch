package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Sources d'une palette : proposée par l'IA ou fixée par l'utilisateur.
const (
	PaletteSourceAI   = "ai"
	PaletteSourceUser = "user"
)

// Bornes du nombre de pages d'ebook.
const (
	EbookMinPagesFloor   = 2
	EbookMaxPagesCeiling = 40
	EbookDefaultMinPages = 6
	EbookDefaultMaxPages = 14
)

// hexColor valide une couleur CSS hexadécimale (#RGB ou #RRGGBB).
var hexColor = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// ValidHexColor indique si s est une couleur hexadécimale valide (ou vide).
func ValidHexColor(s string) bool {
	if s == "" {
		return true
	}
	return hexColor.MatchString(s)
}

// ProjectPalette est l'identité visuelle d'un projet, utilisée par tous les
// assets (cover, ebook, affiches, page de vente).
type ProjectPalette struct {
	Primary    string `json:"primary,omitempty"`
	Secondary  string `json:"secondary,omitempty"`
	Accent     string `json:"accent,omitempty"`
	Background string `json:"background,omitempty"`
	Text       string `json:"text,omitempty"`
	Source     string `json:"source,omitempty"` // "ai" | "user"
}

// Empty indique si la palette n'a aucune couleur définie.
func (p ProjectPalette) Empty() bool {
	return p.Primary == "" && p.Secondary == "" && p.Accent == "" &&
		p.Background == "" && p.Text == ""
}

// Normalize nettoie les couleurs (minuscules) et force la source.
func (p ProjectPalette) Normalize(source string) ProjectPalette {
	p.Primary = strings.ToLower(strings.TrimSpace(p.Primary))
	p.Secondary = strings.ToLower(strings.TrimSpace(p.Secondary))
	p.Accent = strings.ToLower(strings.TrimSpace(p.Accent))
	p.Background = strings.ToLower(strings.TrimSpace(p.Background))
	p.Text = strings.ToLower(strings.TrimSpace(p.Text))
	p.Source = source
	return p
}

// CSS déclare les couleurs en variables CSS pour injection dans un prompt.
func (p ProjectPalette) CSS() string {
	return fmt.Sprintf(
		"--color-primary: %s; --color-secondary: %s; --color-accent: %s; --color-background: %s; --color-text: %s",
		p.Primary, p.Secondary, p.Accent, p.Background, p.Text,
	)
}

// ProjectConfig est la configuration de génération d'un projet (JSONB).
type ProjectConfig struct {
	Palette       *ProjectPalette `json:"palette,omitempty"`
	Style         string          `json:"style,omitempty"`
	EbookMinPages int             `json:"ebook_min_pages,omitempty"`
	EbookMaxPages int             `json:"ebook_max_pages,omitempty"`
}

// ParseProjectConfig décode le JSONB projects.config (fallback : config vide).
func ParseProjectConfig(raw []byte) ProjectConfig {
	var c ProjectConfig
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &c)
	}
	return c
}

// Marshal sérialise la config pour stockage JSONB.
func (c ProjectConfig) Marshal() []byte {
	raw, err := json.Marshal(c)
	if err != nil {
		return []byte("{}")
	}
	return raw
}

// ResolvedPageRange applique les défauts et bornes du nombre de pages ebook.
func (c ProjectConfig) ResolvedPageRange() (minPages, maxPages int) {
	minPages, maxPages = c.EbookMinPages, c.EbookMaxPages
	if minPages < EbookMinPagesFloor {
		minPages = EbookDefaultMinPages
	}
	if maxPages < minPages {
		maxPages = EbookDefaultMaxPages
	}
	if maxPages > EbookMaxPagesCeiling {
		maxPages = EbookMaxPagesCeiling
	}
	if minPages > maxPages {
		minPages = maxPages
	}
	return minPages, maxPages
}

// EffectivePalette renvoie la palette à utiliser (peut être vide : l'IA décide).
func (c ProjectConfig) EffectivePalette() *ProjectPalette {
	if c.Palette == nil || c.Palette.Empty() {
		return nil
	}
	return c.Palette
}
