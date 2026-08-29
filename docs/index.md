# Project Index — AfriLaunch

> SaaS de création de produits digitaux pour les marchés africains.

## Vue d'ensemble

AfriLaunch permet à un entrepreneur/consultant/expert africain de transformer une opportunité de marché en produit digital vendable (ebook/guide + assets marketing + page de vente) en quelques minutes, avec des données de marché locales vérifiées et un paiement adapté (Mobile Money).

## Stack

| Couche | Technologie |
|---|---|
| Backend | Go 1.23 · chi · pgx/v5 · sqlc · goose · asynq |
| Database | PostgreSQL |
| Cache / Queue | Redis (asynq) |
| Frontend | React 19 · Vite · React Router 7 (SPA) · TypeScript |
| Styles | Tailwind CSS · shadcn/ui · Radix UI |
| État serveur | TanStack Query v5 |
| Formulaires | react-hook-form · zod |
| i18n | react-i18next (fr / en) |
| Icônes | lucide-react |
| Tests | Go `testing`/testify · Vitest · Testing Library · Playwright |

**Next.js est strictement interdit.**

## Architecture

- **Backend** : Clean Architecture orientée ports/adapters (`internal/domain` → `application` → `infra` → `server`). API REST versionnée `/api/v1`, erreurs RFC 9457 (Problem Details), auth via **Neon Auth** (JWT EdDSA/JWKS).
- **Frontend** : SPA React (Vite + React Router 7), design system shadcn/ui mappé sur le thème « Emerald & Amber Ledger », dark mode, i18n FR/EN.
- **Asynchrone** : les générations longues passent par Redis + asynq (workers), jamais dans une requête HTTP.
- **IA** : OpenAI (GPT) + HeyGen ; documents générés en HTML puis rendus en PDF/PPTX via chromedp — voir `docs/ai.md`.

Détails : voir `docs/architecture.md` et `docs/ai.md`.

## Key Libraries

- **Go** : `go-chi/chi` (routeur), `jackc/pgx/v5` (driver Postgres), `sqlc` (quêtes typées), `pressly/goose` (migrations), `hibiken/asynq` (queue), `golang-jwt/jwt/v5` (vérif JWT Neon Auth), `chromedp` (rendu HTML→PDF, à venir).
- **React** : `react-router` (v7), `@tanstack/react-query`, `react-hook-form`, `zod`, `i18next`/`react-i18next`, `lucide-react`, `tailwindcss`, `class-variance-authority`, `clsx`/`tailwind-merge`, Radix primitives + shadcn/ui, `@neondatabase/neon-js`/`auth-ui`.

## Essential Conventions

Voir `docs/conventions.md` (conventions de code Go/React, pattern formulaires, pattern listes/pagination/filtres, système d'erreurs, i18n, dark mode, sécurité) et `docs/decisions.md` (ADR).

## Quickstart

```bash
# Backend (Go) — nécessite PostgreSQL + Redis
cd backend
cp .env.example .env            # adapter DATABASE_URL (base locale)
docker compose up -d            # postgres + redis (ou réutiliser des conteneurs existants)
make sqlc                       # génère le code typé (sqlc)
goose -dir db/migrations postgres "$DATABASE_URL" up
APP_ENV=development make run

# Frontend (React)
cd frontend
cp .env.example .env            # VITE_API_URL optionnel (proxy Vite par défaut)
pnpm install
pnpm dev
```

## Folder Map

```
afri-launch/
├── AGENT.md            # mémoire agent (lire avant toute tâche)
├── README.md
├── docs/               # index, architecture, conventions, decisions, glossary
├── design/             # maquettes Stitch
├── prompts/            # master.md + init.md
├── backend/            # API Go
└── frontend/           # SPA React
```
