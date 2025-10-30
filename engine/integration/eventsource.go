package integration

import "context"

type EventSource interface {
	Run(ctx context.Context, out chan<- []byte) error
}

type NoopSource struct{}

func (s *NoopSource) Run(ctx context.Context, out chan<- []byte) error {
	<-ctx.Done()
	return ctx.Err()
}
