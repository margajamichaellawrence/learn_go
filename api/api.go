package api

import (
	"go_auth/auth"

	"github.com/gofiber/fiber/v3"
)

func New() *fiber.App {
	app := fiber.New()
	app.Get("/", hello)
	return app
}

func hello(c fiber.Ctx) error {
	return c.SendString(auth.Hello())
}
