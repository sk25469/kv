package server

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/utils"
)

func StartShard(wg *sync.WaitGroup, shard *models.Shard, shardReady chan bool, shardList *models.ShardsList, shardConfigDb *models.ShardDbConfig) {

	defer wg.Done()
	shardID := utils.GetShardID()
	log.Printf("Starting shard with ID: %v\n", shardID)

	dbStates := models.NewDbState()
	shard.ShardID = shardID

	shardConfigDb.ShardID = shardID

	// Configuration files for master and slaves
	masterConfigFile := shardConfigDb.GetShardMasterPath()
	slaveConfigFiles := shardConfigDb.GetShardSlavesPathList()

	// Load master configuration
	masterConfig, err := models.LoadConfig(masterConfigFile)
	if err != nil {
		fmt.Printf("Error loading master configuration: %v\n", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(utils.HEALTH_CHECK_INTERVAL)
	defer ticker.Stop()

	dbStates.InsertDb(masterConfig)
	dbStates.SetMaster(masterConfig)

	// Start master server in a goroutine
	masterStarted := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		StartServer(masterConfig, true, masterStarted, shardConfigDb, shard)
	}()

	<-masterStarted
	fmt.Println("Master server started")
	shard.AddNode(masterConfig)

	// Start slave servers in separate goroutines
	for _, configFile := range slaveConfigFiles {
		slaveStarted := make(chan bool)
		slaveConfig, err := models.LoadConfig(configFile)
		if err != nil {
			fmt.Printf("Error loading slave configuration from %s: %v\n", configFile, err)
			return
		}
		wg.Add(1)

		go CreateNewSlave(dbStates, slaveConfig, slaveStarted, wg, shardConfigDb, shard)
		<-slaveStarted
		shard.AddNode(slaveConfig)
		fmt.Printf("Slave server started with config: %s\n", configFile)
	}

	shard.DbState = dbStates
	shardList.AddShard(shard)

	// Wait for all servers to start
	fmt.Println("All servers are up and running")

	shardReady <- true

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			log.Printf("printing db state for current shard: %v\n", shardConfigDb.ShardID)
			dbStates.PrintDbState()
			time.Sleep(30 * time.Second)
		}
	}()

	// Periodically check server health
	go StartHealthCheck(dbStates, ticker, shardConfigDb, shard)

	time.Sleep(10 * time.Second)
	ShutdownServer("8001")

	wg.Wait()
}
