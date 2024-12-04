package main

import (
	"log"
	"os"
	"time"

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

	mongoDb := db.ConnectMongoDB(mongoURI, databaseName)
	defer mongoDb.Disconnect()

	handler := api.NewHandler(mongoDb.Client, databaseName)

	// Periodic updates for switch metrics
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
		defer ticker.Stop()

		for range ticker.C {
			handler.UpdateSwitchMetrics()
			handler.UpdateTotalSwitchOnDuration()
		}
	}()

	// Initialize Gin router
	router := gin.Default()

	// metrics := prometheus.NewRegistry()
	// prometheus.MustRegister(prometheus.NewBuildInfoCollector())
	// router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(metrics, promhttp.HandlerOpts{})))

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
