package engine

import (
	"home_automation_server/types"
	"sync"
	"time"
)

// StateCache is a thread-safe in-memory state registry.
type StateCache struct {
	mu    sync.RWMutex
	cache map[string]types.State
}

// NewStateCache creates a new state cache.
func NewStateCache() *StateCache {
	return &StateCache{
		cache: make(map[string]types.State),
	}
}

// Get retrieves the state for the given entityID.
func (s *StateCache) Get(entityID string) (types.State, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.cache[entityID]
	return state, ok
}

func (s *StateCache) GetAll() []types.State {
	res := make([]types.State, 0)
	for _, state := range s.cache {
		res = append(res, state)
	}
	return res
}

// Set updates or creates the state for the given entityID.
// It updates LastChanged only if the main state changes.
func (s *StateCache) Set(entityID string, newState types.State) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	existing, exists := s.cache[entityID]

	if exists {
		if existing.State != newState.State {
			newState.LastChanged = now
		} else {
			newState.LastChanged = existing.LastChanged
		}
		newState.LastUpdated = now
	} else {
		newState.LastChanged = now
		newState.LastUpdated = now
	}

	s.cache[entityID] = newState
}
