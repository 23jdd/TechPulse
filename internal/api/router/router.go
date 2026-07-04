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
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/dashboard.html")
	})
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/login.html")
	})
	r.Get("/login/zh", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/login.zh-CN.html")
	})
	r.Get("/zh", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/dashboard.zh-CN.html")
	})
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/dashboard.html")
	})
	r.Get("/dashboard/zh", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/dashboard.zh-CN.html")
	})
	r.Get("/health", h.Health)
	r.Get("/metrics", observability.NewMetrics().Handler().ServeHTTP)
	r.Get("/ws", h.WebSocket)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.MockAuth(defaultUserID))
		r.Post("/rss", h.CreateRSS)
		r.Get("/rss", h.ListRSS)
		r.Get("/rss/{id}", h.GetRSS)
		r.Put("/rss/{id}", h.UpdateRSS)
		r.Delete("/rss/{id}", h.DeleteRSS)
		r.Post("/rss/{id}/enable", h.EnableRSS)
		r.Post("/rss/{id}/disable", h.DisableRSS)
		r.Post("/rss/{id}/test", h.TestRSS)
		r.Post("/rss/{id}/fetch", h.FetchRSS)
		r.Post("/github/releases/fetch", h.FetchGitHubReleases)
		r.Get("/articles", h.ListArticles)
		r.Get("/articles/{id}", h.GetArticle)
		r.Delete("/articles/{id}", h.DeleteArticle)
		r.Get("/articles/{id}/summary", h.GetArticleSummary)
		r.Post("/articles/{id}/archive", h.ArchiveArticle)
		r.Delete("/articles/{id}/archive", h.UnarchiveArticle)
		r.Post("/articles/{id}/favorite", h.AddFavorite)
		r.Delete("/articles/{id}/favorite", h.RemoveFavorite)
		r.Post("/articles/{id}/read", h.MarkRead)
		r.Post("/articles/{id}/read-later", h.AddReadLater)
		r.Delete("/articles/{id}/read-later", h.RemoveReadLater)
		r.Get("/search", h.Search)
		r.Get("/search/explain", h.SearchExplain)
		r.Post("/search/reindex", h.ReindexSearch)
		r.Post("/summary", h.Summary)
		r.Post("/translate", h.Translate)
		r.Post("/tags", h.Tags)
		r.Post("/chat", h.Chat)
		r.Get("/conversations", h.ListConversations)
		r.Get("/favorites", h.ListFavorites)
		r.Get("/reading-history", h.ReadingHistory)
		r.Get("/prompts", h.ListPrompts)
		r.Post("/prompts", h.UpsertPrompt)
		r.Delete("/prompts/{id}", h.DeletePrompt)
		r.Get("/opml", h.ExportOPML)
		r.Post("/opml", h.ImportOPML)
		r.Get("/auth/github/url", h.GitHubAuthURL)
		r.Get("/auth/github/callback", h.GitHubCallback)
		r.Post("/email/test", h.SendTestEmail)
		r.Post("/daily-reports", h.GenerateDailyReport)
		r.Get("/daily-reports", h.ListDailyReports)
		r.Get("/dashboard", h.Dashboard)
	})
	return r
}
