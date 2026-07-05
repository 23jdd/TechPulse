package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/observability"
	"techpulse/internal/queue"
	"techpulse/internal/storage/mysql"
	"techpulse/pkg/httpclient"
)

func main() {
	mode := flag.String("mode", "run", "run, migrate, or seed")
	flag.Parse()
	cfg := config.Load()
	logger, _ := observability.NewLogger(cfg.AppEnv)
	defer func() { _ = logger.Sync() }()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	db, err := mysql.Open(ctx, cfg.MySQLDSN)
	if err != nil {
		logger.Fatal("open mysql", zap.Error(err))
	}
	defer db.Close()
	repo := mysql.NewRepository(db)
	switch *mode {
	case "migrate":
		if err := mysql.Migrate(ctx, db); err != nil {
			logger.Fatal("migrate", zap.Error(err))
		}
		if err := repo.EnsureDefaultUser(ctx); err != nil {
			logger.Fatal("default user", zap.Error(err))
		}
		fmt.Println("migration complete")
	case "seed":
		if err := repo.EnsureDefaultUser(ctx); err != nil {
			logger.Fatal("default user", zap.Error(err))
		}
		if err := repo.SeedFeeds(ctx, cfg.DefaultUserID); err != nil {
			logger.Fatal("seed", zap.Error(err))
		}
		fmt.Println("seed complete")
	default:
		runWorker(cfg, repo, logger)
	}
}

func runWorker(cfg config.Config, repo *mysql.Repository, logger *zap.Logger) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	broker := queue.NewRabbitMQ(cfg.RabbitMQURL)
	defer broker.Close()
	client := httpclient.New(cfg.RequestTimeout)

	queues := []string{"fetch", "parse", "ai", "index", "daily_report"}
	for _, queueName := range queues {
		messages, err := broker.Consume(ctx, queueName)
		if err != nil {
			logger.Fatal("consume queue", zap.String("queue", queueName), zap.Error(err))
		}
		go func(name string, ch <-chan queue.Message) {
			for msg := range ch {
				handleMessage(ctx, cfg, client, broker, repo, logger, name, msg)
			}
		}(queueName, messages)
	}
	logger.Info("worker started", zap.Strings("queues", queues))
	<-ctx.Done()
	os.Exit(0)
}

func handleMessage(ctx context.Context, cfg config.Config, client *http.Client, broker *queue.RabbitMQ, repo *mysql.Repository, logger *zap.Logger, queueName string, msg queue.Message) {
	start := time.Now()
	raw, _ := json.Marshal(msg.Payload)
	taskID := payloadInt64(msg.Payload, "task_id")
	if taskID == 0 {
		var err error
		taskID, err = repo.CreateTask(ctx, string(msg.Type), string(raw))
		if err != nil {
			logger.Error("create task failed", zap.String("queue", queueName), zap.Error(err))
			return
		}
	}
	if err := repo.MarkTaskRunning(ctx, taskID); err != nil {
		logger.Warn("mark task running failed", zap.Int64("task_id", taskID), zap.Error(err))
	}
	err := processJob(ctx, cfg, client, queueName, msg)
	if err == nil {
		_ = repo.MarkTaskSuccess(ctx, taskID)
		logger.Info("worker job succeeded", zap.Int64("task_id", taskID), zap.String("queue", queueName), zap.String("type", string(msg.Type)), zap.Duration("duration", time.Since(start)))
		return
	}
	retryCount := payloadInt(msg.Payload, "retry_count")
	if retryCount < 3 {
		msg.Payload["retry_count"] = retryCount + 1
		_ = repo.MarkTaskRetrying(ctx, taskID, err.Error())
		if publishErr := broker.Publish(ctx, queueName, msg); publishErr != nil {
			logger.Error("republish retry failed", zap.Int64("task_id", taskID), zap.Error(publishErr))
		}
		logger.Warn("worker job retrying", zap.Int64("task_id", taskID), zap.Int("retry_count", retryCount+1), zap.Error(err))
		return
	}
	_ = repo.MarkTaskFailed(ctx, taskID, err.Error())
	dlq := queueName + ".dlq"
	if publishErr := broker.Publish(ctx, dlq, msg); publishErr != nil {
		logger.Error("publish dlq failed", zap.Int64("task_id", taskID), zap.String("dlq", dlq), zap.Error(publishErr))
	}
	logger.Error("worker job failed", zap.Int64("task_id", taskID), zap.String("queue", queueName), zap.Error(err), zap.Duration("duration", time.Since(start)))
}

func processJob(ctx context.Context, cfg config.Config, client *http.Client, queueName string, msg queue.Message) error {
	if force, ok := msg.Payload["force_error"].(bool); ok && force {
		return fmt.Errorf("forced worker error")
	}
	if queueName == "fetch" || msg.Type == queue.FetchJob {
		return processFetchJob(ctx, cfg, client, msg)
	}
	return nil
}

func processFetchJob(ctx context.Context, cfg config.Config, client *http.Client, msg queue.Message) error {
	feedID := payloadInt64(msg.Payload, "feed_id")
	if feedID == 0 {
		return fmt.Errorf("feed_id is required")
	}
	base := strings.TrimRight(cfg.GatewayURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/rss/%d/fetch", base, feedID), nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("gateway fetch failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func payloadInt(payload map[string]any, key string) int {
	value, ok := payload[key]
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func payloadInt64(payload map[string]any, key string) int64 {
	value, ok := payload[key]
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		return 0
	}
}
