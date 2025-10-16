package engine

import (
	"context"
	"errors"
	"fmt"
	"home_automation_server/engine/rules"
	"home_automation_server/integrations/types"
	"sync"
)

type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]types.ServiceData // key = "domain.service"
}

func newServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		mu:       sync.RWMutex{},
		services: make(map[string]types.ServiceData),
	}
}

func (r *ServiceRegistry) Register(domain, service string, data types.ServiceData) {
	key := getKey(domain, service)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[key] = data
}

func (r *ServiceRegistry) Call(ctx context.Context, domain, service string, action *rules.Action) error {
	key := getKey(domain, service)
	r.mu.RLock()
	defer r.mu.RUnlock()
	serviceData, ok := r.services[key]
	if !ok {
		return errors.New(fmt.Sprintf("service %s not registered yet", key))
	}
	return serviceData.Handler(ctx, action)
}

func getKey(domain, service string) string {
	return fmt.Sprintf("%s.%s", domain, service)
}

func (r *ServiceRegistry) GetAll() []types.ServiceData {
	services := []types.ServiceData{}
	for _, service := range r.services {
		services = append(services, service)
	}
	return services
}
