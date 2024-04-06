package server

import (
	"errors"
	"fmt"
	"sync"
)

type TransactionalKeyValueStore struct {
	data   map[string]string
	mutex  sync.Mutex
	logger *TransactionLogger
}

type TransactionLogger struct {
	logs []string
}

func NewTransactionalKeyValueStore() *TransactionalKeyValueStore {
	return &TransactionalKeyValueStore{
		data:   make(map[string]string),
		logger: &TransactionLogger{},
	}
}

func (kv *TransactionalKeyValueStore) BeginTransaction() {
	kv.logger.logs = nil // Clear transaction log
}

func (kv *TransactionalKeyValueStore) CommitTransaction() {
	kv.logger.logs = nil // Clear transaction log
}

func (kv *TransactionalKeyValueStore) RollbackTransaction() {
	for _, log := range kv.logger.logs {
		parts := kv.parseLog(log)
		if parts == nil {
			continue
		}
		key := parts[0]
		value := parts[1]
		kv.data[key] = value
	}
	kv.logger.logs = nil // Clear transaction log
}

func (kv *TransactionalKeyValueStore) parseLog(log string) []string {
	// Parse log format, e.g., "SET key value"
	return nil
}

func (kv *TransactionalKeyValueStore) Set(key, value string) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	kv.data[key] = value
	kv.logger.logs = append(kv.logger.logs, fmt.Sprintf("SET %s %s", key, value))
}

func (kv *TransactionalKeyValueStore) Get(key string) (string, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	value, ok := kv.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return value, nil
}
