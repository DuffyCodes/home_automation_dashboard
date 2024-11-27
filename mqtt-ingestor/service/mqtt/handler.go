package mqtt

import (
	"context"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"
)

// MessageHandler processes incoming MQTT messages.
func MessageHandler(client mongo.Client, databaseName string, collectionName string) func(mqtt.Client, mqtt.Message) {
	return func(mqttClient mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: Topic=%s Payload=%s\n", msg.Topic(), string(msg.Payload()))

		// Insert the event into MongoDB
		collection := client.Database(databaseName).Collection(collectionName)
		event := map[string]interface{}{
			"topic":   msg.Topic(),
			"payload": string(msg.Payload()),
			"time":    time.Now(),
		}

		_, err := collection.InsertOne(context.Background(), event)
		if err != nil {
			log.Printf("Failed to insert event: %v", err)
		} else {
			fmt.Printf("Event stored: %s -> %s\n", msg.Topic(), msg.Payload())
		}
	}
}
