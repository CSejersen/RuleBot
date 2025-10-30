package engine

import (
	"context"
	"errors"
	"fmt"
	"home_automation_server/automation"
	"home_automation_server/integrations"
	"sync"
)

type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]integrations.ServiceSpec // key = "domain.service"
}

func newServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		mu:       sync.RWMutex{},
		services: make(map[string]integrations.ServiceSpec),
	}
}

func (r *ServiceRegistry) Register(domain, service string, spec integrations.ServiceSpec) {
	key := getKey(domain, service)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[key] = spec
}

func (r *ServiceRegistry) Call(ctx context.Context, domain, service string, action *automation.Action) error {
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

func (r *ServiceRegistry) GetAll() map[string]integrations.ServiceSpec {
	services := make(map[string]integrations.ServiceSpec)
	for name, serviceSpec := range r.services {
		services[name] = serviceSpec
	}
	return services
}
