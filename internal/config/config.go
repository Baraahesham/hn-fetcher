package config

/*type Config struct {
	Port             string
	DBUrl            string
	NatsUrl          string
	RestTimeoutInSec int
}*/

/*
	func Load() Config {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found, using environment variables")
		}

		return Config{
			Port:             getEnv("PORT", "8080"),
			DBUrl:            getEnv("DB_URL", ""),
			NatsUrl:          getEnv("NATS_URL", ""),
			RestTimeoutInSec: getEnvAsInt("REST_TIMEOUT_IN_SEC", "5"),
		}
	}

	func getEnv(key, fallback string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		return fallback
	}
*/
const (
	Port             = "PORT"
	DBUrl            = "DB_URL"
	NatsUrl          = "NATS_URL"
	RestTimeoutInSec = "REST_TIMEOUT_IN_SEC"
	NatsSubject      = "hnfetcher.topstories"
)
