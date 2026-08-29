.DEFAULT_GOAL := help

DEV_DIR := .dev

.PHONY: help dev backend frontend api web dev-stop stop logs install build test migrate-up migrate-down

help: ## Affiche les commandes disponibles
	@printf "\n  \033[1mAfriLaunch\033[0m\n\n"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
	@printf "\n  Nécessite : go, pnpm, goose, setsid.\n  Env lues depuis backend/.env et frontend/.env (déjà en place).\n\n"

backend: ## Démarre l'API Go (http://localhost:8080)
	cd backend && go run ./cmd/api

api: backend ## Alias de `backend`

frontend: ## Démarre le frontend (http://localhost:5173)
	cd frontend && pnpm dev

web: frontend ## Alias de `frontend`

dev: ## Démarre backend + frontend (PIDs et logs dans .dev/)
	@mkdir -p $(DEV_DIR)
	@cd backend && setsid go run ./cmd/api > ../$(DEV_DIR)/backend.log 2>&1 & echo $$! > $(DEV_DIR)/backend.pid
	@cd frontend && setsid pnpm dev > ../$(DEV_DIR)/frontend.log 2>&1 & echo $$! > $(DEV_DIR)/frontend.pid
	@sleep 1
	@echo "  ✔ API      http://localhost:8080  (pid $$(cat $(DEV_DIR)/backend.pid))"
	@echo "  ✔ Frontend http://localhost:5173  (pid $$(cat $(DEV_DIR)/frontend.pid))"
	@echo "  Logs agrégés : make logs   ·   Arrêt : make dev-stop"

dev-stop: ## Arrête proprement backend + frontend
	@-kill -TERM -$$(cat $(DEV_DIR)/backend.pid 2>/dev/null) 2>/dev/null || true
	@-kill -TERM -$$(cat $(DEV_DIR)/frontend.pid 2>/dev/null) 2>/dev/null || true
	@rm -f $(DEV_DIR)/backend.pid $(DEV_DIR)/frontend.pid
	@echo "  ✔ Arrêté."

stop: dev-stop ## Alias de `dev-stop`

logs: ## Suit les logs backend + frontend agrégés en un seul rendu
	@tail -n 30 -f $(DEV_DIR)/backend.log 2>/dev/null | sed 's/^/[api]  /' & tail -n 30 -f $(DEV_DIR)/frontend.log 2>/dev/null | sed 's/^/[web]  /' & wait

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
