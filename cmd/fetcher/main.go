package main

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/fetcher"
	"techpulse/internal/observability"
	"techpulse/internal/service"
	"techpulse/pkg/httpclient"
)

func main() {
	cfg := config.Load()
	logger, err := observability.NewLogger(cfg.AppEnv)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	client := httpclient.New(cfg.RequestTimeout)
	fetchSvc := fetcher.NewService(fetcher.NewRegistry(
		fetcher.NewRSSFetcher(client),
		fetcher.GitHubReleaseFetcher{},
		fetcher.HackerNewsFetcher{},
		fetcher.RedditFetcher{},
		fetcher.ArxivFetcher{},
		fetcher.YouTubeFetcher{},
	))
	service.RegisterSelf(context.Background(), cfg, "fetcher", 8082, logger)
	server := service.NewServer("fetcher", 8082, logger)
	server.Mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			service.Error(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
			return
		}
		var req service.FetchRequest
		if err := service.DecodeJSON(r, &req); err != nil {
			service.Error(w, http.StatusBadRequest, err)
			return
		}
		items, err := fetchSvc.Fetch(r.Context(), req.Source)
		if err != nil {
			logger.Warn("fetch failed", zap.Error(err))
			service.Error(w, http.StatusBadGateway, err)
			return
		}
		service.JSON(w, http.StatusOK, service.FetchResponse{Items: items})
	})
	if err := server.Run(context.Background()); err != nil {
		logger.Fatal("fetcher stopped", zap.Error(err))
	}
}
