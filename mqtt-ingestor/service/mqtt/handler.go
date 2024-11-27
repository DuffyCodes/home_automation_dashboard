package mqtt

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"
)

// MqttMessage represents the MongoDB document structure.
type MqttMessage struct {
	Topic     string                 `bson:"topic"`
	Device    string                 `bson:"device"`
	Value     interface{}            `bson:"value"` // Can handle different data types
	Timestamp time.Time              `bson:"timestamp"`
	Tags      map[string]interface{} `bson:"tags"`
}

// MessageHandler processes incoming MQTT messages and stores them in MongoDB.
func MessageHandler(client *mongo.Client, databaseName, collectionName string) func(mqtt.Client, mqtt.Message) {
	return func(mqttClient mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message: Topic=%s Payload=%s", msg.Topic(), string(msg.Payload()))

		// Parse the payload into an appropriate type
		payload := parsePayload(msg.Payload())

		// Construct the MongoDB document
		mqttMessage := MqttMessage{
			Topic:     msg.Topic(),
			Device:    extractDeviceFromTopic(msg.Topic()),
			Value:     payload,
			Timestamp: time.Now(),
			Tags: map[string]interface{}{
				"room":        extractRoomFromTopic(msg.Topic()),
				"sensor_type": extractSensorTypeFromTopic(msg.Topic()),
			},
		}

		// Insert the document into MongoDB
		collection := client.Database(databaseName).Collection(collectionName)
		_, err := collection.InsertOne(context.Background(), mqttMessage)
		if err != nil {
			log.Printf("Failed to insert message into MongoDB: %v", err)
		} else {
			log.Printf("Successfully stored message: %+v", mqttMessage)
		}
	}
}

// parsePayload attempts to parse the MQTT payload into a proper type (string, float64, or structured JSON).
func parsePayload(payload []byte) interface{} {
	// Try to parse as JSON
	var jsonPayload map[string]interface{}
	if err := json.Unmarshal(payload, &jsonPayload); err == nil {
		return jsonPayload
	}

	// Try to parse as a float
	if numericPayload, err := strconv.ParseFloat(string(payload), 64); err == nil {
		return numericPayload
	}

	// Default to returning the payload as a string
	return string(payload)
}

// extractDeviceFromTopic extracts the device name from the MQTT topic.
func extractDeviceFromTopic(topic string) string {
	// Example: "home/living_room/roku_on_or_off" -> "living_room_roku"
	parts := splitTopic(topic)
	if len(parts) > 1 {
		return strings.Join(parts[1:len(parts)-1], "_")
	}
	return "unknown_device"
}

// extractRoomFromTopic extracts the room name from the MQTT topic.
func extractRoomFromTopic(topic string) string {
	// Example: "home/living_room/roku_on_or_off" -> "living_room"
	parts := splitTopic(topic)
	if len(parts) > 1 {
		return parts[1]
	}
	return "unknown_room"
}

// extractSensorTypeFromTopic extracts the sensor type from the MQTT topic.
func extractSensorTypeFromTopic(topic string) string {
	// Example: "home/kitchen/temperature" -> "temperature"
	parts := splitTopic(topic)
	if len(parts) > 2 {
		return parts[2]
	}
	return "unknown_sensor_type"
}

// splitTopic splits the topic into parts for easier parsing.
func splitTopic(topic string) []string {
	return strings.Split(topic, "/")
}
