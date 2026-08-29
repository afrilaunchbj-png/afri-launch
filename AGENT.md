# AGENT.md

> Mémoire de travail persistante de l'agent. **À lire impérativement avant chaque nouvelle tâche**, puis à mettre à jour après chaque travail significatif.

## Project Overview

**AfriLaunch** est un SaaS B2C/B2B2C qui permet à un créateur, entrepreneur, consultant ou expert **africain** de transformer une opportunité de marché en **produit digital commercialisable** (ebook/guide + assets marketing + page de vente) avec un minimum de travail manuel.

Workflow cœur : Recherche de marché → Identification d'opportunités → Sélection de niche → Génération d'idées → Itération → Création du produit (ebook 4k–6k mots) → Assets marketing → Contrôle qualité → Packaging → Export.

Positionnement : **verticalisation géographique/sectorielle africaine**, **données locales vérifiées** (jamais de statistiques inventées), workflow end-to-end **mobile-first** + paiements locaux (Mobile Money).

## Current Status

**Phase 2 — Fondations + premières features, avec migration de l'auth vers Neon Auth.** L'identité est désormais gérée par **Neon Auth (Managed Better Auth)**, l'infra cible est **Neon** (Postgres + Object Storage) et Redis (Upstash).

Implémenté et validé :
- Migration complète + pool pgx (PostgreSQL).
- **Auth via Neon Auth** : le backend vérifie les JWT EdDSA (JWKS) ; plus de mot de passe/argon2/refresh tokens (voir ADR-011).
- Ledger de crédits idempotent (`reserve → consume|release`, bonus de bienvenue au 1er login).
- Recherche d'opportunités (catalogue + filtres + sauvegarde).
- Frontend : layout applicatif, patterns formulaire/liste, pages login/register (UI Better Auth), dashboard, opportunités, crédits ; i18n FR/EN ; dark mode.

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
│   │   │   └── handler/     # auth, credits, opportunities, markets, health
│   │   ├── domain/          # entités + erreurs métier
│   │   ├── application/     # ports + use cases (auth, credits, opportunities)
│   │   └── infra/           # postgres (sqlc), auth (neon verifier)
│   └── db/
│       ├── migrations/      # goose 00001..00007
│       └── query/           # sqlc -> internal/infra/postgres/db
└── frontend/                # SPA React (Vite + React Router 7)
    └── src/
        ├── app/             # router, root-layout, root-providers
        ├── components/      # ui/ (shadcn), layout/, data-table/, states/
        ├── features/        # auth/, credits/, opportunities/
        ├── i18n/            # locales fr/en
        └── lib/             # api client, auth (Neon Auth), errors, utils
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

## Work In Progress

- Aucune feature en cours.

## Remaining Work

1. **Abstractions IA (Go)** : ports `LLMProvider`/`ImageProvider`/`VideoProvider` + `ModelRouter` + `infra/ai/openai.go` + `infra/ai/heygen.go` + config multi-clés.
2. **Pipeline documents** : contenu JSON → templates HTML (thème + Impeccable) → chromedp → PDF/PPTX image-par-slide.
3. **Génération d'idées** (`ProductIdea`/`IdeaVersion`) puis **création produit** (ebook) + **assets** + **page de vente**.
4. **Object Storage** : adaptateur S3 (Neon) + URLs présignées.
5. **Paiements Mobile Money** (`PaymentProvider`) + recharges de crédits.
6. **Workers asynq** (Research, LLM, Image, Video, Render, QC).
7. **Tests** : unitaires + intégration + E2E.

## Known Issues

- **⚠️ `backend/.env` + `backend/.env.production` contiennent des secrets (Neon Postgres, Upstash Redis).** Gitignorés (`.env.*`), mais **à rotater** si exposés. En dev, forcer `APP_ENV=development` + `DATABASE_URL` locale.
- **Neon Auth non testé de bout en bout** : nécessite `NEON_AUTH_BASE_URL` (backend) et `VITE_NEON_AUTH_URL` (frontend) réelles + activation auth/Google dans la console Neon. Le vérifieur JWT est testé unitairement (EdDSA/JWKS factices).
- `channel_binding=require` dans l'URL Neon : **validé** (pgx 5.7.4 le supporte, `/readyz` OK sur Neon).
- SDK Neon Auth **en beta** (`neon-js 0.7.0-beta`, `auth-ui 0.3.0-beta`) — API volatile.
- Bundle frontend ~1.43 MB (SDK auth lourd) — code-splitting à faire.
- Rate limiting in-memory (à migrer Redis).
- `pnpm-workspace.yaml` créé (`onlyBuiltDependencies: [core-js]`) — sinon pnpm 11 échoue sur les scripts de build ignorés.
- **Aucune couche IA/workers encore implémentée** : `LLMProvider`, `ModelRouter`, chromedp, HeyGen et le pipeline HTML→PDF/PPTX sont documentés (`docs/ai.md`) mais pas codés.

## Tests & Validation

- **Backend** : `go build/vet/test` OK (GOTOOLCHAIN=local). `TestNeonVerifier` (EdDSA/JWKS) OK, `TestCreditLedgerLifecycle` OK (`make test-integration`).
- **Frontend** : `pnpm typecheck` + `pnpm build` OK. Navigateur (Playwright) : home/login rendus (AuthView), dark mode, i18n — vérifié avant la migration auth ; le flux complet Neon Auth reste à tester avec de vraies credentials.

## Database & Migrations

- Base locale `afrilaunch` (schéma v7). `db/migrations/` 00001..00007.
- `sqlc.yaml` : overrides `uuid→string`, `timestamptz→time.Time`, `jsonb→[]byte`. `make sqlc` après modif de `db/query`.
- Note : `users.password_hash` et `refresh_tokens` ne sont **plus utilisés** (légacy) mais restent en base (nettoyage futur).

## Important Files

- `docs/ai.md` — architecture IA (providers, ModelRouter, pipeline HTML→PDF/PPTX, Impeccable).
- `backend/internal/infra/auth/neon.go` — vérifieur JWT Neon (EdDSA/JWKS).
- `backend/internal/application/port/ports.go` — interfaces (dont `TokenVerifier`, `AuthUser`).
- `backend/internal/server/auth.go` + `authctx/` — middleware + identité.
- `backend/internal/infra/postgres/credit_repo.go` — ledger idempotent.
- `frontend/src/lib/auth.ts` — client Neon Auth + `getAccessToken`.
- `frontend/src/app/root-providers.tsx` — `NeonAuthUIProvider` branché sur React Router.
- `frontend/src/lib/api/client.ts` — attache le JWT (Bearer) aux requêtes.

## Next Steps

1. Configurer Neon Auth côté console (enable auth, provider Google, trusted domains) + renseigner `NEON_AUTH_BASE_URL` / `VITE_NEON_AUTH_URL`, puis tester le parcours complet.
2. Implémenter les **abstractions IA** (ports + `infra/ai/openai.go` + `heygen.go` + `ModelRouter` + config).
3. Implémenter le **pipeline documents** (templates HTML + chromedp + export PDF/PPTX), en injectant le skill Impeccable dans le prompt.
4. Génération d'idées → création produit (ebook) → assets → page de vente.
5. Object Storage (Neon S3) + paiements Mobile Money + workers asynq.

## Notes for the Next Agent

- **Relire ce fichier en entier** avant de commencer.
- Priorité : `init.md` > `master.md` > conventions > skills > design > docs.
- Dev backend : `APP_ENV=development DATABASE_URL=<local> GOTOOLCHAIN=local make run` (le `.env` local pointe vers Neon production).
- Dev frontend : `VITE_NEON_AUTH_URL=<url> pnpm dev`. Sans cette var, l'auth (get-session) renvoie 404 (attendu).
- Après modif `db/query/*.sql` → `sqlc generate` ; schéma → nouvelle migration goose.
- SDK Neon Auth beta : vérifier les exports réels dans `node_modules` avant d'étendre l'intégration (l'API a déjà divergé de la doc).
