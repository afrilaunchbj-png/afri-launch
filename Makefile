.DEFAULT_GOAL := help

.PHONY: help dev backend frontend api web install build test migrate-up migrate-down

help: ## Affiche les commandes disponibles
	@printf "\n  \033[1mAfriLaunch\033[0m\n\n"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
	@printf "\n  Nécessite : go, pnpm, goose. Les variables d'env sont lues depuis\n  backend/.env et frontend/.env (déjà en place).\n\n"

backend: ## Démarre l'API Go (http://localhost:8080)
	cd backend && go run ./cmd/api

api: backend ## Alias de `backend`

frontend: ## Démarre le frontend (http://localhost:5173)
	cd frontend && pnpm dev

web: frontend ## Alias de `frontend`

dev: ## Démarre backend + frontend ensemble
	@trap 'kill 0' EXIT INT TERM; \
	cd backend && go run ./cmd/api & \
	cd frontend && pnpm dev & \
	wait

install: ## Installe les dépendances (frontend)
	cd frontend && pnpm install

build: ## Compile backend + frontend
	cd backend && go build ./...
	cd frontend && pnpm build

test: ## Tests backend + typecheck frontend
	cd backend && go test ./...
	cd frontend && pnpm typecheck

migrate-up: ## Applique les migrations (DATABASE_URL de backend/.env)
	cd backend && DATABASE_URL="$$(sed -n 's/^DATABASE_URL=//p' .env)" goose -dir db/migrations postgres "$$DATABASE_URL" up

migrate-down: ## Annule la dernière migration
	cd backend && DATABASE_URL="$$(sed -n 's/^DATABASE_URL=//p' .env)" goose -dir db/migrations postgres "$$DATABASE_URL" down
