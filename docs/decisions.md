# Architecture Decisions (ADR) — AfriLaunch

> Décisions clés avec contexte, options, décision et conséquences.

## ADR-001 — Stack backend : Go + PostgreSQL + Redis (pas Next.js/Prisma)

- **Date** : 2026-08-29
- **Contexte** : `prompts/master.md` recommande Next.js/TypeScript/Prisma ; `prompts/init.md` impose Go/PostgreSQL/Redis backend et interdit Next.js. `init.md` est prioritaire (règle §15).
- **Options** : (A) Go + PostgreSQL + Redis + React SPA ; (B) Next.js + Prisma (master.md) ; (C) Node.js séparé.
- **Décision** : **(A)** — respect de la contrainte prioritaire.
- **Conséquences** : exemples TS de master.md (interfaces providers) transposés en interfaces Go ; BullMQ remplacé par asynq ; Prisma remplacé par sqlc + goose.

## ADR-002 — Accès aux données : sqlc + pgx/v5 + goose

- **Date** : 2026-08-29
- **Contexte** : besoin de requêtes typées, explicites et optimisables (index, transactions) pour un ledger de crédits et des workflows financiers.
- **Options** : (A) sqlc + pgx/v5 + goose ; (B) GORM ; (C) pgx seul (SQL manuel).
- **Décision** : **(A)** (validé par l'utilisateur) — SQL contrôlé, code généré type-safe, migrations versionnées.
- **Conséquences** : SQL source de vérité dans `db/query/` et `db/migrations/` ; génération via `sqlc generate` ; courbe d'apprentissage légère.

## ADR-003 — Monorepo : `backend/` + `frontend/` séparés

- **Date** : 2026-08-29
- **Contexte** : deux stacks hétérogènes (Go / Node).
- **Options** : (A) monorepo workspaces JS + module Go ; (B) deux dossiers séparés.
- **Décision** : **(B)** (validé par l'utilisateur) — simplicité, pas de couplage outillage.
- **Conséquences** : pas de packages partagés au sens npm ; les contrats API partagés passent par OpenAPI (source de vérité unique) et des DTOs dupliqués côté FE typés manuellement si nécessaire.

## ADR-004 — File d'attente : asynq (Redis)

- **Date** : 2026-08-29
- **Contexte** : générations longues (LLM, images, PDF, vidéo) à exécuter hors requête HTTP.
- **Options** : (A) asynq (Go, Redis) ; (B) RabbitMQ/Kafka ; (C) HTTP polling simple.
- **Décision** : **(A)** — idiomatique Go, réutilise Redis déjà dans la stack, retry/deadline natifs.
- **Conséquences** : workers = binaires Go déployés séparément ; jobs idempotents.

## ADR-005 — Authentification : JWT (cookies httpOnly) + Argon2id + Google OAuth

- **Date** : 2026-08-29
- **Contexte** : sécurité (OWASP) + besoin d'OAuth pour l'onboarding.
- **Options** : (A) JWT access/refresh en cookies httpOnly ; (B) sessions serveur ; (C) tokens en localStorage.
- **Décision** : **(A)** — court-lived access + refresh avec rotation, cookies httpOnly/Secure/SameSite, hachage Argon2id, Google OIDC + PKCE.
- **Conséquences** : middleware auth + refresh endpoint ; pas de stockage token côté JS.

## ADR-006 — Erreurs API : RFC 9457 Problem Details

- **Date** : 2026-08-29
- **Contexte** : cohérence et non-exposition des détails internes.
- **Décision** : format unique `{type, title, status, detail, instance, errors[]}` sur toutes les erreurs.
- **Conséquences** : type `APIError` Go + conversion centralisée ; mapping FE en `AppError`.

## ADR-007 — Crédits : ledger double-entrée idempotent

- **Date** : 2026-08-29
- **Contexte** : modèle pay-as-you-go ; éviter double consommation.
- **Décision** : `CreditAccount` + `CreditTransaction` + `CreditReservation` + `GenerationCost`, cycle `reserve → execute → consume|release`, idempotence par `Idempotency-Key` + réservation.
- **Conséquences** : coûts configurables par opération ; refunds en cas d'échec.

## ADR-008 — Paiements : abstraction PaymentProvider (Mobile Money d'abord)

- **Date** : 2026-08-29
- **Contexte** : marchés africains = Mobile Money (Wave, Orange, MTN) ; ne pas se coupler à Stripe.
- **Décision** : interface Go `PaymentProvider` (create/verify/refund), implémentations par provider, webhooks protégés + idempotents.
- **Conséquences** : ajout d'un provider = nouvelle implémentation, sans toucher au cœur.

## ADR-009 — Multi-tenancy : row-level (isolation par user_id)

- **Date** : 2026-08-29
- **Contexte** : un utilisateur ne doit jamais accéder aux données d'un autre.
- **Décision** : row-level (`user_id` sur les tables métier, filtre systématique), RLS PostgreSQL en filet de sécurité.
- **Conséquences** : couche data-access avec filtre tenant ; `Organization` modélisée pour le futur B2B2C.

## ADR-010 — Design system : « Emerald & Amber Ledger » → tokens shadcn

- **Date** : 2026-08-29
- **Contexte** : maquettes Stitch imposent une palette (émeraude/ambre) et typo Lexend/Inter.
- **Décision** : mapper les tokens (`primary #003527`, `secondary #855300`, …) en variables CSS shadcn (`--primary`, `--secondary`, …) + Tailwind `class` pour le dark mode.
- **Conséquences** : `frontend/src/styles/globals.css` définit les tokens ; composants shadcn héritent du thème.

## ADR-011 — Authentification : Neon Auth (Managed Better Auth) au lieu de l'auth maison

- **Date** : 2026-08-29
- **Contexte** : l'utilisateur active Neon Auth sur son projet Neon et veut Google comme provider OAuth, avec Neon pour Postgres + Object Storage. L'ADR-005 (JWT maison HS256 + argon2 + refresh tokens) est remplacé.
- **Options** : (A) Neon Auth managé (Better Auth) ; (B) garder l'auth maison (email/mot de passe) + Google OAuth PKCE.
- **Décision** : **(A)** — identité/sessions gérées par Neon (schéma `neon_auth`), Google en provider OAuth ; le backend **vérifie les JWT** (EdDSA/Ed25519, 15 min) via le JWKS `<NEON_AUTH_URL>/.well-known/jwks.json` (issuer/audience = origine) ; `users` devient une table profil cléée par `sub`, avec bonus de bienvenue au 1er login. Frontend : `@neondatabase/neon-js` + `@neondatabase/auth-ui`.
- **Conséquences** : suppression d'argon2, des refresh tokens, du `JWTManager` HS256 et du Google OAuth PKCE maison ; les JWT partent en `Authorization: Bearer` vers l'API Go (pas de cookies de session cross-origin) ; dépendances beta (`neon-js`, `auth-ui`).

## ADR-012 — IA : deux providers (OpenAI GPT + HeyGen) + routage de modèle

- **Date** : 2026-08-29
- **Contexte** : le backend n'avait qu'une variable `AI_PROVIDER_API_KEY` ; les besoins sont hétérogènes (texte/images/documents vs vidéo avatar).
- **Options** : (A) un provider unique ; (B) OpenAI pour recherche/images/documents + HeyGen pour la vidéo.
- **Décision** : **(B)** — **OpenAI** (famille GPT) pour recherche, images et documents ; **HeyGen** pour la vidéo (avatar). Routage de modèle par tâche (`ModelRouter`) : `gpt-5.6-terra` (recherche & contenu long), `gpt-5.6-luna` (idéation), `gpt-image-2` (images). Config multi-clés (`OPENAI_API_KEY`, `HEYGEN_API_KEY`) + modèles par tâche.
- **Conséquences** : ports `LLMProvider`/`ImageProvider`/`VideoProvider` ; `infra/ai/openai.go` + `infra/ai/heygen.go` ; la vidéo MVP = avatar seul (script + montage en V2).

## ADR-013 — Documents : HTML rendu en PDF/PPTX via chromedp (image-par-slide)

- **Date** : 2026-08-29
- **Contexte** : les LLM produisent du HTML/CSS fiable mais pas des binaires PPTX/PDF. Le rendu Playwright est l'alternative Node, alors que le backend est en Go.
- **Options** : (A) Playwright (worker Node) ; (B) **chromedp** (headless Chrome, 100 % Go) ; (C) `playwright-go` (binding communautaire).
- **Décision** : **(B)** — le LLM génère du **contenu structuré (JSON)** injecté dans des **templates HTML** (thème « Emerald & Amber Ledger »), rendus par **chromedp** : `page.PrintToPDF` pour l'ebook, **slides 16:9 → PNG → PPTX image-par-slide** pour le deck (pas d'export PPTX natif).
- **Conséquences** : pas de runtime Node dans la couche workers ; PPTX = images pleine page (non éditable, acceptable pour du marketing) ; le vocabulaire du skill **Impeccable** (github.com/pbakaus/impeccable) est injecté dans le system prompt pour éviter le « AI slop ».

## ADR-014 — Découverte conversationnelle : chat copilote + canal SSE unique

- **Date** : 2026-09-03
- **Contexte** : le parcours statique (catalogue d'opportunités → génération d'idées en batch → chat inline par idée) est rigide ; un chat conversationnel est plus naturel pour explorer un marché et affiner une idée.
- **Options** : (A) SSE par conversation (stream dans la requête POST, comme l'ancien `/ideas/{id}/messages`) ; (B) **canal SSE unique par utilisateur** (`GET /api/v1/events`) + POST de message en 202.
- **Décision** : **(B)** — un seul tunnel server→client par utilisateur, événements typés (`chat.started|delta|tool|completed|error`, `job.updated`, événements futurs) routés par type + payload. Broker in-process derrière le port `EventBus` (buffer 128/abonné, déconnexion du client lent = resync à la reconnexion) ; migration Redis pub/sub prévue avec asynq. Le tour du copilote est une **boucle agent bornée** (max 2 rounds, 1 recherche/tour) avec outils en marqueurs texte streamés : `@@SEARCH {json}` (recherche web facturée, opportunités créées) et bloc `@@IDEAS … @@END` (idées `product_ideas` créées, retirées du texte visible). La validation d'idée reste un **geste utilisateur explicite** (`POST /ideas/{id}/confirm`) ; « Transformer en projet » réutilise `POST /projects` tel quel.
- **Conséquences** : pages `/opportunities` et `/ideas` supprimées (redirections → `/discover`) ; le polling des jobs devient un filet de sécurité (les events `job.updated` rafraîchissent le cache) ; endpoints job-batch ideas/research legacy conservés mais non appelés par le FE ; chat gratuit, facturation aux actions lourdes (recherche 5 cr, idées 2 cr) ; messages de chat persistés (`conversations`, `conversation_messages`) → perte de connexion sans perte de données.

## ADR-015 — Assets : workflow cover-first + identité visuelle par projet

- **Date** : 2026-09-03
- **Contexte** : les assets étaient générés avec la palette de l'application (emerald/amber codée en dur), sans configuration ni ordre. Le PDF ne contenait pas la cover générée.
- **Options** : (A) statu quo (palette app, assets indépendants) ; (B) **workflow cover-first** : l'IA propose une identité visuelle par projet, l'utilisateur l'ajuste et valide la cover, les autres assets en héritent.
- **Décision** : **(B)** — `projects.config` (JSONB) stocke `palette` (5 couleurs + `source: ai|user`), `style` et `ebook_min_pages`/`ebook_max_pages` (défauts 6–14, clamp 2–40). La génération de cover : si aucune palette → appel LLM (JSON) qui propose palette + style depuis le contexte projet → persisté (`source: ai`) → image `gpt-image-2` avec ces couleurs. La cover est **injectée en première page du PDF** (`PrependCoverPage`, `@page :first` sans marge, image data-URI) et **première slide du deck** (`SlidesToPPTXWithCover`). Gate `ErrCoverRequired` (422) sur ebook/affiches/page de vente. Régénérations = nouveaux jobs = consommation de crédits (`generation_jobs.params` porte les instructions utilisateur).
- **Conséquences** : la palette de l'app n'est qu'un fallback ; tous les prompts d'assets acceptent la palette du projet ; chaque régénération est comptée dans le ledger ; la page projet devient un stepper (cover → ebook → affiches → page de vente).

## ADR-016 — Vidéos publicitaires : pipeline LLM → HeyGen → FFmpeg, stockage S3 prérequis

- **Date** : 2026-09-05
- **Contexte** : nouvelle feature « vidéo publicitaire » (cf. `prompts/video-flow.md`) : transformer une idée confirmée en vidéo pub short-form (TikTok/Reels 9:16). Le repo dispose déjà des ports `LLMProvider`/`ImageProvider`/`VideoProvider` (HeyGen), du worker in-process, du ledger de crédits (`video_generation` = 15 cr) et du canal SSE.
- **Options** : (A) module vidéo dédié avec ses propres tables + hiérarchie de providers ; (B) **réutiliser l'architecture existante** : job `video_ad` sur le worker, storyboard en JSONB, HeyGen via `VideoProvider`, montage via un nouveau port `VideoRenderer` (FFmpeg).
- **Décision** : **(B)** — MVP = 1 variante, 1 scène avatar HeyGen (9:16, 15/30 s), puis montage FFmpeg : sous-titres SRT burnés + overlay cover/texte. Storyboard (scènes, hooks, text overlays) persisté dans `generation_jobs.params`/`result` (JSONB) — tables dédiées (`ad_scenes`…) différées à l'étape « regénération scène par scène ». Avatars/voix HeyGen configurés par env (`HEYGEN_DEFAULT_AVATAR_ID`, `HEYGEN_DEFAULT_VOICE_ID`) + override par génération — pas de hard-code, catalogue API HeyGen post-MVP. **L'adaptateur S3 Neon est un prérequis** : Railway est éphémère et les vidéos sont lourdes — `LocalStorage` remplacé par `S3Storage` quand `S3_BUCKET` est défini (fallback local en dev). Scripts générés dans la langue du compte/opportunité (pattern `chatSystemPrompt(language)`).
- **Conséquences** : migration `00017` (kinds `video_ad` sur jobs/assets) ; nouveau port `VideoRenderer` + `infra/render/ffmpeg.go` (ffmpeg ajouté au Dockerfile) ; download immédiat de l'URL HeyGen (expirable) ; progression multi-étapes via `job.updated` avec champ `stage` ; post-MVP : multi-variantes, B-roll (Veo/Kling → nouveau port), VoiceProvider TTS, scoring LLM, musique libre, regénération scène par scène.

## ADR-017 — Intégrations publicitaires : couche multi-tenant provider-agnostic (Meta d'abord)

- **Date** : 2026-09-05
- **Contexte** : les vidéos générées (ADR-016) doivent pouvoir être publiées en campagnes. Chaque utilisateur connecte ses propres comptes publicitaires (Meta au MVP, Google/TikTok ensuite) ; les tokens OAuth sont des secrets qui ne doivent jamais être exposés ni en clair, ni au frontend, ni dans les logs.
- **Options** : (A) appels Meta directs dans les handlers/le domaine ; (B) **couche d'intégration dédiée** : port `AdPlatformProvider` (+ `Capabilities`) implémenté par plateforme, service applicatif générique, tokens chiffrés AES-256-GCM au repos, états OAuth anti-CSRF/anti-replay en DB.
- **Décision** : **(B)** — `infra/ads/meta` encapsule OAuth (dialogue → code → token courte durée → prolongation `fb_exchange_token` ~60 j, pas de refresh token chez Meta), discovery (`/me/adaccounts`), campagnes (créées **toujours en PAUSED** — garde-fou §32, publication = geste utilisateur), creatives vidéo (`advideos` via `file_url` signée S3 + `adcreatives` avec `page_id` stocké dans les métadonnées de connexion), insights (time_increment=1 j, montants en unités mineures). Tables `00017` : `ad_platform_connections` (1 par user+provider, tokens chiffrés), `oauth_states` (usage unique), `ad_campaigns` (UUID interne + mapping `external_campaign_id`), `ad_creatives` (liées aux `assets` internes), `ad_insights` (upsert campagne+jour), `provider_operations` (traçabilité/reconciliation). Montants toujours en **entiers unités mineures** (jamais de float). Le callback OAuth (`/integrations/{provider}/callback`) est **hors groupe JWT** : l'utilisateur est identifié par l'état consommé puis redirection FE sans token. Providers activés par env (`META_APP_ID/SECRET`...) et résolus par registry — aucun `if provider === "META"` dans le métier.
- **Conséquences** : `ENCRYPTION_KEY` (≥ 32 car., AES-256-GCM dérivé SHA-256, préfixe `v1:` pour rotation) ; `APP_URL` pour les redirections ; creative upload requiert le stockage S3 (URLs présignées) — indisponible en dev local ; campagnes créées synchrones via `provider_operations` (worker asynq = évolution) ; Google Ads (gRPC, developer token) et TikTok Ads suivent le même contrat ; garde-fous budget (`MaxDailySpendMinor`, `MaxCampaigns`) côté service ; audit via `audit_logs`.
