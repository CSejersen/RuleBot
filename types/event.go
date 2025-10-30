package types

import (
	"time"
)

type EventType string

const (
	EventTypeStateChanged EventType = "state_changed"
	EventTypeCallService  EventType = "call_service"
	EventTimeChanged      EventType = "time_changed"
)

// Event is the base event
type Event struct {
	Type      EventType `json:"type"`    // e.g. "state_changed"
	Data      any       `json:"data"`    // event payload, decoded by Type
	Context   *Context  `json:"context"` // metadata about event origin
	TimeFired time.Time `json:"time_fired"`
}

// Context carries event metadata
type Context struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id,omitempty"`
}

// StateChangedData is the data for a state_changed event
type StateChangedData struct {
	EntityID string `json:"entity_id"`
	OldState *State `json:"old_state"`
	NewState *State `json:"new_state"`
}

// State represents an entity's current state and attributes.
type State struct {
	EntityID    string         `json:"entity_id"`
	State       any            `json:"state"`
	Attributes  map[string]any `json:"attributes"`
	LastChanged time.Time      `json:"last_changed"` // last time the main state changed (updated by engine when applying event to stateStore)
	LastUpdated time.Time      `json:"last_updated"` // last time main state or an attribute changed
	Context     *Context       `json:"context"`
}

// CallServiceData is the data for a call_service event.
type CallServiceData struct {
	Domain      string         `json:"domain"`       // e.g. "light"
	Service     string         `json:"service"`      // e.g. "turn_on"
	ServiceData map[string]any `json:"service_data"` // arbitrary service parameters
	EntityID    string         `json:"entity_id"`    // optional target
}

// TimeChangedData is the data for a time_changed event
type TimeChangedData struct {
	Now time.Time `json:"now"`
}
