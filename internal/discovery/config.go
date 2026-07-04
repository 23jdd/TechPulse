package discovery

import "context"

type ConfigStore struct {
	values map[string]string
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{values: map[string]string{}}
}

func (c *ConfigStore) Get(_ context.Context, key string) (string, bool) {
	value, ok := c.values[key]
	return value, ok
}

func (c *ConfigStore) Set(_ context.Context, key, value string) {
	c.values[key] = value
}
