# Database

The MVP uses MySQL and creates these tables automatically on gateway startup or through `make migrate`:

- `users`
- `rss_feeds`
- `articles`
- `tags`
- `article_tags`
- `favorites`
- `summaries`
- `translations`
- `embeddings`
- `tasks`
- `conversations`
- `messages`
- `daily_reports`

Embeddings are stored as JSON text for Phase 1.
