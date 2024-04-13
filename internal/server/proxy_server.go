package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"

	models "github.com/sk25469/kv/internal/model"
)

func RouteRequestsToShards(port string, ch *models.ConsistentHash, shardList *models.ShardsList) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Printf("Server on port %v shutting down", port)
				return // Exit goroutine when context is cancelled
			default:
				log.Printf("Error accepting connection: %v", err)
				continue
			}
		}
		shardID := ch.GetNode(conn.RemoteAddr().String())
		log.Printf("shardID = %v for routing with consistent hash", shardID)
		go sendRequestToShard(shardID, &conn, shardList)
	}
}

func sendRequestToShard(shardID string, conn *net.Conn, shardList *models.ShardsList) {
	// Code
	reader := bufio.NewReader(*conn)
	command, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	shard := shardList.GetShard(shardID)
	shardIP := shard.Nodes[0].Config.IP + ":" + shard.Nodes[0].Config.Port

	shardList.GetShard(shardID).PrintActiveConnections()
	res, err := sendCommand(command, shardIP)
	if err != nil {
		log.Printf("Error sending command to shard: %v", err)
		return
	}

	(*conn).Write([]byte(res))
}

func sendCommand(command string, ip string) (string, error) {
	shardConn, err := net.Dial("tcp", ip)
	if err != nil {
		log.Printf("error connecting to shard: %v", err)
		return "", err
	}
	_, err = shardConn.Write([]byte(command + "\n"))
	if err != nil {
		log.Printf("error writing to shard: %v", err)
		return "", err
	}

	response, err := bufio.NewReader(shardConn).ReadString('\n')
	if err != nil {
		log.Printf("error reading from shard: %v", err)
		return "", err
	}

	return response, nil
}
