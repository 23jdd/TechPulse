# Architecture

TechPulse is organized as a Go monorepo with independently runnable service entrypoints under `cmd/`.

```mermaid
flowchart LR
  Scheduler --> Fetcher
  Fetcher --> Parser
  Parser --> Duplicate
  Duplicate --> Pipeline
  Pipeline --> Search
  Pipeline --> MySQL
  Search --> Gateway
  Gateway --> RAG
  RAG --> Search
  Gateway --> WebSocket
  Gateway --> UserApp
```

Phase 1 runs the real MVP path in the gateway process. Phase 2 exposes scheduler, fetcher, parser, AI pipeline, search, RAG, and worker as independently runnable services.

Service communication:

- HTTP: fetcher `/fetch`, parser `/parse`, ai-pipeline `/process`, search `/index` and `/search`, rag `/chat`.
- RabbitMQ: `fetch`, `parse`, `ai`, `index`, and `daily_report` queues.
- etcd: service registration under `/techpulse/services/*` and distributed locks under `/techpulse/locks/*`.
- Redis: best-effort cache for hot REST responses and service status. Gateway automatically falls back to MySQL/Bleve when Redis is unavailable.
- Hybrid retrieval: Bleve provides lexical recall, then the AI provider embedding API reranks top hits when configured.
- RAG memory: conversations and messages are stored in MySQL and recent turns are included in generation prompts.
