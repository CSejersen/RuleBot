package events

import "home_automation_server/integrations/types"

// TODO: move away from magic strings here.
var Registry = map[string]types.EventData{
	"light": {
		Constructor:  func() types.SourceEvent { return &LightUpdate{} },
		StateChanges: []string{"brightness", "mirek", "power_mode", "color_xy", "effect", "alert", "dynamics_speed", "gradient_mode"},
	},
	"grouped_light": {
		Constructor:  func() types.SourceEvent { return &GroupedLightUpdate{} },
		StateChanges: []string{"brightness", "mirek", "power_mode", "color_xy", "effect", "alert", "dynamics_speed", "gradient_mode"},
	},
	"scene": {
		Constructor:  func() types.SourceEvent { return &SceneUpdate{} },
		StateChanges: []string{"active_status"},
	},
}
