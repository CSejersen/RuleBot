package main

import (
	"context"
	"go.uber.org/zap"
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
	e, err := engine.New(ctx, logger.Named("engine"), nWorkers)
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

func registerIntegrations(ctx context.Context, e *engine.Engine, logger *zap.Logger) {
	// Hue
	hueIntegration, err := hue.NewIntegration(logger)
	if err != nil {
		logger.Fatal("failed to init Hue integration", zap.Error(err))
	}
	e.RegisterIntegration(hueIntegration)

	//Halo
	haloIntegration, err := halo.NewIntegration(ctx, logger)
	if err != nil {
		logger.Fatal("failed to init Halo integration", zap.Error(err))
	}
	e.RegisterIntegration(haloIntegration)

	// Bang and Olufsen
	bangOlufsenIntegration, err := bangandolufsen.NewIntegration(ctx, logger)
	if err != nil {
		logger.Fatal("failed to init Bang and Olufsen integration", zap.Error(err))
	}
	e.RegisterIntegration(bangOlufsenIntegration)
	e.Logger.Debug("All integrations registered successfully", zap.Any("all_services", e.ServiceRegistry.GetAll()))
}

func runEngine(e *engine.Engine, logger *zap.Logger, ctx context.Context) {
	e.RunEventPipelines(ctx)
	e.ProcessEvents(ctx)
}
