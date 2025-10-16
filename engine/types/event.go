package types

import (
	"fmt"
	"time"
)

type Event struct {
	Id          string
	Source      string // "hue", "halo"
	Type        string // "light", "grouped_light", "wheel", "button_press"
	Entity      string // human-readable ID ("flower_pot", "stue")
	StateChange string
	Payload     map[string]any // extra details (brightness=42, scene="movie")
	Time        time.Time
}

type ProcessedEvent struct {
	Event          Event
	TriggeredRules []string
}

func (e *Event) BooleanPayload(key string) (bool, error) {
	raw, ok := e.Payload[key]
	if !ok {
		return false, fmt.Errorf("missing key %s", key)
	}
	val, ok := raw.(bool)
	if !ok {
		return false, fmt.Errorf("value for key %s must be a bool, go %T", key, raw)
	}
	return val, nil
}

func (e *Event) FloatPayload(key string) (float64, error) {
	raw, ok := e.Payload[key]
	if !ok {
		return 0, fmt.Errorf("missing key: %s", key)
	}
	val, ok := raw.(float64)
	if !ok {
		return 0, fmt.Errorf("value for key %s must be a float, go %T", key, raw)
	}
	return val, nil
}

func (e *Event) IntPayload(key string) (int, error) {
	raw, ok := e.Payload[key]
	if !ok {
		return 0, fmt.Errorf("missing key: %s", key)
	}
	val, ok := raw.(int)
	if !ok {
		return 0, fmt.Errorf("value for key %s must be a int, got %T", key, raw)
	}
	return val, nil
}

func (e *Event) StringPayload(key string) (string, error) {
	raw, ok := e.Payload[key]
	if !ok {
		return "", fmt.Errorf("missing key: %s", key)
	}
	val, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("value for key %s must be a string, got %T", key, raw)
	}
	return val, nil
}
