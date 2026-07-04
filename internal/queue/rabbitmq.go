package queue

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker interface {
	Publish(context.Context, string, Message) error
	Consume(context.Context, string) (<-chan Message, error)
	Close() error
}

type RabbitMQ struct {
	URL     string
	mu      sync.Mutex
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQ(url string) *RabbitMQ {
	return &RabbitMQ{URL: url}
}

func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RabbitMQ) ensureChannel() (*amqp.Channel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.channel != nil {
		return r.channel, nil
	}
	conn, err := amqp.Dial(r.URL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	r.conn = conn
	r.channel = ch
	return ch, nil
}

func declareQueue(ch *amqp.Channel, name string) error {
	_, err := ch.QueueDeclare(name, true, false, false, false, nil)
	return err
}
