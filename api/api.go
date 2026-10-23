package api

import (
	"errors"
	"net/http"

	"go_auth/auth"

	"github.com/gofiber/fiber/v3"
)

// New builds the app. Logins are forwarded to the auth backend at backendURL.
func New(client *http.Client, backendURL string) *fiber.App {
	app := fiber.New()
	app.Post("/auth/login", login(client, backendURL))
	return app
}

func login(client *http.Client, backendURL string) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req auth.LoginRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid JSON body")
		}
		if req.Email == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).SendString("email and password are required")
		}

		result, err := auth.Login(client, backendURL, req.Email, req.Password)
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
		case errors.Is(err, auth.ErrForbidden):
			return c.Status(fiber.StatusForbidden).SendString(err.Error())
		case err != nil:
			return c.Status(fiber.StatusBadGateway).SendString("auth backend error")
		}
		return c.JSON(result)
	}
}
