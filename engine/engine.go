package engine

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	"home_automation_server/engine/storage"
	"home_automation_server/engine/types"
	integrationtypes "home_automation_server/integrations/types"
	"log"
	"os"
	"sync"
	"time"
)

type Engine struct {
	RuleSet         *rules.RuleSet
	Integrations    map[string]Integration
	ServiceRegistry *ServiceRegistry

	// storage
	EventStore storage.EventStore
	RuleStore  storage.RuleStore
	StateStore *storage.StateStore

	// Event Transport
	EventChannel      chan types.Event
	ProcessedEventBus *EventBus // For publishing events after they have been processed.

	// Execute actions
	ruleTaskQueue chan *RuleTask
	ActionTimeout time.Duration
	RetryPolicy   RetryPolicy
	wg            sync.WaitGroup
	nWorkers      int

	Logger *zap.Logger
}

type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

func New(ctx context.Context, logger *zap.Logger, nWorkers int) (*Engine, error) {
	db, err := NewMySQLConnection()
	if err != nil {
		return nil, err
	}

	e := &Engine{
		RuleSet:         &rules.RuleSet{},
		Integrations:    make(map[string]Integration),
		ServiceRegistry: newServiceRegistry(),

		StateStore: storage.NewStateStore(logger),
		EventStore: storage.NewMSqlEventStore(db),
		RuleStore:  storage.NewMSqlRuleStore(db),

		ProcessedEventBus: NewEventBus(),          // for transmitting processed events to the ws manager
		EventChannel:      make(chan types.Event), // for receiving events from eventPipelines supplied by the integrations

		ruleTaskQueue: make(chan *RuleTask),
		ActionTimeout: 5 * time.Second,
		RetryPolicy: RetryPolicy{
			MaxAttempts: 3,
			Backoff:     500 * time.Millisecond,
		},
		nWorkers: nWorkers,

		Logger: logger.Named("engine"),
	}

	// TODO: Use migrations or something... we cannot live like this
	if err := e.RuleStore.EnsureTableExists(); err != nil {
		return nil, err
	}
	if err := e.EventStore.EnsureTableExists(); err != nil {
		return nil, err
	}

	if err := e.Init(ctx); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Engine) Init(ctx context.Context) error {
	if err := e.loadRules(ctx); err != nil {
		return err
	}

	go func() {
		e.watchRules(ctx)
	}()

	e.startWorkers()
	return nil
}

func (e *Engine) watchRules(ctx context.Context) {
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

	for service, data := range i.Services {
		e.Logger.Debug("registering service", zap.String("service", service), zap.String("integration", i.Name))
		e.RegisterService(i.Name, service, data)
	}
}

func (e *Engine) RegisterService(domain, service string, data integrationtypes.ServiceData) {
	e.ServiceRegistry.Register(domain, service, data)
}

func (e *Engine) loadRules(ctx context.Context) error {
	rs, err := e.RuleStore.LoadRules(ctx)
	if err != nil {
		return err
	}

	e.Logger.Debug("loaded rules", zap.Any("rules", rs))
	e.RuleSet = rs

	return nil
}

func (e *Engine) Shutdown() {
	close(e.ruleTaskQueue)
	e.wg.Wait()
}

// GetMySQLDSN reads env vars and constructs a connection string
func GetMySQLDSN() (string, error) {
	user := os.Getenv("MYSQL_USER")
	pass := os.Getenv("MYSQL_PASSWORD")
	db := os.Getenv("MYSQL_DATABASE")
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")

	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "3306"
	}

	if user == "" || pass == "" || db == "" {
		return "", fmt.Errorf("MYSQL_USER, MYSQL_PASSWORD, and MYSQL_DATABASE must be set")
	}

	// Format: user:password@tcp(host:port)/dbname?parseTime=true
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, db)
	return dsn, nil
}

func NewMySQLConnection() (*sql.DB, error) {
	dsn, err := GetMySQLDSN()
	if err != nil {
		return nil, fmt.Errorf("failed to get DSN: %w", err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	log.Println("Connected to MySQL successfully")
	return db, nil
}
