package server

import (
	"errors"
	"fmt"
	"strings"
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
		key := parts[0]
		prevValue := parts[2]
		kv.data[key] = prevValue
	}
	kv.logger.logs = nil // Clear transaction log
}

func (kv *TransactionalKeyValueStore) parseLog(log string) []string {
	return strings.Split(log, " ")
}

func (kv *TransactionalKeyValueStore) Set(key, value string) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	prevValue := kv.data[key]
	kv.data[key] = value
	kv.logger.logs = append(kv.logger.logs, fmt.Sprintf("SET %s %s %s", key, value, prevValue))
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
