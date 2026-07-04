package duplicate

import "context"

type Checker interface {
	ArticleExists(ctx context.Context, urlHash, contentHash string) (bool, error)
}

type Service struct {
	checker Checker
}

func NewService(checker Checker) *Service {
	return &Service{checker: checker}
}

func (s *Service) IsDuplicate(ctx context.Context, url, content string) (bool, string, string, error) {
	urlHash := URLHash(url)
	contentHash := ContentHash(content)
	exists, err := s.checker.ArticleExists(ctx, urlHash, contentHash)
	return exists, urlHash, contentHash, err
}
