package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Set up MQTT
	mqttBroker := "tcp://192.168.1.242:1883" // Replace with your broker's IP
	clientID := "go-mqtt-subscriber"
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	mqttClient := connectMQTT(mqttBroker, clientID, os.Getenv("MQTT_UN"), os.Getenv("MQTT_PW"))

	// Topics to subscribe to
	topics := []string{
		"home/#", // Subscribe to all topics under "home/"
	}
	subscribeToTopics(mqttClient, topics)

	// Set up HTTP server
	r := gin.Default()

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API is running!",
		})
	})

	r.POST("/publish", func(c *gin.Context) {
		var payload struct {
			Topic   string `json:"topic"`
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		token := mqttClient.Publish(payload.Topic, 1, false, payload.Message)
		token.Wait()
		c.JSON(200, gin.H{
			"message": "Published successfully",
		})
	})

	r.Run(":8080") // Start the HTTP server on port 8080
}
