package api

import (
	"github.com/Baraahesham/hn-fetcher/internal/db"

	"github.com/rs/zerolog"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	dbClient *db.DBClient
	logger   *zerolog.Logger
}
type NewHandlerParams struct {
	Logger   *zerolog.Logger
	DbClient *db.DBClient
}

func NewHandler(params NewHandlerParams) *Handler {
	logger := params.Logger.With().Str("component", "Handler").Logger()
	return &Handler{
		dbClient: params.DbClient,
		logger:   &logger,
	}
}

func (client *Handler) Health(c *fiber.Ctx) error {
	return c.SendString("Hacker News Fetcher is running")
}

func (client *Handler) GetBrandStats(c *fiber.Ctx) error {
	stats, err := client.dbClient.GetBrandMentionStats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch stats"})
	}
	return c.JSON(stats)
}

func (client *Handler) GetStoriesByBrand(c *fiber.Ctx) error {
	brand := c.Params("brand")
	stories, err := client.dbClient.GetStoriesByBrand(brand)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch stories"})
	}
	return c.JSON(stories)
}
