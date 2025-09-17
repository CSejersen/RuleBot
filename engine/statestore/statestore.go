package statestore

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/pubsub"
	"home_automation_server/utils"
	"strings"
	"sync"
	"time"
)

type StateStore struct {
	mu     sync.RWMutex
	states map[string]*State // key = integration:type:entity
	Logger *zap.Logger
}

type State struct {
	Integration string
	Type        string
	Entity      string
	Fields      map[string]any
	LastSeen    time.Time
}

func NewStateStore(logger *zap.Logger) *StateStore {
	return &StateStore{
		states: make(map[string]*State),
		Logger: logger,
	}
}

// key generates a unique key for the entity
func makeKey(source, typ, entity string) string {
	return fmt.Sprintf("%s:%s:%s", utils.NormalizeString(source), utils.NormalizeString(typ), utils.NormalizeString(entity))
}

// ApplyEvent updates state based on an incoming event
func (s *StateStore) ApplyEvent(e pubsub.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Apply the event to the stateStore
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
		if v != nil {
			state.Fields[utils.NormalizeString(k)] = v
		}
	}
	state.Fields["last_state_change"] = e.StateChange
	state.LastSeen = e.Time

	return nil
}

// GetState returns a state snapshot for an entity
func (s *StateStore) GetState(source, typ, entity string) (*State, bool) {
	key := makeKey(source, typ, entity)
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.states[key]
	return state, ok
}

func (s *StateStore) ResolvePath(path string) (any, bool) {
	split := strings.Split(path, ":")
	if len(split) != 2 {
		s.Logger.Error("Invalid path, expected 'source.type.entity:field'", zap.String("path", path))
		return nil, false
	}
	field := split[1]

	stateObj := strings.Split(split[0], ".")
	if len(stateObj) != 3 {
		s.Logger.Error("Invalid path, expected 'source.type.entity:field'", zap.String("path", path))
		return nil, false
	}
	source := stateObj[0]
	typ := stateObj[1]
	entity := stateObj[2]

	if state, ok := s.GetState(source, typ, entity); ok {
		if val, exists := state.Fields[field]; exists {
			return val, true
		}
		return nil, false
	}

	return nil, false
}
