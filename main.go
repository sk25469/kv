// package main

// import (
// 	"os"
// 	"sync"

// 	"github.com/sirupsen/logrus"
// 	models "github.com/sk25469/kv/internal/model"
// 	"github.com/sk25469/kv/internal/server"
// 	"github.com/sk25469/kv/logger"
// 	"github.com/sk25469/kv/utils"
// )

// // logger
// var log *logrus.Logger

// func main() {
// 	utils.AsciiArt()
// 	shardList := models.NewShardsList()

// 	logger.InitLogger()
// 	log = logger.GetLogger()

// 	// Read the JSON config path
// 	log.Info("Reading shard config file...\n")
// 	jsonData, err := os.ReadFile(utils.SHARD_CONFIG_FILE)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// new consistent hash ch
// 	ch := models.NewConsistentHash()

// 	var shardConfig models.ShardConfig
// 	shardConfig.JsonUnmarshal(jsonData)

// 	log.Info("Starting shard...\n")
// 	checkSnapshotFileAndCreate(shardConfig)

// 	var wg sync.WaitGroup

// 	for _, shardDbConfig := range shardConfig.ShardList {
// 		shard := models.NewShard(&models.DbState{})

// 		shardStarted := make(chan bool)

// 		wg.Add(1)
// 		go server.StartShard(&wg, shard, shardStarted, shardList, shardDbConfig, ch)
// 		<-shardStarted
// 		log.Infof("Shard with ID %v started\n", shardDbConfig.ShardID)
// 	}

// 	wg.Add(1)
// 	go server.RouteRequestsToShards(utils.SERVER_PORT, ch, shardList)

// 	// time.Sleep(30 * time.Second)
// 	// server.ShutdownServer(shardList.Shards[0].Nodes[0])

// 	wg.Wait()
// }

// func checkSnapshotFileAndCreate(shardConfig models.ShardConfig) error {
// 	for _, shard := range shardConfig.ShardList {
// 		snapshotPath := shard.SnapshotPath
// 		if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
// 			log.Infof("Snapshot file not found at %s. Creating a new snapshot...\n", snapshotPath)
// 			// create new snapshotpath.txt file here
// 			_, err := os.Create(snapshotPath)
// 			if err != nil {
// 				log.Infof("error creating snapshot file: %v", err)
// 				return err
// 			}
// 		} else {
// 			log.Infof("Snapshot file found at %s\n", snapshotPath)
// 		}
// 	}
// 	return nil
// }

package main

import (
	"flag"

	"github.com/sk25469/kv/internal/network"
	node_config "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/utils"
)

func main() {
	// Define command-line flags
	configPath := flag.String("config", utils.MASTER_CONFIG_FILE, "Path to master config file")

	// Parse flags
	flag.Parse()
	utils.AsciiArt()
	networkLayer := network.NewNetworkService(network.NetworkServiceParams{
		NodeConfig: node_config.NewNodeConfig(*configPath),
	})

	networkLayer.Start()

}
