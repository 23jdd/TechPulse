package report

import (
	"context"
	"strings"
	"time"

	"techpulse/internal/model"
)

type ArticleLister interface {
	ListArticles(context.Context, int, int) ([]model.Article, error)
	ListArticlesSince(context.Context, time.Time, int) ([]model.Article, error)
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
	articles, err := s.repo.ListArticlesSince(ctx, time.Now().Add(-24*time.Hour), 50)
	if err != nil {
		return nil, err
	}
	if len(articles) == 0 {
		articles, err = s.repo.ListArticles(ctx, 20, 0)
		if err != nil {
			return nil, err
		}
	}
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString("> Generated from recent TechPulse articles. The report is Markdown and can be archived or sent to email/chat tools.\n\n")

	groups := groupArticles(articles)
	order := []string{"Go", "Kubernetes", "AI", "Database", "Security", "Cloud Native", "Linux", "Other"}
	for _, category := range order {
		items := groups[category]
		if len(items) == 0 {
			continue
		}
		b.WriteString("## ")
		b.WriteString(category)
		b.WriteString("\n\n")
		for _, article := range items {
			b.WriteString("- **")
			b.WriteString(article.Title)
			b.WriteString("**")
			if article.URL != "" {
				b.WriteString(" - ")
				b.WriteString(article.URL)
			}
			b.WriteString("\n")
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

func groupArticles(articles []model.Article) map[string][]model.Article {
	groups := map[string][]model.Article{}
	for _, article := range articles {
		category := classify(article.Title + " " + article.CleanContent)
		groups[category] = append(groups[category], article)
	}
	return groups
}

func classify(text string) string {
	text = strings.ToLower(text)
	rules := []struct {
		category string
		terms    []string
	}{
		{"Go", []string{"golang", "go ", " go", "toolchain"}},
		{"Kubernetes", []string{"kubernetes", "k8s", "kubectl"}},
		{"AI", []string{"ai", "llm", "openai", "model", "agent"}},
		{"Database", []string{"mysql", "postgres", "redis", "database", "sqlite"}},
		{"Security", []string{"security", "cve", "vulnerability", "auth"}},
		{"Cloud Native", []string{"cloud", "container", "docker", "grpc", "service mesh"}},
		{"Linux", []string{"linux", "kernel", "systemd"}},
	}
	for _, rule := range rules {
		for _, term := range rule.terms {
			if strings.Contains(text, term) {
				return rule.category
			}
		}
	}
	return "Other"
}
