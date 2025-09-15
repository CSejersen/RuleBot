package client

import (
	"fmt"
	"go.uber.org/zap"
)

// TODO: expand to contain all allowed sources for mozart devices.

var AllowedSources = []string{"lineIn"}

// fetchJID retrieves the JID from a device
func (c *Client) fetchJID(ip string) (string, error) {
	var selfResponse struct {
		FriendlyName string `json:"friendly_name"`
		JID          string `json:"jid"`
	}

	if err := c.get(ip, "beolink/self", &selfResponse); err != nil {
		return "", err
	}

	return selfResponse.JID, nil
}

func (c *Client) SetPlaybackSource(ip string, source string) error {
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
	err := c.post(ip, path, nil, &resp)
	if err != nil {
		c.Logger.Error("set playback source request failed", zap.Error(err), zap.Any("server_response", resp))
		return err
	}

	return nil
}

func (c *Client) ExpandExperience(fromIP string, toJID string) error {
	path := fmt.Sprintf("beolink/expand/%s", toJID)

	resp := ErrorResponse{}
	err := c.post(fromIP, path, nil, &resp)
	if err != nil {
		c.Logger.Error("expand experience request failed", zap.Error(err), zap.Any("server_response", resp))
		return err
	}

	return nil
}
