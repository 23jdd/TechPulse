package main

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/observability"
	"techpulse/internal/pipeline"
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

	provider := service.AIProvider(cfg, httpclient.New(cfg.RequestTimeout))
	pipelineSvc := pipeline.NewService(provider)
	service.RegisterSelf(context.Background(), cfg, "ai-pipeline", 8084, logger)
	server := service.NewServer("ai-pipeline", 8084, logger)
	server.Mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			service.Error(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
			return
		}
		var req service.ProcessRequest
		if err := service.DecodeJSON(r, &req); err != nil {
			service.Error(w, http.StatusBadRequest, err)
			return
		}
		out, err := pipelineSvc.Process(r.Context(), req.Article)
		if err != nil {
			service.Error(w, http.StatusInternalServerError, err)
			return
		}
		service.JSON(w, http.StatusOK, service.ProcessResponse{
			Article: out.Article, Summary: out.Summary, Tags: out.Tags, Keywords: out.Keywords,
			Embedding: out.Embedding, Translation: out.Translation,
		})
	})
	if err := server.Run(context.Background()); err != nil {
		logger.Fatal("ai-pipeline stopped", zap.Error(err))
	}
}
