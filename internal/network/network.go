package network

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/sk25469/kv/internal/codec"
	"github.com/sk25469/kv/internal/comm"
	"github.com/sk25469/kv/internal/core"
	network "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/logger"
)

var log = logger.NewPackageLogger("network")

// network service interface to be implemented by network service
type INetwork interface {
	Start() error
	Stop() error
}

type NetworkServiceParams struct {
	NodeConfig         *network.NodeConfig
	CoreLayer          core.ICore
	CodecLayer         codec.ICodec
	CommunicationLayer comm.ICommunication
}

// NetworkService represents the network service
// whenever a new node is added to the network, it is added to the topology map
// and the data is replicated to the new node
type NetworkService struct {
	nodeConfig         *network.NodeConfig // node configuration
	coreLayer          *core.CoreService
	codecLayer         *codec.CodecLayerService
	communicationLayer *comm.CommunicationService
	listener           net.Listener
}

func NewNetworkService(params NetworkServiceParams) *NetworkService {
	return &NetworkService{
		nodeConfig:         params.NodeConfig,
		coreLayer:          params.CoreLayer.(*core.CoreService),
		codecLayer:         params.CodecLayer.(*codec.CodecLayerService),
		communicationLayer: params.CommunicationLayer.(*comm.CommunicationService),
	}
}

func (n *NetworkService) Start() error {
	// Start TCP server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", n.nodeConfig.Port))
	if err != nil {
		log.Error("Error starting server:", err)
		return err
	}
	defer listener.Close()
	n.listener = listener
	log.Infof("Server is listening on port %v...\n", n.nodeConfig.Port)

	// Signal handling for graceful shutdown
	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// go func() {
	// 	<-sigChan
	// 	log.Println("Received shutdown signal")
	// 	n.Stop()
	// 	cancel()
	// }()

	// Discover other nodes from etcd
	err = n.communicationLayer.DiscoverNodes()
	if err != nil {
		log.Errorf("Error discovering nodes from etcd: %v", err)
		return err
	}

	// Register the node with etcd
	err = n.communicationLayer.RegisterNode(n.nodeConfig)
	if err != nil {
		log.Errorf("Error registering node with etcd: %v", err)
		return err
	}

	// // choose a master
	masterNode, _ := n.communicationLayer.GetMasterNode()
	log.Infof("Master node: %v\n", masterNode.ID)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Printf("Server on port %v shutting down", n.nodeConfig.Port)
				return errors.New("server shutting down")
			default:
				log.Errorf("Error accepting connection: %v", err)
				continue
			}
		}
		// handle connection
		go n.handleConnection(ctx, conn)
	}
}

func (n *NetworkService) Stop() error {
	if n.listener != nil {
		n.listener.Close()
	}
	// remove the node from the topology map
	n.communicationLayer.RemoveNode(n.nodeConfig.ID)
	log.Printf("Node %v removed from the topology map\n", n.nodeConfig.ID)
	// stop the server
	return nil
}

func (n *NetworkService) handleConnection(ctx context.Context, conn net.Conn) {
	// handle connection
	defer conn.Close()

	log.Infof("Connection from %v\n", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)

	for {
		// Read the next line from the connection
		command, err := reader.ReadString('\n')
		// log.Printf("parsed command: %v", command)
		if err != nil || command == "" {
			log.Println("Error reading from connection:", err)
			return
		}
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, finishing last command")
			// Process the last command before shutting down
			n.ProcessCommand(command, conn)
			return
		default:
			n.ProcessCommand(command, conn)
		}
	}
}

func (n *NetworkService) ProcessCommand(command string, conn net.Conn) {
	cmd, err := n.codecLayer.Encode(command, n.nodeConfig, nil)
	if err != nil {
		log.Printf("error encoding command: %v", err)
		return
	}
	log.Infof("encoded command: %v", cmd)
	res, err := n.coreLayer.RunCommand(cmd, n.nodeConfig)
	if err != nil {
		log.Errorf("error running command: %v", err)
		return
	}
	_, err = fmt.Fprintln(conn, string(res))
	if err != nil {
		log.Errorf("error writing to the connection: %v : [%v]", conn, err)
	}
}

func (n *NetworkService) IsListenerActive() bool {
	return n.listener != nil
}
