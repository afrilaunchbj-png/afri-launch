# AGENT.md

> Mémoire de travail persistante de l'agent. **À lire impérativement avant chaque nouvelle tâche**, puis à mettre à jour après chaque travail significatif.

## Project Overview

**AfriLaunch** est un SaaS B2C/B2B2C qui permet à un créateur, entrepreneur, consultant ou expert **africain** de transformer une opportunité de marché en **produit digital commercialisable** (ebook/guide + assets marketing + page de vente) avec un minimum de travail manuel.

Workflow cœur : Recherche de marché → Identification d'opportunités → Sélection de niche → Génération d'idées → Itération → Création du produit (ebook 4k–6k mots) → Assets marketing → Contrôle qualité → Packaging → Export.

Le positionnement distinctif (vs ChatGPT + Canva) : **verticalisation géographique/sectorielle africaine**, **données de marché locales vérifiées** (jamais de statistiques inventées), et un **workflow end-to-end** optimisé mobile + paiements locaux (Mobile Money).

## Current Status

**Phase 2 — Fondations complètes + premières features MVP.** Le backend est fonctionnel de bout en bout (DB, auth JWT/argon2/OAuth, ledger de crédits, opportunités) et le frontend expose un layout applicatif complet (login/register, dashboard, recherche d'opportunités, historique des crédits) en FR/EN + dark mode.

Implémenté et validé (smoke tests HTTP + navigateur Playwright) :
- Migration initiale complète + pool pgx (PostgreSQL).
- Auth backend : register/login/refresh/logout/me + Google OAuth (PKCE, non testé faute de credentials).
- Ledger de crédits idempotent : `reserve → consume|release`, bonus de bienvenue à l'inscription.
- Recherche d'opportunités : catalogue de référence, filtres (pays/secteur/difficulté/recherche), sauvegarde par utilisateur.
- Frontend : layout applicatif (sidebar desktop / top-bar + bottom-nav mobile), pattern formulaire unique (react-hook-form + zod), pattern liste/pagination/filtres (état dans l'URL), états loading/empty/error, i18n FR/EN, dark mode.

## Architecture

```
afri-launch/
├── AGENT.md                 # ce fichier
├── README.md
├── docs/                    # documentation (index, architecture, conventions, decisions, glossary)
├── design/                  # maquettes Stitch (référence visuelle)
├── prompts/                 # master.md (vision produit) + init.md (contraintes)
├── backend/                 # API Go (monorepo séparé — choix utilisateur)
│   ├── cmd/api/             # point d'entrée / composition root
│   ├── internal/
│   │   ├── config/          # chargement config env
│   │   ├── server/          # routeur chi, middlewares, handlers, authctx, rate limiting
│   │   │   └── handler/     # handlers HTTP (auth, credits, opportunities, markets, health)
│   │   ├── domain/          # entités, erreurs métier (zéro dépendance framework)
│   │   ├── application/     # ports (interfaces) + use cases (auth, credits, opportunities)
│   │   └── infra/           # adapters: postgres (sqlc), auth (argon2, jwt, google)
│   └── db/
│       ├── migrations/      # goose (.sql) — 00001..00007
│       └── query/           # sqlc (.sql) -> internal/infra/postgres/db
└── frontend/                # SPA React (Vite + React Router 7)
    └── src/
        ├── app/             # router, root-layout
        ├── components/      # ui/ (shadcn), layout/, data-table/, pagination, states/
        ├── features/        # auth/, credits/, opportunities/ (api, hooks, components, types)
        ├── i18n/            # locales fr/en (common, auth, credits, dashboard, opportunities)
        ├── lib/             # api client, errors, utils
        └── styles/          # tokens CSS
```

Voir `docs/architecture.md` pour le détail complet.

## Tech Stack

| Couche | Technologie |
|---|---|
| Backend | **Go 1.23** (chi, pgx/v5, sqlc, goose, asynq) |
| Base de données | **PostgreSQL** |
| Cache / Queue | **Redis** (asynq pour les jobs asynchrones — pas encore branché) |
| Frontend | **React 19 + Vite + React Router 7 (mode SPA)** |
| Langage FE | TypeScript (strict) |
| Styles | Tailwind CSS + shadcn/ui + Radix UI |
| État serveur | TanStack Query v5 |
| Formulaires | react-hook-form + zod |
| i18n | react-i18next (fr / en) |
| Icônes | lucide-react |
| Tableaux | @tanstack/react-table v8 |
| Tests | Go: `testing` + testify · FE: Vitest + Testing Library + Playwright |

**Next.js est strictement interdit** (contrainte `prompts/init.md`).

## Important Decisions

Consignées dans `docs/decisions.md` (ADR). Synthèse :

1. **Stack** : Go/PostgreSQL/Redis + React SPA (pas Next.js/Prisma) — contrainte init.md prioritaire.
2. **Monorepo** : `backend/` + `frontend/` séparés.
3. **Accès données** : **sqlc + pgx/v5** + **goose** — SQL source de vérité dans `db/query/` + `db/migrations/`.
4. **Queue** : **asynq** (à venir pour les workflows de génération).
5. **Auth** : JWT access (15 min) + refresh (rotation), cookies httpOnly/Secure/SameSite, Argon2id, Google OAuth (PKCE).
6. **Erreurs API** : RFC 9457 Problem Details, format unique.
7. **Crédits** : ledger double-entrée (CreditAccount / CreditTransaction / CreditReservation), cycle idempotent `reserve → execute → consume|release`, coûts configurables (`generation_costs`).
8. **Paiements** : abstraction `PaymentProvider` (à venir), Mobile Money d'abord.
9. **Multi-tenancy** : isolation par `user_id` (filtre systématique dans les requêtes).

## Design & UI Conventions

- Design system **« Emerald & Amber Ledger »** : primary émeraude `#003527`, secondary ambre `#855300`/`#fea619`, typo **Lexend** (titres) + **Inter** (corps).
- Mobile-first, touch targets ≥ 48px, dark mode (`class`) obligatoire, i18n FR/EN, zéro texte hardcodé.
- Tous les composants UI = shadcn/ui + Radix (0 HTML natif quand un équivalent existe). Icônes **lucide-react** uniquement.
- Pattern formulaire unique : `zod` schema → `useForm` + `zodResolver` → `<Form>`/`FormField`/`FormItem`/`FormLabel`/`FormControl`/`FormMessage` → submit `useMutation` (spinner + toast).
- Pattern liste : `useQuery` + `keepPreviousData`, pagination serveur, état des filtres dans l'URL (`useSearchParams`), états loading/empty/error réutilisables (`components/states/`), `DataTable` (@tanstack/react-table v8) pour les tableaux.

## Completed Work

- [x] Analyse complète du repository, des skills, des designs, de master.md.
- [x] `AGENT.md`, `docs/`, `README.md`, `.project-ai/`.
- [x] Squelette `backend/` + `frontend/` (Phase 1).
- [x] **Migration initiale complète** (00001..00007) : users, organizations, refresh_tokens, credit_accounts/transactions/reservations, generation_costs, plans, payments, markets, opportunities, saved_opportunities, audit_logs + seed (marchés, coûts, packs, opportunités de référence).
- [x] **Pool pgx** + `sqlc generate` (code typé dans `internal/infra/postgres/db`).
- [x] **Auth backend** : register/login/refresh/logout/me + Google OAuth (PKCE) + rate limiting (in-memory) + middleware `RequireAuth`.
- [x] **Ledger de crédits** : Reserve/Consume/Release/Grant idempotents (test d'intégration OK).
- [x] **Opportunités** : liste/filtres/pagination + save/unsave + facettes.
- [x] **Frontend** : layout applicatif, pattern formulaires, pattern listes, pages login/register/dashboard/opportunities/credits, i18n FR/EN, dark mode.

## Work In Progress

- Aucune feature en cours (on enchaîne sur la Phase 3 — génération d'idées → produit).

## Remaining Work

1. **Génération d'idées** : `ProductIdea`/`IdeaVersion`, sélection/feedback (liker/rejeter/fusionner).
2. **Création produit** : `Product`/`Chapter`/`ContentVersion`, génération ebook (4–6k mots).
3. **Assets marketing + page de vente** + **paiements Mobile Money** (abstraction `PaymentProvider`, webhooks).
4. **Workers asynq** : orchestrateur + workers LLM/Image/PDF/QC + couche `LLMProvider`.
5. **Tests** : unitaires (scoring, crédits, authz) + intégration (DB, API, queues) + E2E (parcours complet).

## Known Issues

- **⚠️ Un fichier `backend/.env` contient des secrets de production (Neon Postgres, Upstash Redis, JWT).** Il est bien gitignoré, mais **ne jamais** le committer ni l'exposer ; le faire tourner via la base locale (`APP_ENV=development DATABASE_URL=postgres://afrilaunch:afrilaunch@localhost:5432/afrilaunch?sslmode=disable`). Si le fichier a pu fuiter, **rotater les secrets**.
- Bundle frontend ~793 kB (chunk unique) — code-splitting par route (`React.lazy`) à faire.
- Rate limiting **en mémoire** (pas distribué) — à migrer vers Redis pour le multi-instance.
- Google OAuth implémenté mais **non testé** (pas de `GOOGLE_OAUTH_CLIENT_ID`/`SECRET` renseignés).
- Paiements/recharges non implémentés : le bouton « Recharger » est désactivé (`payment` à venir).
- Le port 8080 est occupé par un ancien process de squelette s'il tourne encore — le tuer si « address already in use ».
- PostgreSQL/Redis du projet tournent via les conteneurs existants `postgres16`/`redis8` (base `afrilaunch` créée dedans) ; `backend/docker-compose.yml` reste l'alternative.

## Tests & Validation

Validé (2026-08-29) :
- **Backend** : `go build ./...`, `go vet ./...`, `go test ./...` OK. Test d'intégration ledger `TestCreditLedgerLifecycle` OK (`make test-integration`). Tests unitaires : argon2, `domain.CreditAccount.Available`.
- **API smoke test** (curl) : `/healthz`, `/readyz`, `/api/v1/markets` ; register→me→credits→transactions→opportunities→filters→save/unsave→refresh→logout ; erreurs 401/422 RFC 9457 OK.
- **Frontend** : `pnpm typecheck` OK, `pnpm build` OK. Vérifié au navigateur (Playwright) : login→dashboard, recherche d'opportunités (filtre difficulté → URL `?difficulty=low`, save/unsave), historique crédits, **dark mode** (classe `.dark`), **i18n FR→EN**, **responsive mobile** (bottom-nav + sheet filtres). Console sans erreur JS (hors `/auth/me` 401 attendu sur pages publiques).

## Database & Migrations

- Base locale **`afrilaunch`** (sur le conteneur `postgres16`), schéma version 7 (goose).
- Migrations `backend/db/migrations/` : `00001_init` (users+pgcrypto), `00002_organizations`, `00003_auth_sessions`, `00004_credits`, `00005_billing`, `00006_markets_opportunities`, `00007_seed`.
- `sqlc.yaml` : overrides `uuid→string`, `timestamptz→time.Time`, `jsonb→[]byte`. Régénérer via `make sqlc` après toute modification de `db/query/*.sql`.

## Important Files

- `prompts/init.md`, `prompts/master.md` — contraintes + vision.
- `docs/architecture.md`, `docs/conventions.md`, `docs/decisions.md`.
- `backend/db/migrations/` — schéma ; `backend/db/query/` — requêtes sqlc.
- `backend/internal/domain/credits.go` — entités ledger.
- `backend/internal/application/port/ports.go` — interfaces (ports).
- `backend/internal/infra/postgres/credit_repo.go` — ledger idempotent (cœur économique).
- `backend/internal/server/router.go` — routes `/api/v1`.
- `frontend/src/components/layout/app-layout.tsx` — layout applicatif.
- `frontend/src/components/ui/form.tsx` — pattern formulaire shadcn.
- `frontend/src/features/{auth,credits,opportunities}/` — features.
- `frontend/src/i18n/locales/{fr,en}/` — locales.

## Next Steps

1. Implémenter la **génération d'idées** (ProductIdea + IdeaVersion, itération).
2. Implémenter la **création produit** (ebook) + **assets marketing** + **page de vente**.
3. Implémenter les **paiements Mobile Money** (PaymentProvider) + recharges de crédits.
4. Mettre en place **asynq** (orchestrateur + workers LLM/Image/PDF/QC).
5. **Tests** : unitaires + intégration + E2E (parcours complet).

## Notes for the Next Agent

- **Relire ce fichier en entier** avant de commencer.
- Outillage installé : `goose`, `sqlc` (dans `~/go/bin`). `docker compose up -d` pour Postgres/Redis (ou réutiliser `postgres16`/`redis8` déjà actifs).
- **Toujours** lancer l'API en dev avec `APP_ENV=development` + `DATABASE_URL` locale explicites (le `.env` local pointe vers la production Neon).
- Après modif de `db/query/*.sql` → `sqlc generate`. Après modif du schéma → nouvelle migration goose + `make migrate-up`.
- La priorité des règles : `init.md` > `master.md` > conventions projet > skills > design > docs.
- Exemples TS du master.md (interfaces providers) à transposer en **interfaces Go** (voir `application/port/ports.go`).
- Mapper les icônes Material Symbols → Lucide (tableau dans `docs/conventions.md`).
- Chaque liste → pagination serveur + états loading/empty/error ; chaque formulaire → pattern zod/rhf (pas d'exception).
