// file_hashmap.go
// TODO: Implement a file-based hashmap storage engine
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type FileHashMap struct {
	filepath string
	data     map[string]string
	mu       sync.RWMutex
}

func NewFileHashMap(filepath string) (*FileHashMap, error) {
	f := &FileHashMap{
		filepath: filepath,
		data:     make(map[string]string),
	}

	// Load existing data if file exists
	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		data, err := os.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &f.data); err != nil {
			return nil, err
		}
	}
	return f, nil
}

func (f *FileHashMap) Set(key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.data[key] = value
	return f.persist()
}

func (f *FileHashMap) Get(key string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if val, ok := f.data[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("key not found")
}

func (f *FileHashMap) Delete(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.data, key)
	return f.persist()
}

func (f *FileHashMap) persist() error {
	data, err := json.Marshal(f.data)
	if err != nil {
		return err
	}
	return os.WriteFile(f.filepath, data, 0644)
}
