package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv         string
	HTTPPort       int
	MySQLDSN       string
	RedisAddr      string
	RabbitMQURL    string
	EtcdEndpoints  []string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	AIProvider     string
	AIBaseURL      string
	AIAPIKey       string
	AIModel        string
	BleveIndexPath string
	RequestTimeout time.Duration
	GitHubClientID string
	GitHubSecret   string
	GitHubRedirect string
	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	SMTPFrom       string
	DefaultUserID  int64
}

func Load() Config {
	return Config{
		AppEnv:         env("APP_ENV", "dev"),
		HTTPPort:       envInt("HTTP_PORT", 8080),
		MySQLDSN:       env("MYSQL_DSN", "root:password@tcp(localhost:3306)/techpulse?parseTime=true&charset=utf8mb4&multiStatements=true"),
		RedisAddr:      env("REDIS_ADDR", "localhost:6379"),
		RabbitMQURL:    env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		EtcdEndpoints:  strings.Split(env("ETCD_ENDPOINTS", "localhost:2379"), ","),
		MinIOEndpoint:  env("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: env("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: env("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    env("MINIO_BUCKET", "techpulse"),
		AIProvider:     env("AI_PROVIDER", "mock"),
		AIBaseURL:      env("AI_BASE_URL", "https://api.openai.com/v1"),
		AIAPIKey:       env("AI_API_KEY", ""),
		AIModel:        env("AI_MODEL", "gpt-4o-mini"),
		BleveIndexPath: env("BLEVE_INDEX_PATH", "./data/bleve"),
		RequestTimeout: time.Duration(envInt("REQUEST_TIMEOUT_SECONDS", 20)) * time.Second,
		GitHubClientID: env("GITHUB_CLIENT_ID", ""),
		GitHubSecret:   env("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirect: env("GITHUB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/github/callback"),
		SMTPHost:       env("SMTP_HOST", ""),
		SMTPPort:       envInt("SMTP_PORT", 587),
		SMTPUsername:   env("SMTP_USERNAME", ""),
		SMTPPassword:   env("SMTP_PASSWORD", ""),
		SMTPFrom:       env("SMTP_FROM", ""),
		DefaultUserID:  int64(envInt("DEFAULT_USER_ID", 1)),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
