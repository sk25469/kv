package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sk25469/kv-client/models"
)

func main() {
	client, err := models.NewKVClient("localhost:7001") // Update with your server's address
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Example usage of the client
	if response, err := client.Auth("admin", "password"); err != nil {
		log.Fatalf("Auth failed: %v", err)
	} else {
		fmt.Println("Auth response:", response)
	}

	if response, err := client.Set("myCollection", "key1", "value1"); err != nil {
		log.Fatalf("Set failed: %v", err)
	} else {
		fmt.Println("Set response:", response)
	}

	messages, err := client.Subscribe("myTopic")
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// Listen for messages in a goroutine
	go func() {
		for message := range messages {
			fmt.Println("Received message:", message)
		}
	}()

	// Publish a message to the topic
	if resp, err := client.Publish("myTopic", "Hello, World!"); err != nil {
		log.Fatalf("Failed to publish: %v", err)
	} else {
		fmt.Println("Publish response:", resp)
	}

	// Keep the main goroutine alive for a short duration to receive messages
	time.Sleep(5 * time.Second)
}
