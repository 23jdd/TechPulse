package rag

import (
	"context"

	"techpulse/internal/model"
	"techpulse/internal/search"
)

type Answer struct {
	Answer     string     `json:"answer"`
	Citations  []Citation `json:"citations"`
	Query      string     `json:"query"`
	Confidence float64    `json:"confidence"`
	NoAnswer   bool       `json:"no_answer"`
}

type Service struct {
	retriever *Retriever
	generator *Generator
	memory    MemoryStore
}

type MemoryStore interface {
	CreateConversation(context.Context, int64, string) (int64, error)
	StoreMessage(context.Context, int64, string, string, any) error
	RecentMessages(context.Context, int64, int) ([]model.Message, error)
}

func NewService(retriever *Retriever, generator *Generator) *Service {
	return &Service{retriever: retriever, generator: generator}
}

func (s *Service) WithMemory(memory MemoryStore) *Service {
	s.memory = memory
	return s
}

func (s *Service) Ask(ctx context.Context, question string) (*Answer, error) {
	return s.AskWithConversation(ctx, question, 0, 0)
}

func (s *Service) AskWithConversation(ctx context.Context, question string, userID, conversationID int64) (*Answer, error) {
	if s.memory != nil && conversationID == 0 && userID > 0 {
		id, err := s.memory.CreateConversation(ctx, userID, question)
		if err != nil {
			return nil, err
		}
		conversationID = id
	}
	query := s.generator.RewriteQuery(ctx, question)
	hits, err := s.retriever.Retrieve(ctx, query, 5)
	if err != nil {
		return nil, err
	}
	confidence := confidenceFromHits(hits)
	if len(hits) == 0 || confidence < 0.05 {
		answer := "The knowledge base does not contain enough relevant articles to answer this question yet."
		if s.memory != nil && conversationID > 0 {
			_ = s.memory.StoreMessage(ctx, conversationID, "user", question, nil)
			_ = s.memory.StoreMessage(ctx, conversationID, "assistant", answer, nil)
		}
		return &Answer{Answer: answer, Citations: []Citation{}, Query: query, Confidence: confidence, NoAnswer: true}, nil
	}
	var history []model.Message
	if s.memory != nil && conversationID > 0 {
		history, _ = s.memory.RecentMessages(ctx, conversationID, 8)
		_ = s.memory.StoreMessage(ctx, conversationID, "user", question, nil)
	}
	answer, err := s.generator.GenerateWithHistory(ctx, question, hits, history)
	if err != nil {
		return nil, err
	}
	citations := make([]Citation, 0, len(hits))
	for _, hit := range hits {
		citations = append(citations, Citation{ArticleID: hit.ID, Title: hit.Title, URL: hit.URL, Snippet: citationSnippet(hit), Score: hit.Score})
	}
	if s.memory != nil && conversationID > 0 {
		_ = s.memory.StoreMessage(ctx, conversationID, "assistant", answer, citations)
	}
	return &Answer{Answer: answer, Citations: citations, Query: query, Confidence: confidence, NoAnswer: false}, nil
}

func confidenceFromHits(hits []search.SearchHit) float64 {
	if len(hits) == 0 {
		return 0
	}
	top := hits[0].Score
	if top <= 0 {
		return 0
	}
	if top >= 10 {
		return 1
	}
	return top / 10
}

func citationSnippet(hit search.SearchHit) string {
	for _, fragments := range hit.Highlight {
		if len(fragments) > 0 {
			return fragments[0]
		}
	}
	if hit.Summary != "" {
		return hit.Summary
	}
	return hit.Title
}
