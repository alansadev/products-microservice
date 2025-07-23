package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"log"
	"os"
	"products/database"
	_ "products/docs"
	"products/handlers"
	"products/middleware"
)

// @title Product API - Sabor da Rondônia
// @version 1.0
// @description Microservice responsible for product management.
// @BasePath /api
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	database.Connect()

	loadedApiKey := os.Getenv("API_SECRET_KEY")
	log.Printf("DEBUG: API_SECRET_KEY carregada: [%s]", loadedApiKey)
	if loadedApiKey == "" {
		log.Println("AVISO CRÍTICO: API_SECRET_KEY está vazia. A autenticação irá falhar.")
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Permite todas as origens
		AllowHeaders: "Origin, Content-Type, Accept, X-API-KEY",
	}))

	app.Use(logger.New())

	api := app.Group("/api")
	app.Get("/swagger/*", swagger.HandlerDefault)

	productGroup := api.Group("/products", middleware.AuthMiddleware())

	productGroup.Post("/", handlers.CreateProduct)
	productGroup.Get("/", handlers.GetProducts)
	productGroup.Get("/:id", handlers.GetProductByID)
	productGroup.Patch("/:id", handlers.PatchProduct)
	productGroup.Delete("/:id", handlers.DeleteProduct)
	productGroup.Post("/:id/upload", handlers.UploadProductImage)
	productGroup.Post("/batch", handlers.GetProductsByIDs)
	productGroup.Post("/:id/stock", handlers.UpdateStock)

	app.Listen(":3000")
}
