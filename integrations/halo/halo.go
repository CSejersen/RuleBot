package halo

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/integration"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/integrations/halo/discovery"
	"home_automation_server/integrations/halo/eventaggregator"
	"home_automation_server/integrations/halo/eventsource"
	"home_automation_server/integrations/halo/service"
	"home_automation_server/integrations/halo/translator"
	"home_automation_server/types"
	"os"
)

const (
	HaloIpKey = "halo_ip"
)

func Descriptor() integration.IntegrationDescriptor {
	schema := map[string]integration.ConfigField{
		HaloIpKey: {
			Label:       "Beoremote Halo IP",
			Description: "Can be found in the Halo settings menu",
			Type:        integration.ConfigFieldTypeText,
			Required:    true,
			Placeholder: "192.168.1.100",
			Default:     nil,
			Options:     nil,
		},
	}

	return integration.IntegrationDescriptor{
		Name:         "beoremote_halo",
		DisplayName:  "Beoremote Halo",
		Description:  "Smart remote with touch display and control wheel for home automation control.",
		Version:      "1.0.0",
		Capabilities: []string{integration.CapabilityControl, integration.CapabilityDiscovery},
		ConfigSchema: schema,
		CreateFunc:   NewIntegration,
	}
}

func NewIntegration(ctx context.Context, cfg map[string]any, stateStore types.StateStore, entityRegistry types.EntityRegistry, baseLogger *zap.Logger) (integration.Instance, error) {
	logger := integration.IntegrationLogger(baseLogger, "halo")
	ip, ok := cfg[HaloIpKey].(string)
	if !ok {
		return integration.Instance{}, fmt.Errorf("halo_ip is not a string")
	}

	// TODO: get this from user configuration
	configFile := os.Getenv("HALO_CONFIG")

	haloClient, err := client.New(configFile, logger)
	if err != nil {
		return integration.Instance{}, fmt.Errorf("falied to construct halo integration: %w", err)
	}
	go haloClient.Run(ctx, ip)

	source := eventsource.New(haloClient, logger.Named("event_source"))
	trans := translator.New(haloClient, stateStore, entityRegistry, logger.Named("translator"))
	aggregator := eventaggregator.New(logger.Named("event_aggregator"))

	s := service.Service{
		Client: haloClient,
		Logger: logger.Named("service"),
	}

	discoveryClient := discovery.New(haloClient, logger.Named("discovery"))

	return integration.Instance{
		EventSource: source,
		Translator:  trans,
		Aggregator:  aggregator,
		Discovery:   discoveryClient,
		Services:    s.ExportServices(),
	}, nil
}
