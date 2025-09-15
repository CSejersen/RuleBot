package mozart

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/integrations/mozart/client"
	"home_automation_server/integrations/mozart/services"
	"os"
)

func NewMozartIntegration(baseLogger *zap.Logger) (engine.Integration, error) {
	logger := engine.IntegrationLogger(baseLogger, "mozart")
	coreIP := os.Getenv("BEOCONNECT_CORE_IP")
	if coreIP == "" {
		return engine.Integration{}, fmt.Errorf("BEOCONNECT_CORE_IP environment variable not set")
	}
	a9IP := os.Getenv("A9_IP")
	if a9IP == "" {
		return engine.Integration{}, fmt.Errorf("A9_IP environment variable not set")
	}

	deviceIPs := map[string]string{
		"beconnect_core": coreIP,
		"a9":             a9IP,
	}
	apiClient, err := client.New(deviceIPs, logger.Named("client"))
	if err != nil {
		return engine.Integration{}, fmt.Errorf("falied to construct hue integration: %w", err)
	}

	s := services.Service{
		Client: apiClient,
		Logger: logger.Named("service"),
	}

	return engine.Integration{
		Name:        "mozart",
		EventSource: &engine.NoopSource{},
		Translator:  &engine.NoopTranslator{},
		Aggregator:  &engine.PassThroughAggregator{},
		Services:    s.ExportServices(),
	}, nil
}
