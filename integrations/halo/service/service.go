package service

import (
	"context"
	"home_automation_server/automation"
	"home_automation_server/integrations"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/types"
	"math"

	"go.uber.org/zap"
)

type Service struct {
	Client *client.Client
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]integrations.ServiceSpec {
	return map[string]integrations.ServiceSpec{
		"update_button_value": {
			Handler: s.UpdateButtonValue,
			RequiredParams: map[string]integrations.ParamMetadata{
				"value": {
					DataType:    "int",
					Description: "New value for button, between 0..100",
				},
			},
			AllowedTargets: integrations.TargetSpec{
				Type:        []integrations.TargetType{integrations.TargetTypeEntity},
				EntityTypes: []types.EntityType{types.EntityTypeButton},
			},
		},
	}
}

func (s *Service) UpdateButtonValue(ctx context.Context, action *automation.Action, emitter integrations.EventEmitter) error {
	floatValue, err := action.FloatParam("value")
	if err != nil {
		return err
	}

	value := int(math.Round(floatValue))

	for _, target := range action.Targets {
		err := s.Client.UpdateButtonValue(ctx, target.ExternalID, value)
		if err != nil {
			s.Logger.Error("Failed to update button value", zap.Error(err))
			return err
		}

		// Emit optimistic state_changed event after successful API call
		if emitter != nil {
			newState := &types.State{
				EntityID: target.EntityID,
				State:    value,
			}

			if err := emitter.EmitOptimisticStateChanged(target.EntityID, newState, nil); err != nil {
				s.Logger.Warn("Failed to emit optimistic state_changed event",
					zap.String("entity_id", target.EntityID),
					zap.Error(err))
			}
		}
	}

	return nil
}
