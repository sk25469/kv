package models

import (
	"log"
	"sync"
	"time"
)

// CollectionStore represents a collection with a key-value store
type CollectionStore struct {
	KeyValueStore *KeyValueStore
	collections   map[string]*KeyValueStore // Map to store collections
	mu            sync.RWMutex              // Mutex for thread-safe access to collections map
}

// NewCollectionStore creates a new CollectionStore instance
func NewCollectionStore() *CollectionStore {
	return &CollectionStore{
		KeyValueStore: NewKeyValueStore(),
		collections:   make(map[string]*KeyValueStore),
	}
}

func (cs *CollectionStore) GetCollection() map[string]*KeyValueStore {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.collections
}

func (cs *CollectionStore) UpdateKeyInCollectionWithTTL(collectionName, key string, ttl time.Duration) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Check if the collection exists
	coll, ok := cs.collections[collectionName]
	if !ok {
		// Create a new collection if it doesn't exist
		log.Printf("collection with %v doesn't exist, creating...", collectionName)
	}

	// Set the key-value pair in the collection
	coll.UpdateKeyWithTTL(key, ttl)
}

// SetKeyInCollection sets a key-value pair in the specified collection
func (cs *CollectionStore) SetKeyInCollection(collectionName, key, value string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Check if the collection exists
	coll, ok := cs.collections[collectionName]
	if !ok {
		// Create a new collection if it doesn't exist
		log.Printf("collection with %v doesn't exist, creating...", collectionName)
		coll = NewKeyValueStore()
		cs.collections[collectionName] = coll
	}

	// Set the key-value pair in the collection
	coll.Set(key, value)
}

// GetKeyInCollection retrieves the value for a key in the specified collection
func (cs *CollectionStore) GetKeyInCollection(collectionName, key string) string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// Check if the collection exists
	coll, ok := cs.collections[collectionName]
	if !ok {
		log.Printf("collection with %v not found", collectionName)
		return "" // Collection not found
	}

	// Get the value from the collection
	return coll.Get(key)
}

// DeleteKeyInCollection deletes a key from the specified collection
func (cs *CollectionStore) DeleteKeyInCollection(collectionName, key string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Check if the collection exists
	coll, ok := cs.collections[collectionName]
	if !ok {
		return // Collection not found
	}

	// Delete the key from the collection
	coll.Delete(key)
}

// CollectionExists checks if a collection exists
func (cs *CollectionStore) CollectionExists(collectionName string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	_, ok := cs.collections[collectionName]
	return ok
}

// GetAllKeyValues returns all the key-value pairs in all collections
func (cs *CollectionStore) GetAllKeyValues() map[string]map[string]string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make(map[string]map[string]string)

	for collName, coll := range cs.collections {
		keyValuePairs := make(map[string]string)
		coll.mu.RLock()
		for key, value := range coll.store {
			keyValuePairs[key] = value.Value
		}
		coll.mu.RUnlock()
		result[collName] = keyValuePairs
	}

	return result
}

// GetAllKeyValuesInCollection returns all the key-value pairs in a single collection
func (cs *CollectionStore) GetAllKeyValuesInCollection(collectionName string) map[string]string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make(map[string]string)

	// Check if the collection exists
	coll, ok := cs.collections[collectionName]
	if !ok {
		return result // Collection not found, return empty result
	}

	// Acquire read lock on the collection's KeyValueStore
	coll.mu.RLock()
	defer coll.mu.RUnlock()

	// Copy the key-value pairs from the collection's KeyValueStore
	for key, value := range coll.store {
		result[key] = value.Value
	}

	log.Printf("all keys in collection: %v ----------- %v", collectionName, result)

	return result
}
