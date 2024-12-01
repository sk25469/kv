package storage

import (
	"fmt"

	storage "github.com/sk25469/kv/internal/storage/model"
)

type IStorage interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

type StorageServiceParams struct {
	Type      storage.StorageType
	Structure storage.StorageStructure
	FilePath  string // Used for file-based storage
	MaxSize   int    // Optional size limit
}

func NewStorage(params StorageServiceParams) (IStorage, error) {
	switch params.Type {
	case storage.InMemory:
		switch params.Structure {
		case storage.HashMap:
			return NewInMemoryHashMap(), nil
		case storage.BPlusTree:
			return NewInMemoryBPlusTree(params.MaxSize), nil
		}
	case storage.FileBase:
		switch params.Structure {
		case storage.HashMap:
			return NewFileHashMap(params.FilePath)
		case storage.BPlusTree:
			return NewFileBPlusTree(params.FilePath, params.MaxSize)
		}
	}
	return nil, fmt.Errorf("unsupported storage configuration")
}
