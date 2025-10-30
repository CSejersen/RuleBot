package engine

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/automation"
	"home_automation_server/storage"
	"home_automation_server/types"
	"regexp"
	"strings"
	"time"
)

type AutomationTask struct {
	Actions []*automation.Action
	Event   *types.Event
}

type TemplateRef struct {
	FullMatch string // the full template text, e.g. "{{ state_attr('light.living_room', 'brightness') }}"
	EntityID  string
	Attribute string
	FuncName  string // "state" or "state_attr"
}

var (
	reState     = regexp.MustCompile(`{{\s*state\(['"]([^'"]+)['"]\)\s*}}`)
	reStateAttr = regexp.MustCompile(`{{\s*state_attr\(['"]([^'"]+)['"]\s*,\s*['"]([^'"]+)['"]\)\s*}}`)
)

func (e *Engine) ProcessEvents(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				e.Logger.Info("context cancelled")
				return
			case event := <-e.EventChannel:
				e.processEvent(ctx, event)
			}
		}
	}()
}

func (e *Engine) processEvent(ctx context.Context, event types.Event) {
	e.Logger.Debug("processing event", zap.Any("event", event))
	e.updateStateCache(event)

	for _, a := range e.Automations.Automations {
		if !a.Enabled {
			continue
		}

		triggerFired := false
		for _, baseTrigger := range a.Trigger {
			trigger, err := baseTrigger.AsTrigger()
			if err != nil {
				e.Logger.Error("failed to convert baseTrigger to trigger", zap.Error(err))
				return
			}
			fired, err := trigger.Evaluate(event)
			if err != nil {
				e.Logger.Error("trigger evaluation failed", zap.Error(err), zap.Uint("automation_id", a.Id))
				continue
			}

			if fired {
				triggerFired = true
				break // if any trigger fires, the automation should run
			}
		}

		if !triggerFired {
			continue // skip this automation
		}

		if err := e.queueAutomationTask(&a, &event); err != nil {
			e.Logger.Error("failed to enqueue automation task", zap.Error(err))
		}

		if err := e.AutomationStore.UpdateLastTriggered(ctx, a.Id); err != nil {
			e.Logger.Error("failed to update last triggered", zap.Error(err))
		}
	}
	storageEvent, err := storage.EventToStorage(event)
	if err != nil {
		e.Logger.Error("failed to convert event to storage model", zap.Error(err))
	}

	if err := e.EventStore.SaveEvent(ctx, storageEvent); err != nil {
		e.Logger.Error("failed to save event", zap.Error(err))
	}

	e.ProcessedEventBus.Publish(event) // will be received by the ws-manager for real time updates in the UI.
}

func (e *Engine) executeAutomationTask(task *AutomationTask) {
	for _, action := range task.Actions {
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
					e.Logger.Warn("Actions timed out", zap.String("service", action.Service))
				} else {
					e.Logger.Error("Failed blocking action", zap.String("service", action.Service), zap.Error(err))
				}
				break // stop further actions in this rule
			}
		} else {
			// fire-and-forget
			go func(action *automation.Action) {
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

func (e *Engine) executeActionWithRetry(ctx context.Context, action *automation.Action) error {
	var err error
	for attempt := 1; attempt <= e.RetryPolicy.MaxAttempts; attempt++ {
		err = e.executeAction(ctx, action)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		e.Logger.Warn("Actions failed, retrying",
			zap.String("service", action.Service),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)

		time.Sleep(e.RetryPolicy.Backoff)
	}
	return fmt.Errorf("action failed after %d attempts: %w", e.RetryPolicy.MaxAttempts, err)
}

func (e *Engine) executeAction(ctx context.Context, a *automation.Action) error {
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

func (e *Engine) queueAutomationTask(a *automation.Automation, event *types.Event) error {
	e.Logger.Info("queueing automation task", zap.String("automation", a.Alias))

	task := &AutomationTask{
		Event:   event,
		Actions: make([]*automation.Action, len(a.Actions)),
	}

	for i, action := range a.Actions {
		resolved, err := e.ResolveActionParams(&action)
		if err != nil {
			return fmt.Errorf("failed to resolve action params: %w", err)
		}
		action.Params = resolved

		resolvedTargets, err := e.ResolveTargetsToExternalID(action.Targets)
		if err != nil {
			return err
		}
		action.Targets = resolvedTargets

		task.Actions[i] = &action
	}

	e.AutomationTaskQueue <- task
	return nil
}

func (e *Engine) ResolveTargetsToExternalID(targets []automation.Target) ([]automation.Target, error) {
	resolved := make([]automation.Target, len(targets))
	for i, t := range targets {
		externalID, ok := e.EntityRegistry.ResolveExternalID(t.EntityID)
		if !ok {
			return nil, fmt.Errorf("failed to resolve externalID for entity %s", t.EntityID)
		}
		resolved[i] = t
		resolved[i].EntityID = externalID
	}
	return resolved, nil
}

func (e *Engine) updateStateCache(event types.Event) {
	data, ok := event.Data.(types.StateChangedData)
	if !ok {
		return
	}
	e.StateCache.Set(data.EntityID, *data.NewState)
}

func (e *Engine) ResolveActionParams(action *automation.Action) (map[string]any, error) {
	resolved := make(map[string]any, len(action.Params))

	for name, val := range action.Params {
		// If value is not a string, leave it as-is
		str, ok := val.(string)
		if !ok {
			resolved[name] = val
			continue
		}

		if strings.Contains(str, "{{") {
			ref := ParseTemplateRef(str)
			resolvedVal, err := e.ResolveTemplateRef(ref)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve templateRef '%s': %w", name, err)
			}
			resolved[name] = resolvedVal
			continue
		}

		// If not templated just keep the original string
		resolved[name] = str
	}

	return resolved, nil
}

func ParseTemplateRef(tmpl string) TemplateRef {
	// state_attr matches
	for _, match := range reStateAttr.FindAllStringSubmatchIndex(tmpl, -1) {
		full := tmpl[match[0]:match[1]]
		entity := tmpl[match[2]:match[3]]
		attr := tmpl[match[4]:match[5]]

		return TemplateRef{
			FullMatch: full,
			EntityID:  entity,
			Attribute: attr,
			FuncName:  "state_attr",
		}
	}

	// state() matches
	for _, match := range reState.FindAllStringSubmatchIndex(tmpl, -1) {
		full := tmpl[match[0]:match[1]]
		entity := tmpl[match[2]:match[3]]

		return TemplateRef{
			FullMatch: full,
			EntityID:  entity,
			Attribute: "",
			FuncName:  "state",
		}
	}

	return TemplateRef{}
}

func (e *Engine) ResolveTemplateRef(ref TemplateRef) (any, error) {
	st, ok := e.StateCache.Get(ref.EntityID)
	if !ok {
		return nil, fmt.Errorf("failed to get state for entity_id: %s", ref.EntityID)
	}

	if ref.FuncName == "state_attr" {
		if ref.Attribute == "" {
			return nil, fmt.Errorf("no attribute in 'state_attr' template ref")
		}
		attr, ok := st.Attributes[ref.Attribute]
		if !ok {
			return nil, fmt.Errorf("state does not contain the attribute: %s", ref.Attribute)
		}
		return attr, nil
	}
	if ref.FuncName == "state" {
		state := st.State
		if state == nil {
			return nil, fmt.Errorf("main_state for entity: %s is nil", ref.EntityID)
		}
		return state, nil
	}

	return nil, fmt.Errorf("unknown template function: %s", ref.FuncName)
}
