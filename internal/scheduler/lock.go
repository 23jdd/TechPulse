package scheduler

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Lock interface {
	TryLock(context.Context, string) (func(context.Context) error, bool, error)
}

type MemoryLock struct{}

func (MemoryLock) TryLock(context.Context, string) (func(context.Context) error, bool, error) {
	return func(context.Context) error { return nil }, true, nil
}

type EtcdLock struct {
	client *clientv3.Client
	ttl    int
}

func NewEtcdLock(endpoints []string) (*EtcdLock, error) {
	client, err := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	return &EtcdLock{client: client, ttl: 30}, nil
}

func (l *EtcdLock) Close() error {
	return l.client.Close()
}

func (l *EtcdLock) TryLock(ctx context.Context, key string) (func(context.Context) error, bool, error) {
	session, err := concurrency.NewSession(l.client, concurrency.WithTTL(l.ttl), concurrency.WithContext(ctx))
	if err != nil {
		return nil, false, err
	}
	mutex := concurrency.NewMutex(session, "/techpulse/locks/"+key)
	lockCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := mutex.TryLock(lockCtx); err != nil {
		_ = session.Close()
		return nil, false, nil
	}
	unlock := func(ctx context.Context) error {
		defer session.Close()
		return mutex.Unlock(ctx)
	}
	return unlock, true, nil
}
