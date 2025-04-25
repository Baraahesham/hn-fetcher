package config

const (
	Port                    = "PORT"
	DBUrl                   = "DB_URL"
	NatsUrl                 = "NATS_URL"
	RestTimeoutInSec        = "REST_TIMEOUT_IN_SEC"
	NatsSubject             = "hnfetcher.topstories"
	MaxWorkers              = "MAX_WORKERS"
	MaxCapacity             = "MAX_CAPACITY"
	TopStoriesFetchingLimit = 50
	TopStoriesUrlPath       = "topstories.json"
	FetchStoryByIDUrlPath   = "item/%d.json"
	HNBaseUrl               = "HN_BASE_URL"
)
