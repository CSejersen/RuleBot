package engine

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"home_automation_server/engine/pubsub"
	"home_automation_server/engine/rules"
	"home_automation_server/engine/statestore"
	"os"
	"strings"
	"sync"
)

type Engine struct {
	RuleSet         rules.RuleSet
	Integrations    map[string]Integration
	ServiceRegistry *ServiceRegistry
	StateStore      *statestore.StateStore

	// Collect events from Integrations
	PS           *pubsub.PubSub
	EventChannel <-chan pubsub.Event

	// Execute actions
	actionQueue chan *rules.Action
	wg          sync.WaitGroup
	nWorkers    int

	Logger *zap.Logger
}

func New(ctx context.Context, logger *zap.Logger, nWorkers int) (*Engine, error) {
	ps := pubsub.NewPubSub()

	e := &Engine{
		RuleSet:         rules.RuleSet{},
		Integrations:    make(map[string]Integration),
		ServiceRegistry: newServiceRegistry(),
		StateStore:      statestore.NewStateStore(),

		PS:           ps,
		EventChannel: ps.Subscribe(),

		actionQueue: make(chan *rules.Action),
		nWorkers:    nWorkers,

		Logger: logger.Named("engine"),
	}

	err := e.Init(ctx)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Engine) Init(ctx context.Context) error {
	if err := e.loadRules(); err != nil {
		return err
	}

	go func() {
		e.watchRules(ctx)
	}()

	e.startWorkers()
	return nil
}

func (e *Engine) queueAction(a *rules.Action) {
	e.actionQueue <- a
}

func (e *Engine) startWorkers() {
	for i := 0; i < e.nWorkers; i++ {
		e.wg.Add(1)
		go func(id int) {
			defer e.wg.Done()
			for action := range e.actionQueue {
				if err := e.executeAction(action); err != nil {
					e.Logger.Error("Failed to execute action", zap.Int("worker", id), zap.String("service", action.Service), zap.Error(err))
				}
			}
		}(i)
	}
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

func (e *Engine) RegisterIntegration(i Integration) {
	e.Integrations[i.Name] = i

	for service, handler := range i.Services {
		e.Logger.Debug("registering service", zap.String("service", service))
		e.RegisterService(i.Name, service, handler)
	}
}

func (e *Engine) RegisterService(domain, service string, handler ServiceHandler) {
	e.ServiceRegistry.Register(domain, service, handler)
}

func (e *Engine) loadRules() error {
	path := os.Getenv("RULES_FILE")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var rs rules.RuleSet
	if err := yaml.Unmarshal(data, &rs); err != nil {
		return err
	}

	e.RuleSet = rs

	e.Logger.Info("rules loaded", zap.Int("rule_count", len(rs.Rules)))
	return nil
}

func (e *Engine) watchRules(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		e.Logger.Error("failed to create fs watcher", zap.Error(err))
		return
	}
	defer watcher.Close()

	path := os.Getenv("RULES_FILE")
	if err := watcher.Add(path); err != nil {
		e.Logger.Error("failed to add rules file to watcher", zap.String("path", path), zap.Error(err))
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-watcher.Events:
			if evt.Op&(fsnotify.Write|fsnotify.Create) > 0 {
				e.Logger.Info("rules file changed, reloading", zap.String("file", evt.Name))
				if err := e.loadRules(); err != nil {
					e.Logger.Error("failed to reload rules", zap.Error(err))
				}
			}
		case err := <-watcher.Errors:
			e.Logger.Warn("rules watcher error", zap.Error(err))
		}
	}
}

func (e *Engine) ResolveActionParams(action rules.Action, event pubsub.Event) map[string]interface{} {
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
