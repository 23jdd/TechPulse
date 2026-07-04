package discovery

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegistry struct {
	Endpoints []string
	client    *clientv3.Client
	services  map[string][]ServiceInstance
}

func NewEtcdRegistry(endpoints []string) *EtcdRegistry {
	return &EtcdRegistry{Endpoints: endpoints, services: map[string][]ServiceInstance{}}
}

func NewEtcdRegistryWithClient(ctx context.Context, endpoints []string) (*EtcdRegistry, error) {
	client, err := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	registry := &EtcdRegistry{Endpoints: endpoints, client: client, services: map[string][]ServiceInstance{}}
	go func() {
		<-ctx.Done()
		_ = client.Close()
	}()
	return registry, nil
}

func (e *EtcdRegistry) Register(ctx context.Context, instance ServiceInstance) error {
	if e.client != nil {
		raw, err := json.Marshal(instance)
		if err != nil {
			return err
		}
		lease, err := e.client.Grant(ctx, 30)
		if err != nil {
			return err
		}
		key := "/techpulse/services/" + instance.Name + "/" + instance.Address
		if _, err := e.client.Put(ctx, key, string(raw), clientv3.WithLease(lease.ID)); err != nil {
			return err
		}
		_, err = e.client.KeepAlive(ctx, lease.ID)
		return err
	}
	e.services[instance.Name] = append(e.services[instance.Name], instance)
	return nil
}

func (e *EtcdRegistry) Discover(ctx context.Context, name string) ([]ServiceInstance, error) {
	if e.client != nil {
		resp, err := e.client.Get(ctx, "/techpulse/services/"+name+"/", clientv3.WithPrefix())
		if err != nil {
			return nil, err
		}
		out := make([]ServiceInstance, 0, len(resp.Kvs))
		for _, kv := range resp.Kvs {
			var instance ServiceInstance
			if err := json.Unmarshal(kv.Value, &instance); err == nil && strings.TrimSpace(instance.Address) != "" {
				out = append(out, instance)
			}
		}
		return out, nil
	}
	return e.services[name], nil
}
