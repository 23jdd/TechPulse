package queue

import "context"

func (r *RabbitMQ) Publish(context.Context, string, Message) error {
	return nil
}
