package integration

import "home_automation_server/types"

type EventAggregator interface {
	Aggregate(types.Event) *types.Event
	Flush() *types.Event
}

type PassThroughAggregator struct{}

func (a *PassThroughAggregator) Aggregate(e types.Event) *types.Event {
	return &e
}
func (a *PassThroughAggregator) Flush() *types.Event {
	return nil
}
