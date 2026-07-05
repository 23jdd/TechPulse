package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"techpulse/internal/ai"
	"techpulse/internal/api/dto"
	"techpulse/internal/api/middleware"
	"techpulse/internal/auth"
	"techpulse/internal/duplicate"
	"techpulse/internal/email"
	"techpulse/internal/fetcher"
	"techpulse/internal/model"
	"techpulse/internal/opml"
	"techpulse/internal/parser"
	"techpulse/internal/pipeline"
	"techpulse/internal/queue"
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
	oauth     *auth.GitHubOAuth
	mailer    email.Sender
	hub       *ws.Hub
	broker    queue.Broker
	jwtSecret string
	logger    *zap.Logger
}

func New(repo *mysql.Repository, fetcher *fetcher.Service, parser *parser.Service, duplicate *duplicate.Service, pipeline *pipeline.Service, search search.SearchEngine, rag *rag.Service, ai ai.Provider, cache *redisstore.Client, oauth *auth.GitHubOAuth, mailer email.Sender, hub *ws.Hub, broker queue.Broker, jwtSecret string, logger *zap.Logger) *Handler {
	return &Handler{repo: repo, fetcher: fetcher, parser: parser, duplicate: duplicate, pipeline: pipeline, search: search, rag: rag, ai: ai, cache: cache, oauth: oauth, mailer: mailer, hub: hub, broker: broker, jwtSecret: jwtSecret, logger: logger}
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

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	user, err := h.repo.GetUser(r.Context(), userID)
	if err != nil {
		response.JSON(w, http.StatusOK, response.Envelope{"id": userID, "username": "demo"})
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) GetPreference(w http.ResponseWriter, r *http.Request) {
	pref, err := h.repo.GetPreference(r.Context(), middleware.UserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, pref)
}

func (h *Handler) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	var req dto.PreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	rawTags, _ := json.Marshal(uniqueStrings(req.InterestedTags))
	pref := &model.UserPreference{
		UserID:             middleware.UserID(r),
		InterestedTags:     string(rawTags),
		DailyReportTime:    defaultString(req.DailyReportTime, "09:00"),
		DailyReportEmail:   req.DailyReportEmail,
		DailyReportEnabled: req.DailyReportEnabled,
		Timezone:           defaultString(req.Timezone, "Asia/Shanghai"),
	}
	if err := h.repo.UpsertPreference(r.Context(), pref); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, pref)
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
	feed := &model.RSSFeed{UserID: middleware.UserID(r), URL: req.URL, Title: req.Title, Category: req.Category, Status: "active", FetchInterval: req.FetchIntervalMinutes}
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

func (h *Handler) UpdateRSS(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateRSSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		response.Error(w, http.StatusBadRequest, "url is required")
		return
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "disabled" {
		response.Error(w, http.StatusBadRequest, "status must be active or disabled")
		return
	}
	feed := &model.RSSFeed{
		ID:            idParam(r, "id"),
		UserID:        middleware.UserID(r),
		URL:           req.URL,
		Title:         req.Title,
		Category:      req.Category,
		Status:        status,
		FetchInterval: req.FetchIntervalMinutes,
	}
	if err := h.repo.UpdateFeed(r.Context(), feed); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.deleteCache(r, "rss:list:user:"+strconv.FormatInt(feed.UserID, 10))
	updated, err := h.repo.GetFeed(r.Context(), feed.ID)
	if err != nil {
		response.JSON(w, http.StatusOK, feed)
		return
	}
	response.JSON(w, http.StatusOK, updated)
}

func (h *Handler) DeleteRSS(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteFeed(r.Context(), idParam(r, "id")); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.deleteCache(r, "rss:list:user:"+strconv.FormatInt(middleware.UserID(r), 10))
	response.JSON(w, http.StatusOK, response.Envelope{"deleted": true})
}

func (h *Handler) EnableRSS(w http.ResponseWriter, r *http.Request) {
	h.setFeedStatus(w, r, "active")
}

func (h *Handler) DisableRSS(w http.ResponseWriter, r *http.Request) {
	h.setFeedStatus(w, r, "disabled")
}

func (h *Handler) setFeedStatus(w http.ResponseWriter, r *http.Request, status string) {
	userID := middleware.UserID(r)
	if err := h.repo.SetFeedStatus(r.Context(), userID, idParam(r, "id"), status); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.deleteCache(r, "rss:list:user:"+strconv.FormatInt(userID, 10))
	response.JSON(w, http.StatusOK, response.Envelope{"status": status})
}

func (h *Handler) FetchRSS(w http.ResponseWriter, r *http.Request) {
	result, status := h.fetchFeedContext(r.Context(), idParam(r, "id"))
	response.JSON(w, status, result)
}

func (h *Handler) FetchRSSAsync(w http.ResponseWriter, r *http.Request) {
	feedID := idParam(r, "id")
	payload, _ := json.Marshal(response.Envelope{
		"feed_id":   feedID,
		"user_id":   middleware.UserID(r),
		"queued_at": time.Now().Format(time.RFC3339),
	})
	taskID, err := h.repo.CreateTask(r.Context(), "rss_fetch", string(payload))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	msg := queue.Message{Type: queue.FetchJob, Payload: map[string]any{
		"task_id":   taskID,
		"feed_id":   feedID,
		"user_id":   middleware.UserID(r),
		"queued_at": time.Now().Format(time.RFC3339),
	}}
	if h.broker != nil {
		if err := h.broker.Publish(r.Context(), "fetch", msg); err == nil {
			response.JSON(w, http.StatusAccepted, response.Envelope{"queued": true, "queue": "fetch", "task_id": taskID, "feed_id": feedID})
			return
		} else {
			h.logger.Warn("rabbitmq publish failed; falling back to in-process fetch", zap.Int64("task_id", taskID), zap.Error(err))
			_ = h.repo.MarkTaskRetrying(r.Context(), taskID, err.Error())
		}
	}
	go h.runFetchTask(taskID, feedID)
	response.JSON(w, http.StatusAccepted, response.Envelope{"queued": true, "queue": "in-process", "task_id": taskID, "feed_id": feedID})
}

func (h *Handler) TestRSS(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	feed, err := h.repo.GetFeed(r.Context(), idParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusNotFound, "feed not found")
		return
	}
	out := h.testFeed(r, feed.ID, feed.URL, feed.Title, feed.Category, start)
	status := http.StatusOK
	if !out.OK {
		status = http.StatusBadGateway
	}
	response.JSON(w, status, out)
}

func (h *Handler) fetchFeedContext(ctx context.Context, feedID int64) (*dto.FetchRSSResponse, int) {
	start := time.Now()
	feed, err := h.repo.GetFeed(ctx, feedID)
	if err != nil {
		return &dto.FetchRSSResponse{FeedID: feedID, Errors: []string{"feed not found"}}, http.StatusNotFound
	}
	if feed.Status == "disabled" {
		return &dto.FetchRSSResponse{FeedID: feed.ID, Errors: []string{"feed is disabled"}}, http.StatusConflict
	}
	h.hub.Broadcast(ws.Event{Type: "fetch_started", Message: feed.URL, Time: time.Now()})
	items, err := h.fetcher.Fetch(ctx, fetcher.Source{ID: feed.ID, Type: "rss", URL: feed.URL, Title: feed.Title, Category: feed.Category})
	if err != nil {
		_ = h.repo.MarkFeedHealth(ctx, feed.ID, false, time.Since(start), err.Error())
		h.hub.Broadcast(ws.Event{Type: "task_failed", Message: err.Error(), Time: time.Now()})
		return &dto.FetchRSSResponse{FeedID: feed.ID, Errors: []string{err.Error()}}, http.StatusBadGateway
	}
	inserted, duplicates, errors := h.ingestFetchedItems(ctx, items)
	out := &dto.FetchRSSResponse{FeedID: feed.ID, Fetched: len(items), Inserted: inserted, Duplicates: duplicates, Errors: errors}
	_ = h.repo.MarkFeedFetched(ctx, feed.ID)
	_ = h.repo.MarkFeedHealth(ctx, feed.ID, len(errors) == 0, time.Since(start), strings.Join(errors, "; "))
	h.hub.Broadcast(ws.Event{Type: "fetch_finished", Message: feed.URL, Time: time.Now()})
	return out, http.StatusOK
}

func (h *Handler) runFetchTask(taskID, feedID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if err := h.repo.MarkTaskRunning(ctx, taskID); err != nil {
		h.logger.Warn("mark fetch task running failed", zap.Int64("task_id", taskID), zap.Error(err))
	}
	result, status := h.fetchFeedContext(ctx, feedID)
	if status >= http.StatusBadRequest {
		message := strings.Join(result.Errors, "; ")
		if message == "" {
			message = "fetch failed"
		}
		_ = h.repo.MarkTaskFailed(ctx, taskID, message)
		return
	}
	if len(result.Errors) > 0 && result.Inserted == 0 && result.Fetched > 0 {
		_ = h.repo.MarkTaskFailed(ctx, taskID, strings.Join(result.Errors, "; "))
		return
	}
	_ = h.repo.MarkTaskSuccess(ctx, taskID)
}

func (h *Handler) testFeed(r *http.Request, feedID int64, feedURL, title, category string, start time.Time) dto.TestRSSResponse {
	items, err := h.fetcher.Fetch(r.Context(), fetcher.Source{ID: feedID, Type: "rss", URL: feedURL, Title: title, Category: category})
	out := dto.TestRSSResponse{FeedID: feedID, URL: feedURL, OK: err == nil, Fetched: len(items), Duration: time.Since(start).String()}
	if len(items) > 0 {
		out.Title = items[0].Title
	}
	if err != nil {
		out.Errors = []string{err.Error()}
	}
	return out
}

func (h *Handler) FetchGitHubReleases(w http.ResponseWriter, r *http.Request) {
	var req dto.FetchGitHubReleasesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		response.Error(w, http.StatusBadRequest, "url is required")
		return
	}
	h.hub.Broadcast(ws.Event{Type: "fetch_started", Message: req.URL, Time: time.Now()})
	items, err := h.fetcher.Fetch(r.Context(), fetcher.Source{Type: "github_release", URL: req.URL, Title: req.URL, Category: "GitHub"})
	if err != nil {
		h.hub.Broadcast(ws.Event{Type: "task_failed", Message: err.Error(), Time: time.Now()})
		response.Error(w, http.StatusBadGateway, err.Error())
		return
	}
	inserted, duplicates, errors := h.ingestFetchedItems(r.Context(), items)
	h.hub.Broadcast(ws.Event{Type: "fetch_finished", Message: req.URL, Time: time.Now()})
	response.JSON(w, http.StatusOK, dto.FetchSourceResponse{
		SourceType: "github_release",
		SourceURL:  req.URL,
		Fetched:    len(items),
		Inserted:   inserted,
		Duplicates: duplicates,
		Errors:     errors,
	})
}

func (h *Handler) ListGitHubRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := h.repo.ListGitHubRepos(r.Context(), middleware.UserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": repos})
}

func (h *Handler) MonitorGitHubRepo(w http.ResponseWriter, r *http.Request) {
	var req dto.MonitorGitHubRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	owner, name, err := parseGitHubRepo(req.URL)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	repo, err := h.fetchGitHubRepoSnapshot(r.Context(), middleware.UserID(r), owner, name)
	if err != nil {
		response.Error(w, http.StatusBadGateway, err.Error())
		return
	}
	if err := h.repo.UpsertGitHubRepo(r.Context(), repo); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, repo)
}

func (h *Handler) fetchGitHubRepoSnapshot(ctx context.Context, userID int64, owner, name string) (*model.GitHubRepo, error) {
	var repoResp struct {
		FullName      string `json:"full_name"`
		HTMLURL       string `json:"html_url"`
		Description   string `json:"description"`
		Stars         int64  `json:"stargazers_count"`
		OpenIssues    int64  `json:"open_issues_count"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := fetchGitHubJSON(ctx, "https://api.github.com/repos/"+path.Join(owner, name), &repoResp); err != nil {
		return nil, err
	}
	now := time.Now()
	out := &model.GitHubRepo{
		UserID:        userID,
		Owner:         owner,
		Name:          name,
		HTMLURL:       repoResp.HTMLURL,
		Description:   repoResp.Description,
		Stars:         repoResp.Stars,
		OpenIssues:    repoResp.OpenIssues,
		DefaultBranch: repoResp.DefaultBranch,
		LastCheckedAt: &now,
	}
	var releaseResp struct {
		Name        string     `json:"name"`
		TagName     string     `json:"tag_name"`
		HTMLURL     string     `json:"html_url"`
		Body        string     `json:"body"`
		PublishedAt *time.Time `json:"published_at"`
	}
	if err := fetchGitHubJSON(ctx, "https://api.github.com/repos/"+path.Join(owner, name, "releases/latest"), &releaseResp); err == nil {
		title := strings.TrimSpace(releaseResp.Name)
		if title == "" {
			title = releaseResp.TagName
		}
		out.LatestRelease = title
		out.LatestReleaseURL = releaseResp.HTMLURL
		out.LatestReleaseAt = releaseResp.PublishedAt
		text := strings.ToLower(title + "\n" + releaseResp.Body)
		out.BreakingChange = strings.Contains(text, "breaking") || strings.Contains(text, "migration") || strings.Contains(text, "incompatible")
		out.SecurityUpdate = strings.Contains(text, "security") || strings.Contains(text, "cve-") || strings.Contains(text, "vulnerab")
	}
	if out.HTMLURL == "" {
		out.HTMLURL = "https://github.com/" + path.Join(owner, name)
	}
	return out, nil
}

func fetchGitHubJSON(ctx context.Context, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "TechPulse")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("github api %s returned %d: %s", endpoint, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func parseGitHubRepo(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("url is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	if parsed.Host == "api.github.com" {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) >= 3 && parts[0] == "repos" {
			return parts[1], strings.TrimSuffix(parts[2], ".git"), nil
		}
	}
	if parsed.Host != "github.com" {
		return "", "", fmt.Errorf("github repository URL must look like https://github.com/{owner}/{repo}")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("github repository URL must look like https://github.com/{owner}/{repo}")
	}
	return parts[0], strings.TrimSuffix(parts[1], ".git"), nil
}

func (h *Handler) FetchHackerNews(w http.ResponseWriter, r *http.Request) {
	var req dto.FetchHackerNewsRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	feed := strings.TrimSpace(req.Feed)
	if feed == "" {
		feed = "top"
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	sourceURL := fmt.Sprintf("%s?limit=%d", feed, limit)
	h.hub.Broadcast(ws.Event{Type: "fetch_started", Message: "Hacker News " + feed, Time: time.Now()})
	items, err := h.fetcher.Fetch(r.Context(), fetcher.Source{Type: "hackernews", URL: sourceURL, Title: "Hacker News " + feed, Category: "Hacker News"})
	if err != nil {
		h.hub.Broadcast(ws.Event{Type: "task_failed", Message: err.Error(), Time: time.Now()})
		response.Error(w, http.StatusBadGateway, err.Error())
		return
	}
	inserted, duplicates, errors := h.ingestFetchedItems(r.Context(), items)
	h.hub.Broadcast(ws.Event{Type: "fetch_finished", Message: "Hacker News " + feed, Time: time.Now()})
	response.JSON(w, http.StatusOK, dto.FetchSourceResponse{
		SourceType: "hackernews",
		SourceURL:  sourceURL,
		Fetched:    len(items),
		Inserted:   inserted,
		Duplicates: duplicates,
		Errors:     errors,
	})
}

func (h *Handler) ingestFetchedItems(ctx context.Context, items []fetcher.FetchedItem) (int, int, []string) {
	var inserted int
	var duplicates int
	var errors []string
	for _, item := range items {
		parsed, err := h.parser.Parse(ctx, item)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		content := parsed.CleanContent
		if content == "" {
			content = parsed.Title
		}
		exists, urlHash, contentHash, err := h.duplicate.IsDuplicate(ctx, parsed.URL, content)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		if exists {
			duplicates++
			continue
		}
		enriched, err := h.pipeline.Process(ctx, parsed)
		if err != nil {
			errors = append(errors, err.Error())
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
			errors = append(errors, err.Error())
			continue
		}
		enriched.Article.ID = articleID
		if err := h.search.IndexArticle(ctx, search.DocumentFromArticle(enriched.Article, enriched.Summary.ShortSummary, stored.Tags)); err != nil {
			errors = append(errors, err.Error())
			continue
		}
		inserted++
		h.deleteCacheContext(ctx, "dashboard:v1", "articles:list:page:1:size:20")
		h.hub.Broadcast(ws.Event{Type: "new_article", Message: enriched.Article.Title, ArticleID: articleID, Time: time.Now()})
		h.hub.Broadcast(ws.Event{Type: "index_finished", Message: enriched.Article.Title, ArticleID: articleID, Time: time.Now()})
	}
	return inserted, duplicates, errors
}

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := mysql.ArticleFilter{
		UserID:     middleware.UserID(r),
		SourceType: r.URL.Query().Get("source"),
		Tag:        r.URL.Query().Get("tag"),
		Limit:      page.PageSize,
		Offset:     page.Offset,
	}
	if sourceID, err := strconv.ParseInt(r.URL.Query().Get("source_id"), 10, 64); err == nil && sourceID > 0 {
		filter.SourceID = sourceID
	}
	if v, ok := parseBoolQuery(r, "read"); ok {
		filter.IsRead = &v
	}
	if v, ok := parseBoolQuery(r, "favorite"); ok {
		filter.IsFavorite = &v
	}
	if v, ok := parseBoolQuery(r, "read_later"); ok {
		filter.IsReadLater = &v
	}
	if v, ok := parseBoolQuery(r, "archived"); ok {
		filter.IsArchived = &v
	}
	if from, ok := parseDateQuery(r, "from"); ok {
		filter.From = &from
	}
	if to, ok := parseDateQuery(r, "to"); ok {
		filter.To = &to
	}
	key := fmt.Sprintf("articles:list:user:%d:source:%s:source_id:%d:tag:%s:read:%s:favorite:%s:read_later:%s:archived:%s:from:%s:to:%s:page:%d:size:%d",
		filter.UserID, filter.SourceType, filter.SourceID, filter.Tag, r.URL.Query().Get("read"), r.URL.Query().Get("favorite"),
		r.URL.Query().Get("read_later"), r.URL.Query().Get("archived"), r.URL.Query().Get("from"), r.URL.Query().Get("to"), page.Page, page.PageSize)
	var cached []model.Article
	if h.getCache(r, key, &cached) {
		w.Header().Set("X-Cache", "HIT")
		response.JSON(w, http.StatusOK, response.Envelope{"items": cached, "page": page.Page, "page_size": page.PageSize})
		return
	}
	articles, err := h.repo.ListArticlesFiltered(r.Context(), filter)
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

func (h *Handler) ArchiveArticle(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.AddFavorite(r.Context(), middleware.UserID(r), idParam(r, "id"), "archived"); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.deleteCache(r, "dashboard:v1")
	response.JSON(w, http.StatusOK, response.Envelope{"archived": true})
}

func (h *Handler) UnarchiveArticle(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.RemoveFavorite(r.Context(), middleware.UserID(r), idParam(r, "id"), "archived"); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"archived": false})
}

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := idParam(r, "id")
	if err := h.repo.DeleteArticle(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.search.DeleteArticle(r.Context(), id); err != nil {
		h.logger.Warn("search delete failed", zap.Int64("article_id", id), zap.Error(err))
	}
	h.deleteCache(r, "dashboard:v1", "article:"+strconv.FormatInt(id, 10), "summary:"+strconv.FormatInt(id, 10))
	response.JSON(w, http.StatusOK, response.Envelope{"deleted": true})
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
		Source: r.URL.Query().Get("source"), Page: page.Page, PageSize: page.PageSize,
	}
	if sourceID, err := strconv.ParseInt(r.URL.Query().Get("source_id"), 10, 64); err == nil && sourceID > 0 {
		query.SourceID = sourceID
	}
	if from, ok := parseDateQuery(r, "from"); ok {
		query.DateFrom = &from
	}
	if to, ok := parseDateQuery(r, "to"); ok {
		query.DateTo = &to
	}
	key := fmt.Sprintf("search:q:%s:tag:%s:author:%s:source:%s:source_id:%d:from:%s:to:%s:page:%d:size:%d",
		query.Query, query.Tag, query.Author, query.Source, query.SourceID, r.URL.Query().Get("from"), r.URL.Query().Get("to"), query.Page, query.PageSize)
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

func (h *Handler) SearchExplain(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, response.Envelope{
		"query":  r.URL.Query().Get("q"),
		"fields": []string{"title^3", "summary", "content", "tags"},
		"filters": response.Envelope{
			"tag":       r.URL.Query().Get("tag"),
			"author":    r.URL.Query().Get("author"),
			"source":    r.URL.Query().Get("source"),
			"source_id": r.URL.Query().Get("source_id"),
			"from":      r.URL.Query().Get("from"),
			"to":        r.URL.Query().Get("to"),
		},
		"ranking": []string{
			"Bleve lexical score with title boost",
			"Optional embedding rerank through HybridEngine",
			"Future freshness/tag-weight score can be added behind SearchEngine",
		},
		"highlight": true,
	})
}

func (h *Handler) ReindexSearch(w http.ResponseWriter, r *http.Request) {
	articles, err := h.repo.ListArticles(r.Context(), 1000, 0)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	var indexed int
	var errors []string
	for _, article := range articles {
		summary := ""
		if s, err := h.repo.GetSummary(r.Context(), article.ID); err == nil {
			summary = s.ShortSummary
		}
		tags, _ := h.repo.TagsForArticle(r.Context(), article.ID)
		if err := h.search.IndexArticle(r.Context(), search.DocumentFromArticle(article, summary, tags)); err != nil {
			errors = append(errors, fmt.Sprintf("article %d: %v", article.ID, err))
			continue
		}
		indexed++
	}
	h.deleteCache(r, "dashboard:v1")
	response.JSON(w, http.StatusOK, response.Envelope{"indexed": indexed, "errors": errors})
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
	if h.oauth == nil || !h.oauth.Enabled() {
		response.Error(w, http.StatusServiceUnavailable, "github oauth is not configured")
		return
	}
	state := randomState()
	if r.URL.Query().Get("ui") == "1" {
		state = "ui:" + r.URL.Query().Get("locale") + ":" + state
	}
	authURL, err := h.oauth.AuthURL(state)
	if err != nil {
		response.Error(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"url": authURL, "state": state})
}

func (h *Handler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if h.oauth == nil || !h.oauth.Enabled() {
		response.Error(w, http.StatusServiceUnavailable, "github oauth is not configured")
		return
	}
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}
	githubUser, err := h.oauth.ExchangeAndFetchUser(r.Context(), code)
	if err != nil {
		response.Error(w, http.StatusBadGateway, err.Error())
		return
	}
	user, err := h.repo.UpsertGitHubUser(r.Context(), *githubUser)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if strings.HasPrefix(r.URL.Query().Get("state"), "ui:") {
		sessionToken, err := auth.SignJWT(h.jwtSecret, user.ID, user.Username, 24*time.Hour)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.renderOAuthSuccess(w, user, sessionToken, strings.HasPrefix(r.URL.Query().Get("state"), "ui:zh:"))
		return
	}
	sessionToken, err := auth.SignJWT(h.jwtSecret, user.ID, user.Username, 24*time.Hour)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"user": user, "api_token": githubUser.APIToken, "session_token": sessionToken})
}

func (h *Handler) renderOAuthSuccess(w http.ResponseWriter, user *model.User, token string, zh bool) {
	userJSON, _ := json.Marshal(user)
	tokenJSON, _ := json.Marshal(token)
	redirect := "/app"
	title := "Signed in"
	message := "GitHub login succeeded. Redirecting to app..."
	if zh {
		redirect = "/app/zh"
		title = "登录成功"
		message = "GitHub 登录成功，正在进入控制台..."
	}
	if zh {
		title = "登录成功"
		message = "GitHub 登录成功，正在进入应用..."
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s - TechPulse</title></head>
<body style="font-family:system-ui,sans-serif;background:#f6f7f9;color:#17202a;display:grid;place-items:center;min-height:100vh;margin:0">
<main style="background:white;border:1px solid #d9dee7;border-radius:8px;padding:24px;max-width:420px">
<h1 style="margin:0 0 8px;font-size:20px">%s</h1>
<p style="color:#667085">%s</p>
</main>
<script>
localStorage.setItem("techpulse_api_token", %s);
localStorage.setItem("techpulse_session", JSON.stringify({mode:"github", user:%s}));
setTimeout(function(){ window.location.href = %q; }, 600);
</script></body></html>`, title, title, message, string(tokenJSON), string(userJSON), redirect)
}

func (h *Handler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.mailer == nil || !h.mailer.Enabled() {
		response.Error(w, http.StatusServiceUnavailable, "smtp email is not configured")
		return
	}
	if strings.TrimSpace(req.Subject) == "" {
		req.Subject = "TechPulse test email"
	}
	if strings.TrimSpace(req.Body) == "" {
		req.Body = "TechPulse SMTP delivery is working."
	}
	if err := h.mailer.Send(r.Context(), email.Message{To: req.To, Subject: req.Subject, Body: req.Body}); err != nil {
		response.Error(w, http.StatusBadGateway, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"sent": true})
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
	sent := false
	if req.SendEmail {
		if h.mailer == nil || !h.mailer.Enabled() {
			response.Error(w, http.StatusServiceUnavailable, "smtp email is not configured")
			return
		}
		if err := h.mailer.Send(r.Context(), email.Message{To: req.EmailTo, Subject: daily.Title, Body: daily.Content}); err != nil {
			response.Error(w, http.StatusBadGateway, err.Error())
			return
		}
		sent = true
	}
	response.JSON(w, http.StatusCreated, response.Envelope{"report": daily, "email_sent": sent})
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

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	tasks, err := h.repo.ListTasks(r.Context(), r.URL.Query().Get("status"), page.PageSize, page.Offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, response.Envelope{"items": tasks, "page": page.Page, "page_size": page.PageSize})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, err := h.repo.GetTask(r.Context(), idParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusNotFound, "task not found")
		return
	}
	response.JSON(w, http.StatusOK, task)
}

func (h *Handler) Trends(w http.ResponseWriter, r *http.Request) {
	days := 7
	if raw := strings.TrimSpace(r.URL.Query().Get("days")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			days = parsed
		}
	}
	trends, err := h.repo.Trends(r.Context(), days)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, trends)
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

func parseBoolQuery(r *http.Request, name string) (bool, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return false, false
	}
	value, err := strconv.ParseBool(raw)
	return value, err == nil
}

func parseDateQuery(r *http.Request, name string) (time.Time, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.DateOnly, raw); err == nil {
		if name == "to" {
			t = t.Add(24*time.Hour - time.Nanosecond)
		}
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, true
	}
	return time.Time{}, false
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

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
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
	h.deleteCacheContext(r.Context(), keys...)
}

func (h *Handler) deleteCacheContext(ctx context.Context, keys ...string) {
	if h.cache == nil {
		return
	}
	if err := h.cache.Delete(ctx, keys...); err != nil {
		h.logger.Debug("redis delete failed", zap.Strings("keys", keys), zap.Error(err))
	}
}
