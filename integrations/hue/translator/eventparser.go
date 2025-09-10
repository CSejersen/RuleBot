package translator

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"time"
)

// RegisterEvent registers a constructor for an eventType string
func (t *Translator) RegisterEvent(eventType string, constructor func() Event) {
	t.EventParser.EventRegistry[eventType] = constructor
}

// EventParser parses raw bytes into the correct event type
type EventParser struct {
	Logger        *zap.Logger
	EventRegistry map[string]func() Event
}

func NewEventParser(logger *zap.Logger) EventParser {
	return EventParser{
		Logger:        logger,
		EventRegistry: make(map[string]func() Event),
	}
}

type EventEnvelope struct {
	ID           string            `json:"id"`
	CreationTime time.Time         `json:"creationtime"`
	Type         string            `json:"type"`
	RawData      []json.RawMessage `json:"data"`
}

type Event interface {
	GetType() string
}

type TypeWrapper struct {
	Type string `json:"type"`
}

type EventBatch struct {
	Events    []Event
	TimeStamp time.Time
}

func (p *EventParser) parse(b []byte) (EventBatch, error) {
	var envelopes []EventEnvelope

	// Try to unmarshal as array of envelopes
	if err := json.Unmarshal(b, &envelopes); err != nil {
		// If that fails, try single envelope
		var single EventEnvelope
		if err := json.Unmarshal(b, &single); err != nil {
			return EventBatch{}, fmt.Errorf("failed to unmarshal envelope(s): %w", err)
		}
		envelopes = append(envelopes, single)
	}

	allEvents := []Event{}
	timeStamp := time.Time{}

	for _, envelope := range envelopes {
		if envelope.Type != "update" {
			p.Logger.Info("Ignoring event envelope", zap.String("type", envelope.Type))
			continue
		}

		for _, rawEvent := range envelope.RawData {
			typeWrapper := TypeWrapper{}
			if err := json.Unmarshal(rawEvent, &typeWrapper); err != nil {
				p.Logger.Warn("Failed to read event type", zap.Error(err))
				continue
			}

			constructor, ok := p.EventRegistry[typeWrapper.Type]
			if !ok {
				p.Logger.Info("Unsupported event type, skipping", zap.String("type", typeWrapper.Type))
				continue
			}

			event := constructor()
			if err := json.Unmarshal(rawEvent, event); err != nil {
				p.Logger.Warn("Failed to unmarshal event", zap.String("type", typeWrapper.Type), zap.Error(err))
				continue
			}

			allEvents = append(allEvents, event)
		}
		timeStamp = envelope.CreationTime
	}

	if len(allEvents) == 0 {
		return EventBatch{}, nil
	}

	return EventBatch{
		Events:    allEvents,
		TimeStamp: timeStamp,
	}, nil
}
