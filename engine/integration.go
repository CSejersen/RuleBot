package engine

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	"home_automation_server/pubsub"
)

type EventSource interface {
	Run(ctx context.Context, out chan<- []byte) error
}

type EventTranslator interface {
	Translate(raw []byte) ([]pubsub.Event, error)
}

type ActionExecutor interface {
	ExecuteAction(action *rules.Action) error
}

type Integration struct {
	EventSource    EventSource
	Translator     EventTranslator
	ActionExecutor ActionExecutor
}

func (e *Engine) RegisterIntegration(label string, integration Integration) {
	e.Integrations[label] = integration
}

func IntegrationLogger(base *zap.Logger, name string) *zap.Logger {
	return base.Named(name).With(zap.String("integration", name))
}
