package actionexecutor

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
)

func (e *Executor) setScene(action *rules.Action) error {
	scene, err := action.StringParam("scene")
	if err != nil {
		return err
	}

	switch action.Target.Type {
	case "hue.grouped_light":
		group, ok := e.Client.Registry.Resolve("hue.grouped_light", action.Target.ID)
		if !ok {
			return fmt.Errorf("unable to resolve target for hue.grouped_light:%s", action.Target.ID)
		}
		e.Logger.Debug("Setting scene for group", zap.String("group", group), zap.String("scene", scene))
	default:
		return fmt.Errorf("unknown target type: %s", action.Target.Type)
	}
	return nil
}
