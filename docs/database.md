# Database

The MVP uses MySQL and creates these tables automatically on gateway startup or through `make migrate`:

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

Embeddings are stored as JSON text for Phase 1.

`rss_feeds.fetch_interval_minutes` stores the preferred crawl cadence for each feed. Per-user article state is modeled through `favorites.type` values such as `favorite`, `read_later`, and `archived`, while read events are stored in `reading_history`.
