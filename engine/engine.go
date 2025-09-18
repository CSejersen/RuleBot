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
	"sync"
	"time"
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
	ActionTimeout time.Duration
	RetryPolicy   RetryPolicy
	ruleTaskQueue chan *RuleTask
	wg            sync.WaitGroup
	nWorkers      int

	Logger *zap.Logger
}

type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

func New(ctx context.Context, logger *zap.Logger, nWorkers int) (*Engine, error) {
	ps := pubsub.NewPubSub()

	e := &Engine{
		RuleSet:         rules.RuleSet{},
		Integrations:    make(map[string]Integration),
		ServiceRegistry: newServiceRegistry(),
		StateStore:      statestore.NewStateStore(logger),

		PS:           ps,
		EventChannel: ps.Subscribe(),

		ActionTimeout: 5 * time.Second,
		RetryPolicy: RetryPolicy{
			MaxAttempts: 3,
			Backoff:     2 * time.Second,
		},
		ruleTaskQueue: make(chan *RuleTask),
		nWorkers:      nWorkers,

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

func (e *Engine) startWorkers() {
	for i := 0; i < e.nWorkers; i++ {
		e.wg.Add(1)
		go func(id int) {
			defer e.wg.Done()
			for task := range e.ruleTaskQueue {
				e.executeRuleTask(task, id)
			}
		}(i)
	}
}

func (e *Engine) RegisterIntegration(i Integration) {
	e.Integrations[i.Name] = i

	for service, handler := range i.Services {
		e.Logger.Debug("registering service", zap.String("service", service), zap.String("integration", i.Name))
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

func (e *Engine) Shutdown() {
	close(e.ruleTaskQueue)
	e.wg.Wait()
}
