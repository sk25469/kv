package network

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/sk25469/kv/internal/codec"
	"github.com/sk25469/kv/internal/comm"
	"github.com/sk25469/kv/internal/core"
	network "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/logger"
	"github.com/sk25469/kv/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	etcdClient         *clientv3.Client
}

func NewNetworkService(params NetworkServiceParams) *NetworkService {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   params.NodeConfig.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Error initializing etcd client: %v", err)
	}
	return &NetworkService{
		nodeConfig:         params.NodeConfig,
		coreLayer:          core.NewCoreService(),
		codecLayer:         codec.NewCodecLayerService(),
		communicationLayer: comm.NewCommunicationService(),
		etcdClient:         etcdClient,
	}
}

func (n *NetworkService) Start() error {
	// Start TCP server
	ctx := context.Background()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", n.nodeConfig.Port))
	if err != nil {
		log.Error("Error starting server:", err)
		return err
	}
	defer listener.Close()
	log.Infof("Server is listening on port %v...\n", n.nodeConfig.Port)

	nodeId := n.nodeConfig.SetNodeID()

	// Register the node with etcd
	err = n.registerNode()
	if err != nil {
		log.Errorf("Error registering node with etcd: %v", err)
		return err
	}

	// Discover other nodes from etcd
	err = n.discoverNodes()
	if err != nil {
		log.Errorf("Error discovering nodes from etcd: %v", err)
		return err
	}

	// add the node to its own topology map
	err = n.communicationLayer.AddNode(nodeId, n.nodeConfig)
	if err != nil {
		log.Errorf("Error adding node to the topology map: %v", err)
		return err
	}

	// choose a master
	masterNode := n.communicationLayer.ChooseLeader()
	log.Infof("Master node: %v\n", masterNode)

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
		go n.handleConnection(conn)
	}
}

func (n *NetworkService) Stop() error {
	// remove the node from the topology map
	n.communicationLayer.RemoveNode(n.nodeConfig.ID)
	// stop the server
	return nil
}

func (n *NetworkService) handleConnection(conn net.Conn) {
	// handle connection
	defer conn.Close()
	log.Infof("Connection from %v\n", conn.RemoteAddr().String())
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
		cmd, err := n.codecLayer.Encode(command, n.nodeConfig, nil)
		if err != nil {
			log.Printf("error encoding command: %v", err)
		}
		log.Infof("encoded command: %v", cmd)
		res, err := n.coreLayer.RunCommand(cmd)
		if err != nil {
			log.Errorf("error running command: %v", err)
		}
		_, err = fmt.Fprintln(conn, string(res))
		if err != nil {
			log.Errorf("error writing to the connection: %v : [%v]", conn, err)
		}
	}
}

func (n *NetworkService) registerNode() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("%v/%s", utils.KV_ETCD_ENDPOINT, n.nodeConfig.ID)
	value := fmt.Sprintf("%v:%v", n.nodeConfig.IP, n.nodeConfig.Port)

	_, err := n.etcdClient.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to register node with etcd: %v", err)
	}

	return nil
}

func (n *NetworkService) discoverNodes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := n.etcdClient.Get(ctx, utils.KV_ETCD_ENDPOINT, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to discover nodes from etcd: %v", err)
	}

	for _, kv := range resp.Kvs {
		nodeID := string(kv.Key[len(utils.KV_ETCD_ENDPOINT):])
		address := string(kv.Value)

		// Add node to the topology map
		nodeConfig := network.NodeConfig{
			ID: nodeID,
			IP: address,
		}
		err := n.communicationLayer.AddNode(nodeID, &nodeConfig)
		if err != nil {
			log.Errorf("Error adding node to the topology map: %v", err)
		}
	}
	return nil
}
