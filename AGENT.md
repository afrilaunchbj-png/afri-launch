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
16. **Vidéos publicitaires** (ADR-016) : job `video_ad` (15 cr) = LLM (analyse+storyboard JSONB) → HeyGen (`VideoProvider`) → FFmpeg (`port.VideoRenderer` : sous-titres burnés + cartes intro/outro avec la cover) → assets MP4+vignette ; avatars/voix par env (`HEYGEN_DEFAULT_AVATAR_ID`/`VOICE_ID`), jamais hard-codés ; progression via events `stage` ; provider-agnostic (ports, aucune logique provider dans le métier).
17. **Intégrations publicitaires** (ADR-017) : port `AdPlatformProvider` (+`Capabilities`) implémenté par plateforme (`infra/ads/{meta,googleads,tiktok}`), service générique, tokens OAuth chiffrés AES-256-GCM au repos (`ENCRYPTION_KEY`), états CSRF usage unique en DB, callback public sans JWT (utilisateur identifié par l'état), campagnes créées en pause (garde-fous budget), montants en unités mineures entières, mapping UUID interne ↔ ID externe, opérations tracées (`provider_operations`). Meta complet (créatives vidéo) ; Google Ads (REST/GAQL) et TikTok Ads (Business API v1.3) : campagnes + insights, créatives non supportées (`Capabilities.Creatives=false`).
18. **Paiements** (ADR-018) : port `PaymentProvider` + adapters REST (`infra/payments/{pawapay,fedapay,paydunya}`), provider unique via `PAYMENT_PROVIDER` (vide = désactivé). Règle de sécurité : webhook jamais une preuve → `VerifyStatus` par API avant octroi ; crédits via ledger idempotent (`payment:<id>`) ; retour navigateur = resync uniquement. Webhooks publics `POST /payments/webhook/{provider}` (sans JWT).
19. **Résilience** : aucune variable d'env optionnelle ne doit arrêter le backend — fallbacks explicites (DenyVerifier pour Neon Auth, chiffreur désactivé pour les clés de chiffrement, LocalStorage si S3 invalide, paiements désactivés sans `PAYMENT_PROVIDER`).

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
- [x] **Journal d'activités, discussions tickets, admin filtrable, stats dashboard** : migration `00016` (`support_ticket_messages`) ; fil de discussion des tickets (`GET/POST /support/tickets/{id}/messages` côté user, `GET /admin/tickets/{id}` + `POST /admin/tickets/{id}/messages` côté support ; un ticket résolu est rouvert quand le client répond) ; `audit_logs` (table 00006) alimentée — inscription, promotion admin, cycle de vie des tickets — et exposée via `GET /admin/audit-logs` (filtres action/entity/userId) ; listes admin filtrables côté serveur (users, tickets, projets, conversations, assets, jobs, transactions de crédits — `?search=&status=&role=`) ; `GET /dashboard/stats` (compteurs perso, solde crédits, consommation 30 j, séries journalière 30 j + hebdomadaire 12 semaines) ; FE : sous-navigation admin (`features/admin/admin-nav`) + pages `/admin/{users,tickets,tickets/:id,projects,conversations,assets,jobs,transactions,audit-logs}` (cartes cliquables → liste filtrée, filtres dans l'URL, toolbar debouncée), `/support/:id` (discussion user), dashboard avec 4 StatCards + courbes recharts (crédits 30 j, projets/semaine) ; icône support `Headset`, doublons footer sidebar supprimés ; i18n FR/EN.
- [x] **Stockage S3 (Neon Object Storage)** (ADR-016, Remaining #1) : `infra/storage/s3.go` (SDK AWS v2, endpoint custom + path-style) ; actif dès que `S3_BUCKET` est défini, fallback `LocalStorage` en dev ; config `S3_ENDPOINT/S3_REGION/S3_ACCESS_KEY_ID/S3_SECRET_ACCESS_KEY/S3_BUCKET/S3_PATH_STYLE`.
- [x] **Vidéos publicitaires MVP (ADR-016)** : nouveau job kind `video_ad` (15 crédits, `generation_costs` déjà seedé) — pipeline : analyse marketing LLM (`application/videoad` : `Analyze` → angle/pain points/hook/CTA, `Story` → storyboard JSON scènes avatar/product/text) → vidéo avatar HeyGen (`VideoProvider` existant, poll 5 s avec tolérance erreurs) → téléchargement URL expirable → montage FFmpeg (`port.VideoRenderer` + `infra/render/ffmpeg.go` : carte intro cover+hook, sous-titres burnés recalés sur durée ffprobe, carte outro CTA ; presets 9:16/1:1/16:9 ; cartes HTML→PNG via chromedp) → assets `video_ad` + `video_ad_thumb` ; storyboard dans `generation_jobs.result` (pas de tables dédiées) ; progression SSE `job.updated` avec champ `stage` (analyzing/storyboarding/generating_avatar/rendering) ; endpoint `POST /projects/{id}/video-ads` (gate idée confirmée + cover) ; FE `features/video-ads` (panneau durée/format/instructions, progression multi-étapes via SSE, player + download, stepper étape 5) ; i18n FR/EN ; ffmpeg + chromedp dans le Dockerfile ; tests (storyboard/subtitles/SRT/carte HTML + intégration ffmpeg skip si binaire absent).
- [x] **Intégrations publicitaires — socle + Meta Ads (ADR-017)** : migration `00017` (`ad_platform_connections`, `oauth_states`, `ad_campaigns`, `ad_creatives`, `ad_insights`, `provider_operations`) + sqlc + repos (tokens chiffrés au repos via `infra/crypto` AES-256-GCM) ; port `AdPlatformProvider` + `Capabilities` ; service `application/advertising` (registry de providers, flux OAuth avec état CSRF usage unique en DB, discovery + sélection de compte re-vérifiée, sync campagnes, création **toujours en pause** avec garde-fous budget, creatives vidéo via URLs signées S3, insights normalisés, refresh token auto → statut expired) ; provider `infra/ads/meta` (OAuth prolongation longue durée, adaccounts, campaigns, advideos+adcreatives, insights time_increment=1 j, erreurs Graph typées `Transient()`) ; endpoints `/integrations/*` (callback **public** sans JWT, redirige vers FE) + `/ad-campaigns*` ; FE page `/integrations` (cartes providers, connect/disconnect, sélection compte, campagnes pause/resume + création en pause) + `features/integrations` + nav ; `.env.example` (ENCRYPTION_KEY, META_*, APP_URL) ; docs `docs/integrations/advertising.md` ; tests (service fakes : état CSRF/replay, vérif compte, garde-fou budget ; httptest Meta : OAuth, discovery, mapping campagnes, insights, erreurs ; intégration Postgres : isolation tenant, upserts idempotents).
- [x] **Google Ads + TikTok Ads + créatives depuis les projets (ADR-017 suite)** : providers `infra/ads/googleads` (REST, pas de gRPC : OAuth offline avec refresh_token, `customers:listAccessibleCustomers`, GAQL searchStream campagnes/insights, création en 2 appels budget+campagne SEARCH manualCpc en pause, micros ↔ minor ×10 000) et `infra/ads/tiktok` (Business API v1.3 : auth_code + refresh_token, `/oauth2/advertiser/get/`, campaign create/status/update, report integrated ; spend majeure → minor ×100) — **créatives désactivées** pour ces deux plateformes (`Capabilities.Creatives=false`, refus propre) ; endpoint `POST /ad-campaigns/{id}/creatives` (`PublishCreative` service : asset → URL signée → upload → mapping externe) + `GET /ad-creatives` ; FE : dialog `PublishCreativeDialog` sur la page projet (« Publier en campagne » sur la vidéo), création de campagne multi-provider (plateformes actives), i18n FR/EN ; registry main.go (Meta/Google/TikTok par env) + redirect URIs par provider ; tests httptest des deux providers (OAuth, discovery, mapping, insights, refus créatives).
- [x] **Correctifs d'anomalies (5)** : ① `projects.credits_consumed` désormais alimenté par le worker (`AddCredits` après `Consume` — l'affichage FE restait à 0) ; ② **pièces jointes du support** (migration `00018` `support_attachments`, upload multipart `POST /support/attachments` — 4 fichiers max 5 Mo images/PDF, rattachés au ticket ou au message via `attachment_ids`, downloads user + superadmin `GET /admin/support/attachments/{id}/download`, FE : `AttachmentPicker`/`AttachmentList` sur formulaire, fil et vue admin, `api.upload` multipart) ; ③ **journal d'activités alimenté** : audit des générations (`generation.dispatched/completed/failed` — dispatched/completed/failed via le worker), projets créés, campagnes/créatives/connexions publicitaires ; ④ bloc « Bientôt disponible » supprimé de la sidebar (`futureNav` retiré) ; ⑤ **résilience aux env manquantes** : `ENCRYPTION_KEY` absente → chiffreur désactivé (module pub. en erreur explicite, pas de crash), `NEON_AUTH_*` absente → vérifieur `DenyVerifier` (routes protégées 401, healthz OK), S3 invalide → fallback local — le backend ne s'arrête plus sur ces variables.
- [x] **Fix 500 admin (audit-logs + tickets/{id})** : les requêtes `ListAuditLogs`/`CountAuditLogs` plantaient sur le filtre vide (`''::uuid` → invalid input syntax — la page du journal 500 donc paraissait vide) ; remplacées par `sqlc.narg('user_id')` (pgtype.UUID NULL, helper `optionalUUID`). Le 500 sur `/admin/tickets/{id}` provient de l'absence de la migration `00018` sur la DB ciblée (vérifié : passe en local avec 00018 appliquée — tests de reproduction `TestAuditLogsList`/`TestTicketDetailAttachments`) ; le run des migrations au boot (goose) ou un `goose up` manuel sur Neon corrige.
- [x] **Compat PgBouncer/Neon pooler (SQLSTATE 08P01)** : le mode pgx par défaut prépare des statements nommés qui entrent en collision sur le pooler → `QueryExecModeCacheDescribe` dans `pool.go` (statement sans nom, OIDs en cache — jsonb []byte OK, validé par les tests d'intégration chat).
- [x] **En-têtes de pages harmonisés** : `/discover` et `/admin` ont désormais un header h1 + description comme les autres pages (admin : header avant `AdminNav`).
- [x] **Checkout paiements (ADR-018)** : port `PaymentProvider` (`CreateCheckout`/`VerifyStatus`/`HandleWebhook`) + adapters REST `infra/payments/{pawapay,fedapay,paydunya}` ; provider actif via `PAYMENT_PROVIDER` (vide = désactivé) ; migration `00019` (`payments.checkout_url` + seed 3 packs : 50 cr/5 000 XOF, 120 cr/10 000, 350 cr/25 000) ; service `application/payments` (checkout 202, webhook public reconfirmé par API, octroi crédits idempotent `payment:<id>`, sync FE) ; endpoints `/payments/*` ; FE page crédits : `PlansPanel` (packs + redirection provider) + resync au retour `?payment=` ; i18n FR/EN ; tests (httptest des 3 providers + intégration Postgres du cycle complet avec idempotence).

## Work In Progress

- Aucune feature en cours.

## Remaining Work

1. **Paiements Mobile Money** (`PaymentProvider`) + recharges de crédits.
2. **Workers asynq** (Redis) : remplacer le worker in-process (`application/jobs`) par asynq + **Redis pub/sub pour le broker d'événements** (`infra/events`).
3. **Tests** : unitaires + intégration + E2E du parcours complet (incl. video_ad avec vraies credentials HeyGen).
4. **Nettoyage legacy** : endpoints job-batch idées/recherche (`POST /opportunities/{id}/ideas`, `POST /research`) + table `idea_messages` non appelés par le FE depuis le chat (à supprimer lors d'une prochaine passe).
5. **Liste des conversations** dans le chat (API `GET /conversations` prête, UI à ajouter).
6. **Vidéo post-MVP** : multi-variantes (1 job/angle), provider B-roll (Veo/Kling → nouveau port `BrollProvider`), scoring LLM, musique libre de droits, regénération scène par scène (tables `ad_scenes` dédiées), catalogue avatars HeyGen (`GET /avatars`).

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
- **Worker in-process** (goroutines) pour les générations — pas asynq ; les jobs sont perdus au redémarrage. ~~Storage local~~ → **S3 câblé** (actif si `S3_BUCKET` défini ; sinon local en dev, éphémère sur Railway).
- **Broker d'événements in-process** (`infra/events`) : mono-instance uniquement — avec plusieurs replicas backend, les events du worker n'atteindraient pas la bonne connexion SSE ; Redis pub/sub requis (voir Remaining Work #2).
- Le protocole d'outils du chat repose sur des **marqueurs texte** (`@@SEARCH`, `@@IDEAS`) — dépend de la docilité du modèle ; passer aux **function calling** OpenAI si dérives observées.
- La génération (idées/ebook/assets/vidéo) requiert `OPENAI_API_KEY` + (`HEYGEN_API_KEY` + `HEYGEN_DEFAULT_AVATAR_ID`/`VOICE_ID` pour les vidéos) côté backend, sinon les jobs échouent (erreur 401 OpenAI / avatar manquant).
- **Vidéo pub MVP** : 1 variante / 1 scène avatar HeyGen parlant tout le script ; sous-titres estimés par proportion de mots puis recalés sur la durée ffprobe (approximation acceptable) ; storyboard en JSONB (regénération scène par scène nécessitera des tables dédiées) ; l'URL HeyGen est expirable → téléchargée immédiatement.
- **ffmpeg/ffprobe requis** pour le montage vidéo : installés dans le Dockerfile backend ; absents en dev local (le test d'intégration ffmpeg est skippé) — installer via le gestionnaire de paquets pour tester localement.
- **Intégrations publicitaires** : sans `META_APP_ID/SECRET`, aucun provider n'est enregistré (connect → 422) ; le callback OAuth est public (sans JWT) et identifie l'utilisateur via l'état consommé — ne jamais ajouter de RequireAuth sur cette route ; l'upload de creatives requiert S3 (URLs présignées, indisponible en local) et une page Facebook (`page_id` dans les métadonnées de connexion) ; les campagnes sont créées en pause par garde-fou (§32) ; tokens déchiffrés uniquement en mémoire — ne jamais les logger.

## Tests & Validation

- **Backend** : `go build/vet/test` OK. Tests : `TestNeonVerifier` (EdDSA/JWKS), `TestCreditLedgerLifecycle` (ledger), `TestModelRouter` + OpenAI/HeyGen (httptest), prompts Impeccable, PPTX (zip), chromedp (Chrome réel : PDF + slides→PNG + PPTX). **Vidéo** : `videoad` (storyboard/subtitles/ResolveVideoRequest/normalisation durées), `render` (SRT, canvas, carte HTML échappée ; **intégration ffmpeg skip si binaire absent**). **Chat** : `TestParseSearchLine`/`TestParseIdeasBlock`/`TestStreamAnswer*` (machine à états, marqueurs coupés multi-delta), `TestBroker*` (events, déconnexion client lent), et **tests d'intégration** `TestChatTurnIdeaFlow` + `TestChatTurnSearchFlow` (Postgres réel via `AFRILAUNCH_TEST_DB` : messages, idées liées, crédits, events, opportunités).
- **Frontend** : `pnpm typecheck` + `pnpm build` OK. Section vidéo de la page projet (`features/video-ads`) câblée sur le SSE + polling ; le flux complet (Neon Auth + HeyGen + ffmpeg) reste à tester avec de vraies credentials.

## Database & Migrations

- Schéma v18 (Neon production + local). `db/migrations/` 00001..00018.
- `00018_support_attachments.sql` : `support_attachments` (captures d'écran/PDF liées aux tickets ou messages, max 5 Mo).
- `00017_advertising.sql` : `ad_platform_connections` (tokens chiffrés), `oauth_states` (usage unique), `ad_campaigns`, `ad_creatives`, `ad_insights`, `provider_operations`.
- `00011_conversations.sql` : `conversations`, `conversation_messages` (payload JSONB), `product_ideas.conversation_id`.
- `00012_user_preferences.sql` : `user_preferences` (language, theme) — clée par `user_id`.
- `00013_project_config.sql` : `projects.config` (palette/style/pages) + `generation_jobs.params`.
- `00014_west_africa_markets.sql` : 24 marchés (seed idempotent).
- `00015_admin_support.sql` : `users.role`, `support_tickets`.
- `00016_support_ticket_messages.sql` : `support_ticket_messages` (fil de discussion des tickets ; le message initial reste dans `support_tickets.message`).
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
- `backend/internal/infra/render/chromedp.go` — rendu chromedp (PDF + slides→PNG + HTML→PNG cartes vidéo).
- `backend/internal/infra/pptx/pptx.go` — assemblage PPTX image-par-slide.
- `backend/internal/application/videoad/` — **créatif vidéo pub** : `prompts.go` (analyse marketing + storyboard, langue du compte), `service.go` (`Analyze`/`Story`, `ResolveVideoRequest`, `BuildSubtitleCues`), `service_test.go`.
- `backend/internal/infra/render/ffmpeg.go` — **montage vidéo** (port.VideoRenderer : cartes intro/outro chromedp, sous-titres SRT burnés, concat, vignette) + `ffmpeg_test.go`.
- `backend/internal/domain/video_ad.go` — `VideoAdParams`, `Storyboard`/`Scene`, stages SSE, `ResultVideoAd`.
- `backend/internal/application/jobs/worker.go` — `runVideoAd` (analyse → storyboard → HeyGen → montage → assets) + `awaitVideo` (poll, tolérance erreurs) + `publishStage` (events `stage`).
- `backend/internal/infra/storage/s3.go` — stockage objet S3-compatible (Neon), actif si `S3_BUCKET` défini.
- `frontend/src/features/video-ads/` — api/hooks (`useGenerateVideoAd`, `useVideoJob`) + `video-ads-panel.tsx` (formulaire + progression SSE + player/download).
- `frontend/src/pages/project.tsx` — section vidéo (étape 5 du stepper) avec `VideoAdsPanel` + `VideoPreview`.
- `backend/internal/application/port/advertising.go` — `AdPlatformProvider`/`Capabilities`, `SecretEncryptor`, `OAuthStateStore`, repos advertising, `StorageSigner`.
- `backend/internal/application/advertising/service.go` — orchestration générique (registry providers, OAuth, comptes, campagnes, creatives, insights, garde-fous).
- `backend/internal/infra/ads/meta/meta.go` — provider Meta complet (OAuth, discovery, campagnes, creatives, insights, erreurs Graph).
- `backend/internal/infra/ads/googleads/googleads.go` — provider Google Ads (REST, OAuth offline, GAQL searchStream, budget+campagne, insights micros).
- `backend/internal/infra/ads/tiktok/tiktok.go` — provider TikTok Ads (Business API v1.3, auth_code/refresh, campagnes, report).
- `frontend/src/features/integrations/publish-creative-dialog.tsx` — publication d'une vidéo de projet comme créative (choix campagne, titre, texte, CTA).
- `backend/internal/infra/crypto/crypto.go` — chiffrement AES-256-GCM des tokens (format `v1:nonce:ct`).
- `backend/internal/infra/postgres/ad_repo.go` — repos connexions/campagnes/creatives/insights/opérations + store d'états OAuth.
- `backend/internal/server/handler/integrations.go` — endpoints intégrations (callback public → redirection FE).
- `backend/db/migrations/00017_advertising.sql` — tables advertising (cf. ADR-017).
- `frontend/src/features/integrations/` + `frontend/src/pages/integrations.tsx` — page Publicité (providers, comptes, campagnes).
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
- `backend` : image `backend/Dockerfile` (build + migrations goose au boot + chromium pour chromedp + **ffmpeg pour le montage vidéo**). Healthcheck `/healthz` (`healthcheckTimeout: 300`). Secrets à définir dans le dashboard : `DATABASE_URL`, `NEON_AUTH_BASE_URL`, `NEON_AUTH_JWKS_URL`, `ALLOWED_ORIGINS` (= URL publique du frontend), `OPENAI_API_KEY`, `HEYGEN_API_KEY`, `HEYGEN_DEFAULT_AVATAR_ID`, `HEYGEN_DEFAULT_VOICE_ID`, `S3_ENDPOINT`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_BUCKET` (+ optionnels `S3_REGION`, `S3_PATH_STYLE`).
- `frontend` : image `frontend/Dockerfile` (nginx + fallback SPA). Healthcheck `/`. Variables injectées au build (ARG) : `VITE_API_URL` (= URL publique du backend), `VITE_NEON_AUTH_URL`.
- Workflow : `npm install` (racine, installe le SDK `railway` pour le DSL `railway/iac`) puis `railway login` + `railway link` + `railway config plan` / `railway config apply`. Validé : `railway config plan --json` OK (crée `backend` + `frontend`, supprime l'ancien service `afri-launch`).

## Next Steps

1. Tester les parcours complets avec vraies credentials : vidéo (Neon Auth + `OPENAI_API_KEY` + `HEYGEN_*` + `S3_*`) et Meta Ads (app Meta en mode dev, `META_APP_ID/SECRET`, `META_OAUTH_REDIRECT_URI`, `ENCRYPTION_KEY`).
2. **Intégrations publicitaires suite (ADR-017)** : Google Ads + TikTok Ads — créatives (ad groups/ads Google, upload vidéo + adgroup/ad TikTok), insights UI (courbes), worker async (asynq) pour les opérations provider, attachement page_id via UI (Meta).
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
