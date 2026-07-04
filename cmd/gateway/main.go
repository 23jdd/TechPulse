package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"techpulse/internal/ai"
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
	"techpulse/internal/storage/mysql"
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

	engine, err := search.NewBleveEngine(cfg.BleveIndexPath)
	if err != nil {
		logger.Fatal("open bleve", zap.Error(err))
	}
	defer engine.Close()

	client := httpclient.New(cfg.RequestTimeout)
	provider := aiProvider(cfg, client)
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
	ragSvc := rag.NewService(rag.NewRetriever(engine), rag.NewGenerator(provider))
	handler := apihandler.New(repo, fetchSvc, parserSvc, duplicateSvc, pipelineSvc, engine, ragSvc, provider, hub, logger)

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

func aiProvider(cfg config.Config, client *http.Client) ai.Provider {
	if strings.EqualFold(cfg.AIProvider, "mock") || cfg.AIAPIKey == "" {
		return ai.NewMockProvider()
	}
	return ai.NewOpenAICompatibleProvider(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, client)
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
