package discovery

import "context"

type ServiceInstance struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Registry interface {
	Register(context.Context, ServiceInstance) error
	Discover(context.Context, string) ([]ServiceInstance, error)
}
