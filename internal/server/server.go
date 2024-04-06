package server

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/sk25469/kv/internal/utils"
)

// Start initializes the server
func Start() {
	// Start TCP server
	listener, err := net.Listen("tcp", ":7000")
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	log.Println("Server is listening on port 7000...")

	log.Printf("creating the collection store")
	cs := NewCollectionStore()

	err = handleInitLoad(cs)
	if err != nil {
		log.Printf("error loading dump: %v", err)
		return
	}

	log.Printf("starting TTL cleanups")
	StartKVCleanup(cs, utils.CLEANUP_DURATION)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("connected with client: %v", conn)
		go handleConnection(conn, cs)
	}
}

// handleConnection handles client connections
func handleConnection(conn net.Conn, cs *CollectionStore) {
	defer conn.Close()

	// Create a bufio reader to read from the connection
	reader := bufio.NewReader(conn)

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
		result := ExecuteCommand(cmd, cs)
		log.Printf("result for cmd: %v -------- %v", cmd, result)
		bytesWritten, err := fmt.Fprintln(conn, result)
		if err != nil {
			log.Printf("error writing to the connection: %v : [%v]", conn, err)
		}
		log.Printf("bytes written to conn: %v ----------- %v", conn, bytesWritten)
	}
}

func handleInitLoad(cs *CollectionStore) error {
	cmds, err := ReadCommandsFromFile(utils.DUMP_FILE_NAME)
	if err != nil {
		log.Printf("error reading cmds from file: [%v]", err)
		return err
	}
	for _, cmd := range cmds {
		if ShouldWriteLog(cmd) {
			result := ExecuteCommand(&cmd, cs)
			log.Printf("successfully executed curr cmd: %v ------------ %v", cmd, result)
		}
	}
	return nil
}
