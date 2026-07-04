package queue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) Publish(ctx context.Context, queueName string, msg Message) error {
	ch, err := r.ensureChannel()
	if err != nil {
		return err
	}
	if err := declareQueue(ch, queueName); err != nil {
		return err
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         raw,
	})
}
