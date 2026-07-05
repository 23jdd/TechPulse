# API

基础地址：`http://localhost:8080`

- `GET /health` 返回 gateway 和 Redis 缓存状态。
- `GET /login` 返回英文登录页。
- `GET /login/zh` 返回中文登录页。
- `GET /app` 返回英文用户应用。
- `GET /app/zh` 返回中文用户应用。

## 订阅与抓取

- `POST /api/v1/rss`，请求体：`{"url":"https://go.dev/blog/feed.atom","title":"Go Blog","category":"Go","fetch_interval_minutes":360}`
- `GET /api/v1/rss`
- `PUT /api/v1/rss/{id}`，请求体：`{"url":"...","title":"...","category":"Go","status":"active","fetch_interval_minutes":120}`
- `POST /api/v1/rss/{id}/enable`
- `POST /api/v1/rss/{id}/disable`
- `POST /api/v1/rss/{id}/test`
- `POST /api/v1/rss/{id}/fetch` 同步抓取，返回 fetched / inserted / duplicates。
- `POST /api/v1/rss/{id}/fetch-async` 异步抓取，优先投递 RabbitMQ fetch job 给 worker，返回 `202 Accepted` 和 `task_id`；RabbitMQ 不可用时 gateway 回退为进程内后台任务。
- Feed 返回字段包含 `health_status`、`health_score`、`consecutive_failures`、`last_error`、`last_duration_ms`、`last_checked_at`，连续失败 3 次会自动停用订阅。

## GitHub 与 Hacker News

- `POST /api/v1/github/releases/fetch`，请求体：`{"url":"https://github.com/golang/go"}`。
- `GET /api/v1/github/repos` 查看当前用户监控的 GitHub 仓库。
- `POST /api/v1/github/repos`，请求体：`{"url":"https://github.com/golang/go"}`，会记录 stars、open issues、latest release、breaking change 提示和 security 提示。
- `POST /api/v1/hackernews/fetch`，请求体：`{"feed":"top","limit":20}`，`feed` 支持 `top`、`new`、`best`、`ask`、`show`、`job`。

## 文章与搜索

- `GET /api/v1/articles?tag=Go&source=rss&read=false&favorite=true&archived=false&from=2026-07-01&to=2026-07-04`
- `GET /api/v1/search?q=go&page=1&page_size=20`
- `GET /api/v1/search?tag=Go&source=github_release&from=2026-07-01&to=2026-07-04`
- `GET /api/v1/search/explain?q=go`
- `POST /api/v1/search/reindex`
- `GET /api/v1/trends?days=7` 查询热门标签、来源统计和每日文章数量。

## RAG 问答

- `POST /api/v1/chat`，请求体：`{"question":"What is new in Go?","conversation_id":1}`。
- 返回字段包含 `answer`、`citations`、重写后的 `query`、`confidence` 和 `no_answer`。
- 当知识库没有命中时，`no_answer=true`，不会强行编答案。

## 用户功能

- `GET /api/v1/me`
- `GET /api/v1/preferences`
- `PUT /api/v1/preferences`，请求体：`{"interested_tags":["Go","AI"],"daily_report_time":"09:00","daily_report_email":"you@example.com","daily_report_enabled":true,"timezone":"Asia/Shanghai"}`
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
- `POST /api/v1/opml`，请求体为 OPML XML。

## 登录与邮件

- `GET /api/v1/auth/github/url`
- `GET /api/v1/auth/github/url?ui=1` 返回浏览器登录使用的 GitHub OAuth URL。
- `GET /api/v1/auth/github/callback?code=...&state=...`，JSON 响应包含 `session_token`；UI 登录会保存为 Bearer JWT。
- `POST /api/v1/email/test`，请求体：`{"to":"you@example.com","subject":"TechPulse","body":"SMTP is working"}`
- `POST /api/v1/daily-reports`，请求体：`{"title":"Today AI"}`
- `POST /api/v1/daily-reports`，请求体：`{"title":"Today AI","send_email":true,"email_to":"you@example.com"}`
- `GET /api/v1/daily-reports`
- 设置偏好中的 `daily_report_enabled=true`、`daily_report_email` 和 `daily_report_time` 后，gateway 内置调度器会按用户时区定时生成并发送日报。

## 任务与实时事件

- `GET /api/v1/tasks?page_size=20&status=running`
- `GET /api/v1/tasks/{id}`
- `GET /ws`

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
