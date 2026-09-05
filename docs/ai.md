# Architecture IA — AfriLaunch

> Comment AfriLaunch utilise l'IA générative pour transformer une opportunité en produit digital vendable. Ce document précise `docs/architecture.md` §5/§6.

## 1. Principes

1. **Provider-swappable** : tout accès à un modèle IA passe par une **interface Go** (`LLMProvider`, `ImageProvider`, `VideoProvider`). Le modèle/provider n'est jamais hardcodé (règle master.md §14).
2. **Asynchrone** : toute génération longue passe par **asynq** (jamais dans une requête HTTP).
3. **Observable** : chaque génération consigne `provider, model, tokens, latency, cost, status`.
4. **Crédits** : cycle idempotent `reserve → generate → consume|release` (voir `docs/architecture.md` §8).
5. **Jamais de statistique inventée** : le contenu généré ne doit jamais affirmer une donnée sans source (classification `VERIFIED/ESTIMATED/INFERRED/HYPOTHESIS`).

## 2. Providers

| Tâche | Provider | Modèle(s) |
|---|---|---|
| Recherche de marché / scoring | OpenAI | `gpt-5.6-terra` |
| Contenu long (ebook, chapitres) | OpenAI | `gpt-5.6-terra` |
| Idéation / concepts produits | OpenAI | `gpt-5.6-luna` |
| Images (couvertures, illustrations, visuels) | OpenAI | `gpt-image-2` |
| Vidéo (avatar) | HeyGen | avatar prédéfini |

**Configuration** (env vars, jamais de secret commité) :

| Variable | Rôle |
|---|---|
| `OPENAI_API_KEY` | clé OpenAI |
| `OPENAI_MODEL_RESEARCH` | modèle recherche/contenu long (`gpt-5.6-terra`) |
| `OPENAI_MODEL_IDEATION` | modèle idéation (`gpt-5.6-luna`) |
| `OPENAI_MODEL_IMAGE` | modèle image (`gpt-image-2`) |
| `HEYGEN_API_KEY` | clé HeyGen |
| `HEYGEN_API_URL` | endpoint HeyGen (défaut API publique) |

*(remplace l'ancienne variable unique `AI_PROVIDER_API_KEY` — ADR-012.)*

## 3. Abstractions (ports Go)

```go
// application/port/ai.go (cible)
type LLMProvider interface {
    Complete(ctx context.Context, req LLMRequest) (LLMResponse, error)
}
type ImageProvider interface {
    Generate(ctx context.Context, req ImageRequest) (Image, error)
}
type VideoProvider interface {
    Submit(ctx context.Context, req VideoRequest) (VideoJobID, error)   // async
    Status(ctx context.Context, id VideoJobID) (VideoStatus, error)     // polling/webhook
}
```

- `ModelRouter` : mappe une **tâche** (recherche, idéation, contenu, image…) vers un `(provider, modèle)`.
- `infra/ai/openai.go` : implémentation OpenAI (chat completions + images).
- `infra/ai/heygen.go` : implémentation HeyGen (soumission + suivi du job vidéo).

## 4. Pipeline document (ebook / deck) — HTML → PDF/PPT

**Décision (ADR-013)** : le LLM ne génère **pas** les binaires PDF/PPTX (peu fiables). Il génère un **HTML auto-porteur** (guidé par le design system + le vocabulaire Impeccable injectés dans le prompt), puis **chromedp** (headless Chrome, 100 % Go) le rend et l'exporte.

```
LLM (HTML auto-porteur)  →  chromedp (render)  →  export
                                                  ├─ PDF (ebook) : page.PrintToPDF + @page CSS
                                                  └─ PPTX (deck) : screenshot PNG/slide → assemblage image-par-slide
```

- **Ebook** : une page HTML fluide + `@page { size: A4 }` → `page.PrintToPDF` (chromedp).
- **Deck** : chaque slide = `<section class="slide">` (1280×720) → screenshot PNG → **PPTX image-par-slide** (1 image pleine page / slide, `infra/pptx`).
- **Implémentation** : `application/document` (service + prompts) · `application/port/render.go` (`Renderer`) · `infra/render` (chromedp) · `infra/pptx` (assemblage PPTX). Chrome requis (chromium installé dans l'image Docker).

## 5. Skill « Impeccable » (qualité visuelle)

Pour éviter le « AI slop », on injecte le **vocabulaire design** du skill open-source [Impeccable](https://impeccable.style/) ([github.com/pbakaus/impeccable](https://github.com/pbakaus/impeccable)) dans le **system prompt** du worker de génération HTML :

- **Source** : `skill/SKILL.src.md` + `skill/reference/craft-floor.md` + les règles « slop » (anti-patterns IA : *AI beige*, *italic serif*, *nested cards*, *pulsing dot*, …).
- **Modes** (choix selon la surface) : `Read` (ebook/guide), `Persuade` (page de vente/deck), `Operate`, `Experience`.
- **Intégration** : le contenu du skill est **récupéré et transcris** en instructions de prompt embarquées dans le worker (pas un skill `.opencode` au runtime). Le design system existant (`DESIGN.md` « Emerald & Amber Ledger ») prime toujours sur les goûts du modèle.

## 6. Vidéo publicitaire (HeyGen + FFmpeg) — async (ADR-016)

- Pipeline (job `video_ad`, 15 crédits) : **analyse marketing LLM** (angle, pain points, hook, CTA) → **script + storyboard LLM** (JSON : scènes avatar/product/text, text overlays) → **vidéo avatar** (`VideoProvider` = HeyGen, poll 5 s) → **montage FFmpeg** (`port.VideoRenderer` → `infra/render/ffmpeg.go`) → assets `video_ad` (MP4) + `video_ad_thumb` (JPEG).
- Montage : carte d'intro (mockup cover + hook) + vidéo provider (sous-titres burnés, timing recalé sur la durée ffprobe) + carte d'outro (CTA) — cartes rendues en HTML via chromedp (identité Emerald & Amber). Presets 9:16 (1080x1920), 1:1, 16:9.
- Progression détaillée : events SSE `job.updated` avec champ `stage` (`analyzing` → `storyboarding` → `generating_avatar` → `rendering`).
- Config : `HEYGEN_DEFAULT_AVATAR_ID` / `HEYGEN_DEFAULT_VOICE_ID` (env, overridables par génération), `FFMPEG_PATH` ; storyboard persisté dans `generation_jobs.result` (JSONB). Post-MVP : variantes multiples, B-roll (Veo/Kling → nouveau port), scoring, musique.

## 7. Observabilité & coût

- Chaque `GenerationJob` consigne : `provider`, `model`, `input/output tokens`, `latency`, `estimated cost`, `status`.
- Le coût estimé est comparé à `generation_costs` (déjà seedé : niche_research 5, idea_generation 2, ebook_generation 20, image_generation 3, video_generation 15, sales_page 5).
- Les coûts réels (API providers) sont tracés pour contrôler la marge (master.md §26).

## 8. Workers (asynq)

```
Research → LLM (content) → Image → Video (HeyGen) → Render (chromedp) → QC
```

Chaque worker est idempotent/retryable ; `GenerationJob` porte `id, status, progress, attempts, error, provider, cost, metadata`.

## 9. Stockage des fichiers générés

- **Neon Object Storage** (S3-compatible, SDK AWS v2) : HTML source + PDF/PPTX/vidéos exportés.
- `infra/storage/s3.go` (`S3Storage`) est actif dès que `S3_BUCKET` est défini (sinon fallback `LocalStorage` en dev) ; config `S3_ENDPOINT/S3_REGION/S3_ACCESS_KEY_ID/S3_SECRET_ACCESS_KEY/S3_BUCKET/S3_PATH_STYLE`.
- Bucket privé ; les downloads passent par le backend (`GET /assets/{id}/download`), URLs présignées envisagées plus tard.
