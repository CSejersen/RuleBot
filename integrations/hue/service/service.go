package service

import (
	"context"
	"errors"
	"fmt"
	"home_automation_server/automation"
	"home_automation_server/integrations"
	"home_automation_server/integrations/hue/client"
	"home_automation_server/types"
	"math"

	"go.uber.org/zap"
)

type Service struct {
	Client *client.ApiClient
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]integrations.ServiceSpec {
	return map[string]integrations.ServiceSpec{
		"step_brightness": {
			Handler: s.StepBrightness,
			RequiredParams: map[string]integrations.ParamMetadata{
				"direction": {
					DataType:    "string",
					Description: "One of: up, down",
				},
				"step": {
					DataType:    "int",
					Description: "Maximum 100, clips at Max-level or Min-level.",
				},
			},
			AllowedTargets: integrations.TargetSpec{
				Type:        []integrations.TargetType{integrations.TargetTypeEntity},
				EntityTypes: []types.EntityType{types.EntityTypeLight},
			},
		},
		"set_brightness": {
			Handler: s.SetBrightness,
			RequiredParams: map[string]integrations.ParamMetadata{
				"brightness": {
					DataType:    "float",
					Description: "Brightness value from 0 to 100.",
				},
			},
			AllowedTargets: integrations.TargetSpec{
				Type:        []integrations.TargetType{integrations.TargetTypeEntity},
				EntityTypes: []types.EntityType{types.EntityTypeLight},
			},
		},
		"toggle": {
			Handler:        s.Toggle,
			RequiredParams: map[string]integrations.ParamMetadata{},
			AllowedTargets: integrations.TargetSpec{
				Type:        []integrations.TargetType{integrations.TargetTypeEntity},
				EntityTypes: []types.EntityType{types.EntityTypeLight},
			},
		},
	}
}

func (s *Service) StepBrightness(ctx context.Context, action *automation.Action, emitter integrations.EventEmitter) error {
	step, err := action.IntParam("step")
	if err != nil {
		s.Logger.Error("expected action param: step", zap.Any("params", action.Params))
		return err
	}

	direction, err := action.StringParam("direction")
	if err != nil {
		s.Logger.Error("expected action param: direction", zap.Any("params", action.Params))
		return err
	}

	for _, target := range action.Targets {
		if target.ExternalID == "" {
			return errors.New("external id required in target")
		}

		typ, ok := s.Client.ResourceRegistry.GetTypeByID(target.ExternalID)
		if !ok {
			s.Logger.Warn("Unable to resolve type by id", zap.String("entity_id", target.EntityID), zap.String("external_id", target.ExternalID))
		}

		switch typ {
		case "light":
			return s.Client.LightStepBrightness(ctx, target.ExternalID, math.Abs(float64(step)), direction)
		case "grouped_light":
			return errors.New("step_brightness is not yet supported for grouped lights")
		default:
			return fmt.Errorf("entity type %s is not supported", typ)
		}
	}
	return nil
}

func (s *Service) SetBrightness(ctx context.Context, action *automation.Action, emitter integrations.EventEmitter) error {
	brightness, err := action.FloatParam("brightness")
	if err != nil {
		s.Logger.Error("expected action param: brightness", zap.Any("params", action.Params))
		return err
	}

	for _, target := range action.Targets {
		if target.ExternalID == "" {
			return errors.New("external_id required for target")
		}

		typ, ok := s.Client.ResourceRegistry.GetTypeByID(target.ExternalID)
		if !ok {
			s.Logger.Warn("Unable to resolve type by id", zap.String("entity_id", target.EntityID), zap.String("external_id", target.ExternalID))
			continue
		}

		switch typ {
		case "light":
			// Clamp brightness to valid range
			if brightness > 100 {
				brightness = 100
			}
			if brightness < 0 {
				brightness = 0
			}
			if err := s.Client.LightSetBrightness(ctx, target.ExternalID, brightness); err != nil {
				return err
			}

		case "grouped_light":
			return errors.New("set_brightness is not yet supported for grouped lights")
		default:
			return fmt.Errorf("entity type %s is not supported", typ)
		}
	}
	return nil
}

func (s *Service) Toggle(ctx context.Context, action *automation.Action, emitter integrations.EventEmitter) error {
	for _, target := range action.Targets {
		if target.ExternalID == "" {
			return errors.New("target entity id required")
		}
		typ, ok := s.Client.ResourceRegistry.GetTypeByID(target.ExternalID)
		if !ok {
			s.Logger.Warn("Unable to resolve type by id", zap.String("entity_id", target.EntityID), zap.String("external_id", target.ExternalID))
		}

		switch typ {
		case "light":
			targetLight, err := s.Client.Light(ctx, target.ExternalID)
			if err != nil {
				return fmt.Errorf("failed to get light state: %w", err)
			}
			// flip current on state
			err = s.Client.LightToggle(ctx, target.ExternalID, !targetLight.On.On)
			if err != nil {
				s.Logger.Error("failed to toggle light", zap.String("entity_id", target.EntityID), zap.String("external_id", target.ExternalID))
			}

		case "grouped_light":
			return errors.New("toggle is not yet supported for grouped lights")

		default:
			return fmt.Errorf("entity type %s is not supported", typ)
		}
	}
	return nil
}
