package queue

import (
	"context"
	"encoding/json"
)

func (r *RabbitMQ) Consume(ctx context.Context, queueName string) (<-chan Message, error) {
	ch, err := r.ensureChannel()
	if err != nil {
		return nil, err
	}
	if err := declareQueue(ch, queueName); err != nil {
		return nil, err
	}
	deliveries, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	out := make(chan Message)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-deliveries:
				if !ok {
					return
				}
				var msg Message
				if err := json.Unmarshal(delivery.Body, &msg); err != nil {
					_ = delivery.Nack(false, false)
					continue
				}
				out <- msg
				_ = delivery.Ack(false)
			}
		}
	}()
	return out, nil
}
