package scheduler

import "context"

type Lock interface {
	TryLock(context.Context, string) (func(context.Context) error, bool, error)
}

type MemoryLock struct{}

func (MemoryLock) TryLock(context.Context, string) (func(context.Context) error, bool, error) {
	return func(context.Context) error { return nil }, true, nil
}
