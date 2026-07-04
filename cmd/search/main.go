package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/observability"
	searchengine "techpulse/internal/search"
	"techpulse/internal/service"
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

	service.RegisterSelf(context.Background(), cfg, "search", 8085, logger)
	server := service.NewServer("search", 8085, logger)
	server.Mux.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			service.Error(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
			return
		}
		var req service.IndexRequest
		if err := service.DecodeJSON(r, &req); err != nil {
			service.Error(w, http.StatusBadRequest, err)
			return
		}
		if err := engine.IndexArticle(r.Context(), req.Document); err != nil {
			service.Error(w, http.StatusInternalServerError, err)
			return
		}
		service.JSON(w, http.StatusOK, map[string]any{"indexed": true, "id": req.Document.ID})
	})
	server.Mux.HandleFunc("/index/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			service.Error(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
			return
		}
		id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/index/"), 10, 64)
		if err != nil {
			service.Error(w, http.StatusBadRequest, err)
			return
		}
		if err := engine.DeleteArticle(r.Context(), id); err != nil {
			service.Error(w, http.StatusInternalServerError, err)
			return
		}
		service.JSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
	})
	server.Mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := searchengine.SearchQuery{Page: 1, PageSize: 20}
		if r.Method == http.MethodPost {
			var req service.SearchRequest
			if err := service.DecodeJSON(r, &req); err != nil {
				service.Error(w, http.StatusBadRequest, err)
				return
			}
			query = req.Query
		} else {
			query.Query = r.URL.Query().Get("q")
			query.Tag = r.URL.Query().Get("tag")
			query.Author = r.URL.Query().Get("author")
			query.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
			query.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
		}
		result, err := engine.Search(r.Context(), query)
		if err != nil {
			service.Error(w, http.StatusInternalServerError, err)
			return
		}
		service.JSON(w, http.StatusOK, result)
	})
	if err := server.Run(context.Background()); err != nil {
		logger.Fatal("search stopped", zap.Error(err))
	}
}
