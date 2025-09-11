package halo

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/integrations/halo/actionexecutor"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/integrations/halo/eventaggregator"
	"home_automation_server/integrations/halo/eventsource"
	"home_automation_server/integrations/halo/translator"
	"os"
)

func NewHaloIntegration(baseLogger *zap.Logger) (engine.Integration, error) {
	logger := engine.IntegrationLogger(baseLogger, "halo")
	addr := os.Getenv("HALO_ADDR")
	configFile := os.Getenv("HALO_CONFIG")

	haloClient, err := client.New(addr, configFile, logger)
	if err != nil {
		return engine.Integration{}, fmt.Errorf("falied to construct halo integration: %w", err)
	}
	executor := actionexecutor.New(haloClient, logger.Named("action_executor"))
	source := eventsource.New(haloClient, logger.Named("event_source"))
	trans := translator.New(haloClient, logger.Named("translator"))
	aggregator := eventaggregator.New(logger.Named("event_aggregator"))

	return engine.Integration{
		EventSource:    source,
		Translator:     trans,
		ActionExecutor: executor,
		Aggregator:     aggregator,
	}, nil
}
