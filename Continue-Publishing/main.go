package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
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
	log.Println("USERNAME: ", os.Getenv("USERNAME"))
	log.Println("PASSWORD: ", os.Getenv("PASSWORD"))
}

// Utility function to get environment variables with a fallback default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Configurable variables (loaded after .env is initialized)
var (
	mqttBroker = getEnv("MQTT_BROKER", "tcp://localhost:1883")
	clientID   = getEnv("MQTT_CLIENT_ID", "go_mqtt_client")
	topic      = getEnv("MQTT_TOPIC", "orodje/temp1")
	timeout    = 5 * time.Second
)

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
func connectToMQTT(broker, clientID string, timeout time.Duration, username string, password string) mqtt.Client {
	// Define the MQTT broker options
	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetUsername(username)
	opts.SetPassword(password)

	// Set callback for connection lost (including duplicate client ID detection)
	opts.OnConnectionLost = onConnectionLost

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

// Publish message to the MQTT topic
func publishMessage(client mqtt.Client, topic, message string, timeout time.Duration) {
	token := client.Publish(topic, 0, false, message)
	if err := waitWithTimeout(token, timeout); err != nil {
		log.Fatalf("Failed to publish message to topic %s: %v", topic, err)
	}
	log.Printf("Published message to topic %s: %s", topic, message)
}

func main() {
	// Load environment variables
	mqttBroker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	clientID := getEnv("MQTT_CLIENT_ID", "go_mqtt_client")
	topic := getEnv("MQTT_TOPIC", "orodje/temp1")
	username := getEnv("USERNAME", "admin")
	password := getEnv("PASSWORD", "admin")
	timeout := 5 * time.Second

	// Connect to the MQTT broker
	client := connectToMQTT(mqttBroker, clientID, timeout, username, password)

	// Create a reader to capture input from the user
	reader := bufio.NewReader(os.Stdin)

	// Loop to read input from the user and send it as a message to the topic
	for {
		fmt.Print("Enter message to send (or press ENTER to send a default message): ")
		input, _ := reader.ReadString('\n') // Read user input from terminal

		// Trim the newline characters from the input
		message := strings.TrimSpace(input)

		// If the user just pressed ENTER, send a default message
		if message == "" {
			message = "Hello continuous loop"
		}

		// Publish the message to the MQTT topic
		publishMessage(client, topic, message, timeout)
	}
}

// Handle connection lost event (for duplicate client IDs)
func onConnectionLost(client mqtt.Client, err error) {
	log.Printf("Connection lost: %v", err)
	if err.Error() == "Connection refused: identifier rejected" {
		log.Println("Duplicate client ID detected.")
	}
	if err.Error() == "EOF" {
		log.Println("Duplicate client ID detected.")
	}
}
