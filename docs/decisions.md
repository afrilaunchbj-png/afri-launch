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
