package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	apihandler "techpulse/internal/api/handler"
	apirouter "techpulse/internal/api/router"
	"techpulse/internal/config"
	"techpulse/internal/duplicate"
	"techpulse/internal/fetcher"
	"techpulse/internal/observability"
	"techpulse/internal/parser"
	"techpulse/internal/pipeline"
	"techpulse/internal/rag"
	"techpulse/internal/search"
	svc "techpulse/internal/service"
	"techpulse/internal/storage/mysql"
	redisstore "techpulse/internal/storage/redis"
	ws "techpulse/internal/websocket"
	"techpulse/pkg/httpclient"
)

func main() {
	cfg := config.Load()
	logger, err := observability.NewLogger(cfg.AppEnv)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := openDBWithRetry(ctx, cfg.MySQLDSN, 30*time.Second)
	if err != nil {
		logger.Fatal("open mysql", zap.Error(err))
	}
	defer db.Close()
	repo := mysql.NewRepository(db)
	if err := mysql.Migrate(ctx, db); err != nil {
		logger.Fatal("migrate", zap.Error(err))
	}
	if err := repo.EnsureDefaultUser(ctx); err != nil {
		logger.Fatal("ensure default user", zap.Error(err))
	}

	redisClient := redisstore.New(cfg.RedisAddr)
	if err := redisClient.Ping(ctx); err != nil {
		logger.Warn("redis unavailable; cache disabled", zap.String("addr", cfg.RedisAddr), zap.Error(err))
		_ = redisClient.Close()
		redisClient = nil
	} else {
		defer redisClient.Close()
		logger.Info("redis connected", zap.String("addr", cfg.RedisAddr))
	}

	bleveEngine, err := search.NewBleveEngine(cfg.BleveIndexPath)
	if err != nil {
		logger.Fatal("open bleve", zap.Error(err))
	}
	defer bleveEngine.Close()

	client := httpclient.New(cfg.RequestTimeout)
	provider := svc.AIProvider(cfg, client)
	engine := search.NewHybridEngine(bleveEngine, provider)
	hub := ws.NewHub()
	go hub.Run()

	fetchSvc := fetcher.NewService(fetcher.NewRegistry(
		fetcher.NewRSSFetcher(client),
		fetcher.GitHubReleaseFetcher{},
		fetcher.HackerNewsFetcher{},
		fetcher.RedditFetcher{},
		fetcher.ArxivFetcher{},
		fetcher.YouTubeFetcher{},
	))
	parserSvc := parser.NewService()
	duplicateSvc := duplicate.NewService(repo)
	pipelineSvc := pipeline.NewService(provider)
	ragSvc := rag.NewService(rag.NewRetriever(engine), rag.NewGenerator(provider)).WithMemory(repo)
	handler := apihandler.New(repo, fetchSvc, parserSvc, duplicateSvc, pipelineSvc, engine, ragSvc, provider, redisClient, hub, logger)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           apirouter.New(handler, logger, cfg.DefaultUserID),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("gateway started", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("gateway failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", zap.Error(err))
	}
}

func openDBWithRetry(ctx context.Context, dsn string, timeout time.Duration) (db *sqlx.DB, err error) {
	deadline := time.Now().Add(timeout)
	for {
		db, err = mysql.Open(ctx, dsn)
		if err == nil {
			return db, nil
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}
