package engine

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/engine/types"
	integrationtypes "home_automation_server/integrations/types"
)

type EventSource interface {
	Run(ctx context.Context, out chan<- []byte) error
}

type EventTranslator interface {
	Translate(raw []byte) ([]types.Event, error)
	EventTypes() []string
	EntitiesForType(typ string) []string
	StateChangesForType(typ string) []string
}

type EventAggregator interface {
	Aggregate(types.Event) *types.Event
	Flush() *types.Event
}

type Integration struct {
	Name        string
	EventSource EventSource
	Translator  EventTranslator
	Aggregator  EventAggregator
	Services    map[string]integrationtypes.ServiceData // key = "domain.service"
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

func (s *NoopTranslator) Translate(raw []byte) ([]types.Event, error) {
	return []types.Event{}, nil
}
func (s *NoopTranslator) EventTypes() []string {
	return []string{}
}
func (s *NoopTranslator) EntitiesForType(string) []string {
	return []string{}
}
func (s *NoopTranslator) StateChangesForType(string) []string {
	return []string{}
}

func (a *PassThroughAggregator) Aggregate(e types.Event) *types.Event {
	return &e
}
func (a *PassThroughAggregator) Flush() *types.Event {
	return nil
}
