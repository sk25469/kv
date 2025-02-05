package network

import "sync"

type TopologyMap struct {
	mut   sync.RWMutex
	nodes map[string]*NodeConfig
}

func NewTopologyMap() *TopologyMap {
	return &TopologyMap{
		mut:   sync.RWMutex{},
		nodes: make(map[string]*NodeConfig),
	}
}

func (t *TopologyMap) AddNode(nodeID string, nodeConfig *NodeConfig) {
	t.mut.Lock()
	defer t.mut.Unlock()
	t.nodes[nodeID] = nodeConfig
}

func (t *TopologyMap) RemoveNode(nodeID string) {
	t.mut.Lock()
	defer t.mut.Unlock()
	delete(t.nodes, nodeID)
}

func (t *TopologyMap) GetNode(nodeID string) (*NodeConfig, bool) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	node, ok := t.nodes[nodeID]
	return node, ok
}

func (t *TopologyMap) GetNodes() map[string]*NodeConfig {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return t.nodes
}

func (t *TopologyMap) GetNodeIDs() []string {
	t.mut.RLock()
	defer t.mut.RUnlock()
	nodeIDs := make([]string, 0)
	for nodeID := range t.nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs
}

func (t *TopologyMap) GetNodeCount() int {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return len(t.nodes)
}

func (t *TopologyMap) Clear() {
	t.mut.Lock()
	defer t.mut.Unlock()
	t.nodes = make(map[string]*NodeConfig)
}

func (t *TopologyMap) IsEmpty() bool {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return len(t.nodes) == 0
}

func (t *TopologyMap) ContainsNode(nodeID string) bool {
	t.mut.RLock()
	defer t.mut.RUnlock()
	_, ok := t.nodes[nodeID]
	return ok
}

func (t *TopologyMap) GetMasterNode() (*NodeConfig, bool) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	for _, node := range t.nodes {
		if node.IsMaster {
			return node, true
		}
	}
	return nil, false
}
