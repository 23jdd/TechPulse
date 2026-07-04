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
