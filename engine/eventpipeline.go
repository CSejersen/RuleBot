package engine

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/pubsub"
)

type EventPipeline struct {
	Source     EventSource
	Translator EventTranslator
	PubSub     *pubsub.PubSub
	Logger     *zap.Logger
}

func (e *Engine) RunPipelines(ctx context.Context) {
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case raw := <-rawCh:
			events, err := p.Translator.Translate(raw)
			if err != nil {
				p.Logger.Warn("failed to translate event", zap.Error(err))
				continue
			}
			for _, e := range events {
				p.PubSub.Publish(e)
			}
		}
	}
}

type HueApiInfo struct {
	IP     string
	AppKey string
}

func (e *Engine) constructEventPipeline(label string, i *Integration) EventPipeline {
	return EventPipeline{
		Source:     i.EventSource,
		Translator: i.Translator,
		PubSub:     e.PS,
		Logger:     e.Logger.With(zap.String("integration", label)),
	}
}
