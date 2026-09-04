SHELL := /bin/sh

-include .env
export

APP := ./cmd/api
DATABASE_URL ?= postgres://our_sell:our_sell@localhost:5432/our_sell?sslmode=disable
MIGRATIONS := db/migrations
SQLC_VERSION := v1.31.1
GOOSE_VERSION := v3.27.3

.PHONY: dev build test test-race test-integration lint fmt sqlc migrate-up migrate-down docker-up docker-down start stop logs status

dev:
	docker compose up -d postgres redis
	go run $(APP)

build:
	go build -trimpath -o bin/our-sell-api $(APP)

test:
	go test ./...

test-race:
	go test -race ./...

test-integration:
	RUN_INTEGRATION=1 go test -tags=integration ./tests/integration

lint:
	golangci-lint run

fmt:
	gofmt -w api cmd internal db/sqlc

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS) postgres "$(DATABASE_URL)" up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS) postgres "$(DATABASE_URL)" down

docker-up:
	docker compose up --build -d

docker-down:
	JWT_SECRET=$${JWT_SECRET:-local-development-only-secret-at-least-32-bytes} docker compose down

start: docker-up

stop: docker-down

logs:
	docker compose logs -f api

status:
	docker compose ps
