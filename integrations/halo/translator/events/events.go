package events

import "encoding/json"

type ButtonState = string
type SystemState = string

const (
	ButtonStatePressed  ButtonState = "pressed"
	ButtonStateReleased ButtonState = "released"

	SystemStateActive  SystemState = "active"
	SystemStateStandby SystemState = "standby"
	SystemStateSleep   SystemState = "sleep"
)

type Event interface {
	GetType() string
}

type RawEvent struct {
	Event json.RawMessage `json:"event"`
}

type BaseEvent struct {
	Type string `json:"type"`
}

type ButtonEvent struct {
	ID    string      `json:"id"`
	Type  string      `json:"type"`
	State ButtonState `json:"state"`
}

type WheelEvent struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Counts int    `json:"counts"`
}

type SystemEvent struct {
	Type  string      `json:"type"`
	State SystemState `json:"state"`
}

func (e *WheelEvent) GetType() string  { return e.Type }
func (e *ButtonEvent) GetType() string { return e.Type }
func (e *SystemEvent) GetType() string { return e.Type }
