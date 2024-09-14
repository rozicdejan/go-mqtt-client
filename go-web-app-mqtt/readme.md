
# Go MQTT Web Application
A Go-based web application that uses MQTT to receive and display real-time messages from an MQTT broker. The application serves a web interface where incoming MQTT messages and application settings are displayed. It uses Gin for the web framework, Paho MQTT for the MQTT client, and SQLite to store the received messages.The configuration file config.json needs to be created or adjusted before running the application.


## Prerequisites
Before running the application, ensure you have the following installed:

Go (Golang) – Install Go
Git – Install Git
SQLite – The application uses SQLite to store MQTT messages. Ensure SQLite is installed on your system.
MQTT Broker – You can use a local broker or a public broker like HiveMQ for testing purposes.

# Setup Instructions
## Clone the Repository
```go
git clone https://github.com/rozicdejan/go-mqtt-client.git
cd go-mqtt-cient/go-web-app-mqtt
```
## Install Dependencies

```go
go mod tidy
```
## Create or Use the SQLite Database (Optionally)
The application uses an SQLite database (mqtt_data.db) to store incoming MQTT messages. If the file doesn't exist, the application will create it on startup.


Optionally, you can manually create the SQLite database in the root directory using the following command:
sqlite3 mqtt_data.db
```go
CREATE TABLE IF NOT EXISTS mqtt_data_received (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic TEXT,
    message TEXT,
    received_at DATETIME
);
```

## Configure the application
Edit the config.json file to set up your application's MQTT and web settings:
```json
{
  "web_app_port": 8081,
  "mqtt_broker_url": "tcp://192.168.33.1:1883",
  "mqtt_client_id": "go_mqtt_client_example",
  "mqtt_topics": [
    "/example/topic1"
  ]
}
```
web_app_port: The port on which the web application will run.
mqtt_broker_url: The URL of your MQTT broker.
mqtt_client_id: A unique client ID for connecting to the MQTT broker.
mqtt_topics: A list of MQTT topics to subscribe to.

# Running the Application
## Run the Application

After setting up the configuration file and database, start the application using the following command:

`go run main.go`

This will start the web server and connect to the specified MQTT broker. The web app will be available at http://localhost:8081 (or the port specified in config.json).

## Access the Web Interface

Open a browser and navigate to:

```go
'http://localhost:8081'
```
The webpage will load, displaying:

MQTT Messages: Messages received from the subscribed topics.
Application Settings: Configuration settings fetched from config.json.

# Configuring the Application
The application is configured using a config.json file in the root directory. Here’s an example of how to configure it:

config.json Example:
```go
Copy code
{
    "web_app_port": 8081,
    "mqtt_broker_url": "tcp://192.168.33.1:1883",
    "mqtt_client_id": "go_mqtt_client_example",
    "mqtt_topics": [
        "/example/topic1"
    ]
}
```
web_app_port: The port number on which the web server listens.
mqtt_broker_url: The MQTT broker URL (e.g., tcp://localhost:1883 or tcp://broker.hivemq.com:1883).
mqtt_client_id: A unique client ID used for connecting to the MQTT broker.
mqtt_topics: A list of MQTT topics to subscribe to.
Make sure the values in config.json match your setup.

# Database Explanation (mqtt_data.db)
The application uses an SQLite database (mqtt_data.db) to store MQTT messages it receives. If the database doesn't exist, it will be created automatically when the application starts.

Schema:
Table: mqtt_data_received
Columns:
id: Auto-incremented primary key.
topic: The MQTT topic from which the message was received.
message: The content of the MQTT message.
received_at: Timestamp when the message was received.
The database is useful for logging and displaying the messages on the front-end, allowing you to persist MQTT messages between sessions.