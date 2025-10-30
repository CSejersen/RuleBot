package integration

import types2 "home_automation_server/types"

type EventTranslator interface {
	Translate(raw []byte) ([]types2.Event, error)
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
