package main

import (
	"log"
	"sync"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/internal/server"
)

func main() {
	shardList := models.NewShardsList()
	shard := models.NewShard(&models.DbState{})

	var wg sync.WaitGroup
	shardStarted := make(chan bool)

	wg.Add(1)
	go server.StartShard(&wg, shard, shardStarted, shardList)
	<-shardStarted
	log.Printf("Shard with ID %v started\n", shard.ShardID)

	wg.Wait()
}
