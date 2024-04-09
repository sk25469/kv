package models

import (
	"log"
	"net"
	"sync"
)

type PubSub struct {
	clients map[string]*ClientConfig
	topics  map[string][]chan string
	Mutex   sync.Mutex
}

func NewPubSub() *PubSub {
	return &PubSub{
		topics:  make(map[string][]chan string),
		clients: make(map[string]*ClientConfig),
	}
}

func (ps *PubSub) Subscribe(topic string, conn net.Conn, cc *ClientConfig) {
	ps.Mutex.Lock()
	defer ps.Mutex.Unlock()

	log.Printf("checking if client is already subscribed to topic: %s -------- %v", topic, cc.ClientID)
	if _, ok := ps.clients[cc.ClientID]; ok {
		log.Printf("Client already subscribed to topic: %s", topic)
		return
	}

	if _, ok := ps.topics[topic]; !ok {
		log.Printf("Creating new topic: %s", topic)
		ps.topics[topic] = make([]chan string, 0)
	}

	// Create a channel for this subscription
	ch := make(chan string)
	ps.topics[topic] = append(ps.topics[topic], ch)
	ps.clients[cc.ClientID] = cc

	log.Printf("Subscribed client %s to topic: %s", cc.ClientID, topic)

	// Start a goroutine to listen for messages on this channel and forward them to the client
	go func() {
		log.Printf("inside subsciber goroutine")
		for message := range ch {
			log.Printf("Attempting to write message to client: %s", message) // Add this log
			_, err := conn.Write([]byte(message + "\n"))
			if err != nil {
				// Handle error (e.g., client disconnected)
				log.Printf("Error writing to connection: %v", err)
				break
			}
		}
	}()
}

func (ps *PubSub) Publish(topic, message string) {
	ps.Mutex.Lock()
	defer ps.Mutex.Unlock()

	log.Printf("Publishing message -------------------: %s ----------- %s", topic, message)

	if subscribers, ok := ps.topics[topic]; ok {
		log.Printf("subscribers: %v", subscribers)
		for _, ch := range subscribers {
			// Non-blocking send in case of slow consumers
			go func(ch chan string) {
				select {
				case ch <- message:
					log.Printf("Publishing message to topic: %s", topic)

				default:
					// Log or handle the fact that a message was not sent.
					log.Printf("Message not sent to client")
				}
			}(ch)
		}
	}
}
