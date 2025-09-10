package engine

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	"home_automation_server/pubsub"
)

func (e *Engine) ProcessEvents(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-e.EventChannel:
			// resolve state_change
			if err := e.processEvent(event); err != nil {
				e.Logger.Warn("Failed to process event", zap.Error(err))
			}
		}
	}
}

func (e *Engine) processEvent(event pubsub.Event) error {
	e.Logger.Info("processing event", zap.String("source", event.Source), zap.String("type", event.Type))
	for _, rule := range e.RuleSet.Rules {
		if rule.When.Matches(event) {
			e.Logger.Info("rule match found, executing actions")
			for _, action := range rule.Then {
				action.ResolveParams(event) // resolve templated strings
				err := e.executeAction(&action)
				if err != nil {
					return fmt.Errorf("failed to execute action: %w", err)
				}
			}
		}
	}
	return nil
}

func (e *Engine) executeAction(a *rules.Action) error {
	integration, ok := e.Integrations[a.ResolveExecutorName()]
	if !ok {
		return fmt.Errorf("'%s' integration not found", a.ResolveExecutorName())
	}
	return integration.ActionExecutor.ExecuteAction(a)
}
