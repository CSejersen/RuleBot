package engine

import (
	"context"
	"fmt"
	"home_automation_server/automation"
	"home_automation_server/engine/integration"
	"home_automation_server/integrations"
	"home_automation_server/storage"
	"home_automation_server/storage/models"
	"home_automation_server/types"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Engine struct {
	Automations             *automation.AutomationSet
	Integrations            map[string]integration.Instance      // Enabled integration
	IntegrationDescRegistry *integration.IntegrationDescRegistry // Holds descriptors for all available integration
	ServiceRegistry         *ServiceRegistry

	// storage
	EventStore          storage.EventStore
	AutomationStore     storage.AutomationStore
	IntegrationCfgStore storage.IntegrationCfgStore
	DeviceStore         storage.DeviceStore
	EntityStore         storage.EntityStore
	StateStore          storage.StateStore

	// cache
	StateCache         types.StateStore
	EntityRegistry     types.EntityRegistry // in-memory cache of the entityes
	ServiceCallTracker *ServiceCallTracker  // Tracks recent service calls for causality linking

	// Event Transport
	EventChannel chan types.Event
	EventBus     *EventBus // For broadcasting events to subscribers (e.g., WebSocket)

	// Execute actions
	AutomationTaskQueue chan *AutomationTask
	ActionTimeout       time.Duration
	RetryPolicy         RetryPolicy
	wg                  sync.WaitGroup
	nWorkers            int

	Logger *zap.Logger
}

type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

func New(ctx context.Context, db *gorm.DB, logger *zap.Logger, nWorkers int) (*Engine, error) {
	err := db.AutoMigrate(
		&models.IntegrationConfig{},
		&models.Device{},
		&models.Entity{},
		&models.Automation{},
		&models.Context{},
		&models.Event{},
		&models.State{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate db: %w", err)
	}

	e := &Engine{
		Automations:             &automation.AutomationSet{},
		Integrations:            make(map[string]integration.Instance),
		IntegrationDescRegistry: integration.NewIntegrationRegistry(),
		ServiceRegistry:         newServiceRegistry(),

		// storage
		EventStore:          storage.NewGormEventStore(db),
		AutomationStore:     storage.NewGormRuleStore(db),
		IntegrationCfgStore: storage.NewGormIntegrationCfgStore(db),
		DeviceStore:         storage.NewGormDeviceStore(db),
		EntityStore:         storage.NewGormEntityStore(db),
		StateStore:          storage.NewGormStateStore(db),

		// Cache
		StateCache:         NewStateCache(),
		EntityRegistry:     NewEntityRegistry(),
		ServiceCallTracker: NewServiceCallTracker(5 * time.Second),

		EventBus:     NewEventBus(logger),         // for broadcasting events to subscribers
		EventChannel: make(chan types.Event, 100), // for receiving events from eventPipelines supplied by the integration

		AutomationTaskQueue: make(chan *AutomationTask),
		ActionTimeout:       5 * time.Second,
		RetryPolicy: RetryPolicy{
			MaxAttempts: 3,
			Backoff:     500 * time.Millisecond,
		},
		nWorkers: nWorkers,

		Logger: logger, // logger is already named "engine" in bootstrap.go
	}

	if err := e.RefreshEntityRegistry(ctx); err != nil {
		e.Logger.Error("failed to refresh entity registry", zap.Error(err))
	}

	// Load states from database before initializing
	if err := e.LoadStates(ctx); err != nil {
		e.Logger.Error("failed to load states from database", zap.Error(err))
	}

	if err := e.Init(ctx); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Engine) Init(ctx context.Context) error {
	if err := e.LoadAutomations(ctx); err != nil {
		return err
	}
	e.startWorkers()
	e.startStatePersistence(ctx)
	return nil
}

func (e *Engine) startWorkers() {
	for i := 0; i < e.nWorkers; i++ {
		e.wg.Add(1)
		go func(id int) {
			defer e.wg.Done()
			for task := range e.AutomationTaskQueue {
				e.executeAutomationTask(task)
			}
		}(i)
	}
}

func (e *Engine) RegisterService(domain, service string, spec integrations.ServiceSpec) {
	e.ServiceRegistry.Register(domain, service, spec)
	e.Logger.Info("Registered service", zap.String("domain", domain), zap.String("service", service))
}

func (e *Engine) LoadAutomations(ctx context.Context) error {
	storageAutomations, err := e.AutomationStore.LoadAutomations(ctx)
	if err != nil {
		return err
	}

	automations := make([]automation.Automation, len(storageAutomations))
	for i, storageAutomation := range storageAutomations {
		a, err := storage.AutomationFromStorage(storageAutomation)
		if err != nil {
			return fmt.Errorf("unable to convert automation from storage model: %w", err)
		}
		automations[i] = a
	}

	e.Logger.Info("successfully loaded automations from storage", zap.Int("num_automations", len(automations)))
	e.Automations = &automation.AutomationSet{
		Automations: automations,
	}

	return nil
}

// LoadStates loads all states from the database and populates the StateCache
func (e *Engine) LoadStates(ctx context.Context) error {
	states, err := e.StateStore.LoadAllStates(ctx)
	if err != nil {
		return fmt.Errorf("failed to load states from database: %w", err)
	}

	for _, state := range states {
		e.StateCache.Set(state.EntityID, state)
	}

	e.Logger.Info("successfully loaded states from database", zap.Int("num_states", len(states)))
	return nil
}

// startStatePersistence starts a goroutine that periodically saves all states to the database
func (e *Engine) startStatePersistence(ctx context.Context) {
	// Get interval from environment variable, default to 30 seconds
	intervalStr := os.Getenv("STATE_PERSIST_INTERVAL")
	interval := 30 * time.Second
	if intervalStr != "" {
		if secs, err := strconv.Atoi(intervalStr); err == nil && secs > 0 {
			interval = time.Duration(secs) * time.Second
		}
	}

	e.Logger.Info("starting state persistence", zap.Duration("interval", interval))

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				e.Logger.Info("stopping state persistence due to context cancellation")
				return
			case <-ticker.C:
				states := e.StateCache.GetAll()
				if len(states) > 0 {
					if err := e.StateStore.SaveAllStates(ctx, states); err != nil {
						e.Logger.Error("failed to persist states", zap.Error(err))
					} else {
						e.Logger.Debug("persisted states to database", zap.Int("num_states", len(states)))
					}
				}
			}
		}
	}()
}

func (e *Engine) Shutdown() {
	// Save all states before shutting down
	ctx := context.Background()
	states := e.StateCache.GetAll()
	if len(states) > 0 {
		if err := e.StateStore.SaveAllStates(ctx, states); err != nil {
			e.Logger.Error("failed to save states on shutdown", zap.Error(err))
		} else {
			e.Logger.Info("saved states on shutdown", zap.Int("num_states", len(states)))
		}
	}

	close(e.AutomationTaskQueue)
	e.wg.Wait()
}
