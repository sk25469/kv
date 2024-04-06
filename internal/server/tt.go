package server

import "time"

func StartKVCleanup(cs *CollectionStore, duration time.Duration) {
	for _, collection := range cs.collections {
		go collection.StartExpiryCleanup(duration)
	}
}
