package apiclient

import (
	"fmt"
	"go.uber.org/zap"
)

type Light struct {
	Type     string        `json:"type"`
	ID       string        `json:"id"`
	Metadata LightMetadata `json:"metadata"`
	On       LightOn       `json:"on"`
	Dimming  LightDimming  `json:"dimming"`
}

type LightDimming struct {
	Brightness float64 `json:"brightness"`
}

type LightOn struct {
	On bool
}

type LightMetadata struct {
	Name string `json:"name"`
}

type getLightsResponse struct {
	Data []Light `json:"data"`
}

type setBrightnessResponse struct {
	Errors []SetBrightnessError `json:"errors"`
}

type SetBrightnessError struct {
	Error string `json:"error"`
}

func (c *ApiClient) Lights() ([]Light, error) {
	var resp getLightsResponse
	err := c.get("resource/light", &resp)
	if err != nil {
		return nil, err
	}
	fmt.Println("received light data")
	return resp.Data, nil
}

// LightBrightness sets brightness percentage of a light source.
// value cannot be 0, writing 0 changes it to the lowest possible brightness
func (c *ApiClient) LightBrightness(name string, val float64) error {
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

	resp := setBrightnessResponse{}
	if err := c.put(path, targetLight, &resp); err != nil {
		c.Logger.Error("failed to set Brightness", zap.Any("errs", resp.Errors))
		return err
	}
	return nil
}
