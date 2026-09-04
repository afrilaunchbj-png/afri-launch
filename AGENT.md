# AGENT.md

> Mémoire de travail persistante de l'agent. **À lire impérativement avant chaque nouvelle tâche**, puis à mettre à jour après chaque travail significatif.

## Project Overview

**AfriLaunch** est un SaaS B2C/B2B2C qui permet à un créateur, entrepreneur, consultant ou expert **africain** de transformer une opportunité de marché en **produit digital commercialisable** (ebook/guide + assets marketing + page de vente) avec un minimum de travail manuel.

Workflow cœur : Recherche de marché → Identification d'opportunités → Sélection de niche → Génération d'idées → Itération → Création du produit (ebook ~3000 mots) → Assets marketing → Contrôle qualité → Packaging → Export.

Positionnement : **verticalisation géographique/sectorielle africaine**, **données locales vérifiées** (jamais de statistiques inventées), workflow end-to-end **mobile-first** + paiements locaux (Mobile Money).

## Current Status

**Phase 2 — Fondations + premières features. Découverte conversationnelle (chat copilote) en place.** L'identité est gérée par **Neon Auth (Managed Better Auth)**, l'infra cible est **Neon** (Postgres + Object Storage) et Redis (Upstash).

Implémenté et validé :
- Migration complète + pool pgx (PostgreSQL).
- **Auth via Neon Auth** : le backend vérifie les JWT EdDSA (JWKS) ; plus de mot de passe/argon2/refresh tokens (voir ADR-011).
- Ledger de crédits idempotent (`reserve → consume|release`, bonus de bienvenue au 1er login).
- **Préférences utilisateur en DB** (`user_preferences` : langue, thème) — chargées par le FE au login (`PreferencesSync`), mises à jour par les toggles langue/thème avec optimisme ; le localStorage ne sert qu'avant login.
- **Copilote multilingue** : la langue du compte est injectée dans le prompt système (réponses + champs d'idées) et dans les requêtes de recherche.
- Recherche d'opportunités (catalogue + filtres + sauvegarde) — catalogue conservé comme données/contexte du chat.
- Frontend : layout applicatif, patterns formulaire/liste, pages login/register (UI Better Auth), dashboard, projets, crédits ; i18n FR/EN ; dark mode.
- **Chat copilote (ADR-014)** : parcours opportunités→idées remplacé par une conversation (`/discover`) ; canal SSE unique `GET /api/v1/events` ; boucle agent bornée avec outils `@@SEARCH` (recherche web) et bloc `@@IDEAS` (idées persistées) ; validation d'idée explicite ; « Transformer en projet » inchangé.

## Architecture

```
afri-launch/
├── AGENT.md
├── README.md
├── docs/                    # index, architecture, conventions, decisions, glossary
├── design/                  # maquettes Stitch
├── prompts/                 # master.md + init.md
├── backend/                 # API Go
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── server/          # routeur chi, middlewares, handlers, authctx
│   │   │   └── handler/     # auth, credits, opportunities, markets, health, conversations, events
│   │   ├── domain/          # entités + erreurs métier
│   │   ├── application/     # ports + use cases (auth, credits, opportunities, chat, jobs)
│   │   └── infra/           # postgres (sqlc), auth (neon verifier), events (broker SSE)
│   └── db/
│       ├── migrations/      # goose 00001..00011
│       └── query/           # sqlc -> internal/infra/postgres/db
└── frontend/                # SPA React (Vite + React Router 7)
    └── src/
        ├── app/             # router, root-layout, root-providers
        ├── components/      # ui/ (shadcn), layout/, data-table/, states/
        ├── features/        # auth/, credits/, chat/, projects/, generation/
        ├── i18n/            # locales fr/en
        └── lib/             # api client, events (SSE), auth (Neon Auth), errors, utils
```

## Tech Stack

| Couche | Technologie |
|---|---|
| Backend | **Go 1.23** (chi, pgx/v5, sqlc, goose) |
| Base de données | **PostgreSQL (Neon)** |
| Auth | **Neon Auth (Managed Better Auth)** — JWT EdDSA/JWKS |
| Cache / Queue | **Redis (Upstash)** — asynq à venir |
| Frontend | **React 19 + Vite + React Router 7 (SPA)** |
| Auth FE | `@neondatabase/neon-js` + `@neondatabase/auth-ui` (beta) |
| Styles | Tailwind + shadcn/ui + Radix |
| État serveur | TanStack Query v5 |
| Formulaires | react-hook-form + zod |
| i18n | react-i18next (fr / en) |
| Tests | Go `testing` · (FE Vitest/Playwright à venir) |

**Next.js est strictement interdit.**

## Important Decisions

Voir `docs/decisions.md` (ADR). Synthèse + ajouts récents :

1. **Stack** Go/PostgreSQL/Redis + React SPA (pas Next.js/Prisma).
2. **Monorepo** `backend/` + `frontend/`.
3. **sqlc + pgx/v5 + goose**.
4. **asynq** pour les générations longues (à venir).
5. ~~JWT maison (HS256 + argon2 + refresh)~~ → **ADR-011 : Neon Auth managé**.
6. **Erreurs RFC 9457**.
7. **Crédits** : ledger double-entrée idempotent.
8. **Paiements** : abstraction `PaymentProvider` (à venir), Mobile Money d'abord.
9. **Multi-tenancy** : isolation par `user_id`.
10. **Design system** « Emerald & Amber Ledger ».
11. **Auth = Neon Auth** (voir ADR-011) : identité/sessions dans `neon_auth` (Better Auth), Google en provider OAuth ; backend vérifie les JWT (EdDSA) via JWKS ; `users` = table profil cléée par `sub` ; bonus de bienvenue au 1er login.
12. **Object Storage Neon S3** : S3-compatible, SDK AWS standard — config prête, adaptateur à câbler avec la feature « assets ».
13. **IA = OpenAI (GPT) + HeyGen** (ADR-012) : recherche/images/documents via OpenAI, vidéo avatar via HeyGen ; `ModelRouter` (`gpt-5.6-terra`, `gpt-5.6-luna`, `gpt-image-2`).
14. **Documents = HTML → PDF/PPTX via chromedp** (ADR-013) : contenu structuré (JSON) → templates HTML (thème « Emerald & Amber Ledger » + vocabulaire [Impeccable](https://impeccable.style/)) → `chromedp` → PDF (ebook) ou PPTX image-par-slide (deck).
15. **Découverte conversationnelle + canal SSE unique** (ADR-014) : un flux SSE par utilisateur (`GET /api/v1/events`, événements typés `chat.*`/`job.updated`), POST de message en 202, boucle agent bornée (outils `@@SEARCH` / bloc `@@IDEAS`), broker in-process (`EventBus`) → Redis pub/sub avec asynq. Chat gratuit, facturation aux actions lourdes (recherche 5 cr, idées 2 cr).

## Design & UI Conventions

- Design system **« Emerald & Amber Ledger »** : primary `#003527`, secondary `#855300`/`#fea619`, Lexend/Inter.
- Mobile-first, touch ≥ 48px, dark mode (`class`), i18n FR/EN, zéro texte hardcodé.
- shadcn/ui + Radix uniquement ; icônes lucide-react.
- Pattern formulaire : zod + react-hook-form + shadcn `<Form>` (utilisé pour les formulaires internes ; le login/register passe par l'UI Better Auth).
- Pattern liste : TanStack Query + pagination serveur + filtres dans l'URL + états loading/empty/error + `DataTable` (@tanstack/react-table v8).

## Completed Work

- [x] Analyse, docs, squelettes (Phase 1).
- [x] Migrations 00001..00007 + pool + sqlc.
- [x] Ledger de crédits + opportunités + marché (backend).
- [x] Frontend complet (layout, patterns, pages, i18n, dark mode).
- [x] **Migration auth → Neon Auth** : vérifieur JWT EdDSA/JWKS (Go) + client `@neondatabase/neon-js`/`auth-ui` (FE), suppression argon2/refresh/Google PKCE maison.
- [x] Test unitaire du vérifieur Neon (`TestNeonVerifier`) + test intégration ledger.
- [x] **Abstractions IA (Go)** : ports `LLMProvider`/`ImageProvider`/`VideoProvider` + `ModelRouter` (`gpt-5.6-terra`/`gpt-5.6-luna`/`gpt-image-2`) + `infra/ai/openai.go` + `infra/ai/heygen.go` + config multi-clés + tests (httptest).
- [x] **Pipeline documents (Go)** : `port.Renderer` + `infra/render` (chromedp : HTML → PDF / slides → PNG) + `infra/pptx` (assemblage PPTX image-par-slide) + `application/document` (service `GenerateEbook`/`GenerateDeck` + prompts avec le vocabulaire Impeccable) + tests (chromedp sur Chrome réel, pptx, prompts).
- [x] **Parcours MVP (idées → projet → assets → download)** : migration `00008` (`product_ideas`, `projects`, `assets`, `generation_jobs`), repos + ports, worker asynchrone in-process (`application/jobs`), services `ideas`/`projects`/`assets`, endpoints (idées, projets, ebook/couverture/page de vente, assets, download), crédits `reserve→consume|release` par génération, frontend (pages idées/projets/projet + téléchargement).
- [x] **Chat copilote (ADR-014)** : migration `00011` (`conversations`, `conversation_messages`, `product_ideas.conversation_id`) ; port `EventBus` + broker in-process (`infra/events`, testé) ; service `application/chat` (boucle agent `@@SEARCH`/`@@IDEAS`, crédits, events) ; handlers conversations + `GET /events` ; worker publie `job.updated` ; FE `features/chat` + page `/discover` + client SSE global (`lib/api/events`) avec reconnexion/backoff ; pages `/opportunities` et `/ideas` supprimées (redirections) ; tests d'intégration des deux chemins (idées, recherche) sur Postgres réel.
- [x] **Préférences utilisateur** : migration `00012` (`user_preferences`) ; service + handlers `GET/PUT /preferences` ; FE `features/preferences` (`PreferencesSync` dans AppLayout, toggles persistants avec mise à jour optimiste) ; langue du compte injectée dans le copilote (`chatSystemPrompt(language)`, `ResearchQuery(language)`).
- [x] **Workflow cover-first + identité visuelle (ADR-015)** : migration `00013` (`projects.config` JSONB, `generation_jobs.params`) ; proposition de palette par l'IA (persistée, source `ai`) ajustable par l'utilisateur (source `user`) ; cover en première page du PDF et première slide du deck ; palette injectée dans ebook/affiches/page de vente ; gate `ErrCoverRequired` ; régénérations consomment des crédits ; page projet en stepper (FE) avec aperçu cover, éditeur de couleurs et plage de pages ebook ; tests unitaires (config, prompts, PrependCoverPage) + intégration (gate, config, instructions).
- [x] **Landing page publique** : hero 2 colonnes avec visuel éditorial, chips des marchés chargées depuis `GET /markets`, parcours en 4 étapes, bandeau marché, 3 cartes de valeur, CTA final avec image, footer — i18n FR/EN, dark mode, mobile-first. **3 visuels générés par gpt-image-2** (`frontend/public/images/landing/*.webp`, 89-202 KB, palette Emerald & Amber, sans texte) + animations soft (`components/reveal.tsx` IntersectionObserver, keyframes CSS, `prefers-reduced-motion` respecté, badge hero retiré — interdit Impeccable).
- [x] **Couverture marchés étendue** : migration `00014` — 24 marchés (Afrique de l'Ouest complète : CEDEAO + Mauritanie ; Afrique centrale : CEMAC + RD Congo), seed idempotent ; landing synchronisée sur l'API (plus de liste en dur).
- [x] **Paramètres, Support et Superadmin** : migration `00015` (`users.role`, `support_tickets`) ; promotion superadmin par `SUPERADMIN_EMAILS` (au 1er login et suivants) ; `POST/GET /support/tickets` ; suivi global `GET /admin/stats|users|tickets` + `POST /admin/tickets/{id}/resolve` derrière `RequireSuperadmin` (rôle relu en DB à chaque appel admin) ; `/auth/me` renvoie le rôle ; FE : pages `/settings` (profil + préférences DB + crédits), `/support` (formulaire + historique), `/admin` (stats + users DataTable + tickets, garde client + serveur) ; nav dynamique selon rôle (sidebar + bottom nav mobile).
- [x] **Déploiement Railway production** : `backend` + `frontend` déployés via `railway up` (auto-deploy GitHub inactif — à activer dans le dashboard) ; `.railwayignore` ajouté ; vérifié (/healthz, /readyz, nouveaux endpoints 401, bundle FE à jour).

## Work In Progress

- Aucune feature en cours.

## Remaining Work

1. **Object Storage** : remplacer `LocalStorage` par un adaptateur S3 (Neon) + URLs présignées.
2. **Paiements Mobile Money** (`PaymentProvider`) + recharges de crédits.
3. **Workers asynq** (Redis) : remplacer le worker in-process (`application/jobs`) par asynq + **Redis pub/sub pour le broker d'événements** (`infra/events`).
4. **Tests** : unitaires + intégration + E2E du parcours complet.
5. **Nettoyage legacy** : endpoints job-batch idées/recherche (`POST /opportunities/{id}/ideas`, `POST /research`) + table `idea_messages` non appelés par le FE depuis le chat (à supprimer lors d'une prochaine passe).
6. **Liste des conversations** dans le chat (API `GET /conversations` prête, UI à ajouter).

## Known Issues

- **⚠️ `backend/.env` + `backend/.env.production` contiennent des secrets (Neon Postgres, Upstash Redis).** Gitignorés (`.env.*`), mais **à rotater** si exposés. En dev, forcer `APP_ENV=development` + `DATABASE_URL` locale.
- **Neon Auth non testé de bout en bout** : nécessite `NEON_AUTH_BASE_URL` (backend) et `VITE_NEON_AUTH_URL` (frontend) réelles + activation auth/Google dans la console Neon. Le vérifieur JWT est testé unitairement (EdDSA/JWKS factices).
- `channel_binding=require` dans l'URL Neon : **validé** (pgx 5.7.4 le supporte, `/readyz` OK sur Neon).
- SDK Neon Auth **en beta** (`neon-js 0.7.0-beta`, `auth-ui 0.3.0-beta`) — API volatile.
- Bundle frontend ~1.43 MB (SDK auth lourd) — code-splitting à faire.
- Rate limiting in-memory (à migrer Redis).
- `pnpm-workspace.yaml` créé (`onlyBuiltDependencies: [core-js]`) — sinon pnpm 11 échoue sur les scripts de build ignorés.
- **Abstractions IA codées** (`LLMProvider`, `ModelRouter`, OpenAI, HeyGen) et **pipeline documents codé** (chromedp + prompts Impeccable + PDF/PPTX) mais **pas encore consommés par des workers/endpoints** ; la génération réelle requiert `OPENAI_API_KEY` et `HEYGEN_API_KEY`.
- Le rendu (chromedp) nécessite **Chrome/Chromium** : présent en dev (`/usr/bin/google-chrome`), installé via `chromium` dans le Dockerfile backend (`CHROME_PATH=/usr/bin/chromium`).
- **Worker in-process** (goroutines) pour les générations — pas asynq ; les jobs sont perdus au redémarrage. **Storage local** (fichiers) — pas durable sur Railway (éphémère) ; à remplacer par S3.
- **Broker d'événements in-process** (`infra/events`) : mono-instance uniquement — avec plusieurs replicas backend, les events du worker n'atteindraient pas la bonne connexion SSE ; Redis pub/sub requis (voir Remaining Work #3).
- Le protocole d'outils du chat repose sur des **marqueurs texte** (`@@SEARCH`, `@@IDEAS`) — dépend de la docilité du modèle ; passer aux **function calling** OpenAI si dérives observées.
- La génération (idées/ebook/assets) requiert `OPENAI_API_KEY` côté backend, sinon les jobs échouent (erreur 401 OpenAI).

## Tests & Validation

- **Backend** : `go build/vet/test` OK. Tests : `TestNeonVerifier` (EdDSA/JWKS), `TestCreditLedgerLifecycle` (ledger), `TestModelRouter` + OpenAI/HeyGen (httptest), prompts Impeccable, PPTX (zip), chromedp (Chrome réel : PDF + slides→PNG + PPTX). **Chat** : `TestParseSearchLine`/`TestParseIdeasBlock`/`TestStreamAnswer*` (machine à états, marqueurs coupés multi-delta), `TestBroker*` (events, déconnexion client lent), et **tests d'intégration** `TestChatTurnIdeaFlow` + `TestChatTurnSearchFlow` (Postgres réel via `AFRILAUNCH_TEST_DB` : messages, idées liées, crédits, events, opportunités).
- **Frontend** : `pnpm typecheck` + `pnpm build` OK. Page `/discover` (chat + panneau contexte) câblée sur les events SSE ; le flux complet Neon Auth reste à tester avec de vraies credentials.

## Database & Migrations

- Base locale `afrilaunch` (schéma v13). `db/migrations/` 00001..00013.
- `00011_conversations.sql` : `conversations`, `conversation_messages` (payload JSONB), `product_ideas.conversation_id`.
- `00012_user_preferences.sql` : `user_preferences` (language, theme) — clée par `user_id`.
- `00013_project_config.sql` : `projects.config` (palette/style/pages) + `generation_jobs.params`.
- `sqlc.yaml` : overrides `uuid→string`, `timestamptz→time.Time`, `jsonb→[]byte`. `make sqlc` après modif de `db/query`.
- Note : `users.password_hash` et `refresh_tokens` ne sont **plus utilisés** (légacy) mais restent en base (nettoyage futur).

## Important Files

- `docs/ai.md` — architecture IA (providers, ModelRouter, pipeline HTML→PDF/PPTX, Impeccable).
- `backend/internal/application/chat/` — **copilote conversationnel** : `service.go` (conversations, boucle agent, runSearch, createIdeas), `prompts.go` (system prompt + outils `@@SEARCH`/`@@IDEAS`), `integration_test.go` (2 chemins sur Postgres réel).
- `backend/internal/application/port/events.go` — `AppEvent`, `EventPublisher`, `EventBus` (canal temps réel).
- `backend/internal/infra/events/broker.go` — broker in-process par user (SSE).
- `backend/internal/server/handler/events.go` — `GET /api/v1/events` (SSE, heartbeat 25 s, `X-Accel-Buffering: no`).
- `backend/internal/server/handler/conversations.go` — endpoints conversations (POST message → 202).
- `backend/internal/infra/postgres/conversation_repo.go` + `db/query/conversations.sql` — persistance chat.
- `backend/internal/application/preferences/` + `backend/internal/server/handler/preferences.go` — préférences (GET/PUT `/preferences`).
- `backend/internal/infra/postgres/preference_repo.go` + `db/migrations/00012_user_preferences.sql` — stockage préférences.
- `backend/internal/domain/project_config.go` — `ProjectConfig`/`ProjectPalette` (défauts, clamp, validation hex).
- `backend/internal/application/jobs/worker.go` — `runCover` (proposition de palette IA + image), `runEbook` (pages + cover en 1re page), `latestCoverPNG`, `proposeVisualIdentity`.
- `backend/internal/application/document/` — prompts avec palette/pages + `PrependCoverPage` + `GenerateEbookDeckWithCover`.
- `backend/internal/application/projects/service.go` — `UpdateConfig` + gate `requireCover` (workflow cover-first).
- `frontend/src/pages/project.tsx` — stepper cover-first (aperçu cover, éditeur palette, plage pages, régénérations).
- `frontend/src/features/preferences/` — api/hooks + `preferences-sync.tsx` (application langue/thème au login).
- `frontend/src/lib/api/events.ts` — client SSE global (reconnexion backoff, watchdog, `useAppEvent`/`useEventsConnection`).
- `frontend/src/pages/discover.tsx` — page chat (`?c={id}`), abonnements events, panneau contexte.
- `frontend/src/features/chat/` — api/hooks (conversations, confirm idée) + composants (messages, input, idea-card, context-panel).
- `backend/internal/application/port/ai.go` — interfaces IA (`LLMProvider`, `ImageProvider`, `VideoProvider`).
- `backend/internal/application/ai/` — `ModelRouter` + `Service` (routage modèle par tâche, `StreamMessages` multi-tours).
- `backend/internal/application/port/render.go` — interface `Renderer` (HTML → PDF/PPTX).
- `backend/internal/application/document/` — service de génération ebook/deck + prompts (Impeccable injecté).
- `backend/internal/infra/ai/openai.go` — client OpenAI (chat + images) ; `openai_research.go` (Responses API + web_search).
- `backend/internal/infra/ai/heygen.go` — client HeyGen (vidéo avatar, `/v3/videos`).
- `backend/internal/infra/render/chromedp.go` — rendu chromedp (PDF + slides→PNG).
- `backend/internal/infra/pptx/pptx.go` — assemblage PPTX image-par-slide.
- `backend/internal/application/jobs/worker.go` — worker asynchrone (idées/ebook/couverture/page de vente/recherche) + publication `job.updated`.
- `backend/internal/application/port/workshop.go` — interfaces `IdeaRepository`/`ConversationRepository`/`ProjectRepository`/`AssetRepository`/`JobRepository`/`Storage`.
- `backend/internal/infra/storage/local.go` — stockage local (S3 à venir).
- `backend/db/migrations/00011_conversations.sql` — conversations/messages + lien idée↔conversation.
- `backend/internal/infra/auth/neon.go` — vérifieur JWT Neon (EdDSA/JWKS).
- `backend/internal/server/auth.go` + `authctx/` — middleware + identité.
- `backend/internal/infra/postgres/credit_repo.go` — ledger idempotent.
- `frontend/src/lib/auth.ts` — client Neon Auth + `getAccessToken`.
- `frontend/src/app/root-providers.tsx` — `NeonAuthUIProvider` branché sur React Router.
- `frontend/src/lib/api/client.ts` — attache le JWT (Bearer) aux requêtes.
- `backend/Dockerfile` + `backend/docker-entrypoint.sh` — image API (build + migrations goose au boot).
- `frontend/Dockerfile` + `frontend/nginx.conf.template` — image SPA (nginx + fallback SPA).

## Déploiement (Railway)

- **IaC** : `.railway/railway.ts` (Infrastructure as Code) définit deux services (`backend`, `frontend`) avec leurs `rootDirectory`, healthchecks et variables non-secrètes.
- `backend` : image `backend/Dockerfile` (build + migrations goose au boot + chromium pour chromedp). Healthcheck `/healthz` (`healthcheckTimeout: 300`). Secrets à définir dans le dashboard : `DATABASE_URL`, `NEON_AUTH_BASE_URL`, `NEON_AUTH_JWKS_URL`, `ALLOWED_ORIGINS` (= URL publique du frontend), `OPENAI_API_KEY`, `HEYGEN_API_KEY`.
- `frontend` : image `frontend/Dockerfile` (nginx + fallback SPA). Healthcheck `/`. Variables injectées au build (ARG) : `VITE_API_URL` (= URL publique du backend), `VITE_NEON_AUTH_URL`.
- Workflow : `npm install` (racine, installe le SDK `railway` pour le DSL `railway/iac`) puis `railway login` + `railway link` + `railway config plan` / `railway config apply`. Validé : `railway config plan --json` OK (crée `backend` + `frontend`, supprime l'ancien service `afri-launch`).

## Next Steps

1. Tester le parcours complet avec vraies credentials (Neon Auth + `OPENAI_API_KEY`) : chat → `@@SEARCH` → idées → valider → projet → ebook → download.
2. **Object Storage** : adapter S3 (Neon) + URLs présignées (remplacer `LocalStorage`).
3. **Paiements Mobile Money** (`PaymentProvider`) + recharges de crédits.
4. **Workers asynq** : remplacer le worker in-process par asynq + Redis pub/sub pour le broker d'événements.
5. UI liste des conversations + titres auto affichés dans le chat.

## Notes for the Next Agent

- **Relire ce fichier en entier** avant de commencer.
- Priorité : `init.md` > `master.md` > conventions > skills > design > docs.
- Dev backend : `APP_ENV=development DATABASE_URL=<local> GOTOOLCHAIN=local make run` (le `.env` local pointe vers Neon production).
- Dev frontend : `VITE_NEON_AUTH_URL=<url> pnpm dev`. Sans cette var, l'auth (get-session) renvoie 404 (attendu).
- Après modif `db/query/*.sql` → `sqlc generate` ; schéma → nouvelle migration goose.
- SDK Neon Auth beta : vérifier les exports réels dans `node_modules` avant d'étendre l'intégration (l'API a déjà divergé de la doc).
