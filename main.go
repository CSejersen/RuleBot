package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"home_automation_server/api"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Println("Starting engine")
	os.Stdout.Sync()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx, stop := setupContext()
	defer stop()
	logger := setupLogger()
	defer logger.Sync()

	logger.Info("Bootstrapping engine")

	e, err := setupEngine(ctx, logger, 5)
	if err != nil {
		log.Fatal(err)
	}

	registerIntegrationDescriptors(e)
	if err := LoadIntegrations(ctx, e); err != nil {
		log.Fatal(err)
	}

	// Setup API + WS Server
	eventCh := e.ProcessedEventBus.Subscribe()
	apiServer := api.NewServer(ctx, e, logger, eventCh)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	apiServer.Start(":" + port)

	runEngine(e, logger, ctx)

	logger.Info("Engine bootstrap succeeded")

	<-ctx.Done()
	logger.Info("Shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", zap.Error(err))
	}

	e.Shutdown()
}
