package engine

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/pubsub"
	"home_automation_server/engine/rules"
	"strings"
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
			resolved := action.ResolveTemplatedParams(event)
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

func (e *Engine) executeAction(a *rules.Action) error {
	split := strings.Split(a.Service, ".")
	if len(split) != 2 {
		return fmt.Errorf("invalid service format: %s", a.Service)
	}

	domain := split[0]
	service := split[1]

	e.Logger.Debug("Calling service", zap.String("service", a.Service), zap.Any("params", a.Params))
	if err := e.ServiceRegistry.Call(domain, service, a); err != nil {
		return err
	}
	return nil
}
