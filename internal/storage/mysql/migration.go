package mysql

import "context"
import "database/sql"

const Schema = `
CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  github_id VARCHAR(128) DEFAULT '',
  username VARCHAR(128) NOT NULL,
  email VARCHAR(255) DEFAULT '',
  avatar_url TEXT,
  api_token VARCHAR(255) DEFAULT '',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rss_feeds (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  url TEXT NOT NULL,
  title VARCHAR(255) NOT NULL DEFAULT '',
  category VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  last_fetched_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_rss_user (user_id)
);

CREATE TABLE IF NOT EXISTS articles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  source_type VARCHAR(64) NOT NULL,
  source_id BIGINT NOT NULL DEFAULT 0,
  title TEXT NOT NULL,
  url TEXT NOT NULL,
  url_hash CHAR(64) NOT NULL,
  content_hash CHAR(64) NOT NULL,
  author VARCHAR(255) DEFAULT '',
  language VARCHAR(32) DEFAULT 'en',
  raw_content MEDIUMTEXT,
  clean_content MEDIUMTEXT,
  cover_image TEXT,
  published_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_articles_url_hash (url_hash),
  INDEX idx_articles_source (source_type, source_id),
  INDEX idx_articles_content_hash (content_hash),
  INDEX idx_articles_published_at (published_at)
);

CREATE TABLE IF NOT EXISTS tags (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  type VARCHAR(64) NOT NULL DEFAULT 'topic',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_tags_name_type (name, type)
);

CREATE TABLE IF NOT EXISTS article_tags (
  article_id BIGINT NOT NULL,
  tag_id BIGINT NOT NULL,
  PRIMARY KEY (article_id, tag_id)
);

CREATE TABLE IF NOT EXISTS favorites (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  article_id BIGINT NOT NULL,
  type VARCHAR(64) NOT NULL DEFAULT 'favorite',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_favorites_user_article_type (user_id, article_id, type)
);

CREATE TABLE IF NOT EXISTS summaries (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  article_id BIGINT NOT NULL,
  one_sentence TEXT,
  short_summary TEXT,
  long_summary MEDIUMTEXT,
  bullet_points TEXT,
  tldr TEXT,
  language VARCHAR(32) DEFAULT 'en',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_summaries_article (article_id)
);

CREATE TABLE IF NOT EXISTS translations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  article_id BIGINT NOT NULL,
  target_language VARCHAR(32) NOT NULL,
  translated_title TEXT,
  translated_content MEDIUMTEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS embeddings (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  article_id BIGINT NOT NULL,
  provider VARCHAR(64) NOT NULL,
  model VARCHAR(128) NOT NULL,
  vector MEDIUMTEXT NOT NULL,
  dimension INT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_embeddings_article_provider_model (article_id, provider, model)
);

CREATE TABLE IF NOT EXISTS tasks (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  type VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  payload MEDIUMTEXT,
  retry_count INT NOT NULL DEFAULT 0,
  error_message TEXT,
  scheduled_at TIMESTAMP NULL,
  started_at TIMESTAMP NULL,
  finished_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tasks_status (status)
);

CREATE TABLE IF NOT EXISTS conversations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  conversation_id BIGINT NOT NULL,
  role VARCHAR(32) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  citations MEDIUMTEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_messages_conversation (conversation_id, created_at)
);

CREATE TABLE IF NOT EXISTS reading_history (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  article_id BIGINT NOT NULL,
  read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_reading_history_user_article (user_id, article_id),
  INDEX idx_reading_history_user (user_id, read_at)
);

CREATE TABLE IF NOT EXISTS user_prompts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_user_prompts_name (user_id, name)
);

CREATE TABLE IF NOT EXISTS daily_reports (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  report_date DATE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

func Migrate(ctx context.Context, db Execer) error {
	_, err := db.ExecContext(ctx, Schema)
	return err
}

type Execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}
