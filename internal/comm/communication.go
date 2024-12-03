package comm

import (
	"context"
	"fmt"
	"net"
	"time"

	codec_model "github.com/sk25469/kv/internal/codec/model"
	network "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/logger"
	"github.com/sk25469/kv/utils"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var log = logger.NewPackageLogger("comm")

type ICommunication interface {
	// Send a message to a specific node
	SendMessage(nodeConfig *network.NodeConfig, message interface{}) error
	// Broadcast a message to all nodes
	BroadcastMessage(currentNodeID string, message interface{}) error
	// choose a leader
	ChooseLeader(nodes *network.TopologyMap) string

	AddNode(nodeID string, node *network.NodeConfig) error

	RemoveNode(nodeID string) error

	GetNode(nodeID string) (*network.NodeConfig, error)

	GetMasterNode() (*network.NodeConfig, bool)

	GetTopologyMap() *network.TopologyMap

	// register node with etcd
	RegisterNode(nodeConfig *network.NodeConfig) error

	// discover nodes from etcd
	DiscoverNodes() error
}

type CommunicationServiceParams struct {
}

type CommunicationService struct {
	topologyMap *network.TopologyMap //  map of nodes in the network
	etcdClient  *clientv3.Client
}

func NewCommunicationService() *CommunicationService {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   utils.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Error initializing etcd client: %v", err)
	}
	return &CommunicationService{
		topologyMap: network.NewTopologyMap(),
		etcdClient:  etcdClient,
	}
}

func (c *CommunicationService) SendMessage(nodeConfig *network.NodeConfig, message interface{}) error {
	req := message.(*codec_model.CommunicationModel)
	switch req.Command {
	case codec_model.IAM:
		c.sendMessage(nodeConfig, req)
	}
	return nil
}

func (c *CommunicationService) BroadcastMessage(currentNodeID string, message interface{}) error {
	for _, node := range c.GetTopologyMap().GetNodes() {
		if node.ID != currentNodeID {
			c.SendMessage(node, message)
		}
	}
	return nil
}

func (c *CommunicationService) ChooseLeader() string {
	leader, hasLeader := c.GetMasterNode()
	if hasLeader {
		return leader.ID
	}

	// otherwise choose a leader
	allNodes := c.GetTopologyMap().GetNodes()
	for _, node := range allNodes {
		if !node.IsMaster {
			node.IsMaster = true
			c.AddNode(node.ID, node)
			return node.ID
		}
	}
	return ""
}

func (c *CommunicationService) AddNode(nodeID string, node *network.NodeConfig) error {
	c.topologyMap.AddNode(nodeID, node)
	return nil
}

func (c *CommunicationService) RemoveNode(nodeID string) error {
	c.topologyMap.RemoveNode(nodeID)
	c.ChooseLeader()
	// delete the key from etcd
	nodeConfig, err := c.GetNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node config: %v", err)
	}
	c.removeNode(nodeConfig)

	return nil
}

func (c *CommunicationService) GetNode(nodeID string) (*network.NodeConfig, error) {
	node, ok := c.topologyMap.GetNode(nodeID)
	if ok {
		return node, nil
	}
	return nil, nil
}

func (c *CommunicationService) GetMasterNode() (*network.NodeConfig, bool) {
	return c.topologyMap.GetMasterNode()
}

func (c *CommunicationService) GetTopologyMap() *network.TopologyMap {
	return c.topologyMap
}

func (c *CommunicationService) RegisterNode(nodeConfig *network.NodeConfig) error {
	return c.registerNode(nodeConfig)
}

func (c *CommunicationService) DiscoverNodes() error {
	return c.discoverNodes()
}

func (c *CommunicationService) sendMessage(node *network.NodeConfig, req *codec_model.CommunicationModel) {
	go func(node *network.NodeConfig) {
		conn, err := net.Dial("tcp", node.IP+":"+node.Port)
		if err != nil {
			log.Errorf("Error connecting to node %v: %v", node.ID, err)
			return
		}
		defer conn.Close()

		reqInBytes, err := req.Decode()
		if err != nil {
			log.Errorf("Error decoding IAM message: %v", err)
		}

		_, err = conn.Write(reqInBytes)
		if err != nil {
			log.Errorf("Error sending IAM message to node %v: %v", node.ID, err)
		}
	}(node)

}

func (c *CommunicationService) registerNode(nodeConfig *network.NodeConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	err := c.AddNode(nodeConfig.ID, nodeConfig)
	if err != nil {
		return fmt.Errorf("failed to add node to topology map: %v", err)
	}

	c.ChooseLeader()

	key := fmt.Sprintf("%v/%s", utils.KV_ETCD_KEY, nodeConfig.ID)
	updatedNodeConfig, err := c.GetNode(nodeConfig.ID)
	if err != nil {
		return fmt.Errorf("failed to get node config: %v", err)
	}

	value := updatedNodeConfig.ToJson()

	_, err = c.etcdClient.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to register node with etcd: %v", err)
	}

	return nil
}

func (c *CommunicationService) discoverNodes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	resp, err := c.etcdClient.Get(ctx, utils.KV_ETCD_KEY, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to discover nodes from etcd: %v", err)
	}

	for _, kv := range resp.Kvs {
		nodeID := string(kv.Key[len(utils.KV_ETCD_KEY):])
		nodeConfigString := kv.Value

		nodeConfig, err := network.FromJSON(nodeConfigString)
		if err != nil {
			log.Errorf("Error unmarshalling node config: %v", err)
		}

		err = c.AddNode(nodeID, nodeConfig)
		if err != nil {
			log.Errorf("Error adding node to the topology map: %v", err)
		}
	}
	return nil
}

func (c *CommunicationService) removeNode(nodeConfig *network.NodeConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	key := fmt.Sprintf("%v/%s", utils.KV_ETCD_KEY, nodeConfig.ID)

	_, err := c.etcdClient.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to remove node from etcd: %v", err)
	}

	return nil
}
