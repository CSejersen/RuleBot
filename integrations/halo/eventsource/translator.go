package eventsource

import (
	"errors"
	"go.uber.org/zap"
	"home_automation_server/integrations/halo/eventsource/events"
	"home_automation_server/pubsub"
	"time"
)

type Translator struct {
	Logger *zap.Logger
}

func (t Translator) Translate(raw []byte) ([]pubsub.Event, error) {
	event, err := parseEvent(raw)
	if err != nil {
		return nil, err
	}

	switch event.GetType() {
	case "wheel":
		wheelEvent, ok := event.(*events.WheelEvent)
		if !ok {
			return nil, errors.New("expected a *WheelEvent for event type 'wheel'")
		}
		return []pubsub.Event{
			{
				Source: "halo",
				Type:   "button",
				Entity: wheelEvent.Id,
				Payload: map[string]interface{}{
					"counts": wheelEvent.Counts,
				},
				Time: time.Time{},
			},
		}, nil
	default:

	}
	return []pubsub.Event{}, nil
}
