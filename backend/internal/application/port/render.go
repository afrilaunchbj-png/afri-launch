package port

import "context"

// Renderer transforme du HTML (auto-porteur) en PDF ou en PPTX (image-par-slide).
type Renderer interface {
	HTMLToPDF(ctx context.Context, html []byte) ([]byte, error)
	SlidesToPPTX(ctx context.Context, html []byte) ([]byte, error)
	// SlidesToPPTXWithCover assemble le PPTX en plaçant coverPNG en première
	// slide (workflow cover-first).
	SlidesToPPTXWithCover(ctx context.Context, html []byte, coverPNG []byte) ([]byte, error)
}
