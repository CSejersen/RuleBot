package engine

import (
	"context"
	"errors"
	"fmt"
	"home_automation_server/automation"
	"home_automation_server/storage"
	"home_automation_server/types"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
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

	// Skip automation triggers for optimistic events (origin="service")
	shouldSkipTriggers := event.Context != nil && event.Context.Origin == "service"
	if !shouldSkipTriggers {
		e.evaluateAutomationTriggers(ctx, event)
	} else {
		e.Logger.Debug("Skipping automation triggers for optimistic event",
			zap.String("event_type", string(event.Type)),
			zap.String("context_id", event.Context.ID))
	}

	// Save and publish the event
	e.saveAndPublishEvent(ctx, event)
}

func (e *Engine) evaluateAutomationTriggers(ctx context.Context, event types.Event) {
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
}

func (e *Engine) saveAndPublishEvent(ctx context.Context, event types.Event) {
	storageEvent, err := storage.EventToStorage(event)
	if err != nil {
		e.Logger.Error("failed to convert event to storage model", zap.Error(err))
		return
	}

	savedEvent, err := e.EventStore.SaveEvent(ctx, storageEvent)
	if err != nil {
		e.Logger.Error("failed to save event", zap.Error(err))
		return
	}

	// Populate ID
	if event.Context != nil {
		event.ID = event.Context.ID
	} else {
		event.ID = savedEvent.ID
	}

	e.EventBus.Publish(event) // broadcasted to subscribers (e.g., WebSocket)
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
			if err := e.executeActionWithRetry(ctx, action, task.Event); err != nil {
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

				if err := e.executeActionWithRetry(ctx, action, task.Event); err != nil {
					e.Logger.Error("Failed non-blocking action", zap.String("service", action.Service), zap.Error(err))
				}
			}(action)
		}
	}
}

func (e *Engine) executeActionWithRetry(ctx context.Context, action *automation.Action, parentEvent *types.Event) error {
	var err error
	for attempt := 1; attempt <= e.RetryPolicy.MaxAttempts; attempt++ {
		err = e.executeAction(ctx, action, parentEvent)
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

func (e *Engine) executeAction(ctx context.Context, a *automation.Action, parentEvent *types.Event) error {
	split := strings.Split(a.Service, ".")
	if len(split) != 2 {
		return fmt.Errorf("invalid service format: %s", a.Service)
	}

	domain := split[0]
	service := split[1]

	e.Logger.Debug("Calling service", zap.String("service", a.Service), zap.Any("params", a.Params))

	// Determine parent context ID - use the parent event's context ID if available
	var parentContextID string
	if parentEvent != nil && parentEvent.Context != nil {
		parentContextID = parentEvent.Context.ID
	}

	// Emit call_service event
	serviceCallContextID := uuid.NewString()
	targetEntityIDs := e.extractEntityIDs(a.Targets)
	serviceCallEvent := types.Event{
		Type: types.EventTypeCallService,
		Data: types.CallServiceData{
			Domain:      domain,
			Service:     service,
			ServiceData: a.Params,
			EntityID:    targetEntityIDs,
		},
		Context: &types.Context{
			ID:       serviceCallContextID,
			ParentID: parentContextID,
			Origin:   "automation",
		},
		TimeFired: time.Now(),
	}

	// Emit service call event non-blocking
	select {
	case e.EventChannel <- serviceCallEvent:
		e.Logger.Debug("Emitted call_service event",
			zap.String("service", a.Service),
			zap.String("context_id", serviceCallContextID),
		)
	default:
		e.Logger.Warn("Event channel full, dropped call_service event",
			zap.String("service", a.Service),
		)
	}

	// Record service call context for target entities to enable causality linking
	for _, target := range a.Targets {
		if target.EntityID != "" {
			e.ServiceCallTracker.Record(target.EntityID, serviceCallContextID, a.Params)
		}
	}
	// Also add to a global queue that tracks by domain
	e.ServiceCallTracker.Record(domain, serviceCallContextID, a.Params)

	// Convert targets to external IDs for the service call
	for i, t := range a.Targets {
		externalID, ok := e.EntityRegistry.ResolveExternalID(t.EntityID)
		if !ok {
			return fmt.Errorf("failed to resolve externalID for entity %s", t.EntityID)
		}
		a.Targets[i].ExternalID = externalID
	}

	// Create emitter with the service call context ID as parent
	emitter := NewServiceEventEmitter(e.EventChannel, e.StateCache, e.EntityRegistry, e.Logger, serviceCallContextID)
	if err := e.ServiceRegistry.Call(ctx, domain, service, a, emitter); err != nil {
		return err
	}
	return nil
}

// extractEntityIDs extracts entity IDs from action targets and joins them
// Assumes targets already contain entity IDs (not external IDs)
func (e *Engine) extractEntityIDs(targets []automation.Target) string {
	if len(targets) == 0 {
		return ""
	}

	ids := make([]string, 0, len(targets))
	for _, t := range targets {
		ids = append(ids, t.EntityID)
	}

	if len(ids) == 1 {
		return ids[0]
	}
	return strings.Join(ids, ",")
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
		resolvedVal, err := e.ResolveActionParam(val)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve templateRef '%s': %w", name, err)
		}

		resolved[name] = resolvedVal
	}

	return resolved, nil
}

func (e *Engine) ResolveActionParam(val any) (any, error) {
	str, ok := val.(string)
	if !ok {
		return val, nil
	}

	if strings.Contains(str, "{{") {
		ref, err := ParseTemplateRef(str)
		if err != nil {
			return nil, fmt.Errorf("template parsing error: %w", err)
		}
		return e.ResolveTemplateRef(ref)
	}
	return str, nil
}

func ParseTemplateRef(tmpl string) (TemplateRef, error) {
	// Validate that template contains braces
	openBraces := strings.Count(tmpl, "{{")
	closeBraces := strings.Count(tmpl, "}}")

	if openBraces != closeBraces {
		return TemplateRef{}, fmt.Errorf("template syntax error: unmatched braces (found %d '{{' and %d '}}') in '%s'", openBraces, closeBraces, tmpl)
	}

	if openBraces > 1 {
		return TemplateRef{}, fmt.Errorf("template syntax error: multiple template blocks found, expected exactly one in '%s'", tmpl)
	}

	// Extract the template content
	startIdx := strings.Index(tmpl, "{{")
	endIdx := strings.Index(tmpl, "}}")
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return TemplateRef{}, fmt.Errorf("template syntax error: malformed template braces in '%s'", tmpl)
	}

	templateContent := strings.TrimSpace(tmpl[startIdx+2 : endIdx])
	if templateContent == "" {
		return TemplateRef{}, fmt.Errorf("template syntax error: empty template reference")
	}

	// Validate function name
	if !strings.HasPrefix(templateContent, "state") {
		// Try to extract function name for better error message
		funcEnd := strings.IndexAny(templateContent, "(")
		var funcName string
		if funcEnd > 0 {
			funcName = strings.TrimSpace(templateContent[:funcEnd])
		} else {
			funcName = strings.Fields(templateContent)[0]
		}
		return TemplateRef{}, fmt.Errorf("template syntax error: unknown function '%s', expected 'state' or 'state_attr'", funcName)
	}

	// Try state_attr matches first (more specific)
	matches := reStateAttr.FindAllStringSubmatchIndex(tmpl, -1)
	if len(matches) > 0 {
		match := matches[0]
		full := tmpl[match[0]:match[1]]
		entity := tmpl[match[2]:match[3]]
		attr := tmpl[match[4]:match[5]]

		if entity == "" {
			return TemplateRef{}, fmt.Errorf("template syntax error: empty entity_id in 'state_attr' template")
		}
		if attr == "" {
			return TemplateRef{}, fmt.Errorf("template syntax error: empty attribute in 'state_attr' template")
		}

		return TemplateRef{
			FullMatch: full,
			EntityID:  entity,
			Attribute: attr,
			FuncName:  "state_attr",
		}, nil
	}

	// Try state() matches
	matches = reState.FindAllStringSubmatchIndex(tmpl, -1)
	if len(matches) > 0 {
		match := matches[0]
		full := tmpl[match[0]:match[1]]
		entity := tmpl[match[2]:match[3]]

		if entity == "" {
			return TemplateRef{}, fmt.Errorf("template syntax error: empty entity_id in 'state' template")
		}

		return TemplateRef{
			FullMatch: full,
			EntityID:  entity,
			Attribute: "",
			FuncName:  "state",
		}, nil
	}

	// No match found, but we detected braces - provide helpful error
	if strings.HasPrefix(templateContent, "state_attr") {
		return TemplateRef{}, fmt.Errorf("template syntax error: malformed 'state_attr' template, expected state_attr('entity_id', 'attr') in '%s'", tmpl)
	}
	if strings.HasPrefix(templateContent, "state") {
		return TemplateRef{}, fmt.Errorf("template syntax error: malformed 'state' template, expected state('entity_id') in '%s'", tmpl)
	}

	return TemplateRef{}, fmt.Errorf("template syntax error: malformed template, expected state('entity_id') or state_attr('entity_id', 'attr') in '%s'", tmpl)
}

func (e *Engine) ResolveTemplateRef(ref TemplateRef) (any, error) {
	// Validate that we have a valid function name (should be rare after ParseTemplateRef validation)
	if ref.FuncName == "" {
		return nil, fmt.Errorf("template resolution error: empty function name in template reference")
	}
	if ref.EntityID == "" {
		return nil, fmt.Errorf("template resolution error: empty entity_id in template reference")
	}

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
