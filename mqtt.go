package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler processes incoming MQTT messages.
var MessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
	// Process the message or store it as needed
}

func connectMQTT(broker, clientID string, username string, password string) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(MessageHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}
	fmt.Println("Connected to MQTT broker!")
	return client
}

func subscribeToTopics(client mqtt.Client, topics []string) {
	for _, topic := range topics {
		if token := client.Subscribe(topic, 1, nil); token.Wait() && token.Error() != nil {
			log.Fatalf("Failed to subscribe to topic %s: %v", topic, token.Error())
		}
		fmt.Printf("Subscribed to topic: %s\n", topic)
	}
}
