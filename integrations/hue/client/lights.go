package client

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/integrations/hue/client/types"
)

type LightsGetResponse struct {
	Errors []types.ApiError `json:"errors"`
	Data   []types.LightGet `json:"data"`
}

func (c *Client) LightStepBrightness(ctx context.Context, id string, delta float64, action string) error {
	path := fmt.Sprintf("resource/light/%s", id)
	req := types.LightPut{
		DimmingDelta: &types.DimmingDeltaPut{Action: action, BrightnessDelta: delta},
	}

	resp := types.PutResponse{}
	if err := c.put(ctx, path, req, &resp); err != nil {
		c.Logger.Error("failed to step Brightness", zap.Any("api_errors", resp.Errors), zap.Any("resource_identifiers", resp.Data))
		return err
	}

	return nil
}

func (c *Client) LightToggle(ctx context.Context, id string, on bool) error {
	path := fmt.Sprintf("resource/light/%s", id)
	req := types.LightPut{
		On: &types.OnPut{
			On: on,
		},
	}

	resp := types.PutResponse{}
	if err := c.put(ctx, path, req, &resp); err != nil {
		c.Logger.Error("failed to toggle light", zap.Any("errs", resp.Errors), zap.Any("resource_identifiers", resp.Data))
		return err
	}

	return nil
}
