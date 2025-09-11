package main

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/engine"
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

func setupLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	return logger
}

func setupEngine(ctx context.Context, logger *zap.Logger) *engine.Engine {
	e := engine.New(logger.Named("engine"))
	if err := e.Init(ctx); err != nil {
		log.Fatalf("can't initialize engine: %v", err)
	}
	return e
}

func registerIntegrations(e *engine.Engine, logger *zap.Logger) {
	// Hue
	hueIntegration, err := hue.NewHueIntegration(logger)
	if err != nil {
		logger.Fatal("failed to init Hue integration", zap.Error(err))
	}
	e.RegisterIntegration("hue", hueIntegration)

	// Halo
	haloIntegration, err := halo.NewHaloIntegration(logger)
	if err != nil {
		logger.Fatal("failed to init Halo integration", zap.Error(err))
	}
	e.RegisterIntegration("halo", haloIntegration)
}

func runEngine(e *engine.Engine, logger *zap.Logger, ctx context.Context) {
	e.RunEventPipelines(ctx)
	e.ProcessEvents(ctx)
}
