package main

import (
	"log"
	"os"
	"sync"
	"time"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/internal/server"
	"github.com/sk25469/kv/utils"
)

func main() {
	asciiArt()
	shardList := models.NewShardsList()

	// Read the JSON config path
	log.Print("Reading shard config file...\n")
	jsonData, err := os.ReadFile(utils.SHARD_CONFIG_FILE)
	if err != nil {
		log.Fatal(err)
	}

	var shardConfig models.ShardConfig
	shardConfig.JsonUnmarshal(jsonData)

	log.Print("Starting shard...\n")
	checkSnapshotFileAndCreate(shardConfig)

	var wg sync.WaitGroup

	for _, shardDbConfig := range shardConfig.ShardList {
		shard := models.NewShard(&models.DbState{})

		shardStarted := make(chan bool)

		wg.Add(1)
		go server.StartShard(&wg, shard, shardStarted, shardList, shardDbConfig)
		<-shardStarted
		log.Printf("Shard with ID %v started\n", shardDbConfig.ShardID)
	}

	time.Sleep(10 * time.Second)
	server.ShutdownServer(shardList.Shards[0].Nodes[0])

	wg.Wait()
}

func checkSnapshotFileAndCreate(shardConfig models.ShardConfig) error {
	for _, shard := range shardConfig.ShardList {
		snapshotPath := shard.SnapshotPath
		if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
			log.Printf("Snapshot file not found at %s. Creating a new snapshot...\n", snapshotPath)
			// create new snapshotpath.txt file here
			_, err := os.Create(snapshotPath)
			if err != nil {
				log.Printf("error creating snapshot file: %v", err)
				return err
			}
		} else {
			log.Printf("Snapshot file found at %s\n", snapshotPath)
		}
	}
	return nil
}

func asciiArt() {
	art := `          _____               _____          
         /\    \             /\    \         
        /::\____\           /::\____\        
       /:::/    /          /:::/    /        
      /:::/    /          /:::/    /         
     /:::/    /          /:::/    /          
    /:::/____/          /:::/____/           
   /::::\    \          |::|    |            
  /::::::\____\________ |::|    |     _____  
 /:::/\:::::::::::\    \|::|    |    /\    \ 
/:::/  |:::::::::::\____|::|    |   /::\____\
\::/   |::|~~~|~~~~~    |::|    |  /:::/    /
 \/____|::|   |         |::|    | /:::/    / 
       |::|   |         |::|____|/:::/    /  
       |::|   |         |:::::::::::/    /   
       |::|   |         \::::::::::/____/    
       |::|   |          ~~~~~~~~~~          
       |::|   |                              
       \::|   |                              
        \:|   |                              
         \|___|                              
                                             
			`
	log.Print(art)
}
