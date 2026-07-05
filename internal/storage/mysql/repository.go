package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"techpulse/internal/auth"
	"techpulse/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

type StoredArticle struct {
	Article   model.Article
	Summary   model.Summary
	Tags      []string
	Keywords  []string
	Embedding model.Embedding
}

type Dashboard struct {
	TodayNewArticles int64    `json:"today_new_articles"`
	WeekNewArticles  int64    `json:"week_new_articles"`
	PopularTags      []string `json:"popular_tags"`
	FailedTasks      int64    `json:"failed_tasks"`
	WorkerStatus     string   `json:"worker_status"`
	AITokenCost      string   `json:"ai_token_cost"`
}

type ArticleFilter struct {
	UserID      int64
	SourceType  string
	SourceID    int64
	Tag         string
	From        *time.Time
	To          *time.Time
	IsRead      *bool
	IsFavorite  *bool
	IsReadLater *bool
	IsArchived  *bool
	Limit       int
	Offset      int
}

type UserPrompt struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Content   string    `db:"content" json:"content"`
	IsDefault bool      `db:"is_default" json:"is_default"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) EnsureDefaultUser(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO users (id, username, email, api_token)
VALUES (1, 'demo', 'demo@techpulse.local', 'dev-token')
ON DUPLICATE KEY UPDATE username = VALUES(username)`)
	return err
}

func (r *Repository) UpsertGitHubUser(ctx context.Context, user auth.GitHubUser) (*model.User, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (github_id, username, email, avatar_url, api_token)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE username = VALUES(username), email = VALUES(email), avatar_url = VALUES(avatar_url), api_token = VALUES(api_token), updated_at = NOW()`,
		user.GitHubID, user.Username, user.Email, user.AvatarURL, user.APIToken)
	if err != nil {
		return nil, err
	}
	var stored model.User
	err = r.db.GetContext(ctx, &stored, `SELECT * FROM users WHERE github_id = ?`, user.GitHubID)
	return &stored, err
}

func (r *Repository) CreateFeed(ctx context.Context, feed *model.RSSFeed) error {
	interval := feed.FetchInterval
	if interval <= 0 {
		interval = 60
	}
	res, err := r.db.ExecContext(ctx, `INSERT INTO rss_feeds (user_id, url, title, category, status, fetch_interval_minutes) VALUES (?, ?, ?, ?, ?, ?)`,
		feed.UserID, feed.URL, feed.Title, feed.Category, defaultString(feed.Status, "active"), interval)
	if err != nil {
		return err
	}
	feed.ID, err = res.LastInsertId()
	return err
}

func (r *Repository) UpdateFeed(ctx context.Context, feed *model.RSSFeed) error {
	interval := feed.FetchInterval
	if interval <= 0 {
		interval = 60
	}
	_, err := r.db.ExecContext(ctx, `UPDATE rss_feeds
SET url = ?, title = ?, category = ?, status = ?, fetch_interval_minutes = ?, updated_at = NOW()
WHERE id = ? AND user_id = ?`,
		feed.URL, feed.Title, feed.Category, defaultString(feed.Status, "active"), interval, feed.ID, feed.UserID)
	return err
}

func (r *Repository) SetFeedStatus(ctx context.Context, userID, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE rss_feeds SET status = ?, updated_at = NOW() WHERE id = ? AND user_id = ?`, status, id, userID)
	return err
}

func (r *Repository) ListFeeds(ctx context.Context, userID int64) ([]model.RSSFeed, error) {
	var feeds []model.RSSFeed
	err := r.db.SelectContext(ctx, &feeds, `SELECT * FROM rss_feeds WHERE user_id = ? ORDER BY id DESC`, userID)
	return feeds, err
}

func (r *Repository) GetFeed(ctx context.Context, id int64) (*model.RSSFeed, error) {
	var feed model.RSSFeed
	if err := r.db.GetContext(ctx, &feed, `SELECT * FROM rss_feeds WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &feed, nil
}

func (r *Repository) DeleteFeed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM rss_feeds WHERE id = ?`, id)
	return err
}

func (r *Repository) MarkFeedFetched(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE rss_feeds SET last_fetched_at = NOW(), updated_at = NOW() WHERE id = ?`, id)
	return err
}

func (r *Repository) ArticleExists(ctx context.Context, urlHash, contentHash string) (bool, error) {
	var id int64
	err := r.db.GetContext(ctx, &id, `SELECT id FROM articles WHERE url_hash = ? OR content_hash = ? LIMIT 1`, urlHash, contentHash)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) StoreArticle(ctx context.Context, stored StoredArticle) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `INSERT INTO articles
(source_type, source_id, title, url, url_hash, content_hash, author, language, raw_content, clean_content, cover_image, published_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		stored.Article.SourceType, stored.Article.SourceID, stored.Article.Title, stored.Article.URL,
		stored.Article.URLHash, stored.Article.ContentHash, stored.Article.Author, stored.Article.Language,
		stored.Article.RawContent, stored.Article.CleanContent, stored.Article.CoverImage, stored.Article.PublishedAt)
	if err != nil {
		return 0, err
	}
	articleID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO summaries
(article_id, one_sentence, short_summary, long_summary, bullet_points, tldr, language)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		articleID, stored.Summary.OneSentence, stored.Summary.ShortSummary, stored.Summary.LongSummary,
		stored.Summary.BulletPoints, stored.Summary.TLDR, stored.Summary.Language)
	if err != nil {
		return 0, err
	}

	for _, tag := range stored.Tags {
		tagID, err := upsertTag(ctx, tx, tag, "topic")
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO article_tags (article_id, tag_id) VALUES (?, ?)`, articleID, tagID); err != nil {
			return 0, err
		}
	}
	for _, keyword := range stored.Keywords {
		tagID, err := upsertTag(ctx, tx, keyword, "keyword")
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO article_tags (article_id, tag_id) VALUES (?, ?)`, articleID, tagID); err != nil {
			return 0, err
		}
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO embeddings (article_id, provider, model, vector, dimension) VALUES (?, ?, ?, ?, ?)`,
		articleID, stored.Embedding.Provider, stored.Embedding.Model, stored.Embedding.Vector, stored.Embedding.Dimension)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return articleID, nil
}

func upsertTag(ctx context.Context, tx *sqlx.Tx, name, kind string) (int64, error) {
	_, err := tx.ExecContext(ctx, `INSERT INTO tags (name, type) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)`, name, kind)
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.GetContext(ctx, &id, `SELECT LAST_INSERT_ID()`)
	return id, err
}

func (r *Repository) ListArticles(ctx context.Context, limit, offset int) ([]model.Article, error) {
	return r.ListArticlesFiltered(ctx, ArticleFilter{Limit: limit, Offset: offset})
}

func (r *Repository) ListArticlesFiltered(ctx context.Context, filter ArticleFilter) ([]model.Article, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	var query strings.Builder
	query.WriteString(`SELECT DISTINCT a.* FROM articles a`)
	args := []any{}
	if filter.Tag != "" {
		query.WriteString(` JOIN article_tags at_filter ON at_filter.article_id = a.id JOIN tags t_filter ON t_filter.id = at_filter.tag_id`)
	}
	if filter.IsRead != nil {
		if *filter.IsRead {
			query.WriteString(` JOIN reading_history h_filter ON h_filter.article_id = a.id AND h_filter.user_id = ?`)
		} else {
			query.WriteString(` LEFT JOIN reading_history h_filter ON h_filter.article_id = a.id AND h_filter.user_id = ?`)
		}
		args = append(args, filter.UserID)
	}
	appendFavoriteJoin := func(alias, typ string, active bool) {
		if active {
			query.WriteString(` JOIN favorites ` + alias + ` ON ` + alias + `.article_id = a.id AND ` + alias + `.user_id = ? AND ` + alias + `.type = ?`)
		} else {
			query.WriteString(` LEFT JOIN favorites ` + alias + ` ON ` + alias + `.article_id = a.id AND ` + alias + `.user_id = ? AND ` + alias + `.type = ?`)
		}
		args = append(args, filter.UserID, typ)
	}
	if filter.IsFavorite != nil {
		appendFavoriteJoin("f_filter", "favorite", *filter.IsFavorite)
	}
	if filter.IsReadLater != nil {
		appendFavoriteJoin("rl_filter", "read_later", *filter.IsReadLater)
	}
	archived := false
	if filter.IsArchived != nil {
		archived = *filter.IsArchived
	}
	appendFavoriteJoin("ar_filter", "archived", archived)
	where := []string{}
	if filter.SourceType != "" {
		where = append(where, "a.source_type = ?")
		args = append(args, filter.SourceType)
	}
	if filter.SourceID > 0 {
		where = append(where, "a.source_id = ?")
		args = append(args, filter.SourceID)
	}
	if filter.Tag != "" {
		where = append(where, "t_filter.name = ?")
		args = append(args, filter.Tag)
	}
	if filter.From != nil {
		where = append(where, "COALESCE(a.published_at, a.created_at) >= ?")
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		where = append(where, "COALESCE(a.published_at, a.created_at) <= ?")
		args = append(args, *filter.To)
	}
	if filter.IsRead != nil && !*filter.IsRead {
		where = append(where, "h_filter.id IS NULL")
	}
	if filter.IsFavorite != nil && !*filter.IsFavorite {
		where = append(where, "f_filter.id IS NULL")
	}
	if filter.IsReadLater != nil && !*filter.IsReadLater {
		where = append(where, "rl_filter.id IS NULL")
	}
	if !archived {
		where = append(where, "ar_filter.id IS NULL")
	}
	if len(where) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(where, " AND "))
	}
	query.WriteString(` ORDER BY COALESCE(a.published_at, a.created_at) DESC LIMIT ? OFFSET ?`)
	args = append(args, filter.Limit, filter.Offset)
	var articles []model.Article
	err := r.db.SelectContext(ctx, &articles, query.String(), args...)
	return articles, err
}

func (r *Repository) ListArticlesSince(ctx context.Context, since time.Time, limit int) ([]model.Article, error) {
	var articles []model.Article
	err := r.db.SelectContext(ctx, &articles, `SELECT * FROM articles
WHERE COALESCE(published_at, created_at) >= ?
ORDER BY COALESCE(published_at, created_at) DESC
LIMIT ?`, since, limit)
	return articles, err
}

func (r *Repository) GetArticle(ctx context.Context, id int64) (*model.Article, error) {
	var article model.Article
	if err := r.db.GetContext(ctx, &article, `SELECT * FROM articles WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *Repository) MarkArticleRead(ctx context.Context, userID, articleID int64) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO reading_history (user_id, article_id, read_at)
VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE read_at = NOW()`, userID, articleID)
	return err
}

func (r *Repository) ReadingHistory(ctx context.Context, userID int64, limit, offset int) ([]model.Article, error) {
	var articles []model.Article
	err := r.db.SelectContext(ctx, &articles, `SELECT a.* FROM articles a
JOIN reading_history h ON h.article_id = a.id
WHERE h.user_id = ?
ORDER BY h.read_at DESC
LIMIT ? OFFSET ?`, userID, limit, offset)
	return articles, err
}

func (r *Repository) GetSummary(ctx context.Context, articleID int64) (*model.Summary, error) {
	var summary model.Summary
	if err := r.db.GetContext(ctx, &summary, `SELECT * FROM summaries WHERE article_id = ?`, articleID); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *Repository) TagsForArticle(ctx context.Context, articleID int64) ([]string, error) {
	var tags []string
	err := r.db.SelectContext(ctx, &tags, `SELECT t.name FROM tags t JOIN article_tags at ON at.tag_id = t.id WHERE at.article_id = ?`, articleID)
	return tags, err
}

func (r *Repository) AddFavorite(ctx context.Context, userID, articleID int64, typ string) error {
	_, err := r.db.ExecContext(ctx, `INSERT IGNORE INTO favorites (user_id, article_id, type) VALUES (?, ?, ?)`, userID, articleID, defaultString(typ, "favorite"))
	return err
}

func (r *Repository) ListFavorites(ctx context.Context, userID int64, typ string, limit, offset int) ([]model.Article, error) {
	var articles []model.Article
	err := r.db.SelectContext(ctx, &articles, `SELECT a.* FROM articles a
JOIN favorites f ON f.article_id = a.id
WHERE f.user_id = ? AND f.type = ?
ORDER BY f.created_at DESC
LIMIT ? OFFSET ?`, userID, defaultString(typ, "favorite"), limit, offset)
	return articles, err
}

func (r *Repository) RemoveFavorite(ctx context.Context, userID, articleID int64, typ string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM favorites WHERE user_id = ? AND article_id = ? AND type = ?`, userID, articleID, defaultString(typ, "favorite"))
	return err
}

func (r *Repository) DeleteArticle(ctx context.Context, articleID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	statements := []string{
		`DELETE FROM summaries WHERE article_id = ?`,
		`DELETE FROM translations WHERE article_id = ?`,
		`DELETE FROM embeddings WHERE article_id = ?`,
		`DELETE FROM article_tags WHERE article_id = ?`,
		`DELETE FROM favorites WHERE article_id = ?`,
		`DELETE FROM reading_history WHERE article_id = ?`,
		`DELETE FROM articles WHERE id = ?`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement, articleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) Dashboard(ctx context.Context) (*Dashboard, error) {
	d := &Dashboard{WorkerStatus: "gateway-in-process", AITokenCost: "mock-provider: 0"}
	_ = r.db.GetContext(ctx, &d.TodayNewArticles, `SELECT COUNT(*) FROM articles WHERE DATE(created_at) = CURRENT_DATE()`)
	_ = r.db.GetContext(ctx, &d.WeekNewArticles, `SELECT COUNT(*) FROM articles WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)`)
	_ = r.db.GetContext(ctx, &d.FailedTasks, `SELECT COUNT(*) FROM tasks WHERE status = 'failed'`)
	_ = r.db.SelectContext(ctx, &d.PopularTags, `SELECT t.name FROM tags t JOIN article_tags at ON at.tag_id = t.id GROUP BY t.id ORDER BY COUNT(*) DESC LIMIT 10`)
	return d, nil
}

func (r *Repository) CreateConversation(ctx context.Context, userID int64, title string) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO conversations (user_id, title) VALUES (?, ?)`, userID, defaultString(title, "TechPulse Chat"))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) StoreMessage(ctx context.Context, conversationID int64, role, content string, citations any) error {
	raw, _ := json.Marshal(citations)
	_, err := r.db.ExecContext(ctx, `INSERT INTO messages (conversation_id, role, content, citations) VALUES (?, ?, ?, ?)`,
		conversationID, role, content, string(raw))
	return err
}

func (r *Repository) RecentMessages(ctx context.Context, conversationID int64, limit int) ([]model.Message, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	var messages []model.Message
	err := r.db.SelectContext(ctx, &messages, `SELECT * FROM messages WHERE conversation_id = ? ORDER BY created_at DESC LIMIT ?`, conversationID, limit)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, err
}

func (r *Repository) ListConversations(ctx context.Context, userID int64) ([]model.Conversation, error) {
	var conversations []model.Conversation
	err := r.db.SelectContext(ctx, &conversations, `SELECT * FROM conversations WHERE user_id = ? ORDER BY updated_at DESC`, userID)
	return conversations, err
}

func (r *Repository) UpsertPrompt(ctx context.Context, prompt *UserPrompt) error {
	res, err := r.db.ExecContext(ctx, `INSERT INTO user_prompts (user_id, name, content, is_default)
VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE content = VALUES(content), is_default = VALUES(is_default), id = LAST_INSERT_ID(id)`,
		prompt.UserID, prompt.Name, prompt.Content, prompt.IsDefault)
	if err != nil {
		return err
	}
	prompt.ID, err = res.LastInsertId()
	return err
}

func (r *Repository) ListPrompts(ctx context.Context, userID int64) ([]UserPrompt, error) {
	var prompts []UserPrompt
	err := r.db.SelectContext(ctx, &prompts, `SELECT * FROM user_prompts WHERE user_id = ? ORDER BY is_default DESC, name ASC`, userID)
	return prompts, err
}

func (r *Repository) DeletePrompt(ctx context.Context, userID, promptID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_prompts WHERE user_id = ? AND id = ?`, userID, promptID)
	return err
}

func (r *Repository) StoreDailyReport(ctx context.Context, report *model.DailyReport) error {
	res, err := r.db.ExecContext(ctx, `INSERT INTO daily_reports (user_id, title, content, report_date) VALUES (?, ?, ?, ?)`,
		report.UserID, report.Title, report.Content, report.ReportDate)
	if err != nil {
		return err
	}
	report.ID, err = res.LastInsertId()
	return err
}

func (r *Repository) CreateTask(ctx context.Context, typ, payload string) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO tasks (type, status, payload, scheduled_at) VALUES (?, 'pending', ?, NOW())`, typ, payload)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetTask(ctx context.Context, id int64) (*model.Task, error) {
	var task model.Task
	if err := r.db.GetContext(ctx, &task, `SELECT * FROM tasks WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *Repository) MarkTaskRunning(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tasks SET status = 'running', started_at = NOW(), updated_at = NOW() WHERE id = ?`, id)
	return err
}

func (r *Repository) MarkTaskSuccess(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tasks SET status = 'success', finished_at = NOW(), updated_at = NOW() WHERE id = ?`, id)
	return err
}

func (r *Repository) MarkTaskFailed(ctx context.Context, id int64, message string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tasks SET status = 'failed', error_message = ?, finished_at = NOW(), updated_at = NOW() WHERE id = ?`, message, id)
	return err
}

func (r *Repository) MarkTaskRetrying(ctx context.Context, id int64, message string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tasks SET status = 'retrying', retry_count = retry_count + 1, error_message = ?, updated_at = NOW() WHERE id = ?`, message, id)
	return err
}

func (r *Repository) ListDailyReports(ctx context.Context, userID int64, limit, offset int) ([]model.DailyReport, error) {
	var reports []model.DailyReport
	err := r.db.SelectContext(ctx, &reports, `SELECT * FROM daily_reports WHERE user_id = ? ORDER BY report_date DESC, id DESC LIMIT ? OFFSET ?`, userID, limit, offset)
	return reports, err
}

func (r *Repository) SeedFeeds(ctx context.Context, userID int64) error {
	feeds := []model.RSSFeed{
		{UserID: userID, URL: "https://go.dev/blog/feed.atom", Title: "Go Blog", Category: "Go", Status: "active", FetchInterval: 360},
		{UserID: userID, URL: "https://kubernetes.io/feed.xml", Title: "Kubernetes", Category: "Kubernetes", Status: "active", FetchInterval: 360},
		{UserID: userID, URL: "https://hnrss.org/frontpage", Title: "Hacker News Frontpage", Category: "Tech News", Status: "active", FetchInterval: 60},
		{UserID: userID, URL: "https://github.blog/feed/", Title: "GitHub Blog", Category: "GitHub", Status: "active", FetchInterval: 360},
	}
	for _, feed := range feeds {
		_, err := r.db.ExecContext(ctx, `INSERT INTO rss_feeds (user_id, url, title, category, status, fetch_interval_minutes)
VALUES (?, ?, ?, ?, ?, ?)`, feed.UserID, feed.URL, feed.Title, feed.Category, feed.Status, feed.FetchInterval)
		if err != nil {
			return err
		}
	}
	return nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func ptrTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
