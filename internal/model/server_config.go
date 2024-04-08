package models

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           string
	MaxConnections int
	Username       string
	Password       string
	ProtectedMode  bool
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
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
		case "protected-mode":
			config.ProtectedMode = false
		case "port":
			config.Port = value
		case "max_connections":
			config.MaxConnections = parseMaxConnections(value)
		case "username":
			config.Username = value
		case "password":
			hashedPassword, err := CreateHashedPassword(value)
			if err != nil {
				log.Printf("error generating hashed password: %v", err)
				return &Config{}, err
			}
			config.Password = hashedPassword
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func parseMaxConnections(maxConnStr string) int {
	maxConn, err := strconv.Atoi(maxConnStr)
	if err != nil {
		// Handle error
		log.Printf("unable to parse maxConn: %v", err)
		return 10
	}
	return maxConn
}
