// client.go

package models

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type KVClient struct {
	conn net.Conn
}

func NewKVClient(address string) (*KVClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &KVClient{conn: conn}, nil
}

func (c *KVClient) sendCommand(command string) (string, error) {
	_, err := c.conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", err
	}

	response, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		return "", err
	}

	return response, nil
}

func (c *KVClient) Begin() (string, error) {
	return c.sendCommand("BEGIN")
}

func (c *KVClient) Commit() (string, error) {
	return c.sendCommand("COMMIT")
}

func (c *KVClient) Rollback() (string, error) {
	return c.sendCommand("ROLLBACK")
}

func (c *KVClient) TSet(collectionName, key, value string) (string, error) {
	return c.sendCommand(fmt.Sprintf("TSET %s %s %s", collectionName, key, value))
}

func (c *KVClient) TGet(collectionName, key string) (string, error) {
	return c.sendCommand(fmt.Sprintf("TGET %s %s", collectionName, key))
}

func (c *KVClient) Auth(username, password string) (string, error) {
	return c.sendCommand(fmt.Sprintf("AUTH %s %s", username, password))
}

func (c *KVClient) Set(collectionName, key, value string) (string, error) {
	return c.sendCommand(fmt.Sprintf("SET %s %s %s", collectionName, key, value))
}

func (c *KVClient) Get(collectionName, key string) (string, error) {
	return c.sendCommand(fmt.Sprintf("GET %s %s", collectionName, key))
}

func (c *KVClient) Delete(collectionName, key string) (string, error) {
	return c.sendCommand(fmt.Sprintf("DELETE %s %s", collectionName, key))
}

func (c *KVClient) SetTTL(collectionName, key, ttl string) (string, error) {
	return c.sendCommand(fmt.Sprintf("SET-TTL %s %s %s", collectionName, key, ttl))
}

// Pub-Sub related methods
func (c *KVClient) Subscribe(topic string) (<-chan string, error) {
	// Send subscribe command to server
	if _, err := c.sendCommand(fmt.Sprintf("SUBSCRIBE %s", topic)); err != nil {
		return nil, err
	}

	// Create a channel to receive messages
	messages := make(chan string)

	// Start a goroutine to listen for messages on this topic
	go func() {
		for {
			// Assuming the server sends messages terminated by newline
			response, err := bufio.NewReader(c.conn).ReadString('\n')
			if err != nil {
				close(messages)
				return
			}
			messages <- response
			log.Printf("incoming message: %v", response)
		}
	}()

	return messages, nil
}

func (c *KVClient) Publish(topic, message string) (string, error) {
	return c.sendCommand(fmt.Sprintf("PUBLISH %s %s", topic, message))
}

func (c *KVClient) Close() error {
	return c.conn.Close()
}
