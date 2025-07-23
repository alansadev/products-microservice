package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"products/database"
	_ "products/docs"
	"products/handlers"
	"products/middleware"
)

// @title API de Produtos - Sabor da Rondonia
// @version 1.0
// @description Microserviço responsável pelo gerenciamento de produtos.
// @host localhost:3000
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	database.Connect()

	app := fiber.New()

	api := app.Group("/api")

	app.Get("/swagger/*", swagger.HandlerDefault)

	productGroup := api.Group("/products", middleware.AuthRequired())

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
