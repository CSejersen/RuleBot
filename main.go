package main

import (
	"context"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"home_automation_server/engine"
	hueexecutor "home_automation_server/integrations/hue/actionexecutor"
	hueclient "home_automation_server/integrations/hue/apiclient"
	hueeventsource "home_automation_server/integrations/hue/eventsource"
	huetranslator "home_automation_server/integrations/hue/translator"
	"home_automation_server/integrations/hue/translator/events"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// TODO: implement proper engine bootstrapping
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting home automation engine")

	e := engine.New()
	if err := e.Init(ctx); err != nil {
		panic(err)
	}

	hueIP := os.Getenv("HUE_IP")
	hueAppKey := os.Getenv("HUE_APP_KEY")
	hueLogger := logger.With(zap.String("integration", "hue"))
	hueClient := hueclient.New(hueIP, hueAppKey, hueLogger)
	hueExecutor := hueexecutor.New(hueClient, hueLogger)
	hueEventSource := hueeventsource.New(hueIP, hueAppKey, hueLogger)

	hueTranslator, err := huetranslator.New(hueClient, hueLogger)
	if err != nil {
		panic(err)
	}

	hueTranslator.RegisterEvent("light", func() huetranslator.Event {
		return &events.LightUpdate{}
	})

	hueIntegration := engine.Integration{
		EventSource:    hueEventSource,
		Translator:     hueTranslator,
		ActionExecutor: hueExecutor,
	}

	e.RegisterIntegration("hue", hueIntegration)

	e.RunPipelines(ctx)

	go func(ctx context.Context) {
		if err := e.ProcessEvents(ctx); err != nil {
			logger.Error("error processing events", zap.Error(err))
		}
	}(ctx)

	logger.Info("Engine bootstrap succeeded")
	<-ctx.Done()
	logger.Info("shutting down")
}
