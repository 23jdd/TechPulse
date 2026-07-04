package main

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/observability"
	"techpulse/internal/rag"
	searchengine "techpulse/internal/search"
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

	engine, err := searchengine.NewBleveEngine(cfg.BleveIndexPath)
	if err != nil {
		logger.Fatal("open bleve", zap.Error(err))
	}
	defer engine.Close()

	provider := service.AIProvider(cfg, httpclient.New(cfg.RequestTimeout))
	ragSvc := rag.NewService(rag.NewRetriever(engine), rag.NewGenerator(provider))
	service.RegisterSelf(context.Background(), cfg, "rag", 8086, logger)
	server := service.NewServer("rag", 8086, logger)
	server.Mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			service.Error(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
			return
		}
		var req service.ChatRequest
		if err := service.DecodeJSON(r, &req); err != nil {
			service.Error(w, http.StatusBadRequest, err)
			return
		}
		answer, err := ragSvc.Ask(r.Context(), req.Question)
		if err != nil {
			service.Error(w, http.StatusInternalServerError, err)
			return
		}
		service.JSON(w, http.StatusOK, answer)
	})
	if err := server.Run(context.Background()); err != nil {
		logger.Fatal("rag stopped", zap.Error(err))
	}
}
