package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/observability"
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
		fmt.Println("worker ready: RabbitMQ consumer skeleton is available for Phase 2")
		<-ctx.Done()
		os.Exit(0)
	}
}
