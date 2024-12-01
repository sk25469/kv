package network

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/sk25469/kv/utils"
)

type NodeConfig struct {
	IP             string
	Port           string
	MaxConnections int
	username       string
	password       string
	IsMaster       bool
}

func NewNodeConfig(filename string) *NodeConfig {
	config, err := loadConfig(filename)
	if err != nil {
		log.Printf("error loading config: %v", err)
		return &NodeConfig{}
	}
	return config
}

func loadConfig(filename string) (*NodeConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := NodeConfig{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "ip":
			config.IP = value
		case "port":
			config.Port = value
		case "max_connections":
			config.MaxConnections = utils.ParseMaxConnections(value)
		case "username":
			config.username = value
		case "password":
			hashedPassword, err := utils.CreateHashedPassword(value)
			if err != nil {
				log.Printf("error generating hashed password: %v", err)
				return &NodeConfig{}, err
			}
			config.password = hashedPassword
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &config, nil
}
