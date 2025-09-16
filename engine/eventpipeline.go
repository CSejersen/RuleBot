package engine

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/engine/pubsub"
	"time"
)

type EventPipeline struct {
	Source     EventSource
	Translator EventTranslator
	Aggregator EventAggregator
	PubSub     *pubsub.PubSub
	Logger     *zap.Logger
}

func (e *Engine) RunEventPipelines(ctx context.Context) {
	for label, integration := range e.Integrations {
		go func() {
			p := e.constructEventPipeline(label, &integration)
			if err := p.run(ctx); err != nil {
				e.Logger.Error("event pipeline exited", zap.Error(err))
			}
		}()
	}
}

func (p *EventPipeline) run(ctx context.Context) error {
	rawCh := make(chan []byte, 100)

	go func() {
		if err := p.Source.Run(ctx, rawCh); err != nil {
			p.Logger.Error("event source failed", zap.Error(err))
		}
	}()

	p.Logger.Debug("event pipeline started")

	// periodic flush for aggregators
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if event := p.Aggregator.Flush(); event != nil {
				p.PubSub.Publish(*event)
			}
			return ctx.Err()
		case raw := <-rawCh:
			events, err := p.Translator.Translate(raw)
			if err != nil {
				continue
			}
			for _, e := range events {
				// some events might be buffered by the eventaggregator
				// if event is passed through we publish now
				if event := p.Aggregator.Aggregate(e); event != nil {
					p.PubSub.Publish(e)
				}
			}
		case <-ticker.C:
			if out := p.Aggregator.Flush(); out != nil {
				p.PubSub.Publish(*out)
			}
		}
	}
}

func (e *Engine) constructEventPipeline(label string, i *Integration) EventPipeline {
	return EventPipeline{
		Source:     i.EventSource,
		Translator: i.Translator,
		Aggregator: i.Aggregator,
		PubSub:     e.PS,
		Logger:     e.Logger.With(zap.String("integration", label)),
	}
}
