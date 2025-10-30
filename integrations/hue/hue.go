package hue

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/integration"
	hueclient "home_automation_server/integrations/hue/client"
	"home_automation_server/integrations/hue/discovery"
	"home_automation_server/integrations/hue/eventsource"
	"home_automation_server/integrations/hue/service"
	"home_automation_server/integrations/hue/translator"
	"home_automation_server/types"
)

const (
	BridgeIpKey = "bridge_ip"
	AppKeyKey   = "app_key"
)

func Descriptor() integration.IntegrationDescriptor {
	configSchema := map[string]integration.ConfigField{
		BridgeIpKey: {
			Label:       "Bridge IP",
			Description: "Can be found in the Phillips Hue App",
			Type:        integration.ConfigFieldTypeText,
			Required:    true,
			Placeholder: "192.168.1.100",
			Default:     nil,
		},
		AppKeyKey: {
			Label:       "App Key",
			Description: "Get a key from the Hue V2 API",
			Type:        integration.ConfigFieldTypeText,
			Required:    true,
			Placeholder: "",
			Default:     nil,
		},
	}

	return integration.IntegrationDescriptor{
		Name:         "hue",
		DisplayName:  "Philips Hue",
		Description:  "Controls Philips Hue lights via the Hue Bridge.",
		Version:      "1.0.0",
		Capabilities: []string{integration.CapabilityDiscovery, integration.CapabilityLighting},
		ConfigSchema: configSchema,
		CreateFunc:   NewIntegration,
	}
}

func NewIntegration(ctx context.Context, cfg map[string]any, stateCache types.StateStore, entityRegistry types.EntityRegistry, baseLogger *zap.Logger) (integration.Instance, error) {
	logger := integration.IntegrationLogger(baseLogger, "hue")
	ip, ok := cfg[BridgeIpKey].(string)
	if !ok {
		return integration.Instance{}, fmt.Errorf("bridge_ip is not a string")
	}

	appKey, ok := cfg[AppKeyKey].(string)
	if !ok {
		return integration.Instance{}, fmt.Errorf("app_key is not a string")
	}

	client, err := hueclient.New(ip, appKey, logger.Named("client"))
	if err != nil {
		return integration.Instance{}, fmt.Errorf("falied to construct hue client: %w", err)
	}

	source := eventsource.New(ip, appKey, logger.Named("event_source"))
	trans, err := translator.New(client, stateCache, entityRegistry, logger.Named("translator"))
	if err != nil {
		return integration.Instance{}, fmt.Errorf("failed to construct hue translator: %w", err)
	}

	discoveryClient := discovery.New(client, logger.Named("discovery"))

	s := service.Service{
		Client: client,
		Logger: logger.Named("service"),
	}

	return integration.Instance{
		EventSource: source,
		Translator:  trans,
		Aggregator:  &integration.PassThroughAggregator{},
		Discovery:   discoveryClient,
		Services:    s.ExportServices(),
	}, nil
}
