package models

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ClientConfig represents the configuration details of a connected client
type ClientConfig struct {
	ClientID    string
	IPAddress   string
	ConnectTime time.Time
	ClientState *ClientState
}

// KVServer represents the key-value server
type KVServer struct {
	clients map[string]*ClientConfig // Map to store client configurations
	mu      sync.Mutex               // Mutex for thread-safe access to clients map
}

// NewKVServer creates a new instance of KVServer
func NewKVServer() *KVServer {
	return &KVServer{
		clients: make(map[string]*ClientConfig),
	}
}

// HandleClientConnect handles a new client connection
func (s *KVServer) HandleClientConnect(clientID, ipAddress string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := &ClientConfig{
		ClientID:    clientID,
		IPAddress:   ipAddress,
		ConnectTime: time.Now(),
		ClientState: NewClientState(),
	}

	s.clients[clientID] = config

	log.Printf("Client connected: ID=%s, IP=%s\n", clientID, ipAddress)
}

// HandleClientDisconnect handles a client disconnection
func (s *KVServer) HandleClientDisconnect(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[clientID]; ok {
		delete(s.clients, clientID)
		fmt.Printf("Client disconnected: ID=%s\n", clientID)
	}
}

// GetClientConfig retrieves the configuration details of a client
func (s *KVServer) GetClientConfig(clientID string) (*ClientConfig, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, ok := s.clients[clientID]
	return config, ok
}
