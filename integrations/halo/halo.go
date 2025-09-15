package halo

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/integrations/halo/eventaggregator"
	"home_automation_server/integrations/halo/eventsource"
	"home_automation_server/integrations/halo/service"
	"home_automation_server/integrations/halo/translator"
	"os"
)

func NewIntegration(ctx context.Context, baseLogger *zap.Logger) (engine.Integration, error) {
	logger := engine.IntegrationLogger(baseLogger, "halo")
	addr := os.Getenv("HALO_ADDR")
	configFile := os.Getenv("HALO_CONFIG")

	haloClient, err := client.New(configFile, logger)
	if err != nil {
		return engine.Integration{}, fmt.Errorf("falied to construct halo integration: %w", err)
	}
	go haloClient.Run(ctx, addr)

	source := eventsource.New(haloClient, logger.Named("event_source"))
	trans := translator.New(haloClient, logger.Named("translator"))
	aggregator := eventaggregator.New(logger.Named("event_aggregator"))

	s := service.Service{
		Client: haloClient,
		Logger: logger.Named("service"),
	}

	return engine.Integration{
		Name:        "halo",
		EventSource: source,
		Translator:  trans,
		Aggregator:  aggregator,
		Services:    s.ExportServices(),
	}, nil
}
