package middleware

import (
	"log"
	"time"

	wal "github.com/sk25469/kv/internal/persistence"
	"github.com/sk25469/kv/internal/storage"
)

type StorageMiddleware struct {
	storage storage.IStorage
	wal     wal.WAL
}

func NewStorageMiddleware(storage storage.IStorage, walPath string) (*StorageMiddleware, error) {
	w, err := wal.NewFileWAL(walPath)
	if err != nil {
		return nil, err
	}

	return &StorageMiddleware{
		storage: storage,
		wal:     w,
	}, nil
}

func (sm *StorageMiddleware) Set(key string, value string) error {
	// First append to WAL
	err := sm.wal.AppendLog(wal.LogEntry{
		Operation: wal.SET,
		Key:       key,
		Value:     value,
	})
	if err != nil {
		return err
	}

	// Then perform the actual storage operation
	return sm.storage.Set(key, value)
}

func (sm *StorageMiddleware) Get(key string) (string, error) {
	return sm.storage.Get(key)
}

func (sm *StorageMiddleware) Delete(key string) error {
	err := sm.wal.AppendLog(wal.LogEntry{
		Operation: wal.DELETE,
		Key:       key,
	})
	if err != nil {
		return err
	}

	return sm.storage.Delete(key)
}

func (sm *StorageMiddleware) Recover() error {
	starTime := time.Now()

	entries, err := sm.wal.Recover()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		switch entry.Operation {
		case wal.SET:
			if err := sm.storage.Set(entry.Key, entry.Value); err != nil {
				return err
			}
		case wal.DELETE:
			if err := sm.storage.Delete(entry.Key); err != nil {
				return err
			}
		}
	}

	log.Printf("Recovered %d entries in %v", len(entries), time.Since(starTime))

	return nil
}
