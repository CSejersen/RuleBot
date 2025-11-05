package integration

import types2 "home_automation_server/types"

type EventTranslator interface {
	Translate(raw []byte) ([]types2.Event, error)
	SetServiceCallTracker(ServiceCallTracker)
}

// ServiceCallTracker is the interface for tracking service call contexts per entity
// to enable linking real device events back to their originating service calls
type ServiceCallTracker interface {
	FindBestMatch(entityID string, actualState map[string]any) (contextID string, exists bool)
	PopOldestForDomain(domain string) (contextID string, exists bool)
}

type NoopTranslator struct{}

func (s *NoopTranslator) Translate(raw []byte) ([]types2.Event, error) {
	return []types2.Event{}, nil
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
func (s *NoopTranslator) SetServiceCallTracker(ServiceCallTracker) {
	// No-op implementation
}
