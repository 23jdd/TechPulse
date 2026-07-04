package observability

import "go.uber.org/zap"

func NewLogger(env string) (*zap.Logger, error) {
	if env == "dev" || env == "local" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
