package fetcher

import (
	"context"
	"fmt"
)

type Service struct {
	registry *Registry
}

func NewService(registry *Registry) *Service {
	return &Service{registry: registry}
}

func (s *Service) Fetch(ctx context.Context, source Source) ([]FetchedItem, error) {
	f := s.registry.FetcherFor(source.Type)
	if f == nil {
		return nil, fmt.Errorf("unsupported source type %q", source.Type)
	}
	return f.Fetch(ctx, source)
}
