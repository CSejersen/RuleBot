package eventaggregator

import (
	"go.uber.org/zap"
	"home_automation_server/types"
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
	// let non wheel events pass through
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

	// if we have multiple wheel events in the buffer, only send the newest one as that will contain the up-to-date button value.
	newest := a.EventBuffer[len(a.EventBuffer)-1]
	a.EventBuffer = []types.Event{}
	return &newest
}
