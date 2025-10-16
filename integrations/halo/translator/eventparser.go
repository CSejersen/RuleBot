package translator

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/integrations/halo/translator/events"
	"home_automation_server/integrations/types"
)

// EventParser parses raw bytes into the correct types type
type EventParser struct {
	Logger        *zap.Logger
	EventRegistry map[string]types.EventData
}

func newEventParser(logger *zap.Logger) EventParser {
	return EventParser{
		Logger:        logger,
		EventRegistry: make(map[string]types.EventData),
	}
}

// ParseEvent parses any types based on the eventRegistry
func (p *EventParser) ParseEvent(b []byte) (types.SourceEvent, error) {
	var raw events.RawEvent
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	var base events.BaseEvent
	if err := json.Unmarshal(raw.Event, &base); err != nil {
		return nil, err
	}

	eventData, ok := p.EventRegistry[base.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported event type: %s", base.Type)
	}

	event := eventData.Constructor()
	if err := json.Unmarshal(raw.Event, event); err != nil {
		return nil, err
	}

	return event, nil
}

// RegisterEvent registers a constructor for a type string
func (p *EventParser) RegisterEvent(eventType string, data types.EventData) {
	p.EventRegistry[eventType] = data
}
