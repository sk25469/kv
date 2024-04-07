package server

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	models "github.com/sk25469/kv/internal/model"
)

type TransactionalKeyValueStore struct {
	data   map[string]*KeyValueStore
	mutex  sync.Mutex
	logger *TransactionLogger
}

type TransactionLogger struct {
	logs []string
}

func NewTransactionalKeyValueStore() *TransactionalKeyValueStore {
	return &TransactionalKeyValueStore{
		data:   make(map[string]*KeyValueStore),
		logger: &TransactionLogger{},
	}
}

func (kv *TransactionalKeyValueStore) BeginTransaction() {
	kv.logger.logs = nil // Clear transaction log
}

func (kv *TransactionalKeyValueStore) ExecTransaction() {
	kv.logger.logs = nil // Clear transaction log
}

func (kv *TransactionalKeyValueStore) RollbackTransaction() {
	for i := len(kv.logger.logs) - 1; i >= 0; i-- {
		log := kv.logger.logs[i]
		parts := kv.parseLog(log)
		if parts == nil {
			continue
		}
		collection := parts[1]
		key := parts[2]
		prevValue := parts[4]
		kvStore := kv.data[collection]
		kvStore.store[key] = models.NewKeyValue(prevValue)
	}
	kv.logger.logs = nil // Clear transaction log
}

func (kv *TransactionalKeyValueStore) parseLog(log string) []string {
	return strings.Split(log, " ")
}

func (kv *TransactionalKeyValueStore) Set(collection, key, value string) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	/// get the previous value
	prevKvStore, ok := kv.data[collection]
	prevValue := ""
	if ok {
		log.Printf("prevKvStore: %v", prevKvStore.store)
		kvStoreMap, ok := prevKvStore.store[key]
		if ok {
			prevValue = kvStoreMap.Value
		}
	}

	/// set the new value
	kvStore, ok := kv.data[collection]
	if !ok {
		log.Printf("creating new kvStore for collection: %s", collection)
		kvStore = NewKeyValueStore()
		kvStore.store[key] = models.NewKeyValue(value)
	} else {
		kvStore.store[key] = models.NewKeyValue(value)
	}
	kv.data[collection] = kvStore
	log.Printf("kvStore: %v", kvStore.store)
	kv.logger.logs = append(kv.logger.logs, fmt.Sprintf("SET %s %s %s %s", collection, key, value, prevValue))
}

func (kv *TransactionalKeyValueStore) Get(collection, key string) (string, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	colKv, ok := kv.data[collection]
	if !ok {
		return "", errors.New("key not found in collection")
	}
	log.Printf("colKv: %v", colKv)

	value, ok := colKv.store[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return value.Value, nil
}
