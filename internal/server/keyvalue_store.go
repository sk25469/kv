package server

import (
	"log"
	"sync"
	"time"

	models "github.com/sk25469/kv/internal/model"
)

// KeyValueStore represents the in-memory key-value store
type KeyValueStore struct {
	mu    sync.RWMutex
	store map[string]*models.KeyValue
}

// NewKeyValueStore creates a new instance of KeyValueStore
func NewKeyValueStore() *KeyValueStore {
	return &KeyValueStore{
		store: make(map[string]*models.KeyValue),
	}
}

func (kv *KeyValueStore) UpdateKeyWithTTL(key string, ttl time.Duration) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	keyValue := kv.store[key]
	keyValue.SetExpiration(ttl)
	kv.store[key] = keyValue
}

// set sets a key-value pair with TTL
func (kv *KeyValueStore) SetKeyWithTTL(key, value string, ttl time.Duration) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	keyValue := models.NewKeyValue()
	keyValue.Value = value
	keyValue.SetExpiration(ttl)
	kv.store[key] = keyValue
}

// Set sets a key-value pair in the store
func (kv *KeyValueStore) Set(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	keyValue := models.NewKeyValue()
	keyValue.Value = value
	kv.store[key] = keyValue
}

// Get retrieves the value for a given key from the store
func (kv *KeyValueStore) Get(key string) string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	log.Printf("value for key: %v = %v", key, kv.store[key])
	return kv.store[key].Value
}

// Delete deletes a key from the store
func (kv *KeyValueStore) Delete(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.store, key)
}

func (kv *KeyValueStore) StartExpiryCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			kv.mu.Lock()
			for key, entry := range kv.store {
				if time.Now().After(entry.GetExpiration()) {
					delete(kv.store, key)
				}
			}
			kv.mu.Unlock()
		}
	}
}
