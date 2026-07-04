# Deployment

## One-command Docker Compose deployment

From the repository root:

```bash
docker compose up -d --build
```

Open:

```text
http://localhost:8080/login
http://localhost:8080/login/zh
http://localhost:8080/dashboard
http://localhost:8080/dashboard/zh
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
