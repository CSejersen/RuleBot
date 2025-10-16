package events

import (
	"home_automation_server/integrations/types"
)

var Registry = map[string]types.EventData{
	"wheel": {
		Constructor:  func() types.SourceEvent { return &WheelEvent{} },
		StateChanges: []string{"clockwise", "counter_clockwise"},
	},
	"button": {
		Constructor:  func() types.SourceEvent { return &ButtonEvent{} },
		StateChanges: []string{"pressed", "released"},
	},
	"system": {
		Constructor:  func() types.SourceEvent { return &SystemEvent{} },
		StateChanges: []string{"sleep", "standby", "active"},
	},
}
