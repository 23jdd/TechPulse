package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"techpulse/internal/fetcher"
	"techpulse/internal/parser"
	"techpulse/internal/rag"
	"techpulse/internal/search"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), HTTP: httpClient}
}

func (c *Client) Fetch(ctx context.Context, source fetcher.Source) ([]fetcher.FetchedItem, error) {
	var out FetchResponse
	err := c.call(ctx, http.MethodPost, "/fetch", FetchRequest{Source: source}, &out)
	return out.Items, err
}

func (c *Client) Parse(ctx context.Context, item fetcher.FetchedItem) (*parser.ParsedArticle, error) {
	var out ParseResponse
	err := c.call(ctx, http.MethodPost, "/parse", ParseRequest{Item: item}, &out)
	return out.Article, err
}

func (c *Client) Process(ctx context.Context, article *parser.ParsedArticle) (*ProcessResponse, error) {
	var out ProcessResponse
	err := c.call(ctx, http.MethodPost, "/process", ProcessRequest{Article: article}, &out)
	return &out, err
}

func (c *Client) Index(ctx context.Context, document search.ArticleSearchDocument) error {
	return c.call(ctx, http.MethodPost, "/index", IndexRequest{Document: document}, nil)
}

func (c *Client) Search(ctx context.Context, query search.SearchQuery) (*search.SearchResult, error) {
	var out search.SearchResult
	err := c.call(ctx, http.MethodPost, "/search", SearchRequest{Query: query}, &out)
	return &out, err
}

func (c *Client) Chat(ctx context.Context, question string) (*rag.Answer, error) {
	var out rag.Answer
	err := c.call(ctx, http.MethodPost, "/chat", ChatRequest{Question: question}, &out)
	return &out, err
}

func (c *Client) call(ctx context.Context, method, path string, in, out any) error {
	raw, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return fmt.Errorf("service call %s %s failed: status=%d body=%v", method, path, resp.StatusCode, body)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
