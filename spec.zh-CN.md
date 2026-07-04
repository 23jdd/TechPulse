# TechPulse 项目规格说明（中文版）

本文是 `spec.md` 的中文说明版，用于快速理解项目目标和实现要求。

## 角色要求

你是一名资深 Go 后端架构师和全栈工程师，需要生成一个真实可运行、具备生产风格的项目：**TechPulse**。

项目不能只解释架构，不能只给伪代码，而是需要真实 Go 代码、真实目录结构、真实 API、真实数据库模型、Docker Compose、Makefile，以及可运行 MVP。

## 项目定位

TechPulse 是一个 AI 驱动的开发者技术情报平台。它持续从 RSS、GitHub、HackerNews、Reddit、Arxiv、YouTube 等来源采集技术内容，然后使用 AI 完成清洗、去重、摘要、翻译、标签、索引、搜索和问答。

它不是普通 RSS 阅读器，而是面向开发者的 AI Knowledge Hub。

## 应覆盖的信息类型

- 技术新闻
- GitHub Releases
- AI 新闻
- Go 社区动态
- Linux
- Kubernetes
- 数据库
- 网络
- Cloud Native
- 安全
- 编程语言更新

## 技术栈

项目主体使用 Go，并包含：

- MySQL
- Redis
- RabbitMQ
- etcd
- Bleve 搜索引擎
- MinIO 对象存储
- OpenAI-compatible AI Provider
- Docker Compose
- Makefile
- OpenTelemetry-ready 结构
- Prometheus-ready 指标结构

## 服务模块

必须包含这些可独立运行的服务入口：

- gateway
- scheduler
- fetcher
- parser
- ai-pipeline
- search
- rag
- worker

## 主链路

```text
Scheduler
-> Fetcher
-> Parser
-> Duplicate Detection
-> AI Pipeline
-> Search Index
-> MySQL / Redis / MinIO
-> REST API / WebSocket API
-> Dashboard / AI Chat
```

## etcd 用途

- 服务发现
- 全局配置
- 分布式锁

## RabbitMQ 用途

- 异步文章抓取任务
- 解析任务
- AI 处理任务
- 索引任务
- 日报任务

## MVP 要求

MVP 必须可运行：

```bash
make dev
```

或：

```bash
docker compose up -d
make run
```

MVP 不需要一次性完成所有高级功能，但必须跑通真实路径：

1. 用户添加 RSS Feed
2. 系统抓取文章
3. 解析和清洗文章
4. 去重
5. AI 生成摘要、标签、关键词、embedding
6. 存入 MySQL
7. 建立 Bleve 索引
8. 通过 REST API 搜索
9. 通过 RAG Chat 基于文章回答问题
10. WebSocket 推送任务状态

## 当前实现说明

当前仓库已经实现了一个可运行版本，并额外补充了：

- Feed 管理
- 文章管理
- Redis 缓存
- GitHub Releases 抓取
- GitHub OAuth
- SMTP 邮件
- Tailwind Dashboard
- 中文文档和中文 UI

更详细的中文说明请看：

- [README.zh-CN.md](README.zh-CN.md)
- [docs/api.zh-CN.md](docs/api.zh-CN.md)
- [docs/architecture.zh-CN.md](docs/architecture.zh-CN.md)
- [docs/design-decisions.zh-CN.md](docs/design-decisions.zh-CN.md)
