package main

import (
	"github.com/joho/godotenv"
	"log"
)

func main() {
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
	registerIntegrations(e, logger)
	runEngine(e, logger, ctx)

	logger.Info("Engine bootstrap succeeded")
	<-ctx.Done()
	logger.Info("Shutting down")
	e.Shutdown()
}
