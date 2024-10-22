package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

// EncoderData struct holds angle, voltage, and timestamp in nanoseconds
type EncoderData struct {
	Angle     int     `json:"angle"`
	Voltage   float64 `json:"voltage"`
	Timestamp int64   `json:"timestamp"`
}

// Generate a random voltage between 0.0 and 5.0
func generateVoltage() float64 {
	return rand.Float64() * 5.0
}

// Create JSON format for 360 data points
// It calculates the timestamps based on the current time
func generateEncoderData(startTime int64, nsPerDegree int64) string {
	data := make([]EncoderData, 360)
	for i := 0; i < 360; i++ {
		data[i] = EncoderData{
			Angle:     i,
			Voltage:   generateVoltage(),
			Timestamp: startTime + int64(i)*nsPerDegree, // Calculate timestamp for each degree
		}
	}
	ysonData, _ := json.Marshal(data)
	return string(ysonData)
}

// Non-blocking function to publish data to the MQTT broker with error handling
func publishData(client MQTT.Client, topic string, data string, retryCount int) {
	go func() {
		token := client.Publish(topic, 0, false, data)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Error publishing data: %v", token.Error())
			if retryCount > 0 {
				log.Printf("Retrying... Remaining attempts: %d", retryCount)
				publishData(client, topic, data, retryCount-1) // Retry in case of failure
			}
		} else {
			log.Println("Data successfully published")
		}
	}()
}

// Main loop to send data based on RPS (Revolutions Per Second)
func startPublishing(client MQTT.Client, rps float64, topic string, dataQueue chan string, stopCh <-chan struct{}) {
	// Time per full turn (in nanoseconds)
	nsPerFullTurn := int64((1.0 / rps) * 1e9)
	// Time per degree in nanoseconds
	nsPerDegree := nsPerFullTurn / 360

	for {
		select {
		case <-stopCh:
			// Gracefully exit the loop
			log.Println("Stopping publishing loop...")
			return
		default:
			// Start time in nanoseconds (UNIX time in ns)
			startTime := time.Now().UnixNano()

			// Generate encoder data
			ysonData := generateEncoderData(startTime, nsPerDegree)

			// Send generated data to the queue (non-blocking)
			select {
			case dataQueue <- ysonData:
				log.Println("Data added to queue")
			default:
				log.Println("Data queue is full, dropping data")
			}

			// Wait for the next full turn
			time.Sleep(time.Duration(nsPerFullTurn) * time.Nanosecond)
		}
	}
}

func dataSender(client MQTT.Client, topic string, dataQueue chan string, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			// Gracefully exit the loop
			log.Println("Stopping data sender...")
			return
		case data := <-dataQueue:
			// Send data from the queue with retry logic
			publishData(client, topic, data, 3)
		}
	}
}

func main() {
	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Read MQTT broker, client ID, topic, and RPS from the .env file
	mqttBroker := os.Getenv("MQTT_BROKER")
	clientID := os.Getenv("CLIENT_ID")
	topic := os.Getenv("TOPIC")
	rpsStr := os.Getenv("RPS")
	queueSizeStr := os.Getenv("QUEUE_SIZE")

	// Convert RPS and queue size to appropriate types
	rps, err := strconv.ParseFloat(rpsStr, 64)
	if err != nil {
		log.Printf("Invalid RPS value, defaulting to 1.0")
		rps = 1.0
	}

	queueSize, err := strconv.Atoi(queueSizeStr)
	if err != nil {
		log.Printf("Invalid QUEUE_SIZE value, defaulting to 10")
		queueSize = 10
	}

	// Initialize MQTT connection options
	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqttBroker)
	opts.SetClientID(clientID)

	// Create MQTT client
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}
	defer client.Disconnect(250)
	log.Println("Connected to MQTT broker")

	// Create a buffered channel for data queue
	dataQueue := make(chan string, queueSize) // Buffer size from .env

	// Channel to handle graceful shutdown
	stopCh := make(chan struct{})

	// Capture interrupt signal for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %v, shutting down...", sig)
		close(stopCh) // Stop publishing when we receive the signal
	}()

	// Start goroutine to send data from the queue
	go dataSender(client, topic, dataQueue, stopCh)

	// Start publishing data
	startPublishing(client, rps, topic, dataQueue, stopCh)

	log.Println("Application stopped")
}
