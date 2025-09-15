package hue

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	hueclient "home_automation_server/integrations/hue/client"
	"home_automation_server/integrations/hue/eventsource"
	"home_automation_server/integrations/hue/service"
	"home_automation_server/integrations/hue/translator"
	"os"
)

func NewIntegration(baseLogger *zap.Logger) (engine.Integration, error) {
	logger := engine.IntegrationLogger(baseLogger, "hue")
	ip := os.Getenv("HUE_IP")
	appKey := os.Getenv("HUE_APP_KEY")

	client, err := hueclient.New(ip, appKey, logger.Named("client"))
	if err != nil {
		return engine.Integration{}, fmt.Errorf("falied to construct hue integration: %w", err)
	}

	source := eventsource.New(ip, appKey, logger.Named("event_source"))
	trans, err := translator.New(client, logger.Named("translator"))
	if err != nil {
		return engine.Integration{}, fmt.Errorf("failed to construct hue integration: %w", err)
	}

	s := service.Service{
		Client: client,
		Logger: logger.Named("service"),
	}

	return engine.Integration{
		Name:        "hue",
		EventSource: source,
		Translator:  trans,
		Aggregator:  &engine.PassThroughAggregator{},
		Services:    s.ExportServices(),
	}, nil
}
