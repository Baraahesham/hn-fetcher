package hnfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Baraahesham/hn-fetcher/internal/config"
	"github.com/Baraahesham/hn-fetcher/internal/db"
	"github.com/Baraahesham/hn-fetcher/internal/models"
	"github.com/Baraahesham/hn-fetcher/internal/nats"
	"github.com/Baraahesham/hn-fetcher/internal/rest"

	"github.com/rs/zerolog"
)

type HnFetcherClient struct {
	ctx        context.Context
	restClient rest.RestClient
	logger     *zerolog.Logger
	dbClient   *db.DBClient
	natsClient nats.NatsClient
}
type NewHnFetcherParams struct {
	Ctx        context.Context
	Logger     *zerolog.Logger
	RestClient rest.RestClient
	DbClient   *db.DBClient
	NatsClient *nats.NatsClient
}

func NewHNClient(params NewHnFetcherParams) *HnFetcherClient {
	client := &HnFetcherClient{
		restClient: params.RestClient,
		logger:     params.Logger,
		dbClient:   params.DbClient,
		ctx:        params.Ctx,
		natsClient: *params.NatsClient,
	}
	return client
}

func (client *HnFetcherClient) FetchAndStoreTopStories(ctx context.Context, limit int) error {
	// 1. Get top story IDs
	topIDs, err := client.FetchTopStoryIDs(ctx, limit)
	if err != nil {
		client.logger.Error().Err(err).Msg("Failed to fetch top story IDs")
		return err
	}

	// 2. For each ID, fetch and store story
	for _, id := range topIDs {
		story, err := client.FetchStoryByID(ctx, id)
		if err != nil {
			client.logger.Error().Err(err).Int64("story_id", id).Msg("Failed to fetch story by ID")
			continue
		}
		/*err = client.dbClient.InsertStory(*story)
		if err != nil {
			client.logger.Error().Err(err).Int64("story_id", id).Msg("Failed to insert story into database, skipping publishing")
			continue
		}*/
		natsEvent := models.StoryEvent{
			ID:    story.HnID,
			Title: story.Title,
		}
		err = client.natsClient.Publish(config.NatsSubject, natsEvent)
		if err != nil {
			client.logger.Error().Err(err).Int64("story_id", id).Msg("Failed to publish story to NATS")
			continue
		}
	}

	return nil
}

func (client *HnFetcherClient) FetchTopStoryIDs(ctx context.Context, limit int) ([]int64, error) {
	url := "https://hacker-news.firebaseio.com/v0/topstories.json"
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
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)

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
