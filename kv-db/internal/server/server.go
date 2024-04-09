package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/internal/utils"
)

// Start initializes the server
func Start(config *models.Config, readySignal chan<- bool) {
	// Start TCP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Port))
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	log.Printf("Server is listening on port %v...\n", config.Port)

	log.Printf("creating all the stores for the server: %v", config.Port)
	cs := models.NewCollectionStore()
	ps := models.NewPubSub()
	ts := models.NewTransactionalKeyValueStore()
	kvServer := models.NewKVServer(config)

	err = handleInitLoad(cs)
	if err != nil {
		log.Printf("error loading dump: %v", err)
		return
	}

	go WatchSnapshotAndUpdate(utils.DUMP_FILE_NAME, cs, ts, kvServer, ps)

	log.Printf("starting TTL cleanups")
	StartKVCleanup(cs, utils.CLEANUP_DURATION)

	// Accept client connections

	readySignal <- true
	log.Printf("server ready to accept connections: %v", config.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("connected with client: %v", conn)
		go handleConnection(conn, cs, ts, kvServer, ps)
	}
}

// handleConnection handles client connections
// / there can be 10 types of commands:
// 1. Normal commands: GET, SET, DEL, etc.
// 2. Pub/Sub commands: SUBSCRIBE, PUBLISH
// 3. Replication commands: REPLICATE = create a new slave with a configuration
// 4. Snapshot commands: SNAPSHOT = create a snapshot of the current state
// 5. Transaction commands: BEGIN, COMMIT, ROLLBACK
// 6. Admin commands: SHUTDOWN, MAKE_MASTER, MAKE_SLAVE
// 7. Config commands: CONFIG = get or set configuration
// 8. Health commands: PING
func handleConnection(conn net.Conn, cs *models.CollectionStore, ts *models.TransactionalKeyValueStore, kvServer *models.KVServer, ps *models.PubSub) {

	reader := bufio.NewReader(conn)
	remoteAddress := conn.RemoteAddr().String()
	clientId, err := utils.GenerateBase64ClientID()
	if err != nil {
		log.Printf("error generating client id: %v", err)
		return
	}

	log.Printf("handling connection: %v for slave %v", remoteAddress, kvServer.Config.Port)

	kvServer.HandleClientConnect(clientId, remoteAddress)

	clientConfig, _ := kvServer.GetClientConfig(clientId)

	// handle client disconnection
	defer func(clientId string) {
		conn.Close()
		kvServer.HandleClientDisconnect(clientId)
	}(clientId)

	// Create a bufio reader to read from the connection

	for {
		// Read the next line from the connection
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}
		cmd := ParseCommand(command)

		switch cmd.Name {
		case utils.SUBSCRIBE, utils.PUBLISH:
			handlePubSubMode(cmd, conn, ps, clientConfig)
		case utils.SHUTDOWN, utils.MAKE_MASTER, utils.MAKE_SLAVE:
			handleAdminCommands(conn, kvServer, cmd)
		case utils.CONFIG:
			handleConfigCommands(conn)
		case utils.PING:
			handleHealthCommands(conn)
		default:
			cmd := ParseCommand(command)
			if ShouldWriteLog(*cmd) {
				err = WriteCommandsToFile(*cmd, utils.DUMP_FILE_NAME)
				if err != nil {
					log.Printf("error writing operation to dump")
				}
			}
			result := ExecuteCommand(cmd, cs, ts, clientConfig, kvServer, ps)
			log.Printf("result for cmd: %v -------- %v", cmd, result)
			bytesWritten, err := fmt.Fprintln(conn, result)
			if err != nil {
				log.Printf("error writing to the connection: %v : [%v]", conn, err)
			}
			log.Printf("bytes written to conn: %v ----------- %v", conn, bytesWritten)

		}
	}
}

func handleInitLoad(cs *models.CollectionStore) error {
	cmds, err := ReadCommandsFromFile(utils.DUMP_FILE_NAME)
	if err != nil {
		log.Printf("error reading cmds from file: [%v]", err)
		return err
	}
	for _, cmd := range cmds {
		if ShouldWriteLog(cmd) {
			result := ExecuteCommand(&cmd, cs, nil, &models.ClientConfig{ClientState: &models.ClientState{State: utils.ACTIVE, IsAuthenticated: true}}, &models.KVServer{Config: &models.Config{ProtectedMode: false}}, nil)
			log.Printf("successfully executed curr cmd: %v ------------ %v", cmd, result)
		}
	}
	return nil
}

func handlePubSubMode(cmd *Command, conn net.Conn, pubSub *models.PubSub, cc *models.ClientConfig) {

	// Inform the client that it has entered pub/sub mode
	conn.Write([]byte("Entering pub/sub mode. Ready for SUBSCRIBE and PUBLISH commands.\n"))

	log.Printf("handling pubsub mode: %v", cmd)

	// SUBSCRIBE <topic>
	// PUBLISH <topic> <message>
	if strings.Contains(cmd.Name, utils.SUBSCRIBE) {
		topic := cmd.CollectionName
		subscribeToTopic(topic, conn, pubSub, cc)
	} else if strings.Contains(cmd.Name, utils.PUBLISH) {
		topic := cmd.CollectionName
		message := strings.Join(cmd.Args[0:], " ")
		publishToTopic(topic, message, conn, pubSub)
	} else {
		conn.Write([]byte("Unknown command in pub/sub mode.\n"))
	}
}

func subscribeToTopic(topic string, conn net.Conn, pubSub *models.PubSub, cc *models.ClientConfig) {
	pubSub.Subscribe(topic, conn, cc)
	// Inform the client of successful subscription
	conn.Write([]byte("Subscribed to " + topic + "\n"))
}

func publishToTopic(topic, message string, conn net.Conn, pubSub *models.PubSub) {
	pubSub.Publish(topic, message)
	conn.Write([]byte("Published message to " + topic + "\n"))
}

func ReplicateChanges(jsonCmd string, cs *models.CollectionStore, ts *models.TransactionalKeyValueStore, kvServer *models.KVServer, ps *models.PubSub) string {
	var cmd Command
	err := json.Unmarshal([]byte(jsonCmd), &cmd)
	if err != nil {
		return ""
	}

	log.Printf("parsed command for replication: %v", cmd)
	// Execute the command on the slave server
	result := ExecuteCommand(&cmd, cs, ts, &models.ClientConfig{ClientState: &models.ClientState{State: utils.ACTIVE, IsAuthenticated: true}}, kvServer, ps)
	return result
}

func handleAdminCommands(conn net.Conn, kvServer *models.KVServer, cmd *Command) {
	switch cmd.Name {
	case utils.SHUTDOWN:
		conn.Write([]byte("Shutting down server...\n"))
		// Shutdown the server
		os.Exit(0)
	case utils.MAKE_MASTER:
		conn.Write([]byte("Making server master...\n"))
		// Make the server a master
		if !kvServer.Config.IsMaster {
			kvServer.Config.IsMaster = true
		} else {
			conn.Write([]byte("Server is already a master.\n"))
		}
	case utils.MAKE_SLAVE:
		conn.Write([]byte("Making server slave...\n"))

		if kvServer.Config.IsMaster {
			// Make the server a slave
			kvServer.Config.IsMaster = false
		} else {
			conn.Write([]byte("Server is already a slave.\n"))
		}

	default:
		conn.Write([]byte("Unknown admin command.\n"))
	}

}

// TODO: Implement this function
func handleConfigCommands(conn net.Conn) {

}

func handleHealthCommands(conn net.Conn) {
	conn.Write([]byte("PONG\n"))
}
