package engine

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/engine/pubsub"
)

func (e *Engine) ProcessEvents(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				e.Logger.Info("context cancelled")
				return
			case event := <-e.EventChannel:
				if err := e.processEvent(event); err != nil {
					e.Logger.Warn("failed to process event", zap.Error(err))
				}
			}
		}
	}()
}

func (e *Engine) processEvent(event pubsub.Event) error {
	e.StateStore.ApplyEvent(event)

	for _, rule := range e.RuleSet.Rules {
		if !rule.Trigger.Matches(event) {
			continue
		}

		if !rule.ConditionsMatch(e.StateStore) {
			continue
		}

		for i, action := range rule.Action {
			resolved := e.ResolveActionParams(action, event)
			action.Params = resolved
			e.Logger.Info("match found, queuing action", zap.String("rule", rule.Alias), zap.Int("action_number", i+1))
			e.queueAction(&action)
		}
	}
	return nil
}

func (e *Engine) Shutdown() {
	close(e.actionQueue)
	e.wg.Wait()
}
