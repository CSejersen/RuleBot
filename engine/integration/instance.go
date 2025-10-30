package integration

import (
	"go.uber.org/zap"
	"home_automation_server/integrations"
)

type Instance struct {
	ConfigID    uint
	Descriptor  IntegrationDescriptor
	EventSource EventSource
	Translator  EventTranslator
	Aggregator  EventAggregator
	Discovery   DiscoveryClient
	Services    map[string]integrations.ServiceSpec // key = "domain.service"
}

func IntegrationLogger(base *zap.Logger, name string) *zap.Logger {
	return base.Named(name).With(zap.String("integration", name))
}
