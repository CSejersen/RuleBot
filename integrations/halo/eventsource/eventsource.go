package eventsource

import (
	"context"
	"home_automation_server/integrations/halo/client"
)

type EventSource struct {
	Client *client.Client
}

func New(c *client.Client) *EventSource {
	return &EventSource{
		Client: c,
	}
}

func (s EventSource) Run(ctx context.Context, out chan<- []byte) error {
	_, msg, err := s.Client.Conn.ReadMessage()
	if err != nil {
		return err
	}

	out <- msg
	return nil
}
