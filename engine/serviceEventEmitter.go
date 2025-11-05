package engine

import (
	"home_automation_server/integrations"
	"home_automation_server/types"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ServiceEventEmitter implements the EventEmitter interface for services
// to emit optimistic state_changed events
type ServiceEventEmitter struct {
	eventChannel    chan types.Event
	stateCache      types.StateStore
	entityRegistry  types.EntityRegistry
	logger          *zap.Logger
	parentContextID string // Context ID of the service call event that created this emitter
}

// NewServiceEventEmitter creates a new ServiceEventEmitter
func NewServiceEventEmitter(eventChannel chan types.Event, stateCache types.StateStore, entityRegistry types.EntityRegistry, logger *zap.Logger, parentContextID string) integrations.EventEmitter {
	return &ServiceEventEmitter{
		eventChannel:    eventChannel,
		stateCache:      stateCache,
		entityRegistry:  entityRegistry,
		logger:          logger,
		parentContextID: parentContextID,
	}
}

// EmitOptimisticStateChanged emits an optimistic state_changed event
func (e *ServiceEventEmitter) EmitOptimisticStateChanged(entityID string, newState *types.State, oldState *types.State) error {
	// If oldState is not provided, get it from state cache
	var oldStatePtr *types.State
	if oldState != nil {
		oldStatePtr = oldState
	} else {
		if existingState, ok := e.stateCache.Get(entityID); ok {
			// Create a copy to avoid modifying the cached state
			existingStateCopy := existingState
			oldStatePtr = &existingStateCopy
		} else {
			// No existing state, create an empty state
			emptyState := types.State{
				EntityID:   entityID,
				Attributes: make(map[string]any),
			}
			oldStatePtr = &emptyState
		}
	}

	if newState.Attributes == nil {
		newState.Attributes = make(map[string]any)
	}

	// Create context with service origin and parent ID
	context := &types.Context{
		ID:       uuid.NewString(),
		ParentID: e.parentContextID,
		Origin:   "service",
	}
	newState.Context = context

	// Set timestamps
	now := time.Now()
	newState.LastUpdated = now
	if oldStatePtr.State != newState.State {
		newState.LastChanged = now
	} else {
		newState.LastChanged = oldStatePtr.LastChanged
	}

	// Create event
	event := types.Event{
		Type: types.EventTypeStateChanged,
		Data: types.StateChangedData{
			EntityID: entityID,
			OldState: oldStatePtr,
			NewState: newState,
		},
		Context:   context,
		TimeFired: now,
	}

	// Emit event non-blocking
	select {
	case e.eventChannel <- event:
		e.logger.Debug("Emitted optimistic state_changed event",
			zap.String("entity_id", entityID),
			zap.String("origin", "service"),
		)
	default:
		e.logger.Warn("Event channel full, dropped optimistic state_changed event",
			zap.String("entity_id", entityID),
		)
	}

	return nil
}
