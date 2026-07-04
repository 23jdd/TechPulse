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
