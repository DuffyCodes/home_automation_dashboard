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

type MqttMessage struct {
	Topic     string                 `bson:"topic"`
	Device    string                 `bson:"device"`
	Value     interface{}            `bson:"value"`
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
				"state":       payload,
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

func parsePayload(payload []byte) interface{} {
	var jsonPayload map[string]interface{}
	if err := json.Unmarshal(payload, &jsonPayload); err == nil {
		return jsonPayload
	}

	if numericPayload, err := strconv.ParseFloat(string(payload), 64); err == nil {
		return numericPayload
	}

	return string(payload)
}

func extractDeviceFromTopic(topic string) string {
	parts := splitTopic(topic)
	if len(parts) > 1 {
		return strings.Join(parts[1:len(parts)-1], "_")
	}
	return "unknown_device"
}

func extractRoomFromTopic(topic string) string {
	parts := splitTopic(topic)
	if len(parts) > 1 {
		return parts[1]
	}
	return "unknown_room"
}

func extractSensorTypeFromTopic(topic string) string {
	parts := splitTopic(topic)
	if len(parts) > 2 {
		return parts[2]
	}
	return "unknown_sensor_type"
}

func splitTopic(topic string) []string {
	return strings.Split(topic, "/")
}
