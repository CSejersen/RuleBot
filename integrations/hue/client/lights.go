package client

import (
	"fmt"
	"go.uber.org/zap"
)

func (c *Client) Lights() ([]Light, error) {
	var resp getLightsResponse
	err := c.get("resource/light", &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// LightSetBrightness sets brightness percentage of a light source.
// value cannot be 0, writing 0 changes it to the lowest possible brightness
func (c *Client) LightSetBrightness(name string, val float64) error {
	lights, err := c.Lights()
	if err != nil {
		return fmt.Errorf("failed to get lights: %s", err)
	}

	targetLight := Light{}
	for _, l := range lights {
		if l.Metadata.Name == name {
			targetLight = l
		}
	}

	path := fmt.Sprintf("resource/light/%s", targetLight.ID)
	targetLight.Dimming.Brightness = val

	resp := apiResponse{}
	if err := c.put(path, targetLight, &resp); err != nil {
		c.Logger.Error("failed to set Brightness", zap.Any("errs", resp.Errors))
		return err
	}
	return nil
}

func (c *Client) LightStepBrightness(id string, delta float64, action string) error {
	path := fmt.Sprintf("resource/light/%s", id)
	req := LightStepBrightnessRequest{
		DimmingDelta: DimmingDelta{Action: action, BrightnessDelta: delta},
	}

	resp := apiResponse{}
	if err := c.put(path, req, &resp); err != nil {
		c.Logger.Error("failed to step Brightness", zap.Any("errs", resp.Errors))
		return err
	}

	return nil
}
