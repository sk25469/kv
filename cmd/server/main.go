package main

import (
	"fmt"

	"github.com/sk25469/kv/internal/server"
)

func main() {
	fmt.Println("Starting Key-Value Store Server...")
	server.Start()
}
