package main

import (
	"os"

	"home_automation_dashboard/mqtt-ingestor/service/mqtt"
	"home_automation_dashboard/shared/db"
)

func main() {
	// MongoDB configuration
	mongoURI := os.Getenv("MONGO_URI")
	databaseName := os.Getenv("MONGO_DB")
	collectionName := os.Getenv("MONGO_COLLECTION")

	// MQTT configuration
	mqttBroker := os.Getenv("MQTT_BROKER")
	clientID := os.Getenv("MQTT_CLIENT_ID")
	mqttUsername := os.Getenv("MQTT_UN")
	mqttPassword := os.Getenv("MQTT_PW")
	topics := []string{"home/#"}

	// Connect to MongoDB
	mongoDB := db.ConnectMongoDB(mongoURI, databaseName)
	defer mongoDB.Disconnect()

	// Initialize MQTT message handler with MongoDB integration
	messageHandler := mqtt.MessageHandler(mongoDB.Client, databaseName, collectionName)

	// Connect to MQTT broker
	mqttClient := mqtt.ConnectClient(mqttBroker, clientID, mqttUsername, mqttPassword, messageHandler)

	// Subscribe to topics
	mqtt.SubscribeTopics(mqttClient, topics)

	// Keep the application running
	select {}
}
