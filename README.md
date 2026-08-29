# AfriLaunch

SaaS de création de produits digitaux pour les marchés africains — transformez une opportunité de marché en produit digital vendable (ebook + assets marketing + page de vente), avec des données locales vérifiées et un paiement Mobile Money.

## Stack

- **Backend** : Go 1.23 · chi · PostgreSQL · Redis (asynq) · sqlc · goose
- **Frontend** : React 19 · Vite · React Router 7 · TypeScript · Tailwind CSS · shadcn/ui · Radix UI

## Prérequis

- Go 1.23+, Node 20+, pnpm, Docker (pour Postgres/Redis en local).

## Démarrage

```bash
# 1. Backend
cd backend
cp .env.example .env
docker compose up -d                      # PostgreSQL + Redis
goose -dir db/migrations postgres "$DATABASE_URL" up
go run ./cmd/api                          # API sur :8080

# 2. Frontend
cd frontend
cp .env.example .env
pnpm install
pnpm dev                                 # SPA sur :5173
```

## Documentation

- [docs/index.md](docs/index.md) — vue d'ensemble
- [docs/architecture.md](docs/architecture.md) — architecture + modèle de données
- [docs/conventions.md](docs/conventions.md) — conventions (i18n, dark mode, formulaires, listes, erreurs, sécurité)
- [docs/decisions.md](docs/decisions.md) — ADR
- [docs/glossary.md](docs/glossary.md) — terminologie
- [AGENT.md](AGENT.md) — mémoire agent

## Structure

```
backend/   # API Go (chi + pgx + sqlc + goose)
frontend/  # SPA React (Vite + React Router 7 + Tailwind + shadcn/ui)
docs/      # documentation
design/    # maquettes Stitch (référence visuelle)
prompts/   # master.md (vision produit) + init.md (contraintes)
```

## Licence

Propriétaire — usage interne.
