APP=techpulse
COMPOSE=docker compose -f deploy/docker-compose.yml

.PHONY: dev run run-gateway run-scheduler run-fetcher run-parser run-ai-pipeline run-search run-rag run-worker build test lint docker-up docker-down migrate seed clean

dev: docker-up migrate seed run

run:
	go run ./cmd/gateway

run-gateway:
	go run ./cmd/gateway

run-scheduler:
	go run ./cmd/scheduler

run-fetcher:
	go run ./cmd/fetcher

run-parser:
	go run ./cmd/parser

run-ai-pipeline:
	go run ./cmd/ai-pipeline

run-search:
	go run ./cmd/search

run-rag:
	go run ./cmd/rag

run-worker:
	go run ./cmd/worker

build:
	go build ./...

test:
	go test ./...

lint:
	go vet ./...

docker-up:
	$(COMPOSE) up -d mysql redis rabbitmq etcd minio

docker-down:
	$(COMPOSE) down

migrate:
	go run ./cmd/worker -mode=migrate

seed:
	go run ./cmd/worker -mode=seed

clean:
	-$(COMPOSE) down -v
	-$(RM) -r data bin
