package engine

import (
	"context"
	"errors"
	integration "home_automation_server/engine/integration"
	"home_automation_server/types"
	"time"

	"go.uber.org/zap"
)

type EventPipeline struct {
	Label        string
	Source       integration.EventSource
	Translator   integration.EventTranslator
	Aggregator   integration.EventAggregator
	EventChannel chan types.Event
	StateCache   types.StateStore
	Logger       *zap.Logger

	rawCh chan []byte
	done  chan struct{}
}

func (e *Engine) RunEventPipelines(ctx context.Context) {
	for label, intg := range e.Integrations {
		go func(label string, i *integration.Instance) {
			p := e.constructEventPipeline(label, e.StateCache, i)
			if err := p.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				e.Logger.Error("event pipeline exited with error", zap.Error(err))
			}
		}(label, &intg)
	}
}

// Construct pipeline with per-integration buffered channel
func (e *Engine) constructEventPipeline(label string, stateCache types.StateStore, i *integration.Instance) EventPipeline {
	return EventPipeline{
		Label:        label,
		Source:       i.EventSource,
		Translator:   i.Translator,
		Aggregator:   i.Aggregator,
		EventChannel: e.EventChannel,
		StateCache:   stateCache,
		Logger:       e.Logger.With(zap.String("integration", label)),
		rawCh:        make(chan []byte, 100), // per-integration buffer
		done:         make(chan struct{}),
	}
}

func (p *EventPipeline) Run(ctx context.Context) error {
	// Start source reader
	go func() {
		if err := p.Source.Run(ctx, p.rawCh); err != nil && err != context.Canceled {
			p.Logger.Error("event source failed", zap.Error(err))
		}
		close(p.rawCh) // signal no more events
	}()

	// Aggregator ticker
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.flushAggregator()
			return ctx.Err()

		case raw, ok := <-p.rawCh:
			if !ok {
				// Channel closed, flush aggregator one last time
				p.flushAggregator()
				return nil
			}

			events, err := p.Translator.Translate(raw)
			if err != nil {
				p.Logger.Warn("translator failed, dropping raw event", zap.Error(err))
				continue
			}

			for _, e := range events {
				// if events are passed through the aggregator publish them now
				if aggregated := p.Aggregator.Aggregate(e); aggregated != nil {
					p.sendEvent(*aggregated)
				}
			}

		case <-ticker.C:
			p.flushAggregator()
		}
	}
}

// sendEvent is non-blocking; logs if the EventChannel is full
func (p *EventPipeline) sendEvent(event types.Event) {
	select {
	case p.EventChannel <- event:
	default:
		p.Logger.Warn("event channel full, dropping event", zap.String("type", string(event.Type)), zap.Any("data", event.Data))
	}
}

// flushAggregator flushes aggregator events, if any
func (p *EventPipeline) flushAggregator() {
	if event := p.Aggregator.Flush(); event != nil {
		p.sendEvent(*event)
	}
}
