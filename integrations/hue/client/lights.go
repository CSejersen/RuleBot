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

func (c *ApiClient) Light(ctx context.Context, id string) (types.LightGet, error) {
	path := fmt.Sprintf("/resource/light/%s", id)
	resp := LightsGetResponse{}
	if err := c.get(ctx, path, &resp); err != nil {
		c.Logger.Error("failed to fetch light", zap.String("id", id), zap.Error(err))
		return types.LightGet{}, err
	}
	if len(resp.Data) > 1 {
		return types.LightGet{}, fmt.Errorf("expected 1 light, got %d", len(resp.Data))
	}
	if len(resp.Data) == 0 {
		return types.LightGet{}, fmt.Errorf("no lights found")
	}
	return resp.Data[0], nil
}

func (c *ApiClient) LightStepBrightness(ctx context.Context, id string, delta float64, action string) error {
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

func (c *ApiClient) LightToggle(ctx context.Context, id string, on bool) error {
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

func (c *ApiClient) LightStepColor(ctx context.Context, id string, xy types.XY) error {
	path := fmt.Sprintf("resource/light/%s", id)
	req := types.LightPut{
		Color: &types.ColorPut{XY: xy},
	}

	resp := types.PutResponse{}
	if err := c.put(ctx, path, req, &resp); err != nil {
		c.Logger.Error("failed to step color", zap.Any("errs", resp.Errors))
		return err
	}
	return nil
}
