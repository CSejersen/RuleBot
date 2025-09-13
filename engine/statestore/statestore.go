package statestore

import (
	"fmt"
	"home_automation_server/engine/pubsub"
	"sync"
	"time"
)

type StateStore struct {
	mu     sync.RWMutex
	states map[string]*State // key = integration:type:entity
}

type State struct {
	Integration string
	Type        string
	Entity      string
	Values      map[string]interface{}
	LastSeen    time.Time
}

func NewStateStore() *StateStore {
	return &StateStore{
		states: make(map[string]*State),
	}
}

// key generates a unique key for the entity
func makeKey(source, typ, entity string) string {
	return fmt.Sprintf("%s:%s:%s", source, typ, entity)
}

// ApplyEvent updates state based on an incoming event
func (s *StateStore) ApplyEvent(e pubsub.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := makeKey(e.Source, e.Type, e.Entity)

	state, ok := s.states[key]
	if !ok {
		state = &State{
			Integration: e.Source,
			Type:        e.Type,
			Entity:      e.Entity,
			Values:      make(map[string]interface{}),
			LastSeen:    time.Now(),
		}
		s.states[key] = state
	}

	for k, v := range e.Payload {
		state.Values[k] = v
	}
	state.Values["last_state_change"] = e.StateChange
	state.LastSeen = e.Time
}

// GetState returns a state snapshot for an entity
func (s *StateStore) GetState(source, typ, entity string) (*State, bool) {
	key := makeKey(source, typ, entity)
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.states[key]
	return state, ok
}
