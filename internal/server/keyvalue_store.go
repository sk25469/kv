package server

import (
	"log"
	"sync"
)

// KeyValueStore represents the in-memory key-value store
type KeyValueStore struct {
	mu    sync.RWMutex
	store map[string]string
}

// NewKeyValueStore creates a new instance of KeyValueStore
func NewKeyValueStore() *KeyValueStore {
	return &KeyValueStore{
		store: make(map[string]string),
	}
}

// Set sets a key-value pair in the store
func (kv *KeyValueStore) Set(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.store[key] = value
}

// Get retrieves the value for a given key from the store
func (kv *KeyValueStore) Get(key string) string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	log.Printf("value for key: %v = %v", key, kv.store[key])
	return kv.store[key]
}

// Delete deletes a key from the store
func (kv *KeyValueStore) Delete(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.store, key)
}
