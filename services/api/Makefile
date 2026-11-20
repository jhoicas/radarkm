.PHONY: run test build tidy docker-up docker-down migrate-up migrate-down

run:
	go run ./cmd/server

test:
	go test ./...

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

docker-up:
	docker compose up -d postgres redis

docker-down:
	docker compose down -v

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down
