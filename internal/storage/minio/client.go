package minio

import "context"

type Client struct {
	Endpoint string
	Bucket   string
}

func New(endpoint, bucket string) *Client {
	return &Client{Endpoint: endpoint, Bucket: bucket}
}

func (c *Client) EnsureBucket(context.Context) error {
	return nil
}
