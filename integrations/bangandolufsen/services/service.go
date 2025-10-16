package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	"home_automation_server/integrations/bangandolufsen/client"
	"home_automation_server/integrations/types"
)

type Service struct {
	Client *client.Client
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]types.ServiceData {
	return map[string]types.ServiceData{
		"set_playback_source": {
			FullName: "bang_and_olufsen.set_playback_source",
			Handler:  s.SetPlaybackSource,
			RequiredParams: map[string]types.ParamMetadata{
				"source": {
					DataType:    "string",
					Description: "the name of the source to activate",
				},
			},
			RequiresTargetType: false,
			RequiresTargetID:   true,
		},
		"expand_experience": {
			FullName: "bang_and_olufsen.expand_experience",
			Handler:  s.ExpandExperience,
			RequiredParams: map[string]types.ParamMetadata{
				"to": {
					DataType:    "string",
					Description: "the friendly name of the device to expand the experience to",
				},
			},
			RequiresTargetType: false,
			RequiresTargetID:   true,
		},
	}
}

func (s *Service) SetPlaybackSource(ctx context.Context, action *rules.Action) error {
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
		if err := s.Client.SetPlaybackSource(ctx, device.IP, source); err != nil {
			return fmt.Errorf("failed to set playback source: %w", err)
		}
		return nil

	case false:
		return fmt.Errorf("playback source is not supported for non-mozart devices")
	}

	return nil
}

func (s *Service) ExpandExperience(ctx context.Context, action *rules.Action) error {
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

	if err := s.Client.ExpandExperience(ctx, device.IP, toDevice.JID); err != nil {
		return fmt.Errorf("failed to expand experience: %w", err)
	}
	return nil
}
