package service

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/engine/rules"
	"home_automation_server/integrations/hue/client"
	"math"
)

type Service struct {
	Client *client.Client
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]engine.ServiceHandler {
	return map[string]engine.ServiceHandler{
		"step_brightness": s.StepBrightness,
		"set_brightness":  s.SetBrightness,
	}
}

func (s *Service) StepBrightness(action *rules.Action) error {
	id, ok := s.Client.DeviceRegistry.Resolve(action.Target.Type, action.Target.ID)
	if !ok {
		return fmt.Errorf("unable to resolve device id %s", action.Target.ID)
	}

	step, err := action.IntParam("step")
	if err != nil {
		s.Logger.Error("expected action param: step", zap.Any("params", action.Params))
		return err
	}

	direction, err := action.StringParam("direction")

	switch action.Target.Type {
	case "light":
		return s.Client.LightStepBrightness(id, math.Abs(float64(step)), direction)
	default:
		return fmt.Errorf("unknown target type: %s", action.Target.Type)
	}
}

func (s *Service) SetBrightness(action *rules.Action) error {
	brightness, err := action.FloatParam("brightness")
	if err != nil {
		return err
	}

	target, ok := s.Client.DeviceRegistry.Resolve(action.Target.Type, action.Target.ID)
	if !ok {
		return fmt.Errorf("unable to resolve target for %s:%s", action.Target.Type, action.Target.ID)
	}

	switch action.Target.Type {
	case "light":
		return s.Client.LightSetBrightness(target, brightness)

	default:
		return fmt.Errorf("unknown target type: %s", action.Target.Type)
	}
}
