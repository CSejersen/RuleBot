package eventaggregator

import (
	"home_automation_server/types"
	"home_automation_server/utils"
	"sync"

	"go.uber.org/zap"
)

type Aggregator struct {
	mu          sync.Mutex
	EventBuffer map[string][]types.Event // keyed by entityID
	Logger      *zap.Logger
}

func New(logger *zap.Logger) *Aggregator {
	return &Aggregator{
		mu:          sync.Mutex{},
		EventBuffer: make(map[string][]types.Event),
		Logger:      logger,
	}
}

func (a *Aggregator) Aggregate(e types.Event) *types.Event {
	if e.Type != types.EventTypeStateChanged {
		return &e
	}

	data, ok := e.Data.(types.StateChangedData)
	if !ok {
		return &e
	}

	if !a.isWheelEvent(e) {
		return &e
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.EventBuffer[data.EntityID] = append(a.EventBuffer[data.EntityID], e)
	return nil
}

func (a *Aggregator) isWheelEvent(e types.Event) bool {
	if e.Type != types.EventTypeStateChanged {
		return false
	}
	data := e.Data.(types.StateChangedData)
	oldVal, _ := utils.ToFloat64(data.OldState.State)
	newVal, _ := utils.ToFloat64(data.NewState.State)
	return oldVal != newVal
}

func extractWheelCountDelta(data types.StateChangedData) float64 {
	newVal, ok1 := utils.ToFloat64(data.NewState.State)
	oldVal, ok2 := utils.ToFloat64(data.OldState.State)
	if !ok1 || !ok2 {
		return 0
	}
	return newVal - oldVal
}

func (a *Aggregator) mergeWheelEvents(events []types.Event) *types.Event {
	if len(events) == 0 {
		return nil
	}

	// Use the first event as the base
	firstEvent := events[0]
	firstData := firstEvent.Data.(types.StateChangedData)

	// Sum all deltas
	var totalDelta float64
	for _, e := range events {
		data := e.Data.(types.StateChangedData)
		delta := extractWheelCountDelta(data)
		totalDelta += delta
	}

	// Calculate final state from the first event's oldState
	oldValue, ok := utils.ToFloat64(firstData.OldState.State)
	if !ok {
		a.Logger.Warn("failed to extract old state value for merging")
		return &firstEvent
	}

	newValue := oldValue + totalDelta
	// Clamp to 0-100 (matches translator logic)
	if newValue < 0 {
		newValue = 0
	}
	if newValue > 100 {
		newValue = 100
	}

	// Use the newest event's timestamps and context
	lastEvent := events[len(events)-1]

	// Create merged new state
	mergedNewState := *firstData.NewState
	mergedNewState.State = newValue
	mergedNewState.Context = lastEvent.Context

	// Create merged event
	mergedEvent := types.Event{
		Type: types.EventTypeStateChanged,
		Data: types.StateChangedData{
			EntityID: firstData.EntityID,
			OldState: firstData.OldState,
			NewState: &mergedNewState,
		},
		Context:   lastEvent.Context,
		TimeFired: lastEvent.TimeFired,
	}

	return &mergedEvent
}

func (a *Aggregator) Flush() *types.Event {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Merge wheel events per entity and return the first one found
	for entityID, events := range a.EventBuffer {
		if len(events) == 0 {
			continue
		}

		if len(events) == 1 {
			// Single event, just return it
			delete(a.EventBuffer, entityID)
			return &events[0]
		}

		// Multiple events - merge them
		merged := a.mergeWheelEvents(events)
		delete(a.EventBuffer, entityID)
		return merged
	}

	return nil
}
