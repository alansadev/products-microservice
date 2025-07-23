package database

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"products/models"
)

var DB *gorm.DB

func Connect() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables if available")
	}

	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database! \n", err)
		os.Exit(2)
	}

	fmt.Println("Successfully connected to database!")

	fmt.Println("Running Migrations...")
	err = DB.AutoMigrate(&models.Product{})
	if err != nil {
		log.Fatal("Failed to run migrations! \n", err)
	}
}
