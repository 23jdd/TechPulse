# TechPulse

[English](README.md)

TechPulse 是一个用 Go 构建的 AI 开发者技术情报平台。它可以从 RSS/Atom、GitHub Releases、Hacker News、GitHub 仓库监控等来源采集技术内容，完成解析、清洗、去重、AI 摘要、标签、关键词、embedding、Bleve 全文搜索，并支持带引用来源的 RAG 问答。

核心链路：

```text
添加订阅源 -> 异步/同步抓取 -> 解析文章 -> 去重
-> AI 摘要 / 标签 / Embedding -> 写入 MySQL -> 写入 Bleve
-> 搜索 -> RAG 问答 -> 任务中心 / WebSocket 事件
```

## 项目价值

TechPulse 不是普通博客 CRUD，而是接近真实后端产品的技术信息处理系统，适合用于 Go 后端求职项目展示。

它展示了：

- Go 生产风格分层目录结构。
- 真实 RSS/Atom 抓取。
- Feed 创建、更新、删除、启停、测试、抓取频率、健康评分、OPML 导入导出。
- RabbitMQ 异步抓取任务和 worker 消费。
- 文章已读、收藏、稍后读、归档、删除。
- GitHub Releases 抓取。
- GitHub repo 监控：stars、open issues、latest release、breaking/security 提示。
- Hacker News top/new/best/ask/show/job 抓取。
- GitHub OAuth 登录、用户 upsert、JWT session token。
- 用户偏好设置：兴趣标签、时区、定时日报邮箱和时间。
- SMTP 测试邮件、手动日报、定时日报发送。
- MySQL 持久化、Redis 缓存。
- Bleve 全文搜索、字段权重、过滤、高亮。
- 可插拔 AI Provider：mock、OpenAI-compatible、Ollama-compatible。
- RAG query rewrite、引用、置信度、no-answer guard。
- Docker Compose 一键启动 MySQL、Redis、RabbitMQ、etcd、MinIO 和 Go 服务。

## 3 分钟部署

直接 Docker 部署：

```bash
docker compose up -d --build
```

中国大陆网络构建代理可在 `.env` 中配置：

```env
GOPROXY=https://goproxy.cn,direct
ALPINE_MIRROR=https://mirrors.aliyun.com/alpine
```

Docker Compose 中服务地址要写容器服务名，不要写 `localhost`：

```env
MYSQL_DSN=root:password@tcp(mysql:3306)/techpulse?parseTime=true&charset=utf8mb4&multiStatements=true
REDIS_ADDR=redis:6379
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
ETCD_ENDPOINTS=etcd:2379
MINIO_ENDPOINT=minio:9000
```

打开页面：

```text
http://localhost:8080/login
http://localhost:8080/login/zh
http://localhost:8080/app
http://localhost:8080/app/zh
```

导入 demo Feed：

```bash
docker compose --profile tools run --rm seed
```

## 常用 API

```bash
curl http://localhost:8080/health

curl -X POST http://localhost:8080/api/v1/rss \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev/blog/feed.atom","title":"Go Blog","category":"Go","fetch_interval_minutes":360}'

curl -X POST http://localhost:8080/api/v1/rss/1/test
curl -X POST http://localhost:8080/api/v1/rss/1/fetch-async

curl "http://localhost:8080/api/v1/search?q=go&tag=Go&source=rss&page=1&page_size=20"

curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"question":"最近 Go 有什么重要更新？","conversation_id":1}'

curl -X POST http://localhost:8080/api/v1/github/releases/fetch \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/golang/go"}'

curl -X POST http://localhost:8080/api/v1/github/repos \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/golang/go"}'
```

## Web UI

Gateway 直接提供 Tailwind 用户应用：

- 英文登录页：`http://localhost:8080/login`
- 中文登录页：`http://localhost:8080/login/zh`
- 英文应用：`http://localhost:8080/app`
- 中文应用：`http://localhost:8080/app/zh`

页面支持：

- RSS Feed 添加、测试、异步抓取、启用、停用、删除。
- Feed 健康评分和失败状态展示。
- GitHub Releases 抓取。
- GitHub repo 监控。
- Hacker News 抓取。
- 文章搜索、标签过滤、摘要查看。
- 已读、收藏、稍后读、归档、删除。
- RAG 问答、引用、置信度、证据不足提示。
- 趋势分析。
- 任务状态面板。
- 用户偏好设置。
- 每日报告生成。

## 功能状态

| 模块 | 状态 | 说明 |
| --- | --- | --- |
| Feed 管理 | 已完成 | 创建、更新、删除、启停、测试、频率、健康评分、OPML |
| RSS / Atom 抓取 | 已完成 | 真实 HTTP 抓取、超时、User-Agent |
| RabbitMQ Worker | 已完成 | `/fetch-async` 入队，worker 消费任务 |
| GitHub Releases | 已完成 | 抓取 release 标题、tag、正文、作者、发布时间 |
| GitHub Repo Monitor | 已完成 | stars、open issues、latest release、breaking/security 提示 |
| Hacker News | 已完成 | top/new/best/ask/show/job stories |
| Parser / Cleaner | 已完成 | RSS item 解析和 HTML 清洗 |
| URL / 内容 Hash 去重 | 已完成 | 稳定 SHA-256 hash |
| Mock AI | 已完成 | 无需 API Key 即可运行 |
| OpenAI-compatible Provider | 已完成 | Chat 和 Embedding 接口 |
| Ollama 模式 | 已完成 | 使用 OpenAI 兼容 `/v1` 接口 |
| MySQL 存储 | 已完成 | gateway 启动自动迁移 |
| Redis 缓存 | 已完成 | 热点 REST 响应 best-effort 缓存 |
| 文章管理 | 已完成 | 列表、详情、阅读历史、收藏、稍后读、归档、删除 |
| 趋势分析 | 已完成 | 7/30/90 天热门标签、来源统计、每日数量 |
| Bleve 搜索 | 已完成 | 标题、正文、摘要、标签、来源、日期过滤、高亮 |
| RAG 问答 | 已完成 | query rewrite、引用、confidence、no-answer guard |
| WebSocket 事件 | 已完成 | 抓取、索引、新文章事件 |
| 登录 / GitHub OAuth | 已完成 | 登录页、Auth URL、callback、用户 upsert、JWT session token |
| 用户偏好 | 已完成 | 兴趣标签、时区、日报时间/邮箱/开关 |
| 邮件发送 | 已完成 | SMTP 测试邮件、手动日报、定时日报 |
| Docker Compose | 已完成 | MySQL、Redis、RabbitMQ、etcd、MinIO、Go 服务 |
| Kubernetes | 起步完成 | 基础部署清单 |
| Reddit / Arxiv / YouTube | 预留 | Fetcher 接口已准备 |

## AI 配置

默认本地 mock 模式：

```env
AI_PROVIDER=mock
```

OpenAI-compatible：

```env
AI_PROVIDER=openai
AI_BASE_URL=https://api.openai.com/v1
AI_API_KEY=your-key
AI_MODEL=gpt-4o-mini
```

Ollama：

```env
AI_PROVIDER=ollama
AI_BASE_URL=http://localhost:11434/v1
AI_MODEL=llama3.1
```

## 鉴权配置

默认保留本地 demo 模式：

```env
JWT_SECRET=change-this-to-a-long-random-string
JWT_AUTH_REQUIRED=false
```

开启强鉴权：

```env
JWT_AUTH_REQUIRED=true
```

GitHub OAuth UI 登录成功后会保存 `session_token`，前端请求会自动带上 `Authorization: Bearer <token>`。

## SMTP 配置

```env
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-user
SMTP_PASSWORD=your-password
SMTP_FROM=TechPulse <noreply@example.com>
```

配置 SMTP 后，测试邮件、手动日报和定时日报才会真正发送。

## 已知限制

- 主链路 RSS/GitHub Releases/Hacker News/GitHub repo monitor -> AI -> Search -> RAG 已完成；Reddit、Arxiv、YouTube 仍是扩展 stub。
- RabbitMQ 异步抓取已实现；gateway 保留进程内 fallback，方便本地演示。
- JWT 强鉴权通过 `JWT_AUTH_REQUIRED=true` 开启，默认保留 demo 便利性。
- 邮件发送和定时日报需要 SMTP 环境变量。
- Observability 已具备 Prometheus-ready 结构，但还不是完整 tracing 栈。

## 简历表述

```text
TechPulse - AI-powered Developer Knowledge Hub

- 构建 Go 技术情报平台，采集 RSS/Atom、GitHub Releases、Hacker News 和 GitHub repo 数据，进行去重、AI 摘要、标签、embedding 和 MySQL 存储。
- 实现 RabbitMQ 异步抓取任务、worker 消费、任务中心与前端任务状态面板。
- 实现 GitHub OAuth、JWT session token、用户偏好、SMTP 测试邮件、手动日报和定时日报。
- 使用 Bleve 实现全文搜索，支持标题/正文/摘要/标签搜索、字段权重、分页、过滤和高亮。
- 构建 RAG Chat API，支持 query rewrite、引用、confidence 和 no-answer guard。
- 提供 Docker Compose 一键部署，包含 MySQL、Redis、RabbitMQ、etcd、MinIO 和 Go 服务。
```
