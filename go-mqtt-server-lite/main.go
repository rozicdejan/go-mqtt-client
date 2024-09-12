package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mochi-co/mqtt/server"
	"github.com/mochi-co/mqtt/server/listeners"
)

func main() {
	// Create a new MQTT server instance
	srv := server.New()

	// Create a TCP listener on localhost:1883 (standard MQTT port)
	// If you want the broker to be accessible from other devices on your local network,
	// you need to bind the listener to your local network IP address or `0.0.0.0` (all interfaces).
	// Example:
	// tcp := listeners.NewTCP("tcp1", "0.0.0.0:1883")
	//
	// Alternatively, you can bind to your specific local IP (e.g., 192.168.1.100):
	// tcp := listeners.NewTCP("tcp1", "192.168.1.100:1883")
	//
	// For now, we'll use localhost to run only on the local machine.
	tcp := listeners.NewTCP("tcp", "0.0.0.0:1883") // Change to your local network IP to allow external connections

	// Add the TCP listener to the server
	err := srv.AddListener(tcp, nil)
	if err != nil {
		log.Fatalf("Failed to add listener: %v", err)
	}

	// Start the broker in a goroutine
	go func() {
		err := srv.Serve()
		if err != nil {
			log.Fatalf("Failed to start MQTT broker: %v", err)
		}
	}()

	log.Println("MQTT broker is running on localhost:1883 or on your local IP.")

	// Handle interrupt signal to gracefully shut down the broker
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down the broker...")
	srv.Close()
	log.Println("Broker stopped.")
}
