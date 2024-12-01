package network

type TopologyMap struct {
	nodes map[string]*NodeConfig
}

func NewTopologyMap() *TopologyMap {
	return &TopologyMap{
		nodes: make(map[string]*NodeConfig),
	}
}

func (t *TopologyMap) AddNode(nodeID string, nodeConfig *NodeConfig) {
	t.nodes[nodeID] = nodeConfig
}

func (t *TopologyMap) RemoveNode(nodeID string) {
	delete(t.nodes, nodeID)
}

func (t *TopologyMap) GetNode(nodeID string) (*NodeConfig, bool) {
	node, ok := t.nodes[nodeID]
	return node, ok
}

func (t *TopologyMap) GetNodes() map[string]*NodeConfig {
	return t.nodes
}

func (t *TopologyMap) GetNodeIDs() []string {
	nodeIDs := make([]string, 0)
	for nodeID := range t.nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs
}

func (t *TopologyMap) GetNodeCount() int {
	return len(t.nodes)
}

func (t *TopologyMap) Clear() {
	t.nodes = make(map[string]*NodeConfig)
}

func (t *TopologyMap) IsEmpty() bool {
	return len(t.nodes) == 0
}

func (t *TopologyMap) ContainsNode(nodeID string) bool {
	_, ok := t.nodes[nodeID]
	return ok
}

func (t *TopologyMap) GetMasterNode() (*NodeConfig, bool) {
	for _, node := range t.nodes {
		if node.IsMaster {
			return node, true
		}
	}
	return nil, false
}
