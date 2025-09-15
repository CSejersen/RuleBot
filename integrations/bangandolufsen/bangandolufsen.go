package bangandolufsen

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/integrations/bangandolufsen/client"
	"home_automation_server/integrations/bangandolufsen/services"
	"os"
)

func NewIntegration(baseLogger *zap.Logger) (engine.Integration, error) {
	logger := engine.IntegrationLogger(baseLogger, "bang_and_olufsen")
	configPath := os.Getenv("BO_CONFIG")
	apiClient, err := client.New(configPath, logger.Named("client"))
	if err != nil {
		return engine.Integration{}, fmt.Errorf("falied to construct bang and olufsen integration: %w", err)
	}

	s := services.Service{
		Client: apiClient,
		Logger: logger.Named("service"),
	}

	return engine.Integration{
		Name:        "bang_and_olufsen",
		EventSource: &engine.NoopSource{},
		Translator:  &engine.NoopTranslator{},
		Aggregator:  &engine.PassThroughAggregator{},
		Services:    s.ExportServices(),
	}, nil
}
