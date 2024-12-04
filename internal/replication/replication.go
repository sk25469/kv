package replication

import (
	"github.com/sk25469/kv/internal/comm"
	network "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/logger"
)

var log = logger.NewPackageLogger("replication")

type IReplication interface {
	ReplicateData(nodeConfig *network.NodeConfig, id string, data []byte) error
}

type ReplicationServiceParams struct {
	CommunicationLayer comm.ICommunication
}

type ReplicationService struct {
	communicationService *comm.CommunicationService
	bloomFilter          *BloomFilter
}

func NewReplicationService(params ReplicationServiceParams) *ReplicationService {
	return &ReplicationService{
		communicationService: params.CommunicationLayer.(*comm.CommunicationService),
		bloomFilter:          NewBloomFilter(),
	}
}

func (r *ReplicationService) ReplicateData(nodeConfig *network.NodeConfig, id string, data []byte) error {
	// Check if the data is already present in the Bloom filter
	if r.bloomFilter.Check(string(data)) {
		log.Infof("Data %s already present", string(data))
		return nil
	}

	// If not present, broadcast the data
	err := r.communicationService.BroadcastMessage(nodeConfig.ID, id, data)
	if err != nil {
		return err
	}

	// Update the Bloom filter
	r.bloomFilter.Add(string(data))
	return nil
}
