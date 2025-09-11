package hue

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/integrations/hue/actionexecutor"
	"home_automation_server/integrations/hue/apiclient"
	"home_automation_server/integrations/hue/eventsource"
	"home_automation_server/integrations/hue/translator"
	"home_automation_server/pubsub"
	"os"
)

func NewHueIntegration(baseLogger *zap.Logger) (engine.Integration, error) {
	logger := engine.IntegrationLogger(baseLogger, "hue")
	ip := os.Getenv("HUE_IP")
	appKey := os.Getenv("HUE_APP_KEY")

	client := apiclient.New(ip, appKey, logger.Named("client"))
	executor, err := actionexecutor.New(client, logger.Named("action_executor"))
	if err != nil {
		return engine.Integration{}, fmt.Errorf("failed to construct 'Hue Integration: %w", err)
	}
	source := eventsource.New(ip, appKey, logger.Named("event_source"))
	trans, err := translator.New(client, logger.Named("translator"))
	if err != nil {
		return engine.Integration{}, fmt.Errorf("failed to construct 'Hue Integration: %w", err)
	}

	return engine.Integration{
		EventSource:    source,
		Translator:     trans,
		ActionExecutor: executor,
		Aggregator:     &PassThroughAggregator{},
	}, nil
}

type PassThroughAggregator struct{}

func (a *PassThroughAggregator) Aggregate(e pubsub.Event) *pubsub.Event {
	return &e
}
func (a *PassThroughAggregator) Flush() *pubsub.Event {
	return nil
}
