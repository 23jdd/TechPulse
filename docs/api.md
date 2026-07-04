# API

Base URL: `http://localhost:8080`

- `GET /health` returns gateway and Redis cache status
- `POST /api/v1/rss` with `{"url":"https://go.dev/blog/feed.atom","title":"Go Blog","category":"Go"}`
- `GET /api/v1/rss`
- `POST /api/v1/rss/{id}/fetch`
- `POST /api/v1/github/releases/fetch` with `{"url":"https://github.com/golang/go"}`
- `GET /api/v1/articles`
- `GET /api/v1/search?q=go&page=1&page_size=20`
- `GET /api/v1/search/explain?q=go`
- `POST /api/v1/search/reindex`
- `POST /api/v1/chat` with `{"question":"What is new in Go?","conversation_id":1}`
- `GET /api/v1/dashboard`
- `GET /ws`

User features:

- `POST /api/v1/articles/{id}/read`
- `POST /api/v1/articles/{id}/read-later`
- `DELETE /api/v1/articles/{id}/read-later`
- `GET /api/v1/favorites?type=favorite`
- `GET /api/v1/favorites?type=read_later`
- `GET /api/v1/reading-history`
- `GET /api/v1/conversations`
- `GET /api/v1/prompts`
- `POST /api/v1/prompts` with `{"name":"release analyst","content":"Focus on migrations.","is_default":true}`
- `DELETE /api/v1/prompts/{id}`
- `GET /api/v1/opml`
- `POST /api/v1/opml` with an OPML XML request body
- `GET /api/v1/auth/github/url`
- `GET /api/v1/auth/github/callback?code=...&state=...`
- `POST /api/v1/email/test` with `{"to":"you@example.com","subject":"TechPulse","body":"SMTP is working"}`
- `POST /api/v1/daily-reports` with `{"title":"Today AI"}`
- `POST /api/v1/daily-reports` with `{"title":"Today AI","send_email":true,"email_to":"you@example.com"}`
- `GET /api/v1/daily-reports`

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
