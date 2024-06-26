package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	models "github.com/sk25469/kv/internal/model"
)

func StartHealthCheck(dbState *models.DbState, ticker *time.Ticker, shardConfigDB *models.ShardDbConfig, shard *models.Shard) {
	for {
		select {
		case <-ticker.C:
			checkServerHealth(dbState, shardConfigDB, shard)
		}
	}
}

// SendPing sends a PING message to the specified address and waits for a PONG response
func sendPing(address string) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%v", address))
	if err != nil {
		log.Printf("Error connecting: %v\n", err)
		return false
	}
	defer conn.Close()

	_, err = conn.Write([]byte("PING\n"))
	if err != nil {
		log.Printf("Error sending PING: %v\n", err)
		return false
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Printf("Error reading PONG: %v\n", err)
		return false
	}

	return strings.TrimSpace(response) == "PONG"
}

// checkServerHealth checks the health of all servers and takes appropriate actions
func checkServerHealth(dbState *models.DbState, shardConfigDb *models.ShardDbConfig, shard *models.Shard) {
	var wg sync.WaitGroup

	masterConfig := dbState.GetMaster()
	// Check master server health
	if !sendPing(masterConfig.Port) {
		// Handle master failure
		log.Println("Master server is down. Promoting a slave to master...")

		// get active connections to current master
		activeConnectionsMap := shard.GetNode(masterConfig.IP, masterConfig.Port).GetClientsMap()
		dbState.RemoveFailedDb(masterConfig)
		shard.RemoveNode(shard.GetNode(masterConfig.IP, masterConfig.Port))
		newMasterConfig := promoteSlaveToMaster(dbState)
		newMasterKvServer := shard.GetNode(newMasterConfig.IP, newMasterConfig.Port)

		// Route active connections to new master
		routeActiveConnections(activeConnectionsMap, newMasterKvServer)
	}

	// Check each slave server health
	for _, config := range dbState.State {
		if !config.IsMaster && !sendPing(config.Port) {
			// Handle slave failure
			log.Printf("Slave server at %s:%s is down. Creating a new slave...\n", config.IP, config.Port)
			dbState.RemoveFailedDb(config)
			// Code to create a new slave goes here

			activeConnectionsMap := shard.GetNode(config.IP, config.Port).GetClientsMap()
			shard.RemoveNode(shard.GetNode(config.IP, config.Port))

			slaveStarted := make(chan bool)
			wg.Add(1)
			go CreateNewSlave(dbState, config, slaveStarted, &wg, shardConfigDb, shard)
			<-slaveStarted

			routeActiveConnections(activeConnectionsMap, shard.GetNode(config.IP, config.Port))
		}
	}
	log.Printf("everything working fine")
}

func routeActiveConnections(activeConnectionsMap map[string]*models.ClientConfig, newKvServer *models.KVServer) {
	log.Printf("Routing active connections to new server: %v", newKvServer.Config.Port)
	for clientId, clientConfig := range activeConnectionsMap {
		newKvServer.HandleClientConnect(clientId, clientConfig.IPAddress, *clientConfig.Connection)
	}
}

// This function needs to be called periodically, e.g., using a time.Ticker

func promoteSlaveToMaster(dbState *models.DbState) *models.Config {
	// Code to promote a slave to master goes here
	for _, slaves := range dbState.State {
		if !slaves.IsMaster {
			dbState.SetMaster(slaves)
			return slaves
		}
	}
	return &models.Config{}
}

func CreateNewSlave(dbStates *models.DbState, slaveConfig *models.Config, slaveStarted chan bool, wg *sync.WaitGroup, shardConfigDb *models.ShardDbConfig, shard *models.Shard) {
	defer wg.Done()
	dbStates.InsertDb(slaveConfig)
	StartServer(slaveConfig, false, slaveStarted, shardConfigDb, shard)
}

// TODO: 		Implement the following functions
func RouteActiveConnectionsToOtherServer() {

}
