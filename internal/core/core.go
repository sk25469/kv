package core

import (
	codec_model "github.com/sk25469/kv/internal/codec/model"
	"github.com/sk25469/kv/internal/storage"
)

type ICore interface {
	RunCommand(*codec_model.Command) ([]byte, error)
}

type CoreServiceParams struct {
	StorageLayer storage.IStorage
}

type CoreService struct {
	storageLayer storage.IStorage
}

func NewCoreService() *CoreService {
	return &CoreService{
		storageLayer: storage.NewInMemoryHashMap(),
	}
}

func (c *CoreService) RunCommand(cmd *codec_model.Command) ([]byte, error) {
	switch cmd.Type {
	case codec_model.Set:
		err := c.storageLayer.Set(cmd.Key, cmd.Value)
		if err != nil {
			return nil, err
		}
		return []byte("write successfull"), nil
	case codec_model.Get:
		res, err := c.storageLayer.Get(cmd.Key)
		if err != nil {
			return nil, err
		}
		return []byte(res), nil
	case codec_model.Delete:
		err := c.storageLayer.Delete(cmd.Key)
		if err != nil {
			return nil, err
		}
		return []byte("delete successfull"), nil
	}
	return nil, nil
}
