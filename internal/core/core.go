package core

import (
	codec_model "github.com/sk25469/kv/internal/codec/model"
	"github.com/sk25469/kv/internal/comm"
	"github.com/sk25469/kv/internal/middleware"
	network "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/internal/replication"
	"github.com/sk25469/kv/logger"
)

var log = logger.NewPackageLogger("core")

type ICore interface {
	RunCommand(interface{}, *network.NodeConfig) ([]byte, error)
}

type CoreServiceParams struct {
	StorageLayer       *middleware.StorageMiddleware
	CommunicationLayer comm.ICommunication
	ReplicationLayer   replication.IReplication
}

type CoreService struct {
	storageLayer       *middleware.StorageMiddleware
	communicationLayer *comm.CommunicationService
	replicationLayer   *replication.ReplicationService
}

func NewCoreService(params CoreServiceParams) *CoreService {
	return &CoreService{
		storageLayer:       params.StorageLayer,
		communicationLayer: params.CommunicationLayer.(*comm.CommunicationService),
		replicationLayer:   params.ReplicationLayer.(*replication.ReplicationService),
	}
}

func (c *CoreService) RunCommand(data interface{}, nodeConfig *network.NodeConfig) ([]byte, error) {
	switch v := data.(type) {
	case *codec_model.Command:
		cmdInBytes := v.Decode()
		switch v.Type {
		case codec_model.Set:
			err := c.storageLayer.Set(v.Key, v.Value)
			if err != nil {
				return nil, err
			}
			c.replicationLayer.ReplicateData(nodeConfig, v.ID.String(), cmdInBytes)
			return []byte("write successfull"), nil
		case codec_model.Get:
			res, err := c.storageLayer.Get(v.Key)
			if err != nil {
				return nil, err
			}
			return []byte(res), nil
		case codec_model.Delete:
			err := c.storageLayer.Delete(v.Key)
			if err != nil {
				return nil, err
			}
			c.replicationLayer.ReplicateData(nodeConfig, v.ID.String(), cmdInBytes)

			return []byte("delete successfull"), nil
		}
	case *codec_model.CommunicationModel:
		switch v.Command {
		case codec_model.IAM:
			c.communicationLayer.AddNode(v.SentFrom.ID, v.SentFrom)
			log.Infof("Node added to topology map: %v", v.SentFrom.ID)
		}
	}
	return nil, nil
}
