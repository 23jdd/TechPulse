package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/observability"
	"techpulse/internal/queue"
	"techpulse/internal/storage/mysql"
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
		runWorker(cfg, logger)
	}
}

func runWorker(cfg config.Config, logger *zap.Logger) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	broker := queue.NewRabbitMQ(cfg.RabbitMQURL)
	defer broker.Close()

	queues := []string{"fetch", "parse", "ai", "index", "daily_report"}
	for _, queueName := range queues {
		messages, err := broker.Consume(ctx, queueName)
		if err != nil {
			logger.Fatal("consume queue", zap.String("queue", queueName), zap.Error(err))
		}
		go func(name string, ch <-chan queue.Message) {
			for msg := range ch {
				raw, _ := json.Marshal(msg.Payload)
				logger.Info("worker received job", zap.String("queue", name), zap.String("type", string(msg.Type)), zap.ByteString("payload", raw))
			}
		}(queueName, messages)
	}
	logger.Info("worker started", zap.Strings("queues", queues))
	<-ctx.Done()
	os.Exit(0)
}
