# Deployment

Run local infrastructure:

```bash
make docker-up
make migrate
make seed
make run
```

Run the full compose stack:

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

Kubernetes starter manifests:

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/
```

The manifests include gateway probes and Prometheus scrape annotations. For production, replace demo passwords with Kubernetes Secrets and add persistent volumes for MySQL and Bleve.
