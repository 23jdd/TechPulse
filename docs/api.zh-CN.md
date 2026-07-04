# API

基础地址：`http://localhost:8080`

- `GET /health` 返回 gateway 和 Redis 缓存状态
- `POST /api/v1/rss`，请求体：`{"url":"https://go.dev/blog/feed.atom","title":"Go Blog","category":"Go","fetch_interval_minutes":360}`
- `GET /api/v1/rss`
- `PUT /api/v1/rss/{id}`，请求体：`{"url":"...","title":"...","category":"Go","status":"active","fetch_interval_minutes":120}`
- `POST /api/v1/rss/{id}/enable`
- `POST /api/v1/rss/{id}/disable`
- `POST /api/v1/rss/{id}/test`
- `POST /api/v1/rss/{id}/fetch`
- `POST /api/v1/github/releases/fetch`，请求体：`{"url":"https://github.com/golang/go"}`
- `GET /api/v1/articles?tag=Go&source=rss&read=false&favorite=true&archived=false&from=2026-07-01&to=2026-07-04`
- `GET /api/v1/search?q=go&page=1&page_size=20`
- `GET /api/v1/search?tag=Go&source=github_release&from=2026-07-01&to=2026-07-04`
- `GET /api/v1/search/explain?q=go`
- `POST /api/v1/search/reindex`
- `POST /api/v1/chat`，请求体：`{"question":"What is new in Go?","conversation_id":1}`
- `GET /api/v1/dashboard`
- `GET /ws`

## 用户功能

- `POST /api/v1/articles/{id}/read`
- `POST /api/v1/articles/{id}/favorite`
- `DELETE /api/v1/articles/{id}/favorite`
- `POST /api/v1/articles/{id}/read-later`
- `DELETE /api/v1/articles/{id}/read-later`
- `POST /api/v1/articles/{id}/archive`
- `DELETE /api/v1/articles/{id}/archive`
- `DELETE /api/v1/articles/{id}`
- `GET /api/v1/favorites?type=favorite`
- `GET /api/v1/favorites?type=read_later`
- `GET /api/v1/reading-history`
- `GET /api/v1/conversations`
- `GET /api/v1/prompts`
- `POST /api/v1/prompts`，请求体：`{"name":"release analyst","content":"Focus on migrations.","is_default":true}`
- `DELETE /api/v1/prompts/{id}`
- `GET /api/v1/opml`
- `POST /api/v1/opml`，请求体为 OPML XML
- `GET /api/v1/auth/github/url`
- `GET /api/v1/auth/github/callback?code=...&state=...`
- `POST /api/v1/email/test`，请求体：`{"to":"you@example.com","subject":"TechPulse","body":"SMTP is working"}`
- `POST /api/v1/daily-reports`，请求体：`{"title":"Today AI"}`
- `POST /api/v1/daily-reports`，请求体：`{"title":"Today AI","send_email":true,"email_to":"you@example.com"}`
- `GET /api/v1/daily-reports`

## Phase 2 服务 API

Fetcher 服务，端口 `8082`：

- `POST /fetch`，请求体：`{"source":{"id":1,"type":"rss","url":"https://go.dev/blog/feed.atom"}}`

Parser 服务，端口 `8083`：

- `POST /parse`，请求体：`{"item":{...fetched item...}}`

AI pipeline 服务，端口 `8084`：

- `POST /process`，请求体：`{"article":{...parsed article...}}`

Search 服务，端口 `8085`：

- `POST /index`，请求体：`{"document":{...search document...}}`
- `DELETE /index/{id}`
- `GET /search?q=go&page=1&page_size=20`
- `POST /search`，请求体：`{"query":{"query":"go","page":1,"page_size":20}}`

RAG 服务，端口 `8086`：

- `POST /chat`，请求体：`{"question":"What is new in Go?"}`

Scheduler 服务，端口 `8081`：

- `POST /schedule/fetch?feed_id=1`
- `POST /tick`
