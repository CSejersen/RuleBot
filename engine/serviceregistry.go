package engine

import (
	"context"
	"errors"
	"fmt"
	"home_automation_server/engine/rules"
	"sync"
)

type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]ServiceHandler // key = "domain.service"
}

func newServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		mu:       sync.RWMutex{},
		services: make(map[string]ServiceHandler),
	}
}

func (r *ServiceRegistry) Register(domain, service string, handler ServiceHandler) {
	key := getKey(domain, service)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[key] = handler
}

func (r *ServiceRegistry) Call(ctx context.Context, domain, service string, action *rules.Action) error {
	key := getKey(domain, service)
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.services[key]
	if !ok {
		return errors.New(fmt.Sprintf("service %s not registered yet", key))
	}
	return handler(ctx, action)
}

func getKey(domain, service string) string {
	return fmt.Sprintf("%s.%s", domain, service)
}
