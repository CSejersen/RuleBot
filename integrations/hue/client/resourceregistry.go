package client

import (
	"context"
	"encoding/json"
	"fmt"
	"home_automation_server/integrations/hue/client/types"
	"time"
)

type Resource interface {
	GetType() string
	GetID() string
}

type ResourceRegistry struct {
	byTypeAndID   map[string]map[string]Resource // type -> id -> Resource
	byTypeAndName map[string]map[string]string   // type -> name -> id (if name exists)
	idToType      map[string]string
}

type BaseResource struct {
	Type string `json:"type"`
}

func (r *ResourceRegistry) GetTypeByID(id string) (string, bool) {
	typ, ok := r.idToType[id]
	return typ, ok
}

func (r *ResourceRegistry) EntityNamesForType(typ string) []string {
	entityNames := []string{}

	if byName, ok := r.byTypeAndName[typ]; ok {
		for name := range byName {
			entityNames = append(entityNames, name)
		}
	}

	if len(entityNames) == 0 {
		if byID, ok := r.byTypeAndID[typ]; ok {
			for _, resource := range byID {
				switch res := resource.(type) {
				case *types.GroupedLightGet:
					if ownerName, ok := r.ResolveName(res.Owner.RType, res.Owner.RID); ok {
						entityNames = append(entityNames, ownerName)
					} else {
						continue
					}
				default:
					// fallback, use ExternalID
					entityNames = append(entityNames, res.GetID())
				}
			}
		}
	}

	return entityNames
}

func (r *ResourceRegistry) ResolveName(typ, id string) (string, bool) {
	resource, ok := r.byTypeAndID[typ][id]
	if !ok || resource == nil {
		return "", false
	}

	switch res := resource.(type) {
	case *types.LightGet:
		return res.Metadata.Name, true
	case *types.RoomGet:
		return res.Metadata.Name, true
	case *types.SceneGet:
		return res.Metadata.Name, true
	case *types.GroupedLightGet:
		return r.ResolveName(res.Owner.RType, res.Owner.RID)
	default:
		return "", false
	}
}

func (r *ResourceRegistry) ResolveID(typ, name string) (string, error) {
	if typ == "grouped_light" {
		return "", fmt.Errorf("grouped_lights cannot be resolved by name, only stored by type + id")
	}

	id, ok := r.byTypeAndName[typ][name]
	if !ok {
		return "", fmt.Errorf("id for %s.%s not found in registry", typ, name)
	}

	return id, nil
}

func (r *ResourceRegistry) ResolveGroupedLightForResource(resource Resource) (*types.GroupedLightGet, bool) {
	// Base case
	if resource.GetType() == "grouped_light" {
		if g, ok := resource.(*types.GroupedLightGet); ok {
			return g, true
		}
		return nil, false
	}

	// Get the resource's owner (if it has one)
	ownerIdentifier := getOwner(resource)
	if ownerIdentifier == nil {
		return nil, false
	}

	// Look up the owner resource in the registry
	if owner, ok := r.byTypeAndID[ownerIdentifier.RType][ownerIdentifier.RID]; ok {
		return r.ResolveGroupedLightForResource(owner)
	}

	// Not found
	return nil, false
}

// Helper: extracts owner info from resources that have an Owner field
// For example: scene.Group, grouped_light.Owner, etc.
func getOwner(resource Resource) *types.ResourceIdentifier {
	switch v := resource.(type) {
	case *types.SceneGet:
		// scenes have a Group field that represents the room/zone they belong to
		return &v.Group
	case *types.GroupedLightGet:
		// grouped_lights have an Owner field (room/zone)
		return &v.Owner
	case *types.RoomGet:
		RID, _ := v.GetGroupedLightID()
		return &types.ResourceIdentifier{
			RID:   RID,
			RType: "grouped_light",
		}
	default:
		// no known ownership chain
		return nil
	}
}

// BuildResourceRegistry build the Resource Registry
func (c *ApiClient) BuildResourceRegistry() error {
	c.ResourceRegistry = ResourceRegistry{
		byTypeAndID:   make(map[string]map[string]Resource),
		byTypeAndName: make(map[string]map[string]string),
		idToType:      make(map[string]string),
	}

	getResourceResp := struct {
		Errors []types.ApiError  `json:"errors"`
		Data   []json.RawMessage `json:"data"`
	}{}

	ctx := context.Background()
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := c.get(ctx, "resource", &getResourceResp)
	if err != nil {
		return err
	}

	for _, raw := range getResourceResp.Data {
		base := BaseResource{}
		if err := json.Unmarshal(raw, &base); err != nil {
			return err
		}

		switch base.Type {
		case "light":
			light := &types.LightGet{}
			if err := json.Unmarshal(raw, light); err != nil {
				return err
			}
			if err := c.ResourceRegistry.add(light); err != nil {
				return fmt.Errorf("failed to add resource to registry: %w", err)
			}

		case "room":
			room := &types.RoomGet{}
			if err := json.Unmarshal(raw, room); err != nil {
				return err
			}
			if err := c.ResourceRegistry.add(room); err != nil {
				return fmt.Errorf("failed to add resource to registry: %w", err)
			}

		case "scene":
			scene := &types.SceneGet{}
			if err := json.Unmarshal(raw, scene); err != nil {
				return err
			}
			if err := c.ResourceRegistry.add(scene); err != nil {
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

	r.idToType[resource.GetID()] = typ

	if _, exists := r.byTypeAndID[typ]; !exists {
		r.byTypeAndID[typ] = make(map[string]Resource)
	}
	r.byTypeAndID[typ][resource.GetID()] = resource

	var name string
	switch res := resource.(type) {
	case *types.LightGet:
		name = res.Metadata.Name
	case *types.RoomGet:
		name = res.Metadata.Name
	case *types.SceneGet:
		name = res.Metadata.Name
	case *types.GroupedLightGet:
		// grouped lights do not have names, stored only in byTypeAndID
		return nil
	default:
		return fmt.Errorf("unsupported resource type: %s", typ)
	}

	if _, exists := r.byTypeAndName[typ]; !exists {
		r.byTypeAndName[typ] = make(map[string]string)
	}
	r.byTypeAndName[typ][name] = resource.GetID()

	return nil
}
