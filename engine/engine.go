package engine

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"home_automation_server/automation"
	"home_automation_server/engine/integration"
	"home_automation_server/integrations"
	"home_automation_server/storage"
	"home_automation_server/storage/models"
	"home_automation_server/types"
	"sync"
	"time"
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

	// cache
	StateCache     types.StateStore
	EntityRegistry types.EntityRegistry // in-memory cache of the entityes

	// Event Transport
	EventChannel      chan types.Event
	ProcessedEventBus *EventBus // For publishing events after they have been processed.

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

		// Cache
		StateCache:     NewStateCache(),
		EntityRegistry: NewEntityRegistry(),

		ProcessedEventBus: NewEventBus(),               // for transmitting processed events to the ws manager
		EventChannel:      make(chan types.Event, 100), // for receiving events from eventPipelines supplied by the integration

		AutomationTaskQueue: make(chan *AutomationTask),
		ActionTimeout:       5 * time.Second,
		RetryPolicy: RetryPolicy{
			MaxAttempts: 3,
			Backoff:     500 * time.Millisecond,
		},
		nWorkers: nWorkers,

		Logger: logger.Named("engine"),
	}

	if err := e.RefreshEntityRegistry(ctx); err != nil {
		e.Logger.Error("failed to refresh entity registry", zap.Error(err))
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

func (e *Engine) Shutdown() {
	close(e.AutomationTaskQueue)
	e.wg.Wait()
}
