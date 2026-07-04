APP=techpulse
.PHONY: build run-api run-worker test migrate docker-up docker-down
build:
	go build -o bin/techpulse-api ./cmd/techpulse-api
	go build -o bin/techpulse-worker ./cmd/techpulse-worker
	go build -o bin/techpulse-migrate ./cmd/techpulse-migrate
run-api:
	go run ./cmd/techpulse-api
run-worker:
	go run ./cmd/techpulse-worker
test:
	go test ./...
migrate:
	go run ./cmd/techpulse-migrate
docker-up:
	docker compose up -d
docker-down:
	docker compose down