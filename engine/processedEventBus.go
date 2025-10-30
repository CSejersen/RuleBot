package engine

import (
	"go.uber.org/zap"
	"home_automation_server/types"
	"sync"
)

type EventBus struct {
	subscribers map[chan types.Event]struct{}
	mu          sync.Mutex
	Logger      *zap.Logger
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[chan types.Event]struct{}),
	}
}

func (eb *EventBus) Subscribe() chan types.Event {
	ch := make(chan types.Event, 100)
	eb.mu.Lock()
	eb.subscribers[ch] = struct{}{}
	eb.mu.Unlock()
	return ch
}

func (eb *EventBus) Unsubscribe(ch chan types.Event) {
	eb.mu.Lock()
	delete(eb.subscribers, ch)
	close(ch)
	eb.mu.Unlock()
}

func (eb *EventBus) Publish(event types.Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for ch := range eb.subscribers {
		select {
		case ch <- event:
		default:
			eb.Logger.Warn("dropped event", zap.Any("event", event))
		}
	}
}
