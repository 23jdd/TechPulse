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
	"techpulse/internal/auth"
	"techpulse/internal/config"
	"techpulse/internal/duplicate"
	"techpulse/internal/email"
	"techpulse/internal/fetcher"
	"techpulse/internal/observability"
	"techpulse/internal/parser"
	"techpulse/internal/pipeline"
	"techpulse/internal/queue"
	"techpulse/internal/rag"
	"techpulse/internal/report"
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
	oauth := auth.NewGitHubOAuth(cfg.GitHubClientID, cfg.GitHubSecret, cfg.GitHubRedirect, client)
	mailer := email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)
	engine := search.NewHybridEngine(bleveEngine, provider)
	hub := ws.NewHub()
	go hub.Run()
	go runDailyReportScheduler(ctx, repo, mailer, logger)
	broker := queue.NewRabbitMQ(cfg.RabbitMQURL)
	defer broker.Close()

	fetchSvc := fetcher.NewService(fetcher.NewRegistry(
		fetcher.NewRSSFetcher(client),
		fetcher.NewGitHubReleaseFetcher(client, os.Getenv("GITHUB_TOKEN")),
		fetcher.NewHackerNewsFetcher(client),
		fetcher.RedditFetcher{},
		fetcher.ArxivFetcher{},
		fetcher.YouTubeFetcher{},
	))
	parserSvc := parser.NewService()
	duplicateSvc := duplicate.NewService(repo)
	pipelineSvc := pipeline.NewService(provider)
	ragSvc := rag.NewService(rag.NewRetriever(engine), rag.NewGenerator(provider)).WithMemory(repo)
	handler := apihandler.New(repo, fetchSvc, parserSvc, duplicateSvc, pipelineSvc, engine, ragSvc, provider, redisClient, oauth, mailer, hub, broker, cfg.JWTSecret, logger)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           apirouter.New(handler, logger, cfg.DefaultUserID, cfg.JWTSecret, cfg.JWTRequired),
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

func runDailyReportScheduler(ctx context.Context, repo *mysql.Repository, mailer email.Sender, logger *zap.Logger) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	run := func() {
		prefs, err := repo.EnabledDailyReportPreferences(ctx)
		if err != nil {
			logger.Warn("load report preferences failed", zap.Error(err))
			return
		}
		for _, pref := range prefs {
			location := time.Local
			if pref.Timezone != "" {
				if loaded, err := time.LoadLocation(pref.Timezone); err == nil {
					location = loaded
				}
			}
			now := time.Now().In(location)
			if now.Format("15:04") != pref.DailyReportTime {
				continue
			}
			exists, err := repo.HasDailyReport(ctx, pref.UserID, now)
			if err != nil || exists {
				continue
			}
			daily, err := report.NewService(repo).Generate(ctx, pref.UserID, "TechPulse Daily Report")
			if err != nil {
				logger.Warn("scheduled report generation failed", zap.Int64("user_id", pref.UserID), zap.Error(err))
				continue
			}
			if mailer != nil && mailer.Enabled() && pref.DailyReportEmail != "" {
				if err := mailer.Send(ctx, email.Message{To: pref.DailyReportEmail, Subject: daily.Title, Body: daily.Content}); err != nil {
					logger.Warn("scheduled report email failed", zap.Int64("user_id", pref.UserID), zap.Error(err))
				}
			}
		}
	}
	run()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
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
