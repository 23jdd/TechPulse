package pipeline

import (
	"context"
	"encoding/json"
	"strings"

	"techpulse/internal/ai"
	"techpulse/internal/model"
	"techpulse/internal/parser"
)

type PipelineInput struct {
	Parsed      *parser.ParsedArticle
	Article     model.Article
	Summary     model.Summary
	Translation string
	Tags        []string
	Keywords    []string
	Embedding   []float64
}

type Service struct {
	processor *Processor
}

func NewService(provider ai.Provider) *Service {
	return &Service{processor: NewProcessor(
		CleanStep{},
		LanguageStep{},
		TranslateStep{Provider: provider, TargetLanguage: "zh-CN"},
		SummaryStep{Provider: provider},
		KeywordsStep{Provider: provider},
		TagsStep{Provider: provider},
		EmbeddingStep{Provider: provider},
	)}
}

func (s *Service) Process(ctx context.Context, parsed *parser.ParsedArticle) (*PipelineInput, error) {
	input := &PipelineInput{Parsed: parsed}
	if err := s.processor.Run(ctx, input); err != nil {
		return nil, err
	}
	rawEmbedding, _ := json.Marshal(input.Embedding)
	input.Article = model.Article{
		SourceType:   parsed.SourceType,
		SourceID:     parsed.SourceID,
		Title:        parsed.Title,
		URL:          parsed.URL,
		Author:       parsed.Author,
		Language:     parsed.Language,
		RawContent:   parsed.RawContent,
		CleanContent: parsed.CleanContent,
		CoverImage:   parsed.CoverImage,
		PublishedAt:  parsed.PublishedAt,
	}
	input.Summary.Language = parsed.Language
	_ = rawEmbedding
	return input, nil
}

func JoinBullets(values []string) string {
	return strings.Join(values, "\n")
}
