package rag

import "context"

type Answer struct {
	Answer    string     `json:"answer"`
	Citations []Citation `json:"citations"`
}

type Service struct {
	retriever *Retriever
	generator *Generator
}

func NewService(retriever *Retriever, generator *Generator) *Service {
	return &Service{retriever: retriever, generator: generator}
}

func (s *Service) Ask(ctx context.Context, question string) (*Answer, error) {
	hits, err := s.retriever.Retrieve(ctx, question, 5)
	if err != nil {
		return nil, err
	}
	answer, err := s.generator.Generate(ctx, question, hits)
	if err != nil {
		return nil, err
	}
	citations := make([]Citation, 0, len(hits))
	for _, hit := range hits {
		citations = append(citations, Citation{ArticleID: hit.ID, Title: hit.Title, URL: hit.URL})
	}
	return &Answer{Answer: answer, Citations: citations}, nil
}
