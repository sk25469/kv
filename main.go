package main

import (
	"fmt"
	"os"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/internal/server"
	"github.com/sk25469/kv/utils"
)

// Assuming you have a function like this somewhere in your server package
// func StartServer(config *model.Config, isMaster bool) {
//     // Implementation to start the server based on the config and role
// }

func main() {
	// Configuration files for master and slaves
	masterConfigFile := utils.MASTER_CONFIG_FILE
	slaveConfigFiles := []string{utils.SLAVE_1_CONFIG, utils.SLAVE_2_CONFIG, utils.SLAVE_3_CONFIG}

	// Load master configuration
	masterConfig, err := models.LoadConfig(masterConfigFile)
	if err != nil {
		fmt.Printf("Error loading master configuration: %v\n", err)
		os.Exit(1)
	}

	// Start master server in a goroutine
	masterStarted := make(chan bool)
	// var wg sync.WaitGroup
	// wg.Add(1)
	go func() {
		// defer wg.Done()
		server.StartServer(masterConfig, true, masterStarted)
	}()

	<-masterStarted
	fmt.Println("Master server started")

	// Start slave servers in separate goroutines
	for _, configFile := range slaveConfigFiles {
		slaveStarted := make(chan bool)
		// wg.Add(1)
		go func(confFile string) {
			// defer wg.Done()
			slaveConfig, err := models.LoadConfig(confFile)
			if err != nil {
				fmt.Printf("Error loading slave configuration from %s: %v\n", confFile, err)
				return
			}
			server.StartServer(slaveConfig, false, slaveStarted)
		}(configFile)
		<-slaveStarted
		fmt.Printf("Slave server started with config: %s\n", configFile)
	}

	// Wait for all servers to start
	fmt.Println("All servers are up and running")
	// wg.Wait()
	select {}
}
