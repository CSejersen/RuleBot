package client

import (
	"go.uber.org/zap"
	"home_automation_server/integrations/hue/client/types"
)

type GetScenesResponse struct {
	Errors []types.ApiError `json:"errors"`
	Data   []types.SceneGet `json:"data"`
}

func (c *Client) Scenes() ([]types.SceneGet, error) {
	resp := GetScenesResponse{}
	if err := c.get("resource/scene", &resp); err != nil {
		c.Logger.Error("failed to get scenes", zap.Any("api_errors", resp.Errors))
		return nil, err
	}

	return resp.Data, nil
}
