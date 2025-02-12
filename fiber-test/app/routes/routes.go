package routes

import (
	"context"
	"fiber-test/app/database"
	"fiber-test/app/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")
	api.Get("/", homeHandler)
	api.Get("/hello", helloHandler)
	app.Get("/health", healthCheck)

	api.Post("/db/:id", saveToDB)
	api.Get("/db/:id", getFromDB)
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

func saveToDB(c *fiber.Ctx) error {
	db := database.GetDB()
	collection := db.Collection("data")

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID is required"})
	}

	// 요청 본문을 JSON으로 받아옴
	var body models.DataModel
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	body.ID = id
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": body},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save data"})
	}

	return c.JSON(fiber.Map{"message": "Data saved successfully", "data": body})
}

// MongoDB 데이터 조회
func getFromDB(c *fiber.Ctx) error {
	db := database.GetDB()
	collection := db.Collection("data")

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result models.DataModel
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Data not found"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve data"})
	}

	return c.JSON(result)
}
