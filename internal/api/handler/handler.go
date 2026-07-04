package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	"techpulse/internal/opml"
	"techpulse/internal/parser"
	"techpulse/internal/pipeline"
	"techpulse/internal/rag"
	"techpulse/internal/report"
	"techpulse/internal/search"
	"techpulse/internal/storage/mysql"
	redisstore "techpulse/internal/storage/redis"
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
	cache     *redisstore.Client
	hub       *ws.Hub
	logger    *zap.Logger
}

func New(repo *mysql.Repository, fetcher *fetcher.Service, parser *parser.Service, duplicate *duplicate.Service, pipeline *pipeline.Service, search search.SearchEngine, rag *rag.Service, ai ai.Provider, cache *redisstore.Client, hub *ws.Hub, logger *zap.Logger) *Handler {
	return &Handler{repo: repo, fetcher: fetcher, parser: parser, duplicate: duplicate, pipeline: pipeline, search: search, rag: rag, ai: ai, cache: cache, hub: hub, logger: logger}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	redisStatus := "disabled"
	if h.cache != nil {
		redisStatus = "ok"
		if err := h.cache.Ping(r.Context()); err != nil {
			redisStatus = "error"
		}
	}
	response.JSON(w, http.StatusOK, response.Envelope{"status": "ok", "service": "gateway", "redis": redisStatus})
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
	h.deleteCache(r, "rss:list:user:"+strconv.FormatInt(feed.UserID, 10))
	response.JSON(w, http.StatusCreated, feed)
}

func (h *Handler) ListRSS(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	key := "rss:list:user:" + strconv.FormatInt(userID, 10)
	var cached []model.RSSFeed
	if h.getCache(r, key, &cached) {
		w.Header().Set("X-Cache", "HIT")
		response.JSON(w, http.StatusOK, response.Envelope{"items": cached})
		return
	}
	feeds, err := h.repo.ListFeeds(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.setCache(r, key, feeds, 2*time.Minute)
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
	h.deleteCache(r, "rss:list:user:"+strconv.FormatInt(middleware.UserID(r), 10))
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
		h.deleteCache(r, "dashboard:v1", "articles:list:page:1:size:20")
		h.hub.Broadcast(ws.Event{Type: "new_article", Message: enriched.Article.Title, ArticleID: articleID, Time: time.Now()})
		h.hub.Broadcast(ws.Event{Type: "index_finished", Message: enriched.Article.Title, ArticleID: articleID, Time: time.Now()})
	}
	_ = h.repo.MarkFeedFetched(ctx, feed.ID)
	h.hub.Broadcast(ws.Event{Type: "fetch_finished", Message: feed.URL, Time: time.Now()})
	return out, http.StatusOK
}

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	key := fmt.Sprintf("articles:list:page:%d:size:%d", page.Page, page.PageSize)
	var cached []model.Article
	if h.getCache(r, key, &cached) {
		w.Header().Set("X-Cache", "HIT")
		response.JSON(w, http.StatusOK, response.Envelope{"items": cached, "page": page.Page, "page_size": page.PageSize})
		return
	}
	articles, err := h.repo.ListArticles(r.Context(), page.PageSize, page.Offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.setCache(r, key, articles, time.Minute)
	response.JSON(w, http.StatusOK, response.Envelope{"items": articles, "page": page.Page, "page_size": page.PageSize})
}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id := idParam(r, "id")
	key := "article:" + strconv.FormatInt(id, 10)
	var cached model.Article
	if h.getCache(r, key, &cached) {
		w.Header().Set("X-Cache", "HIT")
		response.JSON(w, http.StatusOK, cached)
		return
	}
	article, err := h.repo.GetArticle(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "article not found")
		return
	}
	h.setCache(r, key, article, 10*time.Minute)
	response.JSON(w, http.StatusOK, article)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.MarkArticleRead(r.Context(), middleware.UserID(r), idParam(r, "id")); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"read": true})
}

func (h *Handler) GetArticleSummary(w http.ResponseWriter, r *http.Request) {
	id := idParam(r, "id")
	key := "summary:" + strconv.FormatInt(id, 10)
	var cached model.Summary
	if h.getCache(r, key, &cached) {
		w.Header().Set("X-Cache", "HIT")
		response.JSON(w, http.StatusOK, cached)
		return
	}
	summary, err := h.repo.GetSummary(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "summary not found")
		return
	}
	h.setCache(r, key, summary, 10*time.Minute)
	response.JSON(w, http.StatusOK, summary)
}

func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.AddFavorite(r.Context(), middleware.UserID(r), idParam(r, "id"), "favorite"); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"favorite": true})
}

func (h *Handler) AddReadLater(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.AddFavorite(r.Context(), middleware.UserID(r), idParam(r, "id"), "read_later"); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"read_later": true})
}

func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.RemoveFavorite(r.Context(), middleware.UserID(r), idParam(r, "id"), "favorite"); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"favorite": false})
}

func (h *Handler) RemoveReadLater(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.RemoveFavorite(r.Context(), middleware.UserID(r), idParam(r, "id"), "read_later"); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"read_later": false})
}

func (h *Handler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	typ := r.URL.Query().Get("type")
	if typ == "" {
		typ = "favorite"
	}
	articles, err := h.repo.ListFavorites(r.Context(), middleware.UserID(r), typ, page.PageSize, page.Offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": articles, "type": typ})
}

func (h *Handler) ReadingHistory(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	articles, err := h.repo.ReadingHistory(r.Context(), middleware.UserID(r), page.PageSize, page.Offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": articles})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	query := search.SearchQuery{
		Query: r.URL.Query().Get("q"), Tag: r.URL.Query().Get("tag"), Author: r.URL.Query().Get("author"),
		Page: page.Page, PageSize: page.PageSize,
	}
	key := fmt.Sprintf("search:q:%s:tag:%s:author:%s:page:%d:size:%d", query.Query, query.Tag, query.Author, query.Page, query.PageSize)
	var cached search.SearchResult
	if h.getCache(r, key, &cached) {
		w.Header().Set("X-Cache", "HIT")
		response.JSON(w, http.StatusOK, cached)
		return
	}
	result, err := h.search.Search(r.Context(), query)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.setCache(r, key, result, 30*time.Second)
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
	answer, err := h.rag.AskWithConversation(r.Context(), req.Question, middleware.UserID(r), req.ConversationID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, answer)
}

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	conversations, err := h.repo.ListConversations(r.Context(), middleware.UserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": conversations})
}

func (h *Handler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	prompts, err := h.repo.ListPrompts(r.Context(), middleware.UserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": prompts})
}

func (h *Handler) UpsertPrompt(w http.ResponseWriter, r *http.Request) {
	var req dto.PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Content) == "" {
		response.Error(w, http.StatusBadRequest, "name and content are required")
		return
	}
	prompt := &mysql.UserPrompt{UserID: middleware.UserID(r), Name: req.Name, Content: req.Content, IsDefault: req.IsDefault}
	if err := h.repo.UpsertPrompt(r.Context(), prompt); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, prompt)
}

func (h *Handler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeletePrompt(r.Context(), middleware.UserID(r), idParam(r, "id")); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"deleted": true})
}

func (h *Handler) ExportOPML(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.repo.ListFeeds(r.Context(), middleware.UserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/x-opml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="techpulse-feeds.opml"`)
	if err := opml.Encode(w, feeds); err != nil {
		h.logger.Warn("opml export failed", zap.Error(err))
	}
}

func (h *Handler) ImportOPML(w http.ResponseWriter, r *http.Request) {
	feeds, err := opml.Decode(r.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	var imported int
	for _, feed := range feeds {
		feed.UserID = middleware.UserID(r)
		if feed.Status == "" {
			feed.Status = "active"
		}
		if err := h.repo.CreateFeed(r.Context(), &feed); err == nil {
			imported++
		}
	}
	response.JSON(w, http.StatusOK, response.Envelope{"imported": imported})
}

func (h *Handler) GitHubAuthURL(w http.ResponseWriter, r *http.Request) {
	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	if clientID == "" {
		clientID = "configure-GITHUB_CLIENT_ID"
	}
	redirectURI := strings.TrimSpace(r.URL.Query().Get("redirect_uri"))
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/api/v1/auth/github/callback"
	}
	state := randomState()
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("scope", "read:user user:email")
	values.Set("state", state)
	response.JSON(w, http.StatusOK, response.Envelope{"url": "https://github.com/login/oauth/authorize?" + values.Encode(), "state": state})
}

func (h *Handler) GenerateDailyReport(w http.ResponseWriter, r *http.Request) {
	var req dto.ReportRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	reportSvc := report.NewService(h.repo)
	daily, err := reportSvc.Generate(r.Context(), middleware.UserID(r), req.Title)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, daily)
}

func (h *Handler) ListDailyReports(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	reports, err := h.repo.ListDailyReports(r.Context(), middleware.UserID(r), page.PageSize, page.Offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": reports})
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	var cached mysql.Dashboard
	if h.getCache(r, "dashboard:v1", &cached) {
		w.Header().Set("X-Cache", "HIT")
		response.JSON(w, http.StatusOK, cached)
		return
	}
	dashboard, err := h.repo.Dashboard(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.setCache(r, "dashboard:v1", dashboard, time.Minute)
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

func randomState() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func (h *Handler) getCache(r *http.Request, key string, dst any) bool {
	if h.cache == nil {
		return false
	}
	ok, err := h.cache.GetJSON(r.Context(), key, dst)
	if err != nil {
		h.logger.Debug("redis get failed", zap.String("key", key), zap.Error(err))
		return false
	}
	return ok
}

func (h *Handler) setCache(r *http.Request, key string, value any, ttl time.Duration) {
	if h.cache == nil {
		return
	}
	if err := h.cache.SetJSON(r.Context(), key, value, ttl); err != nil {
		h.logger.Debug("redis set failed", zap.String("key", key), zap.Error(err))
	}
}

func (h *Handler) deleteCache(r *http.Request, keys ...string) {
	if h.cache == nil {
		return
	}
	if err := h.cache.Delete(r.Context(), keys...); err != nil {
		h.logger.Debug("redis delete failed", zap.Strings("keys", keys), zap.Error(err))
	}
}
