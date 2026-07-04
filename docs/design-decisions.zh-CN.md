# 设计决策

本文解释 TechPulse 背后的工程取舍。重点不只是列技术栈，而是说明为什么这样设计。

## 为什么从单 Gateway 进程开始

仓库从第一天起就按服务包组织，但 MVP 主链路运行在 gateway 进程里。

原因：

- 核心 demo 必须容易在本地运行
- 在 RSS -> AI -> Search -> RAG 主链路稳定前，避免 RabbitMQ/etcd 运维成本
- 包边界已经拆好，未来迁移到 HTTP 或 RabbitMQ 是增量工作

`cmd/` 下的独立服务入口仍保持可运行，用于展示未来服务拆分方向。

## 为什么先用 Bleve

Bleve 是嵌入式、Go 原生搜索引擎，不需要独立搜索集群。

它适合 MVP，因为：

- 本地启动比 Elasticsearch/OpenSearch 更轻
- 对开发者知识库 MVP 来说索引性能足够
- 支持字段搜索、权重、过滤、高亮、分页
- CI 和本地演示更简单

未来替换会隔离在 `SearchEngine` 接口之后。

## 为什么标题权重更高

技术文章中，标题命中通常更能表达用户意图。比如搜索 `Go release` 时，标题为 "Go 1.23 Release Notes" 的文章应该排在正文里偶然提到 Go 的文章前面。

当前排序思路：

- title：最高权重
- summary：压缩后的语义信息
- content：广泛召回
- tags：主题过滤和召回

## 为什么 embedding 先存 JSON 文本

MVP 把向量以 JSON 文本存入 MySQL。

取舍：

- schema 简单
- 方便调试
- 不要求向量数据库
- 足够支持持久化和 demo

后续可以迁移到 pgvector、Milvus、Qdrant、Vespa 或 ANN index，同时继续让 MySQL 作为文章元数据事实源。

## 为什么需要 MockAIProvider

项目必须在没有 API Key 的情况下运行。

MockAIProvider 让 demo 可确定、低成本：

- 测试可以离线运行
- CI 不需要 secrets
- 面试官可以立刻跑起来

真实 Provider 仍通过 OpenAI-compatible 接口和 Ollama-compatible `/v1` endpoint 支持。

## 为什么 RabbitMQ 和 etcd 在 MVP 中是部分完成

系统设计目标是分布式服务，但最核心价值来自采集、搜索、RAG 主链路。RabbitMQ 和 etcd 已作为真实客户端和服务骨架加入，但 gateway 路径在工作流稳定前保持进程内运行。

当前角色：

- RabbitMQ：scheduler/worker 的异步任务传输
- etcd：服务注册和分布式锁

未来角色：

- scheduler 发布 feed 任务
- fetcher 消费 fetch job 并发布 parse job
- parser 发布 AI job
- AI pipeline 发布 index job
- search worker 写索引

## 如何拆成微服务

最终拆分路径很直接：

```text
gateway -> fetcher / parser / ai-pipeline / search / rag over HTTP
scheduler -> RabbitMQ fetch jobs
worker -> consumes fetch/parse/ai/index/daily_report queues
etcd -> service discovery and global locks
```

gateway 可以从直接包调用切换到 `internal/service.Client`，API handler 不需要大改。

## 有意未完成的部分

- GitHub Releases 已实现。Reddit、Arxiv、YouTube、HackerNews 仍是扩展 stub。
- GitHub OAuth callback 已支持用户 upsert，但 session/JWT enforcement 留到后续增强。
- SMTP 邮件发送已支持报告和测试邮件，但模板、退订、偏好设置是后续工作。
- Observability 已具备 Prometheus-ready 结构，还不是完整 tracing 部署。
- Hybrid search reranking 目前较简单，主要作为未来向量索引的过渡。
