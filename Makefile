APP=techpulse
COMPOSE=docker compose -f deploy/docker-compose.yml

.PHONY: dev demo run run-gateway run-scheduler run-fetcher run-parser run-ai-pipeline run-search run-rag run-worker build test lint docker-up docker-down migrate seed reindex clean

dev: docker-up migrate seed run

demo:
	$(COMPOSE) up -d --build mysql redis rabbitmq etcd minio gateway
	go run ./cmd/worker -mode=migrate
	go run ./cmd/worker -mode=seed
	curl http://localhost:8080/health
	curl -X POST http://localhost:8080/api/v1/rss/1/fetch
	curl "http://localhost:8080/api/v1/search?q=go"
	curl -X POST http://localhost:8080/api/v1/chat -H "Content-Type: application/json" -d "{\"question\":\"What changed recently in Go?\",\"conversation_id\":1}"

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

reindex:
	curl -X POST http://localhost:8080/api/v1/search/reindex

clean:
	-$(COMPOSE) down -v
	-$(RM) -r data bin
