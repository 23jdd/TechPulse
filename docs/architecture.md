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
  Gateway --> Dashboard
```

Phase 1 runs the real MVP path in the gateway process. The service packages are split so Phase 2 can move scheduler, fetcher, parser, AI pipeline, search, RAG, and worker behind HTTP or RabbitMQ boundaries.
