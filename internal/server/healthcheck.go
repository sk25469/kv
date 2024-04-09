package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	models "github.com/sk25469/kv/internal/model"
)

func StartHealthCheck(dbState *models.DbState, ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			checkServerHealth(dbState)
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
	log.Printf("response: %v", response)

	return strings.TrimSpace(response) == "PONG"
}

// checkServerHealth checks the health of all servers and takes appropriate actions
func checkServerHealth(dbState *models.DbState) {
	masterConfig := dbState.GetMaster()
	// Check master server health
	if !sendPing(masterConfig.Port) {
		// Handle master failure
		log.Println("Master server is down. Promoting a slave to master...")
		dbState.RemoveFailedDb(masterConfig)
		promoteSlaveToMaster(dbState)
	}

	// Check each slave server health
	for _, config := range dbState.State {
		if !config.IsMaster && !sendPing(config.Port) {
			// Handle slave failure
			log.Printf("Slave server at %s is down. Creating a new slave...\n", config.IP)
			// Code to create a new slave goes here
			CreateNewSlave(dbState)
		}
	}
	log.Printf("everything working fine")
}

// This function needs to be called periodically, e.g., using a time.Ticker

func promoteSlaveToMaster(dbState *models.DbState) {
	// Code to promote a slave to master goes here
	for _, slaves := range dbState.State {
		if !slaves.IsMaster {
			dbState.SetMaster(slaves)
			return
		}
	}
}

// TODO: Implement this function
func CreateNewSlave(dbState *models.DbState) {

}
