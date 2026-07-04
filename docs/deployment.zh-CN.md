# 部署

## 一条命令 Docker Compose 部署

在项目根目录执行：

```bash
docker compose up -d --build
```

中国大陆网络默认使用这些 Docker 构建代理，可以写在 `.env` 中覆盖：

```env
GOPROXY=https://goproxy.cn,direct
ALPINE_MIRROR=https://mirrors.aliyun.com/alpine
```

如果不想替换 Alpine 源，可以设置 `ALPINE_MIRROR=`。

使用 Docker Compose 运行时，服务地址必须写 Compose 服务名，不能写 `localhost`：

```env
MYSQL_DSN=root:password@tcp(mysql:3306)/techpulse?parseTime=true&charset=utf8mb4&multiStatements=true
REDIS_ADDR=redis:6379
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
ETCD_ENDPOINTS=etcd:2379
MINIO_ENDPOINT=minio:9000
```

在容器里，`localhost` 指当前容器自己，不是宿主机，也不是其他服务容器。

打开：

```text
http://localhost:8080/login
http://localhost:8080/login/zh
http://localhost:8080/app
http://localhost:8080/app/zh
```

可选：导入 demo Feed：

```bash
docker compose --profile tools run --rm seed
```

查看状态和日志：

```bash
docker compose ps
docker compose logs -f gateway
```

停止服务：

```bash
docker compose down
```

删除持久化数据：

```bash
docker compose down -v
```

## 本地 Go 开发

只启动本地依赖：

```bash
make docker-up
make migrate
make seed
make run
```

## 旧 compose 文件

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

Kubernetes 起步部署清单：

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/
```

这些清单包含 gateway 探针和 Prometheus scrape annotations。生产环境中需要把 demo 密码替换成 Kubernetes Secrets，并为 MySQL 和 Bleve 增加持久化卷。
