package eventsource

import (
	"encoding/json"
	"fmt"
)

type RawEvent struct {
	Event json.RawMessage `json:"event"`
}

type BaseEvent struct {
	Type string `json:"type"`
}

type Event interface {
	GetType() string
}

var eventRegistry = make(map[string]func() Event)

// RegisterEvent registers a constructor for a type string
func registerEvent(eventType string, constructor func() Event) {
	eventRegistry[eventType] = constructor
}

// ParseEvent parses any event based on the eventRegistry
func parseEvent(b []byte) (Event, error) {
	var raw RawEvent
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	var base BaseEvent
	if err := json.Unmarshal(raw.Event, &base); err != nil {
		return nil, err
	}

	constructor, ok := eventRegistry[base.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported event type: %s", base.Type)
	}

	event := constructor()
	if err := json.Unmarshal(raw.Event, &event); err != nil {
		return nil, err
	}

	return event, nil
}
