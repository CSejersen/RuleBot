package engine

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/engine/pubsub"
	"home_automation_server/engine/rules"
)

type EventSource interface {
	Run(ctx context.Context, out chan<- []byte) error
}

type EventTranslator interface {
	Translate(raw []byte) ([]pubsub.Event, error)
}

type EventAggregator interface {
	Aggregate(pubsub.Event) *pubsub.Event
	Flush() *pubsub.Event
}

type ServiceHandler func(action *rules.Action) error

type Integration struct {
	Name        string
	EventSource EventSource
	Translator  EventTranslator
	Aggregator  EventAggregator
	Services    map[string]ServiceHandler // key = "domain.service"
}

func IntegrationLogger(base *zap.Logger, name string) *zap.Logger {
	return base.Named(name).With(zap.String("integration", name))
}

type NoopSource struct{}
type NoopTranslator struct{}
type PassThroughAggregator struct{}

func (s *NoopSource) Run(ctx context.Context, out chan<- []byte) error {
	<-ctx.Done()
	return ctx.Err()
}

func (s *NoopTranslator) Translate(raw []byte) ([]pubsub.Event, error) {
	return []pubsub.Event{}, nil
}

func (a *PassThroughAggregator) Aggregate(e pubsub.Event) *pubsub.Event {
	return &e
}
func (a *PassThroughAggregator) Flush() *pubsub.Event {
	return nil
}
