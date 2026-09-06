// Package render fournit le rendu HTML → PDF/PNG via chromedp (Chrome headless).
package render

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"afrilaunch/backend/internal/infra/pptx"
)

// waitImagesExpr résout une promesse dès que le document est complet et que
// toutes les images (dont les covers data: URL injectées) sont décodées.
const waitImagesExpr = `new Promise((resolve) => {
  const done = () => {
    if (document.readyState === 'complete' && Array.from(document.images).every(i => i.complete)) {
      return resolve(true);
    }
    setTimeout(done, 60);
  };
  setTimeout(done, 8000); // garde-fou : on imprime même si une image tarde
  done();
})`

// tempHTMLURL écrit le HTML dans un fichier temporaire et renvoie son URL
// file:// — les data: URLs sont limitées (~2 Mo) et échouent (ERR_ABORTED)
// dès que le document embarque la cover en base64.
func tempHTMLURL(html []byte) (string, func(), error) {
	f, err := os.CreateTemp("", "afrilaunch-*.html")
	if err != nil {
		return "", nil, err
	}
	clean := func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}
	if _, err := f.Write(html); err != nil {
		clean()
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		clean()
		return "", nil, err
	}
	return "file://" + filepath.ToSlash(f.Name()), clean, nil
}

// ChromedpRenderer implémente port.Renderer avec Chrome headless.
type ChromedpRenderer struct {
	chromePath string
}

// NewChromedpRenderer construit un renderer chromedp.
// chromePath peut être vide (détection auto via DefaultExecAllocatorOptions).
func NewChromedpRenderer(chromePath string) *ChromedpRenderer {
	return &ChromedpRenderer{chromePath: chromePath}
}

func (r *ChromedpRenderer) allocatorOptions() []chromedp.ExecAllocatorOption {
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	opts = append(opts, chromedp.Flag("no-sandbox", true), chromedp.Flag("disable-gpu", true))
	if r.chromePath != "" {
		opts = append(opts, chromedp.ExecPath(r.chromePath))
	}
	return opts
}

// HTMLToPDF rend un document HTML en PDF (respecte @page CSS).
func (r *ChromedpRenderer) HTMLToPDF(ctx context.Context, html []byte) ([]byte, error) {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, r.allocatorOptions()...)
	defer cancelAlloc()
	cctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	url, clean, err := tempHTMLURL(html)
	if err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	defer clean()

	var pdf []byte
	err = chromedp.Run(cctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Attendre que la page et les images (cover, visuels) soient prêtes
			// avant l'impression, sinon la première page peut sortir vide.
			var ok bool
			_ = chromedp.Evaluate(waitImagesExpr, &ok, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
				return p.WithAwaitPromise(true)
			}).Do(ctx)
			// Petite marge de sécurité de mise en page (polices @page).
			select {
			case <-time.After(150 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
			var err error
			pdf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	return pdf, nil
}

// SlidesToPNG capture chaque élément <section class="slide"> en PNG.
func (r *ChromedpRenderer) SlidesToPNG(ctx context.Context, html []byte) ([][]byte, error) {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, r.allocatorOptions()...)
	defer cancelAlloc()
	cctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	url, clean, err := tempHTMLURL(html)
	if err != nil {
		return nil, fmt.Errorf("render slides: %w", err)
	}
	defer clean()

	if err := chromedp.Run(cctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("section.slide", chromedp.ByQuery),
	); err != nil {
		return nil, fmt.Errorf("render slides: %w", err)
	}

	var count int
	if err := chromedp.Run(cctx, chromedp.Evaluate(`document.querySelectorAll('section.slide').length`, &count)); err != nil {
		return nil, err
	}

	images := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		var buf []byte
		sel := fmt.Sprintf("section.slide:nth-of-type(%d)", i+1)
		if err := chromedp.Run(cctx, chromedp.Screenshot(sel, &buf, chromedp.ByQuery)); err != nil {
			return nil, fmt.Errorf("screenshot slide %d: %w", i+1, err)
		}
		images = append(images, buf)
	}
	return images, nil
}

// SlidesToPPTX rend les slides en PNG puis assemble un PPTX image-par-slide.
func (r *ChromedpRenderer) SlidesToPPTX(ctx context.Context, html []byte) ([]byte, error) {
	pngs, err := r.SlidesToPNG(ctx, html)
	if err != nil {
		return nil, err
	}
	return pptx.Build(pngs, 12192000, 6858000) // 16:9 en EMU
}

// SlidesToPPTXWithCover assemble le PPTX avec coverPNG en première slide
// (workflow cover-first).
func (r *ChromedpRenderer) SlidesToPPTXWithCover(ctx context.Context, html []byte, coverPNG []byte) ([]byte, error) {
	pngs, err := r.SlidesToPNG(ctx, html)
	if err != nil {
		return nil, err
	}
	if coverPNG != nil {
		pngs = append([][]byte{coverPNG}, pngs...)
	}
	return pptx.Build(pngs, 12192000, 6858000)
}

// HTMLToPNG rend un HTML (canvas exact width×height px) en PNG — utilisé
// pour les cartes de montage vidéo (intro/outro).
func (r *ChromedpRenderer) HTMLToPNG(ctx context.Context, html []byte, width, height int) ([]byte, error) {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, r.allocatorOptions()...)
	defer cancelAlloc()
	cctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	url, clean, err := tempHTMLURL(html)
	if err != nil {
		return nil, fmt.Errorf("render png: %w", err)
	}
	defer clean()

	var buf []byte
	err = chromedp.Run(cctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDeviceMetricsOverride(int64(width), int64(height), 1, false).Do(ctx)
		}),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Screenshot("body", &buf, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("render png: %w", err)
	}
	return buf, nil
}

func dataURL(html []byte) string {
	return "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString(html)
}
