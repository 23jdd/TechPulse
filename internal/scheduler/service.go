package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"techpulse/internal/queue"
)

type Publisher interface {
	Publish(context.Context, string, queue.Message) error
}

type Service struct {
	publisher Publisher
	interval  time.Duration
}

func NewService(publisher Publisher, interval time.Duration) *Service {
	return &Service{publisher: publisher, interval: interval}
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case at := <-ticker.C:
			_ = s.publisher.Publish(ctx, "fetch", queue.Message{Type: queue.FetchJob, Payload: map[string]any{"scheduled_at": at}})
		}
	}
}

func RunStandalone(name string, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"status":"ok","service":"%s"}`+"\n", name)
	})
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	fmt.Printf("%s listening on %s\n", name, server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
