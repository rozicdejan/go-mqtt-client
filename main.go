package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	// Subscribe to the topic "test/topic"
	topic := "test/topic"
	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message on topic %s: %s\n", msg.Topic(), string(msg.Payload()))
	}); token.Wait() && token.Error() != nil {
		fmt.Printf("Failed to subscribe: %v\n", token.Error())
		return
	}
	fmt.Println("Subscribed to topic:", topic)

	// Set up a channel to listen for termination signals (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a signal to terminate
	<-sigChan
	fmt.Println("\nReceived termination signal, cleaning up...")

	// Unsubscribe from the topic and disconnect the client
	if token := client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		fmt.Printf("Failed to unsubscribe: %v\n", token.Error())
	}
	client.Disconnect(250)
	fmt.Println("Disconnected from MQTT broker")

	fmt.Println("Application closed")
}
