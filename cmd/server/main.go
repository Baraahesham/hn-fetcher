package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/Baraahesham/hn-fetcher/internal/api"
	"github.com/Baraahesham/hn-fetcher/internal/config"
	"github.com/Baraahesham/hn-fetcher/internal/db"
	hnfetcher "github.com/Baraahesham/hn-fetcher/internal/hn_fetcher"
	"github.com/Baraahesham/hn-fetcher/internal/nats"
	"github.com/Baraahesham/hn-fetcher/internal/rest"
	"github.com/spf13/viper"
)

func main() {

	setupEnv()
	// Initialize the logger
	logger := zerolog.New(os.Stdout).With().Logger()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	//Initialize DB client
	DBClient, err := db.NewClient(db.NewDBClientParams{
		Logger: &logger,
		DbUrl:  viper.GetString(config.DBUrl),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize DB client")
		return
	}
	// Initialize the rest client
	restClient := rest.NewRestClient(rest.NewRestClientParams{
		Logger:  &logger,
		Timeout: viper.GetDuration(config.RestTimeoutInSec) * time.Second,
	})
	natsClient, err := nats.New(nats.NewNatsClientParams{
		Logger:  &logger,
		NatsUrl: viper.GetString(config.NatsUrl),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to NATS")
		return
	}
	logger.Info().Msg("Connected to NATS")

	// Initialize the HN client
	hnClient := hnfetcher.NewHNClient(hnfetcher.NewHnFetcherParams{
		Ctx:        context.Background(),
		Logger:     &logger,
		RestClient: *restClient,
		DbClient:   DBClient,
		NatsClient: natsClient,
		FetchLimit: 50,
	})
	// Initialize the API server
	app := fiber.New()
	// Initialize the API handler
	apiHandler := api.NewHandler(api.NewHandlerParams{
		Logger:   &logger,
		DbClient: DBClient,
	})
	api.RegisterRoutes(app, apiHandler)

	log.Printf("Starting server on port %s", viper.GetString(config.Port))
	// Fetch and store top stories
	go func() {
		if err := app.Listen(":" + viper.GetString(config.Port)); err != nil {
			logger.Error().Err(err).Msg("Failed to start server")
		}
	}()
	// Start the hnClient listener
	hnClient.Init()

	logger.Info().Msg("hn-fetcher is running. Press Ctrl+C to exit...")

	// Wait for interrupt (Ctrl+C)
	<-ctx.Done()
	logger.Warn().Msg("Interrupt received. Cleaning up...")

	// Gracefully shutdown clients
	natsClient.Close()

	if err := DBClient.Close(); err != nil {
		logger.Error().Err(err).Msg("Error closing DB")
	}

	logger.Info().Msg("Shutdown complete")
	app.Shutdown()
}
func setupEnv() {
	viper.SetDefault(config.RestTimeoutInSec, 5)
	viper.SetDefault(config.Port, "8080")
	viper.SetDefault(config.DBUrl, "postgres://hnuser:hnpass@localhost:5432/hackernews?sslmode=disable")
	viper.SetDefault(config.NatsUrl, "nats://localhost:4222")
	viper.SetDefault(config.NatsSubject, "hnfetcher.topstories")
	viper.SetDefault(config.MaxWorkers, 10)
	viper.SetDefault(config.MaxCapacity, 100)

}
