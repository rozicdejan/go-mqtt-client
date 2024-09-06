package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// Define the MQTT broker options
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("go_mqtt_client")

	// Create an MQTT client
	client := mqtt.NewClient(opts)

	// Connect to the MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Failed to connect: %v\n", token.Error())
		return
	}
	fmt.Println("Connected to MQTT broker")

	// Subscribe to a topic
	topic := "test/topic"
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message on topic %s: %s\n", msg.Topic(), string(msg.Payload()))
	})

	// Publish a message to the topic
	token := client.Publish(topic, 0, false, "Hello from Go!")
	token.Wait()

	// Wait for a while to receive messages
	time.Sleep(2 * time.Second)

	// Disconnect from the broker
	client.Disconnect(250)
	fmt.Println("Disconnected from MQTT broker")
}
