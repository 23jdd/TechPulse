package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/observability"
	"techpulse/internal/queue"
	"techpulse/internal/scheduler"
	"techpulse/internal/service"
)

func main() {
	cfg := config.Load()
	logger, err := observability.NewLogger(cfg.AppEnv)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	broker := queue.NewRabbitMQ(cfg.RabbitMQURL)
	defer broker.Close()
	schedulerSvc := scheduler.NewService(broker, time.Hour)
	service.RegisterSelf(context.Background(), cfg, "scheduler", 8081, logger)
	server := service.NewServer("scheduler", 8081, logger)
	server.Mux.HandleFunc("/schedule/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			service.Error(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
			return
		}
		feedID, _ := strconv.ParseInt(r.URL.Query().Get("feed_id"), 10, 64)
		msg := queue.Message{Type: queue.FetchJob, Payload: map[string]any{"feed_id": feedID, "scheduled_at": time.Now()}}
		if err := broker.Publish(r.Context(), "fetch", msg); err != nil {
			service.Error(w, http.StatusBadGateway, err)
			return
		}
		service.JSON(w, http.StatusAccepted, map[string]any{"queued": true, "queue": "fetch", "feed_id": feedID})
	})
	server.Mux.HandleFunc("/tick", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			service.Error(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
			return
		}
		if err := broker.Publish(r.Context(), "fetch", queue.Message{Type: queue.FetchJob, Payload: map[string]any{"scheduled_at": time.Now()}}); err != nil {
			service.Error(w, http.StatusBadGateway, err)
			return
		}
		service.JSON(w, http.StatusAccepted, map[string]any{"queued": true})
	})
	go func() {
		if err := schedulerSvc.Run(context.Background()); err != nil {
			logger.Debug("scheduler ticker stopped", zap.Error(err))
		}
	}()
	if err := server.Run(context.Background()); err != nil {
		logger.Fatal("scheduler stopped", zap.Error(err))
	}
}
