package discovery

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/types"
	"home_automation_server/utils"
)

type DiscoveryClient struct {
	Client *client.Client
	Logger *zap.Logger
}

func New(client *client.Client, logger *zap.Logger) *DiscoveryClient {
	return &DiscoveryClient{Client: client, Logger: logger}
}

func (d *DiscoveryClient) Discover(ctx context.Context) ([]types.Device, []types.Entity, error) {
	haloID := d.Client.Config.ID
	devices := []types.Device{}
	haloDevice := types.Device{
		ID:        haloID,
		Type:      types.DeviceTypeRemote,
		Name:      "Beoremote Halo",
		Metadata:  nil,
		Enabled:   true,
		Available: true,
	}
	devices = append(devices, haloDevice)

	entities := []types.Entity{}
	// Button entities
	for _, page := range d.Client.Config.Pages {
		for _, button := range page.Buttons {
			entity := types.Entity{
				ExternalID: button.ID,
				DeviceID:   haloID,
				EntityID:   fmt.Sprintf("%s.halo_%s", types.EntityTypeButton, utils.NormalizeString(button.Title)),
				Type:       types.EntityTypeButton,
				Name:       button.Title,
				Enabled:    true,
				Available:  true,
			}
			entities = append(entities, entity)
		}
	}
	return devices, entities, nil
}
