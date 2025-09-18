package service

import (
	"context"
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
		"toggle":          s.Toggle,
	}
}

func (s *Service) StepBrightness(ctx context.Context, action *rules.Action) error {
	id, ok := s.Client.ResourceRegistry.ResolveName(action.Target.Typ, action.Target.ID)
	if !ok {
		return fmt.Errorf("unable to resolve device id %s.%s", action.Target.Typ, action.Target.ID)
	}

	step, err := action.IntParam("step")
	if err != nil {
		s.Logger.Error("expected action param: step", zap.Any("params", action.Params))
		return err
	}

	direction, err := action.StringParam("direction")

	switch action.Target.Typ {
	case "light":
		return s.Client.LightStepBrightness(ctx, id, math.Abs(float64(step)), direction)
	default:
		return fmt.Errorf("unknown target type: %s", action.Target.Typ)
	}
}

func (s *Service) Toggle(ctx context.Context, action *rules.Action) error {
	on, err := action.BooleanParam("on")
	if err != nil {
		return err
	}

	target, ok := s.Client.ResourceRegistry.ResolveName(action.Target.Typ, action.Target.ID)
	if !ok {
		return fmt.Errorf("unable to resolve target for %s:%s", action.Target.Typ, action.Target.ID)
	}

	switch action.Target.Typ {
	case "light":
		return s.Client.LightToggle(ctx, target, !on)
	default:
		return fmt.Errorf("unknown target type: %s", action.Target.Typ)
	}
}
