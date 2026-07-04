You are a senior Go backend architect and full-stack engineer.

Your task is to generate a real, runnable production-style project called **TechPulse**.

TechPulse is an AI-powered developer intelligence platform. It continuously collects technical content from RSS, GitHub, HackerNews, Reddit, Arxiv, YouTube and other sources, then uses AI to clean, deduplicate, summarize, translate, tag, index, search, and answer questions over the collected articles.

Do not only explain the architecture.
Do not only give pseudo-code.
Generate an actual project with real Go code, real directory structure, real APIs, real database models, real Docker Compose configuration, real Makefile commands, and a minimal working MVP.

The project must be written mainly in Go.

---

# 1. Main Goal

Build TechPulse as a long-running developer knowledge platform.

It should be more than an RSS reader. It should become an AI Knowledge Hub for developers.

It should support collecting and organizing:

* Tech news
* GitHub releases
* AI news
* Go community updates
* Linux
* Kubernetes
* Databases
* Networking
* Cloud Native
* Security
* Programming language updates

---

# 2. Required Architecture

Design the system as a microservice-style project.

Each major module should be independently runnable and independently deployable.

Use a monorepo layout.

Services communicate through HTTP.

Use:

* Go
* MySQL
* Redis
* RabbitMQ
* etcd
* Bleve search engine
* MinIO object storage
* OpenAI-compatible AI provider
* Docker Compose
* Makefile
* OpenTelemetry-ready structure
* Prometheus-ready metrics structure

Required services:

* gateway
* scheduler
* fetcher
* parser
* ai-pipeline
* search
* rag
* worker

The architecture should follow this flow:

Scheduler
→ Fetcher
→ Parser
→ Duplicate Detection
→ AI Pipeline
→ Search Index
→ MySQL / Redis / MinIO
→ REST API / WebSocket API
→ Dashboard / AI Chat

Use etcd for:

* service discovery
* global config
* distributed lock

Use RabbitMQ for:

* async article fetching jobs
* parsing jobs
* AI processing jobs
* indexing jobs
* daily report jobs

---

# 3. Important Rule

Generate an MVP first.

The MVP must be runnable with:

```bash
make dev
```

or:

```bash
docker compose up -d
make run
```

The MVP does not need to implement every advanced feature fully, but it must contain a real working path:

1. User adds RSS feed
2. Scheduler creates fetch task
3. Fetcher fetches RSS articles
4. Parser extracts article title, URL, content, published time
5. Deduplicate engine checks duplicate by URL hash and content hash
6. AI pipeline generates mock or real summary, tags, keywords, translation and embedding
7. Article is stored in MySQL
8. Article is indexed into Bleve
9. User can search articles through REST API
10. User can ask a RAG question over stored articles
11. WebSocket sends task status updates

---

# 4. Tech Stack

Use the following stack:

## Backend

* Go
* Gin or Chi for HTTP router
* sqlx or GORM for MySQL
* Redis
* RabbitMQ
* etcd
* Bleve
* MinIO
* zap or slog for logging
* viper or envconfig for config
* OpenTelemetry-ready interfaces

## AI Provider

Support OpenAI-compatible API.

Create a pluggable AI provider interface.

Support providers:

* OpenAI
* Gemini
* Claude
* DeepSeek
* Ollama
* OpenRouter

For MVP, implement:

* MockAIProvider
* OpenAICompatibleProvider

The AI interface must support:

* Summarize
* Translate
* GenerateTags
* GenerateKeywords
* GenerateEmbedding
* ChatCompletion

---

# 5. Required Project Directory

Generate this structure:

```text
techpulse/
  cmd/
    gateway/
      main.go
    scheduler/
      main.go
    fetcher/
      main.go
    parser/
      main.go
    ai-pipeline/
      main.go
    search/
      main.go
    rag/
      main.go
    worker/
      main.go

  internal/
    api/
      handler/
      middleware/
      router/
      dto/

    scheduler/
      service.go
      lock.go
      job.go

    fetcher/
      service.go
      rss.go
      github.go
      hackernews.go
      reddit.go
      arxiv.go
      youtube.go
      plugin.go

    parser/
      service.go
      rss_parser.go
      html_parser.go
      markdown_parser.go
      pdf_parser.go
      cleaner.go

    duplicate/
      service.go
      hash.go
      simhash.go

    pipeline/
      service.go
      processor.go
      steps.go

    ai/
      provider.go
      mock.go
      openai_compatible.go
      prompt.go

    search/
      engine.go
      bleve.go
      query.go
      indexer.go

    rag/
      service.go
      retriever.go
      generator.go
      citation.go

    storage/
      mysql/
        db.go
        migration.go
        repository.go
      redis/
        client.go
      minio/
        client.go

    queue/
      rabbitmq.go
      message.go
      producer.go
      consumer.go

    discovery/
      etcd.go
      registry.go
      config.go

    websocket/
      hub.go
      client.go
      message.go

    model/
      user.go
      rss.go
      article.go
      tag.go
      favorite.go
      embedding.go
      task.go
      summary.go
      translation.go
      conversation.go
      daily_report.go

    config/
      config.go

    observability/
      logger.go
      metrics.go
      tracing.go

  pkg/
    errors/
    hash/
    httpclient/
    pagination/
    response/
    validator/

  web/
    README.md

  docs/
    architecture.md
    api.md
    database.md
    deployment.md
    roadmap.md

  deploy/
    docker-compose.yml
    mysql/
      init.sql
    prometheus/
      prometheus.yml
    grafana/
      README.md

  scripts/
    migrate.sh
    seed.sh

  Makefile
  go.mod
  go.sum
  README.md
  .env.example
  .gitignore
```

---

# 6. Database Design

Use MySQL.

Create migrations or init SQL for these tables:

## users

Fields:

* id
* github_id
* username
* email
* avatar_url
* api_token
* created_at
* updated_at

## rss_feeds

Fields:

* id
* user_id
* url
* title
* category
* status
* last_fetched_at
* created_at
* updated_at

## articles

Fields:

* id
* source_type
* source_id
* title
* url
* url_hash
* content_hash
* author
* language
* raw_content
* clean_content
* cover_image
* published_at
* created_at
* updated_at

## tags

Fields:

* id
* name
* type
* created_at

## article_tags

Fields:

* article_id
* tag_id

## favorites

Fields:

* id
* user_id
* article_id
* type
* created_at

## summaries

Fields:

* id
* article_id
* one_sentence
* short_summary
* long_summary
* bullet_points
* tldr
* language
* created_at

## translations

Fields:

* id
* article_id
* target_language
* translated_title
* translated_content
* created_at

## embeddings

Fields:

* id
* article_id
* provider
* model
* vector
* dimension
* created_at

For MVP, vector can be stored as JSON text.

## tasks

Fields:

* id
* type
* status
* payload
* retry_count
* error_message
* scheduled_at
* started_at
* finished_at
* created_at
* updated_at

## conversations

Fields:

* id
* user_id
* title
* created_at
* updated_at

## messages

Fields:

* id
* conversation_id
* role
* content
* citations
* created_at

## daily_reports

Fields:

* id
* user_id
* title
* content
* report_date
* created_at

---

# 7. API Requirements

Implement REST APIs.

## Health

```http
GET /health
```

Returns service health.

## RSS

```http
POST /api/v1/rss
GET /api/v1/rss
GET /api/v1/rss/:id
DELETE /api/v1/rss/:id
POST /api/v1/rss/:id/fetch
```

## Articles

```http
GET /api/v1/articles
GET /api/v1/articles/:id
GET /api/v1/articles/:id/summary
POST /api/v1/articles/:id/favorite
DELETE /api/v1/articles/:id/favorite
```

## Search

```http
GET /api/v1/search?q=golang&tag=Go&page=1&page_size=20
```

Support:

* title search
* content search
* tag search
* author filter
* date filter
* highlight result

## AI

```http
POST /api/v1/summary
POST /api/v1/translate
POST /api/v1/tags
```

## RAG

```http
POST /api/v1/chat
```

Request:

```json
{
  "question": "最近 Go 社区有什么重要更新？",
  "conversation_id": 1
}
```

Response:

```json
{
  "answer": "根据最近的文章，Go 社区主要有这些更新...",
  "citations": [
    {
      "article_id": 1,
      "title": "Go 1.23 Release Notes",
      "url": "https://go.dev/blog/..."
    }
  ]
}
```

## Dashboard

```http
GET /api/v1/dashboard
```

Return:

* today new articles
* week new articles
* popular tags
* failed tasks
* worker status
* AI token cost placeholder

## WebSocket

```http
GET /ws
```

Send real-time events:

* fetch_started
* fetch_finished
* parse_finished
* ai_finished
* index_finished
* new_article
* task_failed
* worker_status

---

# 8. Fetcher Requirements

Implement a plugin-based fetcher system.

Create this interface:

```go
type Fetcher interface {
    Name() string
    Supports(sourceType string) bool
    Fetch(ctx context.Context, source Source) ([]FetchedItem, error)
}
```

Implement at least:

* RSSFetcher
* GitHubReleaseFetcher stub
* HackerNewsFetcher stub
* RedditFetcher stub
* ArxivFetcher stub
* YouTubeFetcher stub

RSSFetcher must really work.

It should support:

* RSS
* Atom
* JSON Feed if possible

Use real HTTP client with timeout, retry and user-agent.

---

# 9. Parser Requirements

Create a parser engine.

Parser should extract:

* title
* author
* publish time
* raw content
* clean content
* language
* image
* tags

For MVP, implement:

* RSS item parser
* simple HTML cleaner

Create parser interfaces so HTML, Markdown and PDF parser can be added later.

---

# 10. Duplicate Detection Requirements

Implement duplicate detection by:

* URL hash
* content hash
* SimHash stub
* Embedding similarity stub

For MVP, URL hash and content hash must work.

If duplicate article already exists, skip insert and log reason.

---

# 11. AI Pipeline Requirements

Create pipeline steps:

1. Clean
2. Language Detect
3. Translate
4. Summary
5. Keywords
6. Tags
7. Embedding
8. Store
9. Index

Each step should be independent.

Create interfaces:

```go
type Step interface {
    Name() string
    Execute(ctx context.Context, input *PipelineInput) error
}
```

The pipeline should be extensible.

For MVP:

* Mock summary works without external API
* Mock tags work
* Mock keywords work
* Mock embedding works
* OpenAI-compatible provider can be enabled by config

---

# 12. Search Requirements

Use Bleve for MVP.

Implement:

* index article
* delete article
* search article
* highlight
* pagination
* query by title/content/tag

Create clean interface:

```go
type SearchEngine interface {
    IndexArticle(ctx context.Context, article ArticleSearchDocument) error
    DeleteArticle(ctx context.Context, articleID int64) error
    Search(ctx context.Context, query SearchQuery) (*SearchResult, error)
}
```

Prepare future replacement by self-made inverted index.

---

# 13. RAG Requirements

Implement simple RAG:

1. Receive question
2. Search related articles using Bleve
3. Select top 5 articles
4. Build prompt with citations
5. Call AI provider
6. Return answer and sources

For MVP, if AI provider is mock, generate a simple answer using retrieved article summaries.

RAG answer must include citations.

---

# 14. Scheduler Requirements

Scheduler should support intervals:

* 5 min
* 10 min
* 30 min
* 1 hour

Use etcd distributed lock.

Use RabbitMQ to publish fetch jobs.

Support:

* retry
* backoff
* worker pool
* task status

For MVP, implement a simple ticker scheduler and manual fetch API.

---

# 15. User System Requirements

Implement basic user model.

GitHub OAuth should be designed.

For MVP:

* Provide GitHub OAuth config structure
* Provide auth middleware structure
* Provide mock user or API token auth

Later support:

* GitHub OAuth login
* user RSS management
* favorite articles
* user prompts

---

# 16. Daily Report Requirements

Design daily report system.

Generate:

* Morning Brief
* Today Go
* Today AI
* Today Linux
* Today Kubernetes
* Today Database

For MVP, provide service skeleton and API-ready design.

Future delivery channels:

* Email
* Telegram
* Discord
* Webhook

---

# 17. Non-Functional Requirements

The code should be:

* modular
* readable
* testable
* benchmark-ready
* production-style
* not over-engineered
* easy to run locally

Add support structure for:

* logging
* metrics
* tracing
* config hot reload placeholder
* graceful shutdown
* context cancellation
* error wrapping
* retries
* worker pool
* service registry
* distributed lock
* Docker
* Kubernetes-ready deployment

---

# 18. Configuration

Create `.env.example` with:

```env
APP_ENV=dev
HTTP_PORT=8080

MYSQL_DSN=root:password@tcp(localhost:3306)/techpulse?parseTime=true&charset=utf8mb4
REDIS_ADDR=localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
ETCD_ENDPOINTS=localhost:2379
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=techpulse

AI_PROVIDER=mock
AI_BASE_URL=https://api.openai.com/v1
AI_API_KEY=
AI_MODEL=gpt-4o-mini

BLEVE_INDEX_PATH=./data/bleve
```

---

# 19. Docker Compose

Create Docker Compose with:

* mysql
* redis
* rabbitmq
* etcd
* minio
* gateway
* scheduler
* fetcher
* parser
* ai-pipeline
* search
* rag
* worker

For MVP, it is acceptable that all Go services use the same image with different command args.

---

# 20. Makefile

Create Makefile commands:

```makefile
make dev
make run
make build
make test
make lint
make docker-up
make docker-down
make migrate
make seed
make clean
```

---

# 21. README Requirements

README must include:

* project introduction
* architecture diagram
* features
* tech stack
* how to run
* environment variables
* API examples
* development roadmap
* service explanation
* database explanation
* future improvements

---

# 22. Testing Requirements

Add basic tests for:

* RSS fetcher
* parser
* duplicate hash
* search engine
* AI mock provider
* RAG retrieval

Add benchmark placeholder for:

* parser
* search
* duplicate detection

---

# 23. Output Format

Generate the project step by step.

First output:

1. Final architecture
2. Directory tree
3. Database schema
4. API design
5. MVP execution path

Then generate real code files.

When generating code, always use this format:

```text
// file: path/to/file.go
```

followed by the complete file content.

Do not skip important files.

Do not write “left as exercise”.

Do not use fake imports that cannot compile.

Do not write only fragments.

If a feature is too large for MVP, implement a working simplified version and add TODO comments for future extension.

---

# 24. MVP Acceptance Criteria

The MVP is accepted only if:

1. `go mod tidy` works
2. `go test ./...` works
3. `make run` starts gateway successfully
4. Docker Compose starts MySQL, Redis, RabbitMQ, etcd and MinIO
5. User can add RSS feed by API
6. User can trigger RSS fetch by API
7. RSS articles are stored into MySQL
8. Articles are indexed into Bleve
9. Search API returns results
10. RAG chat API returns an answer with citations
11. WebSocket endpoint accepts connection
12. Logs are readable
13. README explains how to run everything

---

# 25. Development Priority

Implement in this order:

## Phase 1: Runnable MVP

* project skeleton
* config
* logger
* MySQL connection
* models
* repositories
* RSS API
* RSS fetcher
* parser
* duplicate detection
* mock AI pipeline
* Bleve index
* search API
* simple RAG API
* WebSocket hub
* Docker Compose
* Makefile
* README

## Phase 2: Microservice Split

* scheduler service
* fetcher service
* parser service
* ai-pipeline service
* search service
* rag service
* worker service
* RabbitMQ messaging
* etcd service discovery

## Phase 3: Advanced AI

* OpenAI-compatible provider
* real translation
* real summary
* real embedding
* hybrid search
* better RAG prompt
* citations
* conversation memory

## Phase 4: User Features

* GitHub OAuth
* favorites
* read later
* reading history
* user prompts
* OPML import/export

## Phase 5: Production

* metrics
* tracing
* Prometheus
* Grafana
* Kubernetes manifests
* CI/CD
* benchmark
* stress test
* config hot reload

---

# 26. Coding Style

Use clean Go style.

Prefer:

* small interfaces
* dependency injection
* context.Context
* structured logging
* repository pattern
* service layer
* clear error handling
* graceful shutdown
* no global mutable state unless necessary

Avoid:

* giant files
* hard-coded config
* circular dependencies
* unnecessary complexity
* pseudo-code
* incomplete functions

---

# 27. Important Implementation Detail

For MVP, you may start with one gateway process that directly calls internal services in-process.

But the code must be organized so that later each module can become an independent microservice.

That means:

* keep service packages independent
* define HTTP clients for future service-to-service calls
* define queue message types
* define service registry abstraction
* avoid tight coupling

---

# 28. Example RSS Feeds for Seed Data

Use these default feeds:

```text
https://go.dev/blog/feed.atom
https://kubernetes.io/feed.xml
https://hnrss.org/frontpage
https://github.blog/feed/
```

Create seed script to insert them.

---

# 29. Final Instruction

Start now.

Generate a real TechPulse repository.

Do not only describe it.

First generate the architecture and file tree.

Then generate the MVP code files in order.

Make sure the code can compile and the project can run locally.

Start with Phase 1 only. Generate complete runnable MVP code. Do not generate Phase 2 until I say continue.