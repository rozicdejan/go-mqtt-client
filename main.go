package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

// Load .env file if it exists in the init function
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Using default values or environment variables.")
	} else {
		log.Println(".env file loaded successfully.")
	}
	log.Println("MQTT_BROKER:", os.Getenv("MQTT_BROKER"))
	log.Println("MQTT_CLIENT_ID:", os.Getenv("MQTT_CLIENT_ID"))
	log.Println("MQTT_TOPIC:", os.Getenv("MQTT_TOPIC"))
}

// Utility function to get environment variables with a fallback default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Handle subscription messages
func messageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
}

// Wait for a token to complete or a timeout
func waitWithTimeout(token mqtt.Token, timeout time.Duration) error {
	done := make(chan struct{})

	// Wait in a separate goroutine
	go func() {
		token.Wait()
		close(done)
	}()

	// Use a select statement to either wait for the token or timeout
	select {
	case <-done:
		return token.Error()
	case <-time.After(timeout):
		return fmt.Errorf("operation timed out after %s", timeout)
	}
}

// Setup MQTT client and return the client object
func connectToMQTT(broker, clientID string, timeout time.Duration) mqtt.Client {
	// Define the MQTT broker options
	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)

	// Create an MQTT client
	client := mqtt.NewClient(opts)

	// Connect to the MQTT broker with a timeout
	token := client.Connect()
	if err := waitWithTimeout(token, timeout); err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	log.Println("Connected to MQTT broker")
	return client
}

// Subscribe to the MQTT topic with a timeout
func subscribeToTopic(client mqtt.Client, topic string, timeout time.Duration) {
	// Subscribe to the topic and handle incoming messages
	token := client.Subscribe(topic, 0, messageHandler)
	if err := waitWithTimeout(token, timeout); err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}
	log.Printf("Subscribed to topic: %s", topic)
}

// Gracefully handle system signals and disconnect the client
func handleShutdown(client mqtt.Client, topic string, timeout time.Duration) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	sig := <-sigChan
	log.Printf("Received signal: %v. Cleaning up...", sig)

	// Unsubscribe from the topic
	token := client.Unsubscribe(topic)
	if err := waitWithTimeout(token, timeout); err != nil {
		log.Printf("Failed to unsubscribe from topic: %v", err)
	}

	// Disconnect the client
	client.Disconnect(250) // 250 ms to complete any pending operations
	log.Println("Disconnected from MQTT broker")

	log.Println("Application exited gracefully")
}

func main() {
	// Load environment variables after .env has been loaded
	mqttBroker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	clientID := getEnv("MQTT_CLIENT_ID", "go_mqtt_client")
	topic := getEnv("MQTT_TOPIC", "test/topic")
	timeout := 5 * time.Second

	// Log to see if the environment variables are correctly loaded
	log.Println("Using MQTT_BROKER:", mqttBroker)
	log.Println("Using MQTT_CLIENT_ID:", clientID)
	log.Println("Using MQTT_TOPIC:", topic)

	// Connect to the MQTT broker
	client := connectToMQTT(mqttBroker, clientID, timeout)

	// Subscribe to the topic
	subscribeToTopic(client, topic, timeout)

	// Wait for termination signal and handle shutdown
	handleShutdown(client, topic, timeout)
}
