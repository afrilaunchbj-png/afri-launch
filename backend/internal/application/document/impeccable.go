// Package document orchestre la génération de documents (ebook/deck) :
// LLM → HTML (guidé par le vocabulaire design « Impeccable ») → chromedp → PDF/PPTX.
package document

// impeccableGuidelines est une transcription condensée du skill open-source
// « Impeccable » (github.com/pbakaus/impeccable), injectée dans le prompt du
// LLM pour éviter le « AI slop » et respecter le design system.
const impeccableGuidelines = `
# Design guidelines (Impeccable — craft floor)

You are producing a self-contained HTML document. Follow these rules strictly.

## Respect the design system (never invent a new palette)
- Colors (light, sober, high contrast): deep emerald primary #003527, warm amber
  accent #855300 / #FEA619, background #F8F9FA, ink #191C1D, muted #404944.
  Success emerald #059669, error #BA1A1A. Support dark mode is NOT required here.
- Typography: titles in Lexend, body in Inter. Clear hierarchy by weight, not color.
- 8px spacing rhythm. Sharp, consistent alignment. 8px radius (cards 16px).

## Avoid AI "slop" (absolute bans)
- No "AI beige" gradient blobs, no italic-serif display headlines, no nested
  cards inside cards, no pulsing dots, no generic drop-shadows, no
  "hero eyebrow chip" labels, no icon-tile stacks, no numbered section labels.
- No vague headlines. No generic CTAs. No lorem ipsum. Every sentence earns its place.

## Typography & layout
- Generous whitespace. Short paragraphs. Meaningful headings. No orphan labels.
- Contrast WCAG AA. Body 16px minimum for readability.

## Mode: READ (ebook / guide)
- Structure for comprehension: title page, clear chapters, short sections,
  checklist, conclusion. The reading experience must feel crafted, not generated.
`
