package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRoutes initializes all API routes
func SetupRoutes(router *gin.Engine, client *mongo.Client, databaseName string) {
	handler := NewHandler(client, databaseName)

	api := router.Group("/api/v1")
	{
		temperature := api.Group("/temperature")
		temperature.GET("/average", handler.GetAverageTemperature)
		temperature.GET("/max", handler.GetMaxTemperature)
		temperature.GET("/min", handler.GetMinTemperature)

	}
}
