package observability

import "context"

type Tracer struct{}

func NewTracer() *Tracer {
	return &Tracer{}
}

func (t *Tracer) Shutdown(context.Context) error {
	return nil
}
