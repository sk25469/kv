package server

import (
	"bufio"
	"fmt"
	"log"
	"net"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/internal/utils"
)

// Start initializes the server
func Start(config *models.Config) {
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

	err = handleInitLoad(cs)
	if err != nil {
		log.Printf("error loading dump: %v", err)
		return
	}

	log.Printf("starting TTL cleanups")
	StartKVCleanup(cs, utils.CLEANUP_DURATION)

	// Accept client connections
	kvServer := models.NewKVServer(config)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("connected with client: %v", conn)
		go handleConnection(conn, cs, kvServer)
	}
}

// handleConnection handles client connections
func handleConnection(conn net.Conn, cs *models.CollectionStore, kvServer *models.KVServer) {

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
		cmd := ParseCommand(command)
		if ShouldWriteLog(*cmd) {
			err = WriteCommandsToFile(*cmd, utils.DUMP_FILE_NAME)
			if err != nil {
				log.Printf("error writing operation to dump")
			}
		}
		result := ExecuteCommand(cmd, cs, ts, clientConfig, kvServer)
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
			result := ExecuteCommand(&cmd, cs, nil, &models.ClientConfig{ClientState: &models.ClientState{State: utils.ACTIVE, IsAuthenticated: true}}, &models.KVServer{Config: &models.Config{ProtectedMode: false}})
			log.Printf("successfully executed curr cmd: %v ------------ %v", cmd, result)
		}
	}
	return nil
}
