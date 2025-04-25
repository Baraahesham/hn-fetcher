package hnfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Baraahesham/hn-fetcher/internal/config"
	"github.com/Baraahesham/hn-fetcher/internal/db"
	Err "github.com/Baraahesham/hn-fetcher/internal/errors"
	"github.com/Baraahesham/hn-fetcher/internal/models"
	"github.com/Baraahesham/hn-fetcher/internal/nats"
	"github.com/Baraahesham/hn-fetcher/internal/rest"
	"github.com/alitto/pond"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type HnFetcherClient struct {
	ctx        context.Context
	restClient rest.RestClient
	logger     *zerolog.Logger
	dbClient   *db.DBClient
	natsClient nats.NatsClient
	fetchLimit int
	workerPool *pond.WorkerPool
}
type NewHnFetcherParams struct {
	Ctx        context.Context
	Logger     *zerolog.Logger
	RestClient rest.RestClient
	DbClient   *db.DBClient
	NatsClient *nats.NatsClient
	FetchLimit int
}

func NewHNClient(params NewHnFetcherParams) *HnFetcherClient {
	pool := pond.New(
		viper.GetInt(config.MaxWorkers),
		viper.GetInt(config.MaxCapacity),
		pond.Context(params.Ctx),
		pond.Strategy(pond.Balanced()),
	)
	client := &HnFetcherClient{
		restClient: params.RestClient,
		logger:     params.Logger,
		dbClient:   params.DbClient,
		ctx:        params.Ctx,
		natsClient: *params.NatsClient,
		fetchLimit: params.FetchLimit,
		workerPool: pool,
	}
	return client
}
func (client *HnFetcherClient) Init() {
	go client.FetchAndStoreTopStories(client.ctx)
}
func (client *HnFetcherClient) FetchAndStoreTopStories(ctx context.Context) error {
	// 1. Get top story IDs
	topIDs, err := client.FetchTopStoriesIDs(ctx, client.fetchLimit)
	if err != nil {
		client.logger.Error().Err(err).Msg("Failed to fetch top story IDs")
		return err
	}

	// 2. For each ID, fetch and store story
	for _, id := range topIDs {
		client.workerPool.Submit(func() {
			err := client.FetchAndPublishStory(client.ctx, id)
			if err != nil {
				client.logger.Error().Err(err).Int64("story_id", id).Msg("Failed to fetch and publish story")
			}
		})

	}

	return nil
}
func (client *HnFetcherClient) FetchAndPublishStory(ctx context.Context, id int64) error {
	story, err := client.FetchStoryByID(ctx, id)
	if err != nil {
		client.logger.Error().Err(err).Int64("story_id", id).Msg("Failed to fetch story by ID")
		return err
	}
	err = client.dbClient.InsertStory(*story)
	if err != nil {
		var msg string
		if err == Err.ErrStoryAlreadyExists {
			msg = "Story already exists in the database, skipping publish"
		} else {
			msg = "Failed to insert story into database, skipping publish"
		}
		client.logger.Error().Err(err).Int64("story_id", id).Msg(msg)
		return err
	}
	natsEvent := models.StoryEvent{
		HnID:  story.HnID,
		Title: story.Title,
	}
	err = client.natsClient.Publish(config.NatsSubject, natsEvent)
	if err != nil {
		client.logger.Error().Err(err).Int64("story_id", id).Msg("Failed to publish story to NATS")
		return err
	}
	client.logger.Info().Int64("story_id", id).Msg("Story successfully published to NATS")
	return nil
}

func (client *HnFetcherClient) FetchTopStoriesIDs(ctx context.Context, limit int) ([]int64, error) {
	url := viper.GetString(config.HNBaseUrl) + config.TopStoriesUrlPath
	params := map[string]string{"print": "pretty"}

	body, _, err := client.restClient.ExecuteHttpRequest(ctx, "GET", url, params, nil, nil)
	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := json.Unmarshal(body, &ids); err != nil {
		client.logger.Error().Err(err).Msg("Failed to unmarshal top story IDs")
		return nil, err
	}

	if limit > 0 && limit < len(ids) {
		return ids[:limit], nil
	}

	return ids, nil
}

func (client *HnFetcherClient) FetchStoryByID(ctx context.Context, id int64) (*models.StoryDbModel, error) {
	url := viper.GetString(config.HNBaseUrl) + fmt.Sprintf(config.FetchStoryByIDUrlPath, id)

	body, _, err := client.restClient.ExecuteHttpRequest(ctx, "GET", url, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	story := models.Story{}
	err = json.Unmarshal(body, &story)
	if err != nil {
		client.logger.Error().Err(err).Msg("Failed to unmarshal story")
		return nil, err
	}
	mappedStory := client.MaptoStoryDbModel(story)

	return &mappedStory, nil
}

func (client *HnFetcherClient) MaptoStoryDbModel(story models.Story) models.StoryDbModel {
	return models.StoryDbModel{
		HnID:  story.HnID,
		Title: story.Title,
		Time:  time.Unix(story.Time, 0),
		URL:   story.URL,
	}
}
