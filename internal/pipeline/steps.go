package pipeline

import (
	"context"

	"techpulse/internal/ai"
	"techpulse/internal/parser"
)

type CleanStep struct{}

func (CleanStep) Name() string { return "clean" }
func (CleanStep) Execute(_ context.Context, input *PipelineInput) error {
	input.Parsed.CleanContent = parser.CleanHTML(input.Parsed.CleanContent)
	return nil
}

type LanguageStep struct{}

func (LanguageStep) Name() string { return "language_detect" }
func (LanguageStep) Execute(_ context.Context, input *PipelineInput) error {
	input.Parsed.Language = parser.DetectLanguage(input.Parsed.CleanContent)
	return nil
}

type TranslateStep struct {
	Provider       ai.Provider
	TargetLanguage string
}

func (TranslateStep) Name() string { return "translate" }
func (s TranslateStep) Execute(ctx context.Context, input *PipelineInput) error {
	translated, err := s.Provider.Translate(ctx, input.Parsed.CleanContent, s.TargetLanguage)
	if err != nil {
		return err
	}
	input.Translation = translated
	return nil
}

type SummaryStep struct{ Provider ai.Provider }

func (SummaryStep) Name() string { return "summary" }
func (s SummaryStep) Execute(ctx context.Context, input *PipelineInput) error {
	summary, err := s.Provider.Summarize(ctx, input.Parsed.CleanContent)
	if err != nil {
		return err
	}
	input.Summary.OneSentence = summary.OneSentence
	input.Summary.ShortSummary = summary.ShortSummary
	input.Summary.LongSummary = summary.LongSummary
	input.Summary.BulletPoints = JoinBullets(summary.BulletPoints)
	input.Summary.TLDR = summary.TLDR
	return nil
}

type KeywordsStep struct{ Provider ai.Provider }

func (KeywordsStep) Name() string { return "keywords" }
func (s KeywordsStep) Execute(ctx context.Context, input *PipelineInput) error {
	keywords, err := s.Provider.GenerateKeywords(ctx, input.Parsed.CleanContent)
	if err != nil {
		return err
	}
	input.Keywords = keywords
	return nil
}

type TagsStep struct{ Provider ai.Provider }

func (TagsStep) Name() string { return "tags" }
func (s TagsStep) Execute(ctx context.Context, input *PipelineInput) error {
	tags, err := s.Provider.GenerateTags(ctx, input.Parsed.CleanContent)
	if err != nil {
		return err
	}
	input.Tags = append(input.Parsed.Tags, tags...)
	return nil
}

type EmbeddingStep struct{ Provider ai.Provider }

func (EmbeddingStep) Name() string { return "embedding" }
func (s EmbeddingStep) Execute(ctx context.Context, input *PipelineInput) error {
	embedding, err := s.Provider.GenerateEmbedding(ctx, input.Parsed.CleanContent)
	if err != nil {
		return err
	}
	input.Embedding = embedding
	return nil
}
