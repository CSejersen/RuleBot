package events

import "encoding/json"

type Event interface {
	GetType() string
}

type RawEvent struct {
	Event json.RawMessage `json:"event"`
}

type BaseEvent struct {
	Type string `json:"type"`
}
