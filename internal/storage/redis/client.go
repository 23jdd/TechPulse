package redis

import "context"

type Client struct {
	Addr string
}

func New(addr string) *Client {
	return &Client{Addr: addr}
}

func (c *Client) Ping(context.Context) error {
	return nil
}
