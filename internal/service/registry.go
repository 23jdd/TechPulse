package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"techpulse/internal/config"
	"techpulse/internal/discovery"
)

func RegisterSelf(parent context.Context, cfg config.Config, name string, port int, logger *zap.Logger) {
	registry, err := discovery.NewEtcdRegistryWithClient(parent, cfg.EtcdEndpoints)
	if err != nil {
		logger.Warn("etcd registry unavailable", zap.String("service", name), zap.Error(err))
		return
	}
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	if err := registry.Register(ctx, discovery.ServiceInstance{Name: name, Address: fmt.Sprintf("localhost:%d", port)}); err != nil {
		logger.Warn("service registration failed", zap.String("service", name), zap.Error(err))
		return
	}
	logger.Info("service registered", zap.String("service", name))
}
