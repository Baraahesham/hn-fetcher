package main

import (
	"log"

	//"github.com/Baraahesham/hn-fetcher/internal/config"
	"hn-fetcher/internal/config/config.go"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hacker News Fetcher is running")
	})

	log.Printf("Starting server on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
