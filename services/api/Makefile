.PHONY: run test build tidy docker-up docker-down

run:
	go run ./cmd/server

test:
	go test ./...

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v
