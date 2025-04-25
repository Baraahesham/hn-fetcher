package api

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, handler *Handler) {

	app.Get("/", handler.Health)
	app.Get("/brands/stats", handler.GetBrandStats)
	app.Get("/brands/:brand/stories", handler.GetStoriesByBrand)
}
