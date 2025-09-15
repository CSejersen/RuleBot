package engine

import (
	"context"
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

		e.queueActionsForRule(&rule, &event)
	}
	return nil
}

func (e *Engine) queueActionsForRule(rule *rules.Rule, event *pubsub.Event) {
	for i, action := range rule.Action {
		e.Logger.Info("queuing action", zap.String("rule_alias", rule.Alias), zap.Int("rule_num", i))
		resolved := e.ResolveActionParams(&action, event)
		action.Params = resolved
		e.actionQueue <- &action
	}
}

func (e *Engine) Shutdown() {
	close(e.actionQueue)
	e.wg.Wait()
}

func (e *Engine) ResolveActionParams(action *rules.Action, event *pubsub.Event) map[string]interface{} {
	e.Logger.Debug("Resolving action params", zap.Any("actionParams", action.Params), zap.Any("eventPayload", event.Payload))
	resolved := make(map[string]interface{})
	for k, v := range action.Params {
		switch val := v.(type) {
		case string:
			if ref, ok := parseTemplate(val); ok {
				// param is templated
				switch ref.Source {
				case "payload":
					e.Logger.Debug("resolving param from Payload")
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

func parseTemplate(val string) (*rules.TemplateRef, bool) {
	val = strings.TrimSpace(val)

	if !strings.HasPrefix(val, "${") || !strings.HasSuffix(val, "}") {
		return nil, false
	}

	inner := strings.TrimSuffix(strings.TrimPrefix(val, "${"), "}")

	// optional defaultVal, split on "|"
	var path string
	var defaultVal any
	parts := strings.Split(inner, "|")
	path = strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		defaultVal = strings.TrimSpace(parts[1])
	}

	switch {
	case strings.HasPrefix(path, "payload."):
		return &rules.TemplateRef{
			Source:  "payload",
			Path:    strings.TrimPrefix(path, "payload."),
			Default: defaultVal,
		}, true
	case strings.HasPrefix(path, "state."):
		return &rules.TemplateRef{
			Source:  "state",
			Path:    strings.TrimPrefix(path, "state."),
			Default: defaultVal,
		}, true
	default:
		return nil, false
	}
}
