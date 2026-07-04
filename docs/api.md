# API

Base URL: `http://localhost:8080`

- `GET /health`
- `POST /api/v1/rss` with `{"url":"https://go.dev/blog/feed.atom","title":"Go Blog","category":"Go"}`
- `GET /api/v1/rss`
- `POST /api/v1/rss/{id}/fetch`
- `GET /api/v1/articles`
- `GET /api/v1/search?q=go&page=1&page_size=20`
- `POST /api/v1/chat` with `{"question":"What is new in Go?","conversation_id":1}`
- `GET /api/v1/dashboard`
- `GET /ws`

## Phase 2 Service APIs

Fetcher service, port `8082`:

- `POST /fetch` with `{"source":{"id":1,"type":"rss","url":"https://go.dev/blog/feed.atom"}}`

Parser service, port `8083`:

- `POST /parse` with `{"item":{...fetched item...}}`

AI pipeline service, port `8084`:

- `POST /process` with `{"article":{...parsed article...}}`

Search service, port `8085`:

- `POST /index` with `{"document":{...search document...}}`
- `DELETE /index/{id}`
- `GET /search?q=go&page=1&page_size=20`
- `POST /search` with `{"query":{"query":"go","page":1,"page_size":20}}`

RAG service, port `8086`:

- `POST /chat` with `{"question":"What is new in Go?"}`

Scheduler service, port `8081`:

- `POST /schedule/fetch?feed_id=1`
- `POST /tick`
