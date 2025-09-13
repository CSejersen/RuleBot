package eventsource

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/integrations/halo/client"
)

type EventSource struct {
	ReceiveCh chan []byte
	Logger    *zap.Logger
}

func New(c *client.Client, logger *zap.Logger) *EventSource {
	return &EventSource{
		ReceiveCh: c.ReceiveCh,
		Logger:    logger,
	}
}

func (s EventSource) Run(ctx context.Context, out chan<- []byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-s.ReceiveCh:
			out <- msg
		}
	}
}
