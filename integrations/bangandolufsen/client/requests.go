package client

import (
	"context"
	"fmt"
	"go.uber.org/zap"
)

// TODO: expand to contain all allowed sources for bangandolufsen devices.

var AllowedSources = []string{"lineIn"}

func (c *Client) SetPlaybackSource(ctx context.Context, deviceIP, source string) error {
	path := fmt.Sprintf("playback/sources/active/%s", source)

	allowed := false
	for _, allowedSource := range AllowedSources {
		if source == allowedSource {
			allowed = true
		}
	}
	if !allowed {
		return fmt.Errorf("%s is not an allowed source", source)
	}

	resp := ErrorResponse{}
	err := c.post(ctx, deviceIP, path, nil, &resp)
	if err != nil {
		c.Logger.Error("set playback source request failed", zap.Error(err), zap.Any("server_response", resp))
		return err
	}

	return nil
}

func (c *Client) ExpandExperience(ctx context.Context, deviceIP, toJID string) error {
	path := fmt.Sprintf("beolink/expand/%s", toJID)

	resp := ErrorResponse{}
	err := c.post(ctx, deviceIP, path, nil, &resp)
	if err != nil {
		c.Logger.Error("expand experience request failed", zap.Error(err), zap.Any("server_response", resp))
		return err
	}

	return nil
}
