package report

import (
	"context"
	"strings"
	"time"

	"techpulse/internal/model"
)

type ArticleLister interface {
	ListArticles(context.Context, int, int) ([]model.Article, error)
	StoreDailyReport(context.Context, *model.DailyReport) error
}

type Service struct {
	repo ArticleLister
}

func NewService(repo ArticleLister) *Service {
	return &Service{repo: repo}
}

func (s *Service) Generate(ctx context.Context, userID int64, title string) (*model.DailyReport, error) {
	if title == "" {
		title = "Morning Brief"
	}
	articles, err := s.repo.ListArticles(ctx, 20, 0)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")
	for _, article := range articles {
		b.WriteString("- ")
		b.WriteString(article.Title)
		if article.URL != "" {
			b.WriteString(" (")
			b.WriteString(article.URL)
			b.WriteString(")")
		}
		b.WriteString("\n")
	}
	if len(articles) == 0 {
		b.WriteString("No articles collected yet.\n")
	}
	report := &model.DailyReport{UserID: userID, Title: title, Content: b.String(), ReportDate: time.Now()}
	if err := s.repo.StoreDailyReport(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}
