package client

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/integrations/hue/client/types"
)

type Resource interface {
	GetType() string
	GetID() string
}

type ResourceRegistry struct {
	ByTypeAndID   map[string]map[string]Resource // type -> id -> Resource
	ByTypeAndName map[string]map[string]string   // type -> name -> id (if name exists)
}

type BaseResource struct {
	Type string `json:"type"`
}

func (r *ResourceRegistry) ResolveName(typ, id string) (string, bool) {
	resource, ok := r.ByTypeAndID[typ][id]
	if !ok || resource == nil {
		return "", false
	}

	switch res := resource.(type) {
	case *types.LightGet:
		return res.Metadata.Name, true
	case *types.RoomGet:
		return res.Metadata.Name, true
	case *types.GroupedLightGet:
		return r.ResolveName(res.Owner.RType, res.Owner.RID)
	default:
		return "", false
	}
}

// BuildResourceRegistry build the Resource Registry
func (c *Client) BuildResourceRegistry() error {
	c.ResourceRegistry = ResourceRegistry{
		ByTypeAndID:   make(map[string]map[string]Resource),
		ByTypeAndName: make(map[string]map[string]string),
	}

	getResourceResp := struct {
		Errors []types.ApiError  `json:"errors"`
		Data   []json.RawMessage `json:"data"`
	}{}

	err := c.get("resource", &getResourceResp)
	if err != nil {
		return err
	}

	for _, raw := range getResourceResp.Data {
		base := BaseResource{}
		if err := json.Unmarshal(raw, &base); err != nil {
			return err
		}

		c.Logger.Debug("adding resource", zap.String("type", base.Type))
		switch base.Type {
		case "light":
			light := &types.LightGet{}
			if err := json.Unmarshal(raw, light); err != nil {
				return err
			}
			err := c.ResourceRegistry.add(light)
			if err != nil {
				return fmt.Errorf("failed to add resource to registry: %w", err)
			}

		case "room":
			room := &types.RoomGet{}
			if err := json.Unmarshal(raw, room); err != nil {
				return err
			}
			err := c.ResourceRegistry.add(room)
			if err != nil {
				return fmt.Errorf("failed to add resource to registry: %w", err)
			}

		case "grouped_light":
			groupedLight := &types.GroupedLightGet{}
			if err := json.Unmarshal(raw, groupedLight); err != nil {
				return err
			}
			err := c.ResourceRegistry.add(groupedLight)
			if err != nil {
				return fmt.Errorf("failed to add resource to registry: %w", err)
			}
		}

	}

	return nil
}

func (r *ResourceRegistry) add(resource Resource) error {
	typ := resource.GetType()

	if _, exists := r.ByTypeAndID[typ]; !exists {
		r.ByTypeAndID[typ] = make(map[string]Resource)
	}
	r.ByTypeAndID[typ][resource.GetID()] = resource

	var name string
	switch res := resource.(type) {
	case *types.LightGet:
		name = res.Metadata.Name
	case *types.RoomGet:
		name = res.Metadata.Name
	case *types.GroupedLightGet:
		// grouped lights do not have names, stored only in ByTypeAndID
		return nil
	default:
		return fmt.Errorf("unsupported resource type: %s", typ)
	}

	if _, exists := r.ByTypeAndName[typ]; !exists {
		r.ByTypeAndName[typ] = make(map[string]string)
	}
	r.ByTypeAndName[typ][name] = resource.GetID()

	return nil
}
