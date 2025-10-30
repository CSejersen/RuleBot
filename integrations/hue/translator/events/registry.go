package events

import "home_automation_server/types"

// TODO: move away from magic strings here.
var Registry = map[string]types.ExternalEventDescriptor{
	"light": {
		Constructor: func() types.ExternalEvent { return &LightUpdate{} },
	},
	"grouped_light": {
		Constructor: func() types.ExternalEvent { return &GroupedLightUpdate{} },
	},
	"scene": {
		Constructor: func() types.ExternalEvent { return &SceneUpdate{} },
	},
}
