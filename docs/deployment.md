# Deployment

## One-command Docker Compose deployment

From the repository root:

```bash
docker compose up -d --build
```

For China mainland networks, Docker build uses these defaults from `.env` or `compose.yaml`:

```env
GOPROXY=https://goproxy.cn,direct
ALPINE_MIRROR=https://mirrors.aliyun.com/alpine
```

Set `ALPINE_MIRROR=` to disable Alpine mirror rewriting.

When running inside Docker Compose, service addresses must use Compose service names, not `localhost`:

```env
MYSQL_DSN=root:password@tcp(mysql:3306)/techpulse?parseTime=true&charset=utf8mb4&multiStatements=true
REDIS_ADDR=redis:6379
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
ETCD_ENDPOINTS=etcd:2379
MINIO_ENDPOINT=minio:9000
```

Inside a container, `localhost` means the current container itself.

Open:

```text
http://localhost:8080/login
http://localhost:8080/login/zh
http://localhost:8080/app
http://localhost:8080/app/zh
```

Optional demo feed seed:

```bash
docker compose --profile tools run --rm seed
```

Check status and logs:

```bash
docker compose ps
docker compose logs -f gateway
```

Stop the stack:

```bash
docker compose down
```

Remove persisted data:

```bash
docker compose down -v
```

## Local Go development

Run local infrastructure only:

```bash
make docker-up
make migrate
make seed
make run
```

## Legacy compose file

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

Kubernetes starter manifests:

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/
```

The manifests include gateway probes and Prometheus scrape annotations. For production, replace demo passwords with Kubernetes Secrets and add persistent volumes for MySQL and Bleve.
