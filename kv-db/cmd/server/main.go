package main

import (
	"flag"
	"fmt"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/internal/server"
)

func main() {
	fmt.Println("Starting Key-Value Store Server...")

	configFile := flag.String("config", "", "config file path")

	flag.Parse()
	fmt.Printf("Loading configuration from file: %v\n", configFile)

	config, err := models.LoadConfig(*configFile)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	fmt.Printf("Loading KV Store Master with following config: ----------------- \n")

	fmt.Println("Port:", config.Port)
	fmt.Println("Max Connections:", config.MaxConnections)
	fmt.Println("Username:", config.Username)
	fmt.Println("Password:", config.GetPassword())
	fmt.Println("Protected mode:", config.ProtectedMode)

	// Channel to signal that the master server is ready
	masterReady := make(chan bool)

	// Start the master server in a goroutine
	go func() {
		server.Start(config, masterReady)
	}()

	// Wait for the master server to signal it's ready
	<-masterReady
	fmt.Println("Master server is ready, starting slave servers...")

	// Start each slave server in its own goroutine
	for _, slave := range config.Slaves {
		slaveReady := make(chan bool)
		go func(slaveConfig *models.Config) {
			fmt.Println("Loading KV Store Slave with following config: -----------------")
			fmt.Println("Port:", slaveConfig.Port)
			fmt.Println("Max Connections:", slaveConfig.MaxConnections)
			fmt.Println("Username:", slaveConfig.Username)
			fmt.Println("Password:", slaveConfig.GetPassword())
			fmt.Println("Protected mode:", slaveConfig.ProtectedMode)
			server.Start(slaveConfig, slaveReady)
		}(slave)
		<-slaveReady
	}

	// Use an infinite loop or a more sophisticated method to keep the main goroutine alive
	select {}
}
