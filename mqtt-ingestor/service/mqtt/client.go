package mqtt

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ConnectClient connects to the MQTT broker and returns the client instance.
func ConnectClient(broker, clientID, username, password string, messageHandler mqtt.MessageHandler) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(messageHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}
	fmt.Println("Connected to MQTT broker!")
	return client
}

// SubscribeTopics subscribes the MQTT client to the specified topics.
func SubscribeTopics(client mqtt.Client, topics []string) {
	for _, topic := range topics {
		if token := client.Subscribe(topic, 1, nil); token.Wait() && token.Error() != nil {
			log.Fatalf("Failed to subscribe to topic %s: %v", topic, token.Error())
		}
		fmt.Printf("Subscribed to topic: %s\n", topic)
	}
}
