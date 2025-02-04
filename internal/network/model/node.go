package network

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sk25469/kv/utils"
)

type NodeConfig struct {
	ID              string `json:"id"`
	IP              string `json:"ip"`
	Port            string `json:"port"`
	MaxConnections  int    `json:"max_connections"`
	username        string `json:"username"`
	password        string `json:"password"`
	IsMaster        bool   `json:"is_master"`
	HealthCheckPort int    `json:"health_check_port"`
	LogPath         string `json:"log_file_path"`
}

func NewNodeConfig(filename string) *NodeConfig {
	config, err := loadConfig(filename)
	if err != nil {
		log.Printf("error loading config: %v", err)
		return &NodeConfig{}
	}
	config.ID = setNodeID()
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
		if strings.HasPrefix(parts[0], "#") || line == "" {
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
		case "health_check_port":
			port, err := strconv.Atoi(value)
			if err != nil {
				log.Printf("error converting health_check_port to int: %v", err)
				return &NodeConfig{}, err
			}
			config.HealthCheckPort = port
		case "password":
			hashedPassword, err := utils.CreateHashedPassword(value)
			if err != nil {
				log.Printf("error generating hashed password: %v", err)
				return &NodeConfig{}, err
			}
			config.password = hashedPassword
		case "log_file":
			config.LogPath = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &config, nil
}

func setNodeID() string {
	return utils.GenerateBase64ClientID()
}

func (n *NodeConfig) ToJson() string {
	json, err := utils.ConvertStructToJSON(n)
	if err != nil {
		log.Printf("error converting struct to json: %v", err)
		return ""
	}
	return json
}

func FromJSON(bytes []byte) (*NodeConfig, error) {
	var res *NodeConfig
	err := json.Unmarshal(bytes, &res)
	if err != nil {
		log.Printf("error unmarshalling json: %v", err)
		return nil, err
	}
	return res, nil
}
