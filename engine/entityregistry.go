package engine

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"sync"
)

type EntityRegistry struct {
	mu      sync.RWMutex
	mapping map[string]string // externalID -> entityID
}

func NewEntityRegistry() *EntityRegistry {
	return &EntityRegistry{
		mapping: make(map[string]string),
	}
}

func (r *EntityRegistry) Register(externalID, entityID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mapping[externalID] = entityID
}

func (r *EntityRegistry) Resolve(externalID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entityID, ok := r.mapping[externalID]
	return entityID, ok
}

func (r *EntityRegistry) ResolveExternalID(entityID string) (string, bool) {
	for externalID, entID := range r.mapping {
		if entityID == entID {
			return externalID, true
		}
	}
	return "", false
}

func (e *Engine) RefreshEntityRegistry(ctx context.Context) error {
	entities, err := e.EntityStore.GetAllEntities(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh entity registry: %w", err)
	}

	for _, entity := range entities {
		e.EntityRegistry.Register(entity.ExternalID, entity.EntityID)
	}
	e.Logger.Info("refreshed entity registry", zap.Int("entity_count", len(entities)))
	return nil
}
