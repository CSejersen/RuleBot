package service

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/automation"
	"home_automation_server/integrations"
	"home_automation_server/integrations/hue/client"
	"home_automation_server/types"
	"math"
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

func (s *Service) StepBrightness(ctx context.Context, action *automation.Action) error {
	step, err := action.IntParam("step")
	if err != nil {
		s.Logger.Error("expected action param: step", zap.Any("params", action.Params))
		return err
	}

	direction, err := action.StringParam("direction")

	for _, target := range action.Targets {
		if target.EntityID == "" {
			return errors.New("target entity id required")
		}

		typ, ok := s.Client.ResourceRegistry.GetTypeByID(target.EntityID)
		if !ok {
			s.Logger.Warn("Unable to resolve type by id", zap.String("id", target.EntityID))
		}

		switch typ {
		case "light":
			return s.Client.LightStepBrightness(ctx, target.EntityID, math.Abs(float64(step)), direction)
		case "grouped_light":
			return errors.New("step_brightness is not yet supported for grouped lights")
		default:
			return fmt.Errorf("entity type %s is not supported", typ)
		}
	}
	return nil
}

func (s *Service) Toggle(ctx context.Context, action *automation.Action) error {
	for _, target := range action.Targets {
		if target.EntityID == "" {
			return errors.New("target entity id required")
		}
		typ, ok := s.Client.ResourceRegistry.GetTypeByID(target.EntityID)
		if !ok {
			s.Logger.Warn("Unable to resolve type by id", zap.String("id", target.EntityID))
		}

		switch typ {
		case "light":
			targetLight, err := s.Client.Light(ctx, target.EntityID)
			if err != nil {
				return fmt.Errorf("failed to get light state: %w", err)
			}
			// flip current on state
			err = s.Client.LightToggle(ctx, target.EntityID, !targetLight.On.On)
			if err != nil {
				s.Logger.Error("failed to toggle light", zap.String("id", target.EntityID))
			}

		case "grouped_light":
			err := errors.New("toggle is not yet supported for grouped lights")
			if err != nil {
				s.Logger.Error("failed to toggle grouped_light", zap.String("id", target.EntityID))
			}

		default:
			s.Logger.Error("entity type is not supported", zap.String("type", typ))
		}
	}

	return nil
}
