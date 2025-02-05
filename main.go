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
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sk25469/kv/internal/codec"
	"github.com/sk25469/kv/internal/comm"
	"github.com/sk25469/kv/internal/core"
	"github.com/sk25469/kv/internal/middleware"
	"github.com/sk25469/kv/internal/network"
	node_config "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/internal/replication"
	"github.com/sk25469/kv/internal/storage"
	storage_model "github.com/sk25469/kv/internal/storage/model"
	"github.com/sk25469/kv/utils"
)

func main() {
	// Define command-line flags
	configPath := flag.String("config", utils.MASTER_CONFIG_FILE, "Path to master config file")

	// Parse flags
	flag.Parse()
	utils.AsciiArt()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	communicationService := comm.NewCommunicationService()
	replicationService := replication.NewReplicationService(replication.ReplicationServiceParams{
		CommunicationLayer: communicationService,
	})

	storage, err := storage.NewStorage(storage.StorageServiceParams{
		Type:      storage_model.InMemory,
		Structure: storage_model.HashMap,
	})
	if err != nil {
		log.Fatalf("Error creating storage: %v", err)
	}

	nodeConfig := node_config.NewNodeConfig(*configPath)

	storageMiddleware, err := middleware.NewStorageMiddleware(storage, nodeConfig.LogPath)
	if err != nil {
		log.Fatalf("Error creating storage middleware: %v", err)
	}

	err = storageMiddleware.Recover()
	if err != nil {
		log.Fatalf("Error recovering storage: %v", err)
	}

	coreLayer := core.NewCoreService(
		core.CoreServiceParams{
			CommunicationLayer: communicationService,
			ReplicationLayer:   replicationService,
			StorageLayer:       storageMiddleware,
		},
	)

	codecLayer := codec.NewCodecLayerService()

	networkLayer := network.NewNetworkService(network.NetworkServiceParams{
		NodeConfig:         nodeConfig,
		CoreLayer:          coreLayer,
		CommunicationLayer: communicationService,
		CodecLayer:         codecLayer,
	})

	// Create a context that is cancelled on termination signals

	go func() {
		if err := networkLayer.Start(); err != nil {
			log.Fatalf("Error starting network layer: %v", err)
		}
	}()

	// start the health check service
	healthCheckService := network.NewHealthCheckService(network.HealthCheckServiceParams{
		Port:                 nodeConfig.HealthCheckPort,
		NetworkService:       networkLayer,
		CommunicationService: communicationService,
	})

	healthCheckService.StartHealthCheck()
	healthCheckService.StartPeriodicHealthChecks(utils.HEALTH_CHECK_INTERVAL)

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Stop the network service
	if err := networkLayer.Stop(); err != nil {
		log.Fatalf("Error stopping network layer: %v", err)
	}

	log.Println("Server stopped gracefully")

}
