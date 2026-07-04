package ai

import (
	"context"
	"sync/atomic"
	"time"
)

type ResilientProvider struct {
	base        Provider
	timeout     time.Duration
	retries     int
	tokenBudget atomic.Int64
}

func NewResilientProvider(base Provider, timeout time.Duration, retries int) *ResilientProvider {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	if retries < 0 {
		retries = 0
	}
	return &ResilientProvider{base: base, timeout: timeout, retries: retries}
}

func (p *ResilientProvider) Name() string { return p.base.Name() }

func (p *ResilientProvider) EstimatedTokens() int64 {
	return p.tokenBudget.Load()
}

func (p *ResilientProvider) Summarize(ctx context.Context, content string) (*Summary, error) {
	p.addEstimate(content)
	return retry(ctx, p.timeout, p.retries, func(ctx context.Context) (*Summary, error) {
		return p.base.Summarize(ctx, content)
	})
}

func (p *ResilientProvider) Translate(ctx context.Context, content, targetLanguage string) (string, error) {
	p.addEstimate(content)
	return retry(ctx, p.timeout, p.retries, func(ctx context.Context) (string, error) {
		return p.base.Translate(ctx, content, targetLanguage)
	})
}

func (p *ResilientProvider) GenerateTags(ctx context.Context, content string) ([]string, error) {
	p.addEstimate(content)
	return retry(ctx, p.timeout, p.retries, func(ctx context.Context) ([]string, error) {
		return p.base.GenerateTags(ctx, content)
	})
}

func (p *ResilientProvider) GenerateKeywords(ctx context.Context, content string) ([]string, error) {
	p.addEstimate(content)
	return retry(ctx, p.timeout, p.retries, func(ctx context.Context) ([]string, error) {
		return p.base.GenerateKeywords(ctx, content)
	})
}

func (p *ResilientProvider) GenerateEmbedding(ctx context.Context, content string) ([]float64, error) {
	p.addEstimate(content)
	return retry(ctx, p.timeout, p.retries, func(ctx context.Context) ([]float64, error) {
		return p.base.GenerateEmbedding(ctx, content)
	})
}

func (p *ResilientProvider) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	for _, message := range messages {
		p.addEstimate(message.Content)
	}
	return retry(ctx, p.timeout, p.retries, func(ctx context.Context) (string, error) {
		return p.base.ChatCompletion(ctx, messages)
	})
}

func (p *ResilientProvider) addEstimate(text string) {
	p.tokenBudget.Add(int64(len([]rune(text))/4 + 1))
}

func retry[T any](ctx context.Context, timeout time.Duration, retries int, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, timeout)
		value, err := fn(callCtx)
		cancel()
		if err == nil {
			return value, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 150 * time.Millisecond):
		}
	}
	return zero, lastErr
}
