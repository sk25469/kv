package models

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Username string
	Password []byte
}

// ClientConfig represents the configuration details of a connected client
// KVServer represents the key-value server
type KVServer struct {
	Config  *Config
	auth    *Auth
	clients map[string]*ClientConfig // Map to store client configurations
	mu      sync.Mutex               // Mutex for thread-safe access to clients map
}

// NewKVServer creates a new instance of KVServer
func NewKVServer(config *Config) *KVServer {
	return &KVServer{
		Config:  config,
		clients: make(map[string]*ClientConfig),
		auth: &Auth{
			Username: config.Username,
			Password: []byte(config.GetPassword()),
		},
	}
}

// HandleClientConnect handles a new client connection
func (s *KVServer) HandleClientConnect(clientID, ipAddress string, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := &ClientConfig{
		ClientID:    clientID,
		IPAddress:   ipAddress,
		ConnectTime: time.Now(),
		ClientState: NewClientState(),
		Connection:  &conn,
	}

	s.clients[clientID] = config

	log.Printf("Client connected: ID=%s, IP=%s\n", clientID, ipAddress)
}

// HandleClientDisconnect handles a client disconnection
func (s *KVServer) HandleClientDisconnect(clientID string, conn *net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[clientID]; ok {
		delete(s.clients, clientID)
		fmt.Printf("Client disconnected: ID=%s\n", clientID)
	}
}

func (s *KVServer) GetClientsMap() map[string]*ClientConfig {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.clients
}

// GetClientConfig retrieves the configuration details of a client
func (s *KVServer) GetClientConfig(clientID string) (*ClientConfig, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, ok := s.clients[clientID]
	return config, ok
}

func (s *KVServer) Authenticate(username, password string) (string, bool) {
	if s.auth.Username != username {
		return "username didn't match", false
	}
	// Retrieve hashed password for the given username
	hashedPassword := s.auth.Password

	// Compare the provided password with the hashed password
	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		return "password didn't match", false
	}

	return "auth successful", true
}

func CreateHashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error generating hashed password: %v", err)
		return "", err
	}
	return string(hashedPassword), nil
}
