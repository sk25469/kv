package comm

import (
	"net"

	codec_model "github.com/sk25469/kv/internal/codec/model"
	network "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/logger"
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
}

type CommunicationServiceParams struct {
}

type CommunicationService struct {
	topologyMap *network.TopologyMap //  map of nodes in the network

}

func NewCommunicationService() *CommunicationService {
	return &CommunicationService{
		topologyMap: network.NewTopologyMap(),
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
