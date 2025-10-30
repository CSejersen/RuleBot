package automation

import (
	"encoding/json"
	"errors"
	"fmt"
	"home_automation_server/types"
	"home_automation_server/utils"
)

type TriggerType string

const (
	TriggerTypeState TriggerType = "state"
	TriggerTypeEvent TriggerType = "event"
)

type Trigger interface {
	Type() TriggerType
	Evaluate(e types.Event) (bool, error)
}

type BaseTrigger struct {
	Type TriggerType `json:"type"`
	Data any         `json:"data"`
}

// AsTrigger Converts a BaseTrigger into a concrete Trigger implementation
func (b BaseTrigger) AsTrigger() (Trigger, error) {
	dataBytes, err := json.Marshal(b.Data)
	if err != nil {
		return nil, err
	}

	switch b.Type {
	case TriggerTypeState:
		var st StateTrigger
		if err := json.Unmarshal(dataBytes, &st); err != nil {
			return nil, err
		}
		return st, nil
	case TriggerTypeEvent:
		var et EventTrigger
		if err := json.Unmarshal(dataBytes, &et); err != nil {
			return nil, err
		}
		return et, nil
	default:
		return nil, fmt.Errorf("unknown trigger type: %s", b.Type)
	}
}

// ---------- Concrete trigger types ----------

// StateTrigger triggers on state_change events that match the constraints.
type StateTrigger struct {
	EntityID  string  `json:"entity_id"`
	Attribute *string `json:"attribute,omitempty"`
	From      any     `json:"from,omitempty"`
	To        any     `json:"to,omitempty"`
}

func (t StateTrigger) Type() TriggerType { return TriggerTypeState }

func (t StateTrigger) Evaluate(e types.Event) (bool, error) {
	if e.Type != types.EventTypeStateChanged {
		return false, nil
	}

	eventData, ok := e.Data.(types.StateChangedData)
	if !ok {
		return false, errors.New("unable to type assert state changed data")
	}

	if eventData.EntityID != t.EntityID {
		return false, nil
	}

	var oldVal, newVal any
	if t.Attribute != nil {
		if eventData.OldState != nil {
			oldVal = eventData.OldState.Attributes[*t.Attribute]
		}
		if eventData.NewState != nil {
			newVal = eventData.NewState.Attributes[*t.Attribute]
		}
	} else {
		if eventData.OldState != nil {
			oldVal = eventData.OldState.State
		}
		if eventData.NewState != nil {
			newVal = eventData.NewState.State
		}
	}

	if utils.AnyEqual(oldVal, newVal) {
		return false, nil
	}

	if t.From != nil && !utils.AnyEqual(oldVal, t.From) {
		return false, nil
	}
	if t.To != nil && !utils.AnyEqual(newVal, t.To) {
		return false, nil
	}

	return true, nil
}

// EventTrigger triggers on any event of a given type.
type EventTrigger struct {
	EventType types.EventType `json:"event_type"`
}

func (t EventTrigger) Type() TriggerType { return TriggerTypeEvent }

func (t EventTrigger) Evaluate(e types.Event) (bool, error) {
	return e.Type == t.EventType, nil
}
