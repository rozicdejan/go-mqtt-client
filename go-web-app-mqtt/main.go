package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

type Config struct {
	WebAppPort    int      `json:"web_app_port"`
	MQTTBrokerURL string   `json:"mqtt_broker_url"`
	MQTTClientID  string   `json:"mqtt_client_id"`
	MQTTTopics    []string `json:"mqtt_topics"`
}

var receivedMessages []string // Store all received messages in memory
var lastSentIndex int         // Index to track last sent message
var mutex sync.Mutex          // Ensure safe access to receivedMessages
var db *sql.DB                // SQLite database connection

func main() {
	// Load configuration
	config := loadConfig("config.json")

	// Initialize the database
	initDatabase()

	// Setup MQTT client
	mqttClient := connectToMQTTBroker(config)

	// Subscribe to topics
	for _, topic := range config.MQTTTopics {
		subscribeToTopic(mqttClient, topic)
	}

	// Start the web server
	startWebServer(config.WebAppPort)
}

// Load the configuration from config.json
func loadConfig(filePath string) Config {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
	return config
}

// Initialize the SQLite database and create the table if it doesn't exist
func initDatabase() {
	var err error
	db, err = sql.Open("sqlite3", "./mqtt_data.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Create the table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS mqtt_data_received (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		topic TEXT,
		message TEXT,
		received_at DATETIME
	);
	`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	fmt.Println("Database and table initialized.")
}

// Insert the MQTT message into the database
func saveMessageToDB(topic string, message string) {
	insertQuery := `
	INSERT INTO mqtt_data_received (topic, message, received_at)
	VALUES (?, ?, ?);
	`

	_, err := db.Exec(insertQuery, topic, message, time.Now().Format(time.RFC3339))
	if err != nil {
		log.Printf("Error inserting message into database: %v", err)
	}
}

// Connect to the MQTT broker
func connectToMQTTBroker(config Config) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MQTTBrokerURL)
	opts.SetClientID(config.MQTTClientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}
	fmt.Println("Connected to MQTT broker:", config.MQTTBrokerURL)
	return client
}

// Subscribe to the MQTT topics and handle incoming messages
func subscribeToTopic(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		mutex.Lock()
		receivedMessages = append(receivedMessages, fmt.Sprintf("Topic: %s, Message: %s", msg.Topic(), string(msg.Payload())))
		mutex.Unlock()

		// Save the message to the database
		saveMessageToDB(msg.Topic(), string(msg.Payload()))

		fmt.Printf("Received message from topic %s: %s\n", msg.Topic(), string(msg.Payload()))
	})
	token.Wait()
	if token.Error() != nil {
		log.Fatalf("Error subscribing to topic %s: %v", topic, token.Error())
	}
	fmt.Println("Subscribed to topic:", topic)
}

// Start the web server to serve the frontend and messages
func startWebServer(port int) {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	// Serve the index.html page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Serve only new messages as JSON
	router.GET("/messages", func(c *gin.Context) {
		mutex.Lock()
		data := receivedMessages[lastSentIndex:] // Get only new messages since lastSentIndex
		lastSentIndex = len(receivedMessages)    // Update lastSentIndex to the current message count
		mutex.Unlock()

		c.JSON(http.StatusOK, data)
	})

	addr := fmt.Sprintf(":%d", port)
	router.Run(addr)
}
