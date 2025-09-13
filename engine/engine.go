package engine

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"home_automation_server/engine/pubsub"
	"home_automation_server/engine/rules"
	"home_automation_server/engine/statestore"
	"os"
)

type Engine struct {
	RuleSet         rules.RuleSet
	Integrations    map[string]Integration
	ServiceRegistry *ServiceRegistry
	StateStore      *statestore.StateStore
	PS              *pubsub.PubSub
	Logger          *zap.Logger
	EventChannel    <-chan pubsub.Event
}

func New(logger *zap.Logger) *Engine {
	ps := pubsub.NewPubSub()

	return &Engine{
		RuleSet:         rules.RuleSet{},
		Integrations:    make(map[string]Integration),
		ServiceRegistry: newServiceRegistry(),
		StateStore:      statestore.NewStateStore(),
		PS:              ps,
		Logger:          logger.Named("engine"),
		EventChannel:    ps.Subscribe(),
	}
}

func (e *Engine) Init(ctx context.Context) error {
	if err := e.loadRules(); err != nil {
		return err
	}

	go func() {
		e.watchRules(ctx)
	}()
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
