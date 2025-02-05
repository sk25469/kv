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
	SendMessage(nodeConfig *network.NodeConfig, id string, message []byte) error
	// Broadcast a message to all nodes
	BroadcastMessage(currentNodeID string, id string, message []byte) error
	// choose a leader
	ChooseLeader() string

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

func (c *CommunicationService) SendMessage(nodeConfig *network.NodeConfig, id string, message []byte) error {
	c.sendMessage(nodeConfig, id, message)
	return nil
}

func (c *CommunicationService) BroadcastMessage(currentNodeID string, id string, message []byte) error {
	for _, node := range c.GetTopologyMap().GetNodes() {
		if node.ID != currentNodeID {
			c.SendMessage(node, id, message)
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
	c.ChooseLeader()
	// delete the key from etcd
	nodeConfig, err := c.GetNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node config: %v", err)
	}
	c.removeNode(nodeConfig)
	c.topologyMap.RemoveNode(nodeID)

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

func (c *CommunicationService) sendMessage(node *network.NodeConfig, msgId string, req []byte) {
	go func(node *network.NodeConfig) {
		conn, err := net.Dial("tcp", node.IP+":"+node.Port)
		if err != nil {
			log.Errorf("Error connecting to node %v: %v", node.ID, err)
			return
		}
		defer conn.Close()

		_, err = conn.Write(req)
		if err != nil {
			log.Errorf("Error sending IAM message to node %v: %v", node.ID, err)
		}
		log.Info("Message sent to node: ", node.ID)
	}(node)

}

func (c *CommunicationService) registerNode(nodeConfig *network.NodeConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), utils.DEFAULT_CTX_TIMEOUT)
	defer cancel()

	err := c.AddNode(nodeConfig.ID, nodeConfig)
	if err != nil {
		return fmt.Errorf("failed to add node to topology map: %v", err)
	}

	c.ChooseLeader()

	key := fmt.Sprintf("%v%s", utils.KV_ETCD_KEY, nodeConfig.ID)
	updatedNodeConfig, err := c.GetNode(nodeConfig.ID)
	if err != nil {
		return fmt.Errorf("failed to get node config: %v", err)
	}

	value := updatedNodeConfig.ToJson()

	_, err = c.etcdClient.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to register node with etcd: %v", err)
	}

	// send IAM message to all nodes

	for _, node := range c.GetTopologyMap().GetNodes() {
		if node.ID != updatedNodeConfig.ID {
			c.sendIAMMessage(node, &codec_model.CommunicationModel{
				Command:  codec_model.IAM,
				SendTo:   node,
				SentFrom: updatedNodeConfig,
			})
		}
	}

	return nil
}

func (c *CommunicationService) discoverNodes() error {
	ctx, cancel := context.WithTimeout(context.Background(), utils.DEFAULT_CTX_TIMEOUT)
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
	ctx, cancel := context.WithTimeout(context.Background(), utils.DEFAULT_CTX_TIMEOUT)
	defer cancel()

	key := fmt.Sprintf("%v%s", utils.KV_ETCD_KEY, nodeConfig.ID)

	_, err := c.etcdClient.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to remove node from etcd: %v", err)
	}

	return nil
}

func (c *CommunicationService) sendIAMMessage(nodeConfig *network.NodeConfig, message interface{}) {
	req := message.(*codec_model.CommunicationModel)
	switch req.Command {
	case codec_model.IAM:
		messageToSend, err := req.ToBytes()
		if err != nil {
			log.Errorf("Error marshalling IAM message: %v", err)
		}
		c.sendMessage(nodeConfig, req.ID.String(), messageToSend)
	}
}

func (c *CommunicationService) GetEtcdClientHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), utils.DEFAULT_CTX_TIMEOUT)
	defer cancel()

	_, err := c.etcdClient.Get(ctx, "health")
	if err != nil {
		return fmt.Errorf("failed to get etcd client health: %v", err)
	}
	return nil
}
