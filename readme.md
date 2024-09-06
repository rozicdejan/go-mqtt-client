# MQTT Publisher in Go

This is a simple Go program that connects to an MQTT broker, reads messages from the terminal, and publishes those messages to a specified MQTT topic. If the user presses ENTER without entering a message, a default message is sent.

## Features

- Connects to an MQTT broker specified in a `.env` file.
- Continuously reads messages from the terminal and publishes them to an MQTT topic.
- If no message is provided, sends a default message ("Hello continuous loop").

## Requirements

- Go (version 1.16 or later)
- Mosquitto (or any other MQTT broker)

## Setup

### 1. Install Go

If you don’t have Go installed, download and install it from the official Go website: https://golang.org/dl/

### 2. Install Mosquitto

You need an MQTT broker like Mosquitto to test this application.

- Install Mosquitto from [Mosquitto website](https://mosquitto.org/download/) or use a cloud MQTT broker like [HiveMQ](https://www.hivemq.com).

### 3. Clone the Repository

```bash
git clone https://github.com/yourusername/mqtt-go-publisher.git
cd mqtt-go-publisher
```
## 4. Install Dependencies
This project uses the following Go libraries:

Paho MQTT Go Client
godotenv
Run the following commands to install the required dependencies:

```bash
go get github.com/eclipse/paho.mqtt.golang
go get github.com/joho/godotenv
```
## 5. Create a .env File
Create a .env file in the project root with the following content:

```bash
MQTT_BROKER=tcp://localhost:1883
MQTT_CLIENT_ID=custom_mqtt_client
MQTT_TOPIC=orodje/temp1
```
MQTT_BROKER: The address of the MQTT broker (replace localhost with your broker's address).
MQTT_CLIENT_ID: The MQTT client ID used to identify this client on the broker.
MQTT_TOPIC: The MQTT topic to which messages will be published.

## 6. Run the Application
Run the application using the following command:

```bash
go run main.go
```