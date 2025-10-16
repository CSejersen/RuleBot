package service

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/integrations/types"
)

type Service struct {
	Client *client.Client
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]types.ServiceData {
	return map[string]types.ServiceData{
		"update_button_value": {
			FullName: "halo.update_button_value",
			Handler:  s.UpdateButtonValue,
			RequiredParams: map[string]types.ParamMetadata{
				"value": {
					DataType:    "float",
					Description: "New value for button, between 0..100",
				},
			},
			RequiresTargetType: false,
			RequiresTargetID:   true,
		},
	}
}

func (s *Service) UpdateButtonValue(ctx context.Context, action *rules.Action) error {
	id, err := s.Client.ResolveBtnId(action.Target.ID)
	if err != nil {
		s.Logger.Error("Failed to resolve button id", zap.String("id", id), zap.Error(err))
		return err
	}

	value, err := action.FloatParam("value")
	if err != nil {
		return err
	}

	return s.Client.UpdateButtonValue(ctx, id, int(value))
}
