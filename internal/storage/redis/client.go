package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

type Client struct {
	Addr   string
	client *redisv9.Client
}

func New(addr string) *Client {
	return &Client{
		Addr: addr,
		client: redisv9.NewClient(&redisv9.Options{
			Addr: addr,
		}),
	}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	raw, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redisv9.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(raw), dst)
}

func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, raw, ttl).Err()
}

func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := c.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	_, err := pipe.Exec(ctx)
	return incr.Val(), err
}

func (c *Client) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, ttl).Result()
}
