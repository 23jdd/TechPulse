package main

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/observability"
	"techpulse/internal/parser"
	"techpulse/internal/service"
)

func main() {
	cfg := config.Load()
	logger, err := observability.NewLogger(cfg.AppEnv)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	parserSvc := parser.NewService()
	service.RegisterSelf(context.Background(), cfg, "parser", 8083, logger)
	server := service.NewServer("parser", 8083, logger)
	server.Mux.HandleFunc("/parse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			service.Error(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
			return
		}
		var req service.ParseRequest
		if err := service.DecodeJSON(r, &req); err != nil {
			service.Error(w, http.StatusBadRequest, err)
			return
		}
		article, err := parserSvc.Parse(r.Context(), req.Item)
		if err != nil {
			service.Error(w, http.StatusBadRequest, err)
			return
		}
		service.JSON(w, http.StatusOK, service.ParseResponse{Article: article})
	})
	if err := server.Run(context.Background()); err != nil {
		logger.Fatal("parser stopped", zap.Error(err))
	}
}
