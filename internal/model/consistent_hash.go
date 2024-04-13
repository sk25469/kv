package models

import (
	"hash/crc32"
	"sort"
	"sync"
)

type ConsistentHash struct {
	Ring  []int          // Sorted list of hashes
	Nodes map[int]string // Map hash to node ID
	hash  func(data []byte) int
	mutex sync.RWMutex
}

func NewConsistentHash() *ConsistentHash {
	return &ConsistentHash{
		Ring:  make([]int, 0),
		Nodes: make(map[int]string),
		hash: func(data []byte) int {
			return int(crc32.ChecksumIEEE(data))
		},
	}
}

func (ch *ConsistentHash) AddNode(nodeID string) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	hash := ch.hash([]byte(nodeID))
	ch.Nodes[hash] = nodeID
	ch.Ring = append(ch.Ring, hash)
	sort.Ints(ch.Ring)

	// TODO: handle data migration
}

func (ch *ConsistentHash) RemoveNode(nodeID string) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	hash := ch.hash([]byte(nodeID))
	idx := sort.Search(len(ch.Ring), func(i int) bool {
		return ch.Ring[i] == hash
	})
	if idx < len(ch.Ring) && ch.Ring[idx] == hash {
		ch.Ring = append(ch.Ring[:idx], ch.Ring[idx+1:]...)
		delete(ch.Nodes, hash)
	}

	// TODO: handle data migration
}

func (ch *ConsistentHash) GetNode(key string) string {
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()

	if len(ch.Ring) == 0 {
		return ""
	}

	hash := ch.hash([]byte(key))
	idx := sort.Search(len(ch.Ring), func(i int) bool {
		return ch.Ring[i] >= hash
	})

	if idx == len(ch.Ring) {
		idx = 0
	}

	return ch.Nodes[ch.Ring[idx]]
}
