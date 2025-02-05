// memory_hashmap.go
package storage

import (
	"fmt"

	"github.com/sk25469/kv/logger"
)

var log = logger.NewPackageLogger("storage")

type InMemoryHashMap struct {
	data map[string]string
}

func NewInMemoryHashMap() *InMemoryHashMap {
	return &InMemoryHashMap{
		data: make(map[string]string),
	}
}

func (s *InMemoryHashMap) Set(key string, value string) error {
	s.data[key] = value
	return nil
}

func (s *InMemoryHashMap) Get(key string) (string, error) {
	if val, exists := s.data[key]; exists {
		return val, nil
	}
	return "", fmt.Errorf("key not found")
}

func (s *InMemoryHashMap) Delete(key string) error {
	delete(s.data, key)
	return nil
}
