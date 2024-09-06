# Go MQTT Publisher and Continuous Publisher

This repository contains two Go programs:

1. **MQTT Publisher in Go**: A simple Go program that connects to an MQTT broker, reads messages from the terminal, and publishes those messages to a specified MQTT topic. If the user presses ENTER without entering a message, a default message is sent.
2. **Continuous MQTT Publisher in Go**: A Go program that allows the user to continuously input messages from the terminal and publish them to an MQTT topic. If the user presses ENTER without input, a default message is sent.

## Features

- Connects to an MQTT broker specified in a `.env` file.
- Publishes messages to an MQTT topic based on user input or sends default messages when no input is provided.
- Continuously reads messages from the terminal and publishes them to a specified MQTT topic.
- Automatically reconnects to the MQTT broker if the connection drops.

## Requirements

- **Go** (version 1.16 or later)
- **Mosquitto** (or any other MQTT broker)

## Setup

### 1. Install Go

If you don’t have Go installed, download and install it from the official Go website: https://golang.org/dl/

### 2. Install Mosquitto

You need an MQTT broker like Mosquitto to test this application.

- Install Mosquitto from [Mosquitto website](https://mosquitto.org/download/) or use a cloud MQTT broker like [HiveMQ](https://www.hivemq.com).

### 3. Clone the Repository

```bash
git clone https://github.com/rozicdejan/go-mqtt-client
cd Continue-Publishing
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
## 5. Create a .env File (in main folder and in continue-publishing folder)
Create a .env file in the project root with the following content:
```bash
1. file in /.env
2. file in Continue-Publishing/.env
```

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
## 7. Testing with Mosquitto
To test the messages sent to the MQTT broker, you can use mosquitto_sub to subscribe to the topic:

```bash
mosquitto_sub -h localhost -t "orodje/temp1"
```