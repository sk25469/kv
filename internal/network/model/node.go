package network

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/sk25469/kv/utils"
)

type NodeConfig struct {
	ID             string `json:"id"`
	IP             string `json:"ip"`
	Port           string `json:"port"`
	MaxConnections int    `json:"max_connections"`
	username       string `json:"username"`
	password       string `json:"password"`
	IsMaster       bool   `json:"is_master"`
	EtcdEndpoints  []string
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
		if strings.HasPrefix(parts[0], "#") || line == "" {
			continue
		}
		key := strings.TrimSpace(parts[0])
		var etcdEnpointValues []string

		for i := 1; i < len(parts); i++ {
			etcdEnpointValues = append(etcdEnpointValues, strings.TrimSpace(parts[i]))
		}

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
		case "etcd_endpoints":
			config.EtcdEndpoints = etcdEnpointValues
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (n *NodeConfig) SetNodeID() string {
	n.ID = utils.GenerateBase64ClientID()
	return n.ID
}
