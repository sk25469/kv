package main

import (
	"fmt"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/internal/server"
	"github.com/sk25469/kv/internal/utils"
)

func main() {
	fmt.Println("Starting Key-Value Store Server...")
	config, err := models.LoadConfig(utils.CONFIG_FILE)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	fmt.Printf("Loading KV Store with following config: ----------------- ")

	fmt.Println("Port:", config.Port)
	fmt.Println("Max Connections:", config.MaxConnections)
	fmt.Println("Username:", config.Username)
	fmt.Println("Password:", config.Password)
	fmt.Println("Protected mode:", config.ProtectedMode)

	server.Start(config)
}
