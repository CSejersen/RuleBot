package client

import (
	"context"
	"encoding/json"
	"fmt"
	"home_automation_server/integrations/hue/client/types"
)

func (c *ApiClient) GetResources(ctx context.Context) ([]Resource, error) {
	getResourceResp := struct {
		Errors []types.ApiError  `json:"errors"`
		Data   []json.RawMessage `json:"data"`
	}{}

	err := c.get(ctx, "resource", &getResourceResp)
	if err != nil {
		return nil, err
	}

	res := []Resource{}
	for _, raw := range getResourceResp.Data {
		base := BaseResource{}
		if err := json.Unmarshal(raw, &base); err != nil {
			return nil, err
		}

		switch base.Type {
		case "light":
			light := &types.LightGet{}
			if err := json.Unmarshal(raw, light); err != nil {
				return nil, fmt.Errorf("failed to unmarshal lightGet")
			}
			res = append(res, light)

		case "room":
			room := &types.RoomGet{}
			if err := json.Unmarshal(raw, room); err != nil {
				return nil, fmt.Errorf("failed to unmarshal roomGet")
			}
			res = append(res, room)

		case "scene":
			scene := &types.SceneGet{}
			if err := json.Unmarshal(raw, scene); err != nil {
				return nil, fmt.Errorf("failed to unmarshal sceneGet")
			}
			res = append(res, scene)

		case "grouped_light":
			groupedLight := &types.GroupedLightGet{}
			if err := json.Unmarshal(raw, groupedLight); err != nil {
				return nil, fmt.Errorf("failed to unmarshal groupedLightGet")
			}
			res = append(res, groupedLight)
		}
	}
	return res, nil
}
