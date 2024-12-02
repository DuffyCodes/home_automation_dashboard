package main

import (
	"log"
	"os"

	"home_automation_dashboard/mqtt-api/services/api"
	"home_automation_dashboard/shared/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	// MongoDB configuration
	mongoURI := os.Getenv("MONGO_URI")
	databaseName := os.Getenv("MONGO_DB")
	log.Printf("MONGO_DATABASE: %s", databaseName)

	mongoDb := db.ConnectMongoDB(mongoURI, databaseName)
	defer mongoDb.Disconnect()

	// Initialize Gin router
	router := gin.Default()

	// Set up routes
	api.SetupRoutes(router, mongoDb.Client, databaseName)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
