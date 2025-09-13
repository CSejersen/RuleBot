package pubsub

import (
	"fmt"
	"time"
)

type Event struct {
	Source      string // "hue", "halo"
	Type        string // "light", "group", "wheel", "button_press"
	Entity      string // human-readable ID ("flower_pot", "movie scene")
	StateChange string
	Payload     map[string]any // extra details (brightness=42, scene="movie")
	Time        time.Time
}

func (e *Event) FloatPayload(key string) (float64, error) {
	raw, ok := e.Payload[key]
	if !ok {
		return 0, fmt.Errorf("missing param: %s", key)
	}
	val, ok := raw.(float64)
	if !ok {
		return 0, fmt.Errorf("param %s must be float, got %T", key, raw)
	}
	return val, nil
}

func (e *Event) IntPayload(key string) (int, error) {
	raw, ok := e.Payload[key]
	if !ok {
		return 0, fmt.Errorf("missing param: %s", key)
	}
	val, ok := raw.(int)
	if !ok {
		return 0, fmt.Errorf("param %s must be int, got %T", key, raw)
	}
	return val, nil
}

func (e *Event) StringParam(key string) (string, error) {
	raw, ok := e.Payload[key]
	if !ok {
		return "", fmt.Errorf("missing param: %s", key)
	}
	val, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("param %s must be string, got %T", key, raw)
	}
	return val, nil
}
