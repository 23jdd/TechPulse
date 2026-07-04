package queue

import "context"

type Broker interface {
	Publish(context.Context, string, Message) error
	Consume(context.Context, string) (<-chan Message, error)
	Close() error
}

type RabbitMQ struct {
	URL string
}

func NewRabbitMQ(url string) *RabbitMQ {
	return &RabbitMQ{URL: url}
}

func (r *RabbitMQ) Close() error {
	return nil
}
