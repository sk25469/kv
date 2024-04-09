package models

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	IP             string
	Port           string
	MaxConnections int
	Username       string
	password       string
	ProtectedMode  bool
	IsMaster       bool
}

func NewConfig(ip, port, username, password string) *Config {
	return &Config{
		IP:             ip,
		Port:           port,
		MaxConnections: 10,
		Username:       username,
		password:       password,
		IsMaster:       false,
		ProtectedMode:  false,
	}
}

func (c *Config) GetPassword() string {
	return c.password
}

func (c *Config) SetPassword(password string) {
	c.password = password
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := NewConfig("", "", "", "")
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
			config.SetPassword(hashedPassword)
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
