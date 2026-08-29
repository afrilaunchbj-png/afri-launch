package port

import "context"

// Renderer transforme du HTML (auto-porteur) en PDF ou en PPTX (image-par-slide).
type Renderer interface {
	HTMLToPDF(ctx context.Context, html []byte) ([]byte, error)
	SlidesToPPTX(ctx context.Context, html []byte) ([]byte, error)
}
