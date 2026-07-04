package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"techpulse/internal/ai"
	"techpulse/internal/api/dto"
	"techpulse/internal/api/middleware"
	"techpulse/internal/duplicate"
	"techpulse/internal/fetcher"
	"techpulse/internal/model"
	"techpulse/internal/parser"
	"techpulse/internal/pipeline"
	"techpulse/internal/rag"
	"techpulse/internal/search"
	"techpulse/internal/storage/mysql"
	ws "techpulse/internal/websocket"
	"techpulse/pkg/pagination"
	"techpulse/pkg/response"
)

type Handler struct {
	repo      *mysql.Repository
	fetcher   *fetcher.Service
	parser    *parser.Service
	duplicate *duplicate.Service
	pipeline  *pipeline.Service
	search    search.SearchEngine
	rag       *rag.Service
	ai        ai.Provider
	hub       *ws.Hub
	logger    *zap.Logger
}

func New(repo *mysql.Repository, fetcher *fetcher.Service, parser *parser.Service, duplicate *duplicate.Service, pipeline *pipeline.Service, search search.SearchEngine, rag *rag.Service, ai ai.Provider, hub *ws.Hub, logger *zap.Logger) *Handler {
	return &Handler{repo: repo, fetcher: fetcher, parser: parser, duplicate: duplicate, pipeline: pipeline, search: search, rag: rag, ai: ai, hub: hub, logger: logger}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, response.Envelope{"status": "ok", "service": "gateway"})
}

func (h *Handler) CreateRSS(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRSSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		response.Error(w, http.StatusBadRequest, "url is required")
		return
	}
	feed := &model.RSSFeed{UserID: middleware.UserID(r), URL: req.URL, Title: req.Title, Category: req.Category, Status: "active"}
	if err := h.repo.CreateFeed(r.Context(), feed); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, feed)
}

func (h *Handler) ListRSS(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.repo.ListFeeds(r.Context(), middleware.UserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": feeds})
}

func (h *Handler) GetRSS(w http.ResponseWriter, r *http.Request) {
	feed, err := h.repo.GetFeed(r.Context(), idParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusNotFound, "feed not found")
		return
	}
	response.JSON(w, http.StatusOK, feed)
}

func (h *Handler) DeleteRSS(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteFeed(r.Context(), idParam(r, "id")); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"deleted": true})
}

func (h *Handler) FetchRSS(w http.ResponseWriter, r *http.Request) {
	result, status := h.fetchFeed(r, idParam(r, "id"))
	response.JSON(w, status, result)
}

func (h *Handler) fetchFeed(r *http.Request, feedID int64) (*dto.FetchRSSResponse, int) {
	ctx := r.Context()
	feed, err := h.repo.GetFeed(ctx, feedID)
	if err != nil {
		return &dto.FetchRSSResponse{FeedID: feedID, Errors: []string{"feed not found"}}, http.StatusNotFound
	}
	h.hub.Broadcast(ws.Event{Type: "fetch_started", Message: feed.URL, Time: time.Now()})
	items, err := h.fetcher.Fetch(ctx, fetcher.Source{ID: feed.ID, Type: "rss", URL: feed.URL, Title: feed.Title, Category: feed.Category})
	if err != nil {
		h.hub.Broadcast(ws.Event{Type: "task_failed", Message: err.Error(), Time: time.Now()})
		return &dto.FetchRSSResponse{FeedID: feed.ID, Errors: []string{err.Error()}}, http.StatusBadGateway
	}
	out := &dto.FetchRSSResponse{FeedID: feed.ID, Fetched: len(items)}
	for _, item := range items {
		parsed, err := h.parser.Parse(ctx, item)
		if err != nil {
			out.Errors = append(out.Errors, err.Error())
			continue
		}
		content := parsed.CleanContent
		if content == "" {
			content = parsed.Title
		}
		exists, urlHash, contentHash, err := h.duplicate.IsDuplicate(ctx, parsed.URL, content)
		if err != nil {
			out.Errors = append(out.Errors, err.Error())
			continue
		}
		if exists {
			out.Duplicates++
			continue
		}
		enriched, err := h.pipeline.Process(ctx, parsed)
		if err != nil {
			out.Errors = append(out.Errors, err.Error())
			continue
		}
		enriched.Article.URLHash = urlHash
		enriched.Article.ContentHash = contentHash
		rawEmbedding, _ := json.Marshal(enriched.Embedding)
		stored := mysql.StoredArticle{
			Article:  enriched.Article,
			Summary:  enriched.Summary,
			Tags:     uniqueStrings(enriched.Tags),
			Keywords: uniqueStrings(enriched.Keywords),
			Embedding: model.Embedding{
				Provider:  h.ai.Name(),
				Model:     "mock-embedding",
				Vector:    string(rawEmbedding),
				Dimension: len(enriched.Embedding),
			},
		}
		articleID, err := h.repo.StoreArticle(ctx, stored)
		if err != nil {
			out.Errors = append(out.Errors, err.Error())
			continue
		}
		enriched.Article.ID = articleID
		if err := h.search.IndexArticle(ctx, search.DocumentFromArticle(enriched.Article, enriched.Summary.ShortSummary, stored.Tags)); err != nil {
			out.Errors = append(out.Errors, err.Error())
			continue
		}
		out.Inserted++
		h.hub.Broadcast(ws.Event{Type: "new_article", Message: enriched.Article.Title, ArticleID: articleID, Time: time.Now()})
		h.hub.Broadcast(ws.Event{Type: "index_finished", Message: enriched.Article.Title, ArticleID: articleID, Time: time.Now()})
	}
	_ = h.repo.MarkFeedFetched(ctx, feed.ID)
	h.hub.Broadcast(ws.Event{Type: "fetch_finished", Message: feed.URL, Time: time.Now()})
	return out, http.StatusOK
}

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	articles, err := h.repo.ListArticles(r.Context(), page.PageSize, page.Offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": articles, "page": page.Page, "page_size": page.PageSize})
}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	article, err := h.repo.GetArticle(r.Context(), idParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusNotFound, "article not found")
		return
	}
	response.JSON(w, http.StatusOK, article)
}

func (h *Handler) GetArticleSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.repo.GetSummary(r.Context(), idParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusNotFound, "summary not found")
		return
	}
	response.JSON(w, http.StatusOK, summary)
}

func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.AddFavorite(r.Context(), middleware.UserID(r), idParam(r, "id"), "favorite"); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"favorite": true})
}

func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.RemoveFavorite(r.Context(), middleware.UserID(r), idParam(r, "id"), "favorite"); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"favorite": false})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	result, err := h.search.Search(r.Context(), search.SearchQuery{
		Query: r.URL.Query().Get("q"), Tag: r.URL.Query().Get("tag"), Author: r.URL.Query().Get("author"),
		Page: page.Page, PageSize: page.PageSize,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	var req dto.TextRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	summary, err := h.ai.Summarize(r.Context(), req.Text)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, summary)
}

func (h *Handler) Translate(w http.ResponseWriter, r *http.Request) {
	var req dto.TextRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	translated, err := h.ai.Translate(r.Context(), req.Text, req.TargetLanguage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"translation": translated})
}

func (h *Handler) Tags(w http.ResponseWriter, r *http.Request) {
	var req dto.TextRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	tags, err := h.ai.GenerateTags(r.Context(), req.Text)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"tags": tags})
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	var req dto.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Question) == "" {
		response.Error(w, http.StatusBadRequest, "question is required")
		return
	}
	answer, err := h.rag.Ask(r.Context(), req.Question)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, answer)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.repo.Dashboard(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, dashboard)
}

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	if err := ws.Serve(h.hub, w, r); err != nil {
		h.logger.Warn("websocket upgrade failed", zap.Error(err))
	}
}

func idParam(r *http.Request, name string) int64 {
	id, _ := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	return id
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
