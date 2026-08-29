.PHONY: run build test vet tidy sqlc migrate-up migrate-down test-integration

run:
	go run ./cmd/api

build:
	go build ./...

test:
	go test ./...

test-integration:
	AFRILAUNCH_TEST_DB="$$DATABASE_URL" go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

sqlc:
	sqlc generate

migrate-up:
	goose -dir db/migrations postgres "$$DATABASE_URL" up

migrate-down:
	goose -dir db/migrations postgres "$$DATABASE_URL" down
