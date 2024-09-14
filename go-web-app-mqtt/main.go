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

// Config struct to hold the application configuration
type Config struct {
	WebAppPort    int      `json:"web_app_port"`
	MQTTBrokerURL string   `json:"mqtt_broker_url"`
	MQTTClientID  string   `json:"mqtt_client_id"`
	MQTTTopics    []string `json:"mqtt_topics"`
}

var (
	receivedMessages []string                       // Store all received messages in memory
	lastSentIndex    int                            // Index to track last sent message
	db               *sql.DB                        // SQLite database connection
	messageChan      = make(chan mqtt.Message, 100) // Buffered channel for MQTT messages
	mutex            sync.RWMutex                   // RWMutex for handling shared resources
)

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

	// Start a background goroutine to process messages from the channel
	go processMessages()

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

// Save the MQTT message into the database (run this in a separate goroutine)
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
		// Pass the received message to the channel
		messageChan <- msg
	})
	token.Wait()
	if token.Error() != nil {
		log.Fatalf("Error subscribing to topic %s: %v", topic, token.Error())
	}
	fmt.Println("Subscribed to topic:", topic)
}

// Process messages from the channel in the background
func processMessages() {
	for msg := range messageChan {
		messageContent := fmt.Sprintf("Topic: %s, Message: %s", msg.Topic(), string(msg.Payload()))

		// Append the message to the receivedMessages slice in a thread-safe way
		mutex.Lock()
		receivedMessages = append(receivedMessages, messageContent)
		mutex.Unlock()

		// Save the message to the database (non-blocking)
		go saveMessageToDB(msg.Topic(), string(msg.Payload()))

		fmt.Printf("Processed message from topic %s: %s\n", msg.Topic(), string(msg.Payload()))
	}
}

// Start the web server to serve the frontend and messages
func startWebServer(port int) {
	router := gin.Default()

	// Serve static files from the "css" directory
	router.Static("/css", "./css")

	router.LoadHTMLGlob("templates/*")

	// Serve the index.html page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Serve config.json as a JSON response via the /config endpoint
	router.GET("/config", serveConfigJSON)

	// Serve only new messages as JSON
	router.GET("/messages", func(c *gin.Context) {
		mutex.RLock()                            // Use RLock to allow concurrent reads
		data := receivedMessages[lastSentIndex:] // Get only new messages since lastSentIndex
		lastSentIndex = len(receivedMessages)    // Update lastSentIndex to the current message count
		mutex.RUnlock()

		c.JSON(http.StatusOK, data)
	})

	addr := fmt.Sprintf(":%d", port)
	router.Run(addr)
}

// serveConfigJSON reads the config.json file and serves it as JSON
func serveConfigJSON(c *gin.Context) {
	// Read the config.json file from the root directory
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config.json: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read config file"})
		return
	}

	// Unmarshal the JSON file into the Config struct
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling config.json: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid config format"})
		return
	}

	// Serve the config as JSON
	c.JSON(http.StatusOK, config)
}
