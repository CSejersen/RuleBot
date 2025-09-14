package client

import (
	"go.uber.org/zap"
	"home_automation_server/integrations/hue/client/types"
)

type GetRoomsResponse struct {
	Errors []types.ApiError `json:"errors"`
	Data   []types.RoomGet  `json:"data"`
}

func (c *Client) Rooms() ([]types.RoomGet, error) {
	resp := GetRoomsResponse{}
	if err := c.get("resource/room", &resp); err != nil {
		c.Logger.Error("failed to get rooms", zap.Any("api_errors", resp.Errors))
		return nil, err
	}

	return resp.Data, nil
}
