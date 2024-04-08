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
	Slaves         []*Config
}

func NewConfig(ip, port, username, password string) *Config {
	return &Config{
		IP:             ip,
		Port:           port,
		MaxConnections: 10,
		Username:       username,
		password:       password,
		ProtectedMode:  false,
		Slaves:         []*Config{},
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
		case "slave:port":
			slave_port_list := strings.Split(value, ",")
			log.Printf("slave_port_list: %v", slave_port_list)
			for _, slave_port := range slave_port_list {
				config.Slaves = append(config.Slaves, NewConfig(config.IP, slave_port, config.Username, config.GetPassword()))
			}
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

	// log.Printf("slaves: %v", config.Slaves)
	// for _, slave := range config.Slaves {
	// 	log.Printf("slave: %v", slave)
	// }

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
