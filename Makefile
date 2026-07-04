APP=techpulse
COMPOSE=docker compose -f deploy/docker-compose.yml

.PHONY: dev run build test lint docker-up docker-down migrate seed clean

dev: docker-up migrate seed run

run:
	go run ./cmd/gateway

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
