package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
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

	log.Printf("creating the collection store")
	cs := models.NewCollectionStore()
	ps := models.NewPubSub()

	err = handleInitLoad(cs)
	if err != nil {
		log.Printf("error loading dump: %v", err)
		return
	}

	log.Printf("starting TTL cleanups")
	StartKVCleanup(cs, utils.CLEANUP_DURATION)

	// Accept client connections
	kvServer := models.NewKVServer(config)

	readySignal <- true

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("connected with client: %v", conn)
		go handleConnection(conn, cs, kvServer, ps)
	}
}

// handleConnection handles client connections
func handleConnection(conn net.Conn, cs *models.CollectionStore, kvServer *models.KVServer, ps *models.PubSub) {

	reader := bufio.NewReader(conn)
	remoteAddress := utils.GetParsedIP(conn.RemoteAddr().String())
	clientId, err := utils.GenerateBase64ClientID(remoteAddress)
	if err != nil {
		log.Printf("error generating client id: %v", err)
		return
	}
	kvServer.HandleClientConnect(clientId, remoteAddress)

	clientConfig, _ := kvServer.GetClientConfig(clientId)

	// handle client disconnection
	defer func(clientId string) {
		conn.Close()
		kvServer.HandleClientDisconnect(clientId)
	}(clientId)

	// Create a bufio reader to read from the connection
	ts := models.NewTransactionalKeyValueStore()

	for {
		// Read the next line from the connection
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}
		if strings.TrimSpace(command) == "PUBSUB_MODE" {
			handlePubSubMode(conn, ps, clientConfig)
			return // Exit the main command handler
		}
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

func handlePubSubMode(conn net.Conn, pubSub *models.PubSub, cc *models.ClientConfig) {
	reader := bufio.NewReader(conn)

	// Inform the client that it has entered pub/sub mode
	conn.Write([]byte("Entering pub/sub mode. Ready for SUBSCRIBE and PUBLISH commands.\n"))

	for {
		rawCommand, err := reader.ReadString('\n')
		if err != nil {
			// Handle error or client disconnect
			log.Printf("Error reading from connection: %v", err)
			break
		}

		cmd := ParseCommand(rawCommand)
		if cmd == nil {
			continue
		}

		// SUBSCRIBE <topic>
		// PUBLISH <topic> <message>
		if cmd.Name == "SUBSCRIBE" {
			topic := cmd.CollectionName
			subscribeToTopic(topic, conn, pubSub, cc)
		} else if cmd.Name == "PUBLISH" {
			topic := cmd.CollectionName
			message := strings.Join(cmd.Args[0:], " ")
			publishToTopic(topic, message, conn, pubSub)
		} else {
			conn.Write([]byte("Unknown command in pub/sub mode.\n"))
		}
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
