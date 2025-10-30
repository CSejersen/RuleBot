package bangandolufsen

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/integration"
	"home_automation_server/integrations/bangandolufsen/client"
	"home_automation_server/integrations/bangandolufsen/services"
	"home_automation_server/types"
	"os"
)

func Descriptor() integration.IntegrationDescriptor {
	schema := map[string]integration.ConfigField{}

	return integration.IntegrationDescriptor{
		Name:         "bang_and_olufsen_mozart",
		DisplayName:  "Bang & Olufsen: Mozart",
		Description:  "Controls playback, source selection, and multiroom audio synchronization through the Mozart API.",
		Version:      "1.0.0",
		Capabilities: []string{integration.CapabilityDiscovery, integration.CapabilityAudio},
		ConfigSchema: schema,
		CreateFunc:   NewIntegration,
	}
}

func NewIntegration(ctx context.Context, cfg map[string]any, stateStore types.StateStore, entityRegistry types.EntityRegistry, baseLogger *zap.Logger) (integration.Instance, error) {
	logger := integration.IntegrationLogger(baseLogger, "bang_and_olufsen")
	configPath := os.Getenv("BO_CONFIG")
	apiClient, err := client.New(ctx, configPath, logger.Named("client"))
	if err != nil {
		return integration.Instance{}, fmt.Errorf("falied to construct bang and olufsen integration: %w", err)
	}

	s := services.Service{
		Client: apiClient,
		Logger: logger.Named("service"),
	}

	return integration.Instance{
		EventSource: &integration.NoopSource{},
		Translator:  &integration.NoopTranslator{},
		Aggregator:  &integration.PassThroughAggregator{},
		Services:    s.ExportServices(),
	}, nil
}
