package service

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/automation"
	"home_automation_server/integrations"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/types"
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
					DataType:    "float",
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

func (s *Service) UpdateButtonValue(ctx context.Context, action *automation.Action) error {
	value, err := action.FloatParam("value")
	if err != nil {
		return err
	}

	s.Logger.Debug("n targets", zap.Int("n", len(action.Targets)))

	for _, target := range action.Targets {
		s.Logger.Info("Updating button value", zap.Float64("value", value), zap.String("id", target.EntityID))
		err := s.Client.UpdateButtonValue(ctx, target.EntityID, int(value))
		if err != nil {
			s.Logger.Error("Failed to update button value")
		}
	}

	return nil
}
