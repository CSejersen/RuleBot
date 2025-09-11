package eventsource

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/integrations/halo/client"
)

type EventSource struct {
	Client *client.Client
	Logger *zap.Logger
}

func New(c *client.Client, logger *zap.Logger) *EventSource {
	return &EventSource{
		Client: c,
		Logger: logger,
	}
}

func (s EventSource) Run(ctx context.Context, out chan<- []byte) error {
	for {
		_, msg, err := s.Client.Conn.ReadMessage()
		if err != nil {
			s.Logger.Error("failed to read message", zap.Error(err))
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- msg:
		}
	}
}
