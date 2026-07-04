package queue

import "context"

func (r *RabbitMQ) Consume(ctx context.Context, _ string) (<-chan Message, error) {
	ch := make(chan Message)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}
