package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

type Config struct {
	WebAppPort    int      `json:"web_app_port"`
	MQTTBrokerURL string   `json:"mqtt_broker_url"`
	MQTTClientID  string   `json:"mqtt_client_id"`
	MQTTTopics    []string `json:"mqtt_topics"`
}

var receivedMessages []string // Store all received messages
var lastSentIndex int         // Index to track last sent message
var mutex sync.Mutex          // Ensure safe access to receivedMessages

func main() {
	// Load configuration
	config := loadConfig("config.json")

	// Setup MQTT client
	mqttClient := connectToMQTTBroker(config)

	// Subscribe to topics
	for _, topic := range config.MQTTTopics {
		subscribeToTopic(mqttClient, topic)
	}

	// Start the web server
	startWebServer(config.WebAppPort)
}

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

func subscribeToTopic(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		mutex.Lock()
		receivedMessages = append(receivedMessages, fmt.Sprintf("Topic: %s, Message: %s", msg.Topic(), string(msg.Payload())))
		mutex.Unlock()
		fmt.Printf("Received message from topic %s: %s\n", msg.Topic(), string(msg.Payload()))
	})
	token.Wait()
	if token.Error() != nil {
		log.Fatalf("Error subscribing to topic %s: %v", topic, token.Error())
	}
	fmt.Println("Subscribed to topic:", topic)
}

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
