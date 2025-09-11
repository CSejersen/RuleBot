package actionexecutor

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	"math"
)

func (e *Executor) setBrightness(action *rules.Action) error {
	brightness, err := action.FloatParam("brightness")
	if err != nil {
		return err
	}

	switch action.Target.Type {
	case "hue.light":
		target, ok := e.Registry.Resolve("light", action.Target.ID)
		if !ok {
			return fmt.Errorf("unable to resolve target for %s:%s", action.Target.Type, action.Target.ID)
		}
		return e.Client.LightSetBrightness(target, brightness)

	default:
		return fmt.Errorf("unknown target type: %s", action.Target.Type)
	}
}

func (e *Executor) stepBrightness(action *rules.Action) error {
	step, err := action.IntParam("step")
	if err != nil {
		e.Logger.Error("expected action param: step", zap.Any("params", action.Params))
		return err
	}

	direction, err := action.StringParam("direction")

	switch action.Target.Type {
	case "hue.light":
		target, ok := e.Registry.Resolve("light", action.Target.ID)
		if !ok {
			return fmt.Errorf("unable to resolve target for %s:%s", action.Target.Type, action.Target.ID)
		}
		return e.Client.LightStepBrightness(target, math.Abs(float64(step)), direction)

	default:
		return fmt.Errorf("unknown target type: %s", action.Target.Type)
	}
}
