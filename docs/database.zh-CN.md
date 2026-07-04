# 数据库

MVP 使用 MySQL。gateway 启动或执行 `make migrate` 时会自动创建这些表：

- `users`
- `rss_feeds`
- `articles`
- `tags`
- `article_tags`
- `favorites`
- `reading_history`
- `summaries`
- `translations`
- `embeddings`
- `tasks`
- `conversations`
- `messages`
- `daily_reports`

Phase 1 中 embedding 以 JSON 文本形式存储。

`rss_feeds.fetch_interval_minutes` 保存每个 feed 的抓取频率。用户级文章状态通过 `favorites.type` 建模，例如 `favorite`、`read_later`、`archived`。阅读事件存储在 `reading_history`。
