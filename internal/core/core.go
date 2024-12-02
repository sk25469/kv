package core

import (
	codec_model "github.com/sk25469/kv/internal/codec/model"
	"github.com/sk25469/kv/internal/comm"
	"github.com/sk25469/kv/internal/storage"
)

type ICore interface {
	RunCommand(interface{}) ([]byte, error)
}

type CoreServiceParams struct {
	StorageLayer       storage.IStorage
	CommunicationLayer comm.ICommunication
}

type CoreService struct {
	storageLayer       storage.IStorage
	communicationLayer *comm.CommunicationService
}

func NewCoreService() *CoreService {
	return &CoreService{
		storageLayer:       storage.NewInMemoryHashMap(),
		communicationLayer: comm.NewCommunicationService(),
	}
}

func (c *CoreService) RunCommand(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case *codec_model.Command:
		switch v.Type {
		case codec_model.Set:
			err := c.storageLayer.Set(v.Key, v.Value)
			if err != nil {
				return nil, err
			}
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
			return []byte("delete successfull"), nil
		}
	case *codec_model.CommunicationModel:
		switch v.Command {
		case codec_model.IAM:
			// node added in the network layer
			err := c.communicationLayer.AddNode(v.SendTo.ID, v.SendTo)
			if err != nil {
				return nil, err
			}
			return []byte("node added successfully"), nil
		}
	}
	return nil, nil
}
