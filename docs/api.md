# API

Base URL: `http://localhost:8080`

- `GET /health` returns gateway and Redis cache status
- `GET /login` serves the English login page
- `GET /login/zh` serves the Chinese login page
- `POST /api/v1/rss` with `{"url":"https://go.dev/blog/feed.atom","title":"Go Blog","category":"Go","fetch_interval_minutes":360}`
- `GET /api/v1/rss`
- `PUT /api/v1/rss/{id}` with `{"url":"...","title":"...","category":"Go","status":"active","fetch_interval_minutes":120}`
- `POST /api/v1/rss/{id}/enable`
- `POST /api/v1/rss/{id}/disable`
- `POST /api/v1/rss/{id}/test`
- `POST /api/v1/rss/{id}/fetch` runs a synchronous fetch and returns fetched/inserted/duplicate counts
- `POST /api/v1/rss/{id}/fetch-async` queues a RabbitMQ fetch job for the worker and returns `202 Accepted` with `task_id`; gateway falls back to an in-process background task if RabbitMQ is unavailable
- `POST /api/v1/github/releases/fetch` with `{"url":"https://github.com/golang/go"}`
- `GET /api/v1/github/repos`
- `POST /api/v1/github/repos` with `{"url":"https://github.com/golang/go"}` to monitor stars, open issues, latest release, breaking-change hints, and security hints
- `POST /api/v1/hackernews/fetch` with `{"feed":"top","limit":20}`; `feed` supports `top`, `new`, `best`, `ask`, `show`, and `job`
- `GET /api/v1/articles?tag=Go&source=rss&read=false&favorite=true&archived=false&from=2026-07-01&to=2026-07-04`
- `GET /api/v1/search?q=go&page=1&page_size=20`
- `GET /api/v1/search?tag=Go&source=github_release&from=2026-07-01&to=2026-07-04`
- `GET /api/v1/search/explain?q=go`
- `POST /api/v1/search/reindex`
- `POST /api/v1/chat` with `{"question":"What is new in Go?","conversation_id":1}`
  - response includes `answer`, `citations`, rewritten `query`, `confidence`, and `no_answer`
- `GET /api/v1/dashboard`
- `GET /api/v1/tasks?page_size=20&status=running`
- `GET /api/v1/tasks/{id}`
- `GET /api/v1/trends?days=7`
- `GET /ws`

User features:

- `POST /api/v1/articles/{id}/read`
- `GET /api/v1/me`
- `GET /api/v1/preferences`
- `PUT /api/v1/preferences` with `{"interested_tags":["Go","AI"],"daily_report_time":"09:00","daily_report_email":"you@example.com","daily_report_enabled":true,"timezone":"Asia/Shanghai"}`
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
- `POST /api/v1/prompts` with `{"name":"release analyst","content":"Focus on migrations.","is_default":true}`
- `DELETE /api/v1/prompts/{id}`
- `GET /api/v1/opml`
- `POST /api/v1/opml` with an OPML XML request body
- `GET /api/v1/auth/github/url`
- `GET /api/v1/auth/github/url?ui=1` returns a GitHub OAuth URL for browser login
- `GET /api/v1/auth/github/callback?code=...&state=...`
  - JSON response includes `session_token`; UI login stores it as a Bearer JWT
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
