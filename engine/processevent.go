package engine

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/pubsub"
	"home_automation_server/engine/rules"
	"strings"
	"time"
)

type RuleTask struct {
	Actions []*rules.Action
	Event   *pubsub.Event
}

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
	// Update cached state
	e.StateStore.ApplyEvent(event)

	for _, rule := range e.RuleSet.Rules {
		if !rule.Trigger.Matches(event) {
			continue
		}

		if !rule.ConditionsMatch(e.StateStore) {
			continue
		}

		e.queueRuleTask(&rule, &event)
	}
	return nil
}

func (e *Engine) executeRuleTask(task *RuleTask, workerID int) {
	for _, action := range task.Actions {
		resolved := e.ResolveActionParams(action, task.Event)
		action.Params = resolved

		if action.Blocking {
			ctx := context.Background()
			if e.ActionTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, e.ActionTimeout)
				defer cancel()
			}
			// Wait for completion
			if err := e.executeActionWithRetry(ctx, action); err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					e.Logger.Warn("Action timed out", zap.String("service", action.Service))
				} else {
					e.Logger.Error("Failed blocking action", zap.String("service", action.Service), zap.Error(err))
				}
				break // stop further actions in this rule
			}
		} else {
			// fire-and-forget
			go func(action *rules.Action) {
				// each goroutine gets its own cancel
				ctx := context.Background()
				if e.ActionTimeout > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, e.ActionTimeout)
					defer cancel()
				}

				if err := e.executeActionWithRetry(ctx, action); err != nil {
					e.Logger.Error("Failed non-blocking action", zap.String("service", action.Service), zap.Error(err))
				}
			}(action)
		}
	}
}

func (e *Engine) executeActionWithRetry(ctx context.Context, action *rules.Action) error {
	var err error
	for attempt := 1; attempt <= e.RetryPolicy.MaxAttempts; attempt++ {
		err = e.executeAction(ctx, action)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		e.Logger.Warn("Action failed, retrying",
			zap.String("service", action.Service),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)

		time.Sleep(e.RetryPolicy.Backoff)
	}
	return fmt.Errorf("action failed after %d attempts: %w", e.RetryPolicy.MaxAttempts, err)
}

func (e *Engine) executeAction(ctx context.Context, a *rules.Action) error {
	split := strings.Split(a.Service, ".")
	if len(split) != 2 {
		return fmt.Errorf("invalid service format: %s", a.Service)
	}

	domain := split[0]
	service := split[1]

	e.Logger.Debug("Calling service", zap.String("service", a.Service), zap.Any("params", a.Params))
	if err := e.ServiceRegistry.Call(ctx, domain, service, a); err != nil {
		return err
	}
	return nil
}

func (e *Engine) queueRuleTask(rule *rules.Rule, event *pubsub.Event) {
	e.Logger.Info("queueing rule task", zap.String("rule", rule.Alias))

	task := &RuleTask{
		Event:   event,
		Actions: make([]*rules.Action, len(rule.Action)),
	}

	for i, action := range rule.Action {
		resolved := e.ResolveActionParams(&action, event)
		action.Params = resolved
		task.Actions[i] = &action
	}

	e.ruleTaskQueue <- task
}

func (e *Engine) ResolveActionParams(action *rules.Action, event *pubsub.Event) map[string]interface{} {
	e.Logger.Debug("Resolving action params", zap.Any("actionParams", action.Params), zap.Any("eventPayload", event.Payload))
	resolved := make(map[string]interface{})
	for k, v := range action.Params {
		switch val := v.(type) {
		case string:
			if ref, ok := rules.ParseTemplateParam(val); ok {
				// param is templated
				switch ref.Source {
				case "payload":
					if value, ok := event.Payload[ref.Path]; ok {
						resolved[k] = value
					} else {
						e.Logger.Debug("param not found in payload")
						resolved[k] = nil
					}
				case "state":
					if value, ok := e.StateStore.ResolvePath(ref.Path); ok {
						resolved[k] = value
					}
					resolved[k] = ref.Default // if no default is defined this will just be nil
				}
			} else {
				// param is a non templated string
				resolved[k] = val
			}
		default:
			resolved[k] = val
		}
	}
	return resolved
}
