package service

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/engine/rules"
	"home_automation_server/integrations/halo/client"
)

type Service struct {
	Client *client.Client
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]engine.ServiceHandler {
	return map[string]engine.ServiceHandler{
		"update_button_value": s.UpdateButtonValue,
	}
}

func (s *Service) UpdateButtonValue(action *rules.Action) error {
	if action.Target == nil {
		return fmt.Errorf("no target found")
	}

	id, err := s.Client.ResolveBtnId(action.Target.ID)
	if err != nil {
		s.Logger.Error("Failed to resolve button id", zap.String("id", id), zap.Error(err))
		return err
	}

	value, err := action.FloatParam("value")
	if err != nil {
		return err
	}

	return s.Client.UpdateButtonValue(id, int(value))
}
