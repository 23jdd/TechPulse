# Design Decisions

This document explains the tradeoffs behind TechPulse. The goal is to show the main engineering reasoning, not only the list of technologies.

## Why Start With One Gateway Process

The repository is organized as service packages from day one, but the MVP path runs in the gateway process.

Reasons:

- The core demo must be easy to run locally.
- It avoids RabbitMQ/etcd operational friction while the RSS -> AI -> Search -> RAG path is still being hardened.
- The package boundaries are already split, so moving calls behind HTTP or RabbitMQ is incremental.

The independent service entrypoints under `cmd/` are kept runnable to demonstrate the future decomposition.

## Why Bleve First

Bleve is embedded, Go-native, and does not require a separate search cluster.

This is a good MVP fit because:

- local setup is lighter than Elasticsearch or OpenSearch
- indexing is fast enough for a developer knowledge base MVP
- it supports fielded search, boosting, filters, highlighting, and pagination
- it keeps CI and local demo simple

Future replacement is isolated behind `SearchEngine`.

## Why Title Boost Is Higher

For technical articles, title matches usually indicate stronger intent than body matches. A query for `Go release` should rank "Go 1.23 Release Notes" above an article that mentions Go once in the body.

Current ranking:

- title: highest boost
- summary: useful compressed meaning
- content: broad recall
- tags: topical filtering and recall

## Why Store Embeddings As JSON Text

The MVP stores vectors in MySQL as JSON text.

Tradeoff:

- simple schema
- easy debugging
- no vector database requirement
- enough for persistence and demo

Future improvement:

- move embeddings to pgvector, Milvus, Qdrant, Vespa, or an ANN index
- keep MySQL as the article metadata source of truth

## Why MockAIProvider Exists

The project must run without API keys.

MockAIProvider makes the demo deterministic and cheap:

- tests can run offline
- CI does not need secrets
- interviewers can run the project immediately

Real providers are still supported through the OpenAI-compatible interface and Ollama-compatible `/v1` endpoints.

## Why RabbitMQ And etcd Are Partial In MVP

The system is designed for distributed services, but the strongest value comes from the main ingestion/search/RAG chain. RabbitMQ and etcd are included as real clients and service skeletons, but the gateway path remains in-process until the workflow is stable.

Current role:

- RabbitMQ: async job transport for scheduler/worker services
- etcd: service registration and distributed locks

Future role:

- scheduler publishes feed tasks
- fetcher consumes fetch jobs and publishes parse jobs
- parser publishes AI jobs
- AI pipeline publishes index jobs
- search worker indexes documents

## How It Can Split Into Microservices

The eventual split is straightforward:

```text
gateway -> fetcher / parser / ai-pipeline / search / rag over HTTP
scheduler -> RabbitMQ fetch jobs
worker -> consumes fetch/parse/ai/index/daily_report queues
etcd -> service discovery and global locks
```

The gateway can switch from direct package calls to `internal/service.Client` calls without changing API handlers heavily.

## What Is Intentionally Not Finished

- GitHub Releases are implemented. Reddit, Arxiv, YouTube, and HackerNews are still extension stubs.
- GitHub OAuth callback is planned, not a full auth system.
- Observability is Prometheus-ready, not a complete tracing deployment.
- Hybrid search reranking is simple and designed as a stepping stone toward a vector index.
