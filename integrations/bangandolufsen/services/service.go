package services

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"home_automation_server/automation"
	"home_automation_server/integrations"
	"home_automation_server/integrations/bangandolufsen/client"
	"home_automation_server/types"
)

type Service struct {
	Client *client.Client
	Logger *zap.Logger
}

func (s *Service) ExportServices() map[string]integrations.ServiceSpec {
	return map[string]integrations.ServiceSpec{
		"set_playback_source": {
			Handler: s.SetPlaybackSource,
			RequiredParams: map[string]integrations.ParamMetadata{
				"source": {
					DataType:    "string",
					Description: "the name of the source to activate",
				},
			},
			AllowedTargets: integrations.TargetSpec{
				Type:        []integrations.TargetType{integrations.TargetTypeEntity},
				EntityTypes: []types.EntityType{types.EntityTypeSpeaker},
			},
		},
		"expand_experience": {
			Handler: s.ExpandExperience,
			RequiredParams: map[string]integrations.ParamMetadata{
				"to": {
					DataType:    "string",
					Description: "the friendly name of the device to expand the experience to",
				},
			},
			AllowedTargets: integrations.TargetSpec{
				Type:        []integrations.TargetType{integrations.TargetTypeEntity},
				EntityTypes: []types.EntityType{types.EntityTypeSpeaker},
			},
		},
	}
}

func (s *Service) SetPlaybackSource(ctx context.Context, action *automation.Action) error {
	source, err := action.StringParam("source")
	if err != nil {
		s.Logger.Error("Error getting 'source' param", zap.Error(err))
		return err
	}

	for _, target := range action.Targets {
		if target.EntityID == "" {
			return errors.New("target entity id should not be empty")
		}

		ip, ok := s.Client.IpForDevice(target.EntityID)
		if !ok {
			s.Logger.Error("failed to find device_ip", zap.String("jid", target.EntityID))
		}

		if err := s.Client.SetPlaybackSource(ctx, ip, source); err != nil {
			s.Logger.Error("failed to set playback source", zap.Error(err))
		}
	}

	return nil
}

func (s *Service) ExpandExperience(ctx context.Context, action *automation.Action) error {
	expandTo, err := action.StringParam("to")
	if err != nil {
		s.Logger.Error("Error getting 'to' param", zap.Error(err))
	}

	for _, target := range action.Targets {
		if target.EntityID == "" {
			return errors.New("target entity id should not be empty")
		}

		ip, ok := s.Client.IpForDevice(target.EntityID)
		if !ok {
			s.Logger.Error("failed to find device_ip", zap.String("entityID", target.EntityID))
		}
		if err := s.Client.ExpandExperience(ctx, ip, expandTo); err != nil {
			s.Logger.Error("failed to set playback source", zap.Error(err))
		}
	}

	return nil
}
