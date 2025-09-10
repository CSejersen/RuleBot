package actionexecutor

import (
	"fmt"
	"home_automation_server/engine/rules"
)

func (e *Executor) adjustBrightness(action *rules.Action) error {
	brightness, err := action.FloatParam("brightness")
	if err != nil {
		return err
	}

	switch action.Target.Type {
	case "hue.light":
		target, ok := e.Client.Registry.Resolve("hue.light", action.Target.ID)
		if !ok {
			return fmt.Errorf("unable to resolve target for %s:%s", action.Target.Type, action.Target.ID)
		}
		return e.Client.LightBrightness(target, brightness)

	default:
		return fmt.Errorf("unknown target type: %s", action.Target.Type)
	}
}
