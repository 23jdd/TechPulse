# 部署

启动本地依赖：

```bash
make docker-up
make migrate
make seed
make run
```

启动完整 compose stack：

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

Kubernetes 起步部署清单：

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/
```

这些清单包含 gateway 探针和 Prometheus scrape annotations。生产环境中需要把 demo 密码替换成 Kubernetes Secrets，并为 MySQL 和 Bleve 增加持久化卷。
