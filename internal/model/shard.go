package models

import (
	"fmt"
)

type Shard struct {
	// Fields
	ShardID int
	Nodes   []*KVServer
	DbState *DbState
}

type ShardsList struct {
	// Fields
	Shards []*Shard
}

func (shard *Shard) GetNode(ip, port string) *Config {
	// Code
	for _, node := range shard.Nodes {
		if fmt.Sprintf("%v:%v", ip, port) == fmt.Sprintf("%v:%v", node.Config.IP, node.Config.Port) {
			return node.Config
		}
	}
	return nil
}

// NewShard creates a new Shard object
func NewShard(dbState *DbState) *Shard {
	// Code
	return &Shard{
		Nodes:   make([]*KVServer, 0),
		DbState: dbState,
	}
}

func (shard *Shard) AddNode(node *KVServer) {
	// Code
	shard.Nodes = append(shard.Nodes, node)
}

func (shard *Shard) RemoveNode(node *KVServer) {
	// Code
	for i, n := range shard.Nodes {
		if n == node {
			shard.Nodes = append(shard.Nodes[:i], shard.Nodes[i+1:]...)
			break
		}
	}
}

// NewShardsList creates a new ShardsList object
func NewShardsList() *ShardsList {
	// Code
	return &ShardsList{
		Shards: make([]*Shard, 0),
	}
}

func (shardsList *ShardsList) AddShard(shard *Shard) {
	// Code
	shardsList.Shards = append(shardsList.Shards, shard)
}

func (shardsList *ShardsList) RemoveShard(shard *Shard) {
	// Code
	for i, s := range shardsList.Shards {
		if s == shard {
			shardsList.Shards = append(shardsList.Shards[:i], shardsList.Shards[i+1:]...)
			break
		}
	}
}

func (shardsList *ShardsList) GetShard(shardID int) *Shard {
	// Code
	for _, shard := range shardsList.Shards {
		if shard.ShardID == shardID {
			return shard
		}
	}
	return nil
}
