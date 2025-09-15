package services

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/engine/rules"
	"home_automation_server/integrations/bangandolufsen/client"
)

type Service struct {
	Client *client.Client
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]engine.ServiceHandler {
	return map[string]engine.ServiceHandler{
		"set_playback_source": s.SetPlaybackSource,
		"expand_experience":   s.ExpandExperience,
	}
}

func (s *Service) SetPlaybackSource(action *rules.Action) error {
	source, err := action.StringParam("source")
	if err != nil {
		s.Logger.Error("Error getting 'source' param", zap.Error(err))
		return err
	}

	device, ok := s.Client.Config.Devices[action.Target.ID]
	if !ok {
		return fmt.Errorf("device %s not found", action.Target.ID)
	}

	switch device.IsMozart {
	case true:
		if err := s.Client.SetPlaybackSource(device.IP, source); err != nil {
			return fmt.Errorf("failed to set playback source: %w", err)
		}
		return nil

	case false:
		return fmt.Errorf("playback source is not supported for non-mozart devices")
	}

	return nil
}

func (s *Service) ExpandExperience(action *rules.Action) error {
	device, ok := s.Client.Config.Devices[action.Target.ID]
	if !ok {
		return fmt.Errorf("device %s not found", action.Target.ID)
	}

	expandTo, err := action.StringParam("to")
	if err != nil {
		s.Logger.Error("Error getting 'to' param", zap.Error(err))
	}

	toDevice, ok := s.Client.Config.Devices[expandTo]
	if !ok {
		return fmt.Errorf("device %s not found", expandTo)
	}

	if err := s.Client.ExpandExperience(device.IP, toDevice.JID); err != nil {
		return fmt.Errorf("failed to expand experience: %w", err)
	}
	return nil
}
