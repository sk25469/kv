package server

import (
	"time"

	models "github.com/sk25469/kv/internal/model"
)

func StartKVCleanup(cs *models.CollectionStore, duration time.Duration) {
	for _, collection := range cs.GetCollection() {
		go collection.StartExpiryCleanup(duration)
	}
}
