// Package render fournit le rendu HTML → PDF/PNG via chromedp (Chrome headless).
package render

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"afrilaunch/backend/internal/infra/pptx"
)

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

	var pdf []byte
	err := chromedp.Run(cctx,
		chromedp.Navigate(dataURL(html)),
		chromedp.ActionFunc(func(ctx context.Context) error {
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

	if err := chromedp.Run(cctx,
		chromedp.Navigate(dataURL(html)),
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

func dataURL(html []byte) string {
	return "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString(html)
}
