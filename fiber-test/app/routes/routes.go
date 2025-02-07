package routes

import "github.com/gofiber/fiber/v2"

func Setup(app *fiber.App) {
	api := app.Group("/api")
	api.Get("/", homeHandler)
	api.Get("/hello", helloHandler)
	app.Get("/health", healthCheck)
}

func healthCheck(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func homeHandler(c *fiber.Ctx) error {
	return c.SendString("Welcome to the home page!")
}

func helloHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Hello, Fiber!",
	})
}
