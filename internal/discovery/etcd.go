package discovery

import "context"

type EtcdRegistry struct {
	Endpoints []string
	services  map[string][]ServiceInstance
}

func NewEtcdRegistry(endpoints []string) *EtcdRegistry {
	return &EtcdRegistry{Endpoints: endpoints, services: map[string][]ServiceInstance{}}
}

func (e *EtcdRegistry) Register(_ context.Context, instance ServiceInstance) error {
	e.services[instance.Name] = append(e.services[instance.Name], instance)
	return nil
}

func (e *EtcdRegistry) Discover(_ context.Context, name string) ([]ServiceInstance, error) {
	return e.services[name], nil
}
