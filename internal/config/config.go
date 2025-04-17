package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    string
	DBUrl   string
	NatsUrl string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return Config{
		Port:    getEnv("PORT", "8080"),
		DBUrl:   getEnv("DB_URL", ""),
		NatsUrl: getEnv("NATS_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
