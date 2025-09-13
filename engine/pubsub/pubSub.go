package pubsub

import "go.uber.org/zap"

type PubSub struct {
	subscribers []chan Event
	Logger      *zap.Logger
}

func NewPubSub() *PubSub {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
		return nil
	}

	return &PubSub{
		Logger: logger,
	}
}

// Subscribe returns a channel that receives events
func (ps *PubSub) Subscribe() <-chan Event {
	ch := make(chan Event, 50)
	ps.subscribers = append(ps.subscribers, ch)
	return ch
}

// Publish sends the event to all subscribers
func (ps *PubSub) Publish(e Event) {
	ps.Logger.Info("publishing event", zap.Int("subscribers", len(ps.subscribers)), zap.Any("event", e))

	for _, sub := range ps.subscribers {
		select {
		case sub <- e:
		default:
			// buffer full: drop oldest and insert newest
			select {
			case <-sub:
			default:
			}
			sub <- e
			ps.Logger.Warn("subscriber too slow, dropped oldest event", zap.Any("event", e))
		}
	}
}
