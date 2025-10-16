package service

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	"home_automation_server/integrations/hue/client"
	"home_automation_server/integrations/types"
	"math"
)

type Service struct {
	Client *client.Client
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]types.ServiceData {
	return map[string]types.ServiceData{
		"step_brightness": {
			FullName: "hue.step_brightness",
			Handler:  s.StepBrightness,
			RequiredParams: map[string]types.ParamMetadata{
				"direction": {
					DataType:    "string",
					Description: "One of: up, down",
				},
				"step": {
					DataType:    "int",
					Description: "Maximum 100, clips at Max-level or Min-level.",
				},
			},
			RequiresTargetType: true,
			RequiresTargetID:   true,
		},
		"toggle": {
			FullName: "hue.toggle",
			Handler:  s.Toggle,
			RequiredParams: map[string]types.ParamMetadata{
				"on": {
					DataType:    "bool",
					Description: "The current on_state (get from stateStore)",
				},
			},
			RequiresTargetType: true,
			RequiresTargetID:   true,
		},
	}
}

func (s *Service) StepBrightness(ctx context.Context, action *rules.Action) error {
	id, ok := s.Client.ResourceRegistry.ByTypeAndName[action.Target.Typ][action.Target.ID]
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
