package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

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
	//intialize DB client
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
	natsClient.Publish(viper.GetString(config.NatsSubject), "test message")

	// Initialize the HN client
	hnClient := hnfetcher.NewHNClient(hnfetcher.NewHnFetcherParams{
		Ctx:        context.Background(),
		Logger:     &logger,
		RestClient: *restClient,
		DbClient:   DBClient,
		NatsClient: natsClient,
	})
	// Fetch and store top stories
	go func() {
		if err := hnClient.FetchAndStoreTopStories(ctx, 50); err != nil {
			logger.Error().Err(err).Msg("Fetcher exited with error")
		}
	}()

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

	/*
		app := fiber.New()

		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendString("Hacker News Fetcher is running")
		})

		log.Printf("Starting server on port %s", viper.GetString(config.Port))
		if err := app.Listen(":" + viper.GetString(config.Port)); err != nil {
			log.Fatal(err)
		}*/
}
func setupEnv() {
	viper.SetDefault(config.RestTimeoutInSec, 5)
	viper.SetDefault(config.Port, "8080")
	viper.SetDefault(config.DBUrl, "postgres://hnuser:hnpass@localhost:5432/hackernews?sslmode=disable")
	viper.SetDefault(config.NatsUrl, "nats://localhost:4222")
	viper.SetDefault(config.NatsSubject, "hnfetcher.topstories")

}
