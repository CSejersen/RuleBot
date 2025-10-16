package eventaggregator

import (
	"go.uber.org/zap"
	"home_automation_server/engine/types"
	"home_automation_server/integrations/halo/translator"
	"sync"
)

type Aggregator struct {
	mu          sync.Mutex
	EventBuffer []types.Event
	Logger      *zap.Logger
}

func New(logger *zap.Logger) *Aggregator {
	return &Aggregator{
		mu:          sync.Mutex{},
		EventBuffer: []types.Event{},
		Logger:      logger,
	}
}

func (a *Aggregator) Aggregate(e types.Event) *types.Event {
	a.Logger.Debug("received event", zap.Any("event", e))
	if e.Type != "wheel" {
		return &e
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.EventBuffer = append(a.EventBuffer, e)
	return nil // don’t emit yet
}

func (a *Aggregator) Flush() *types.Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.EventBuffer) == 0 {
		return nil
	}

	totalCount := 0
	for _, e := range a.EventBuffer {
		step, err := e.IntPayload("step")
		if err != nil {
			a.Logger.Error("Failed to parse step payload", zap.Error(err))
		}
		totalCount += step
	}

	first := a.EventBuffer[0]
	a.EventBuffer = []types.Event{}
	return &types.Event{
		Source:      first.Source,
		Type:        first.Type,
		Entity:      first.Entity,
		StateChange: translator.ResolveWheelStateChange(totalCount),
		Payload: map[string]interface{}{
			"step": totalCount,
		},
		Time: first.Time,
	}
}
