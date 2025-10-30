package events

import (
	"home_automation_server/types"
)

var Registry = map[string]types.ExternalEventDescriptor{
	"wheel": {
		Constructor: func() types.ExternalEvent { return &WheelEvent{} },
	},
	"button": {
		Constructor: func() types.ExternalEvent { return &ButtonEvent{} },
	},
	"system": {
		Constructor: func() types.ExternalEvent { return &SystemEvent{} },
	},
}
