package render

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func chromePath() string {
	for _, p := range []string{"/usr/bin/google-chrome", "/usr/bin/chromium", "/usr/bin/chromium-browser"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func TestSlidesToPNG(t *testing.T) {
	path := chromePath()
	if path == "" {
		t.Skip("chrome non disponible")
	}

	r := NewChromedpRenderer(path)
	html := []byte(`<!doctype html><html><head><style>
		body{margin:0}
		section.slide{width:1280px;height:720px;overflow:hidden;background:#003527;color:#fff}
	</style></head><body>
		<section class="slide"><h1>Slide 1</h1></section>
		<section class="slide"><h1>Slide 2</h1></section>
	</body></html>`)

	pngs, err := r.SlidesToPNG(context.Background(), html)
	if err != nil {
		t.Fatalf("SlidesToPNG: %v", err)
	}
	if len(pngs) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(pngs))
	}
	for i, p := range pngs {
		if len(p) < 8 || !bytes.HasPrefix(p, []byte{0x89, 'P', 'N', 'G'}) {
			t.Errorf("slide %d is not a PNG (len=%d)", i+1, len(p))
		}
	}
}

func TestHTMLToPDF(t *testing.T) {
	path := chromePath()
	if path == "" {
		t.Skip("chrome non disponible")
	}

	r := NewChromedpRenderer(path)
	html := []byte(`<!doctype html><html><head><style>@page{size:A4;margin:2cm}</style></head><body><h1>Hello</h1><p>world</p></body></html>`)

	pdf, err := r.HTMLToPDF(context.Background(), html)
	if err != nil {
		t.Fatalf("HTMLToPDF: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatalf("not a PDF (first bytes: %q)", pdf[:min(len(pdf), 8)])
	}
}

func TestSlidesToPPTX(t *testing.T) {
	path := chromePath()
	if path == "" {
		t.Skip("chrome non disponible")
	}

	r := NewChromedpRenderer(path)
	html := []byte(`<!doctype html><html><head><style>
		body{margin:0}
		section.slide{width:1280px;height:720px;background:#855300}
	</style></head><body><section class="slide"><h1>One</h1></section></body></html>`)

	out, err := r.SlidesToPPTX(context.Background(), html)
	if err != nil {
		t.Fatalf("SlidesToPPTX: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("PK")) {
		t.Fatalf("not a zip/pptx (first bytes: %q)", out[:min(len(out), 4)])
	}
}
