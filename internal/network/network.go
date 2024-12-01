package network

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/sk25469/kv/internal/codec"
	"github.com/sk25469/kv/internal/core"
	network "github.com/sk25469/kv/internal/network/model"
)

// network service interface to be implemented by network service
type INetwork interface {
	Start() error
	Stop() error
	GetTopologyMap() *network.TopologyMap
	GetNodeInfo(nodeIP string) (*network.NodeConfig, error)
}

type NetworkServiceParams struct {
	NodeConfigPath string
	CoreLayer      core.ICore
	CodecLayer     codec.ICodec
}

// NetworkService represents the network service
// whenever a new node is added to the network, it is added to the topology map
// and the data is replicated to the new node
type NetworkService struct {
	nodeConfig  *network.NodeConfig  // node configuration
	topologyMap *network.TopologyMap //  map of nodes in the network
	coreLayer   *core.CoreService
	codecLayer  *codec.CodecLayer
}

func NewNetworkService(params NetworkServiceParams) *NetworkService {
	return &NetworkService{
		nodeConfig:  network.NewNodeConfig(params.NodeConfigPath),
		topologyMap: network.NewTopologyMap(),
		coreLayer:   core.NewCoreService(),
		codecLayer:  codec.NewCodecLayer(),
	}
}

func (n *NetworkService) Start() error {
	// Start TCP server
	ctx := context.Background()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", n.nodeConfig.Port))
	if err != nil {
		log.Println("Error starting server:", err)
		return err
	}
	defer listener.Close()
	log.Printf("Server is listening on port %v...\n", n.nodeConfig.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Printf("Server on port %v shutting down", n.nodeConfig.Port)
				return errors.New("server shutting down")
			default:
				log.Printf("Error accepting connection: %v", err)
				continue
			}
		}
		// handle connection
		go n.handleConnection(conn)
	}
}

func (n *NetworkService) Stop() error {
	return nil
}

func (n *NetworkService) GetTopologyMap() *network.TopologyMap {
	return n.topologyMap
}

func (n *NetworkService) GetNodeInfo(nodeIP string) (*network.NodeConfig, error) {
	return nil, nil
}

func (n *NetworkService) handleConnection(conn net.Conn) {
	// handle connection
	defer conn.Close()
	log.Printf("Connection from %v\n", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	// remoteAddress := conn.RemoteAddr().String()
	// clientId := utils.GenerateBase64ClientID()

	for {
		// Read the next line from the connection
		command, err := reader.ReadString('\n')
		// log.Printf("parsed command: %v", command)
		if err != nil || command == "" {
			// fmt.Println("Error reading from connection:", err)
			return
		}
		cmd, err := n.codecLayer.Encode(command)
		if err != nil {
			log.Printf("error encoding command: %v", err)
		}
		log.Printf("encoded command: %v", cmd)
		res, err := n.coreLayer.RunCommand(cmd)
		if err != nil {
			log.Printf("error running command: %v", err)
		}
		_, err = fmt.Fprintln(conn, string(res))
		if err != nil {
			log.Printf("error writing to the connection: %v : [%v]", conn, err)
		}
	}
}
