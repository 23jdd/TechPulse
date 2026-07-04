package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"techpulse/internal/api/handler"
	"techpulse/internal/api/middleware"
	"techpulse/internal/observability"
)

func New(h *handler.Handler, logger *zap.Logger, defaultUserID int64) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(middleware.Logging(logger))
	r.Get("/health", h.Health)
	r.Get("/metrics", observability.NewMetrics().Handler().ServeHTTP)
	r.Get("/ws", h.WebSocket)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.MockAuth(defaultUserID))
		r.Post("/rss", h.CreateRSS)
		r.Get("/rss", h.ListRSS)
		r.Get("/rss/{id}", h.GetRSS)
		r.Delete("/rss/{id}", h.DeleteRSS)
		r.Post("/rss/{id}/fetch", h.FetchRSS)
		r.Get("/articles", h.ListArticles)
		r.Get("/articles/{id}", h.GetArticle)
		r.Get("/articles/{id}/summary", h.GetArticleSummary)
		r.Post("/articles/{id}/favorite", h.AddFavorite)
		r.Delete("/articles/{id}/favorite", h.RemoveFavorite)
		r.Get("/search", h.Search)
		r.Post("/summary", h.Summary)
		r.Post("/translate", h.Translate)
		r.Post("/tags", h.Tags)
		r.Post("/chat", h.Chat)
		r.Get("/dashboard", h.Dashboard)
	})
	return r
}
