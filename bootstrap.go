package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"home_automation_server/engine"
	"home_automation_server/integrations/bangandolufsen"
	"home_automation_server/integrations/halo"
	"home_automation_server/integrations/hue"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func setupContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func setupEngine(ctx context.Context, logger *zap.Logger, nWorkers int) (*engine.Engine, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("MYSQL_DSN environment variable not set")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	e, err := engine.New(ctx, db, logger.Named("engine"), nWorkers)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func setupLogger() *zap.Logger {
	//mongoURI := os.Getenv("MONGO_URI")
	//if mongoURI == "" {
	//	panic("MONGO_URI environment variable not set")
	//}
	//mongoDB := os.Getenv("MONGO_DATABASE")
	//mongoColl := os.Getenv("MONGO_COLL")
	//
	//mongoCore, err := logging.NewMongoCore(mongoURI, mongoDB, mongoColl, zapcore.InfoLevel)
	//if err != nil {
	//	panic(err)
	//}
	//
	//return zap.New(mongoCore)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	return logger
}

func registerIntegrationDescriptors(e *engine.Engine) {
	reg := e.IntegrationDescRegistry
	reg.Register(hue.Descriptor())
	reg.Register(halo.Descriptor())
	reg.Register(bangandolufsen.Descriptor())

	e.Logger.Info("Integration descriptors registered successfully", zap.Int("num_descriptors", len(reg.List())))
}

func LoadIntegrations(ctx context.Context, e *engine.Engine) error {
	integrationCfgs, err := e.IntegrationCfgStore.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("unable to load integration configurations: %w", err)
	}
	for _, cfg := range integrationCfgs {
		err := e.LoadIntegration(ctx, cfg.IntegrationName)
		if err != nil {
			return fmt.Errorf("unable to load integration %s: %w", cfg.IntegrationName, err)
		}
	}

	e.Logger.Info("Successfully loaded integration", zap.Int("num_active_integrations", len(integrationCfgs)))
	return nil
}

func runEngine(e *engine.Engine, logger *zap.Logger, ctx context.Context) {
	e.ProcessEvents(ctx)
}
