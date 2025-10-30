package discovery

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/integrations/hue/client"
	huetypes "home_automation_server/integrations/hue/client/types"
	"home_automation_server/types"
	"home_automation_server/utils"
	"time"
)

const (
	LightDeviceMetadataArchetypeKey  = "archetype"
	LightDeviceMetadataOwnerRIDKey   = "owner_rid"
	LightDeviceMetadataOwnerRTypeKey = "owner_rtype"
)

type DiscoveryClient struct {
	ApiClient *client.ApiClient
	Logger    *zap.Logger
}

func New(apiClient *client.ApiClient, logger *zap.Logger) *DiscoveryClient {
	return &DiscoveryClient{ApiClient: apiClient, Logger: logger}
}

func (d *DiscoveryClient) Discover(ctx context.Context) ([]types.Device, []types.Entity, error) {
	resources, err := d.ApiClient.GetResources(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch resources")
	}

	devices := []types.Device{}
	entities := []types.Entity{}

	for _, resource := range resources {
		switch resource.GetType() {
		case "light":
			lightGet, ok := resource.(*huetypes.LightGet)
			if !ok {
				return nil, nil, fmt.Errorf("failed to cast lightGet resource")
			}

			// Create Device
			deviceMetadata := map[string]any{
				LightDeviceMetadataArchetypeKey:  lightGet.Metadata.Archetype,
				LightDeviceMetadataOwnerRIDKey:   lightGet.Owner.RID,
				LightDeviceMetadataOwnerRTypeKey: lightGet.Owner.RType,
			}

			device := types.Device{
				ID:        lightGet.GetID(),
				Type:      types.DeviceTypeLight,
				Name:      lightGet.Metadata.Name,
				Metadata:  deviceMetadata,
				Enabled:   true,
				Available: true,
			}
			devices = append(devices, device)

			// Create Entity
			entity := types.Entity{
				ExternalID: lightGet.GetID(),
				DeviceID:   lightGet.GetID(),
				EntityID:   fmt.Sprintf("%s.%s", types.EntityTypeLight, utils.NormalizeString(lightGet.Metadata.Name)),
				Type:       types.EntityTypeLight,
				Name:       lightGet.Metadata.Name,
				Enabled:    true,
				Available:  true,
			}
			entities = append(entities, entity)

		case "grouped_light":
			groupedLightGet, ok := resource.(*huetypes.GroupedLightGet)
			if !ok {
				return nil, nil, fmt.Errorf("failed to cast groupedLightGet resource")
			}

			deviceMetadata := map[string]any{}

			name, ok := d.ApiClient.ResourceRegistry.ResolveName(groupedLightGet.GetType(), groupedLightGet.GetID())
			if !ok {
				d.Logger.Info("No owner name for grouped light, ignoring during discovery")
				continue
			}

			device := types.Device{
				ID:        groupedLightGet.ID,
				Type:      types.DeviceTypeGroupedLight,
				Name:      name,
				Metadata:  deviceMetadata,
				Enabled:   false,
				Available: true,
			}
			devices = append(devices, device)

			// Create Entity
			entity := types.Entity{
				ExternalID: groupedLightGet.GetID(),
				DeviceID:   groupedLightGet.GetID(),
				EntityID:   fmt.Sprintf("%s.%s", types.EntityTypeLight, utils.NormalizeString(name)),
				Type:       types.EntityTypeLight,
				Name:       name,
				Enabled:    true,
				Available:  true,
			}
			entities = append(entities, entity)

		case "scene":
			SceneGet, ok := resource.(*huetypes.SceneGet)
			if !ok {
				return nil, nil, fmt.Errorf("failed to cast SceneGet resource")
			}

			// find the grouped_light that owns the scene.
			// SceneGet.Group is the room or zone, we have to resolve the grouped_light resource tied to that.
			owner, ok := d.ApiClient.ResourceRegistry.ResolveGroupedLightForResource(resource)
			if !ok {
				d.Logger.Warn("No grouped_light owner for scene, ignoring during discovery")
				continue
			}

			name, ok := d.ApiClient.ResourceRegistry.ResolveName(owner.GetType(), owner.GetID())
			if !ok {
				d.Logger.Info("No name for grouped light, skipping it for scene entity_id")
			}

			entity := types.Entity{
				ExternalID: SceneGet.GetID(),
				DeviceID:   owner.ID,
				EntityID:   fmt.Sprintf("%s.%s_%s", types.EntityTypeScene, utils.NormalizeString(name), utils.NormalizeString(SceneGet.Metadata.Name)),
				Type:       types.EntityTypeScene,
				Name:       SceneGet.Metadata.Name,
				Enabled:    false,
				Available:  true,
				CreatedAt:  time.Time{},
			}
			entities = append(entities, entity)

		default:
			continue
		}
	}

	return devices, entities, nil
}
