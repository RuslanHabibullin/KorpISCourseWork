BINARY_API  := bin/api
BINARY_ETL  := bin/etl
MIGRATIONS  := ./migrations
GOOSE       := goose
DSN         ?= "host=localhost port=5432 user=carservice password=carservice dbname=carservice sslmode=disable"

.PHONY: all build build-api build-etl run-api run-etl test lint \
        migrate-up migrate-down migrate-status docker-up docker-down clean

all: build

## ── Build ─────────────────────────────────────────────────────────────────────
build: build-api build-etl

build-api:
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY_API) ./cmd/api

build-etl:
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY_ETL) ./cmd/etl

## ── Run local ─────────────────────────────────────────────────────────────────
run-api: build-api
	./$(BINARY_API)

run-etl: build-etl
	./$(BINARY_ETL)

## ── Tests ─────────────────────────────────────────────────────────────────────
test:
	go test ./... -v -race -count=1

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

## ── Lint ──────────────────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

## ── Migrations ────────────────────────────────────────────────────────────────
migrate-up:
	$(GOOSE) -dir $(MIGRATIONS) postgres $(DSN) up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS) postgres $(DSN) down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS) postgres $(DSN) status

migrate-reset:
	$(GOOSE) -dir $(MIGRATIONS) postgres $(DSN) reset

## ── Docker ────────────────────────────────────────────────────────────────────
docker-up:
	docker compose -p carservice up --build -d

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f

## ── Misc ──────────────────────────────────────────────────────────────────────
tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out coverage.html reports/