package statestore

import (
	"fmt"
	"home_automation_server/engine/pubsub"
	"home_automation_server/utils"
	"strings"
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
	Fields      map[string]any
	LastSeen    time.Time
}

func NewStateStore() *StateStore {
	return &StateStore{
		states: make(map[string]*State),
	}
}

// key generates a unique key for the entity
func makeKey(source, typ, entity string) string {
	return fmt.Sprintf("%s:%s:%s", utils.NormalizeString(source), utils.NormalizeString(typ), utils.NormalizeString(entity))
}

// ApplyEvent updates state based on an incoming event
func (s *StateStore) ApplyEvent(e pubsub.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := makeKey(e.Source, e.Type, e.Entity)

	state, ok := s.states[key]
	if !ok {
		state = &State{
			Integration: utils.NormalizeString(e.Source),
			Type:        utils.NormalizeString(e.Type),
			Entity:      utils.NormalizeString(e.Entity),
			Fields:      make(map[string]interface{}),
			LastSeen:    time.Now(),
		}
		s.states[key] = state
	}

	for k, v := range e.Payload {
		state.Fields[utils.NormalizeString(k)] = v
	}
	state.Fields["last_state_change"] = e.StateChange
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

func (s *StateStore) ResolvePath(path string) any {
	split := strings.Split(path, ":")
	if len(split) == 2 {
		return nil
	}
	value := split[1]

	obj := strings.Split(split[0], ".")
	source := obj[0]
	typ := obj[1]
	entity := obj[2]

	if state, ok := s.GetState(source, typ, entity); ok {
		return state.Fields[utils.NormalizeString(value)]
	}
	return nil
}
