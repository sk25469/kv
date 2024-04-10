package models

import (
	"encoding/json"
)

type Node struct {
	MasterConfPath string   `json:"master_conf_path"`
	SlaveConfPath  []string `json:"slaves_conf_path"`
}

type ShardDbConfig struct {
	ShardID      int    `json:"shard_id"`
	SnapshotPath string `json:"snapshot_path"`
	Nodes        *Node  `json:"nodes"`
}

type ShardConfig struct {
	NumberOfShards int              `json:"number_of_shards"`
	ShardList      []*ShardDbConfig `json:"shards"`
}

func (sc *ShardDbConfig) GetShardSlavesPathList() []string {
	return sc.Nodes.SlaveConfPath
}

func (sc *ShardDbConfig) GetShardMasterPath() string {
	return sc.Nodes.MasterConfPath
}

func (sc *ShardDbConfig) GetSnapshotPath() string {
	return sc.SnapshotPath
}

func (sc *ShardConfig) JsonMarshal() ([]byte, error) {
	return json.Marshal(sc)
}

func (sc *ShardConfig) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, sc)
}
