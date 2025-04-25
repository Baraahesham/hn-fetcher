package db

import (
	"errors"

	Err "github.com/Baraahesham/hn-fetcher/internal/errors"
	"github.com/Baraahesham/hn-fetcher/internal/models"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBClient struct {
	Gorm   *gorm.DB
	logger *zerolog.Logger
}
type NewDBClientParams struct {
	Logger *zerolog.Logger
	DbUrl  string
}

func NewClient(params NewDBClientParams) (*DBClient, error) {
	// Initialize the logger
	logger := params.Logger.With().Str("component", "DBClient").Logger()

	logger.Info().Str("db_url", params.DbUrl).Msg("Connecting to PostgreSQL")

	gormDB, err := gorm.Open(postgres.Open(params.DbUrl), &gorm.Config{})

	if err != nil {
		logger.Error().Err(err).Str("db_url", params.DbUrl).Msg("Failed to connect to PostgreSQL")
		return nil, err
	}

	err = gormDB.AutoMigrate(&models.StoryDbModel{})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to auto-migrate Story model")
		return nil, err
	}

	logger.Info().Str("db_url", params.DbUrl).Msg("Successfully Connected to PostgreSQL")
	return &DBClient{
		Gorm:   gormDB,
		logger: &logger,
	}, nil
}

// InsertStory inserts a story into the database
func (client *DBClient) InsertStory(story models.StoryDbModel) error {
	//var existing models.StoryDbModel
	res := client.Gorm.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "hn_id"}},
		DoNothing: true,
	}).Where("hn_id = ?", story.HnID).First(&story)

	if res.Error == nil {
		client.logger.Warn().Int64("story_id", story.HnID).Msg("Story already exists in database, skipping insert")
		return Err.ErrStoryAlreadyExists
	}

	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		client.logger.Error().Err(res.Error).Int64("story_id", story.HnID).Msg("Failed to check if story exists")
		return res.Error
	}

	// Story doesn't exist → insert
	if err := client.Gorm.Create(&story).Error; err != nil {
		client.logger.Error().Err(err).Int64("story_id", story.HnID).Msg("Failed to insert story into database")
		return err
	}

	client.logger.Info().Err(res.Error).Int64("story_id", story.HnID).Msg("Story inserted into database")
	return nil
}
func (client *DBClient) GetBrandMentionStats() ([]models.BrandStats, error) {
	var stats []models.BrandStats
	err := client.Gorm.Raw(`
		SELECT brand, COUNT(*) AS mentions 
		FROM brand_mentions 
		GROUP BY brand 
		ORDER BY mentions DESC
	`).Scan(&stats).Error
	return stats, err
}

func (client *DBClient) GetStoriesByBrand(brand string) ([]models.StoryDbModel, error) {
	var stories []models.StoryDbModel
	client.logger.Info().Str("brand", brand).Msg("Fetching stories by brand")
	err := client.Gorm.Raw(`
		SELECT s.* 
		FROM stories s 
		JOIN brand_mentions bm ON s.hn_id = bm.hn_id 
		WHERE bm.brand = ? 
		ORDER BY s.time DESC
	`, brand).Scan(&stories).Error
	client.logger.Info().Any("story", stories).Msg("Fetched stories by brand")
	return stories, err
}

func (client *DBClient) Close() error {
	sqlDB, err := client.Gorm.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
