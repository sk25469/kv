package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/sk25469/kv/internal/server"
)

func BenchmarkSet(b *testing.B) {
	// Initialize your key-value database
	cs := server.NewCollectionStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform set operation
		cs.SetKeyInCollection(fmt.Sprintf("collection%d", i%b.N), fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
}

func BenchmarkGet(b *testing.B) {
	// Initialize your key-value database
	cs := server.NewCollectionStore()

	// Preload the database with test data
	for i := 0; i < b.N; i++ {
		cs.SetKeyInCollection(fmt.Sprintf("collection%d", i%b.N), fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform get operation
		cs.GetKeyInCollection(fmt.Sprintf("collection%d", i%b.N), fmt.Sprintf("key%d", i))
	}
}

// Example concurrent read-write test
func TestConcurrentReadWrite(t *testing.T) {
	// Initialize your key-value database
	cs := server.NewCollectionStore()

	// Number of concurrent readers and writers
	numReaders := 100
	numWriters := 50

	// Wait group to synchronize completion of all goroutines
	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	// Concurrent readers
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cs.GetKeyInCollection(fmt.Sprintf("collection%d", j%1000), fmt.Sprintf("key%d", j))
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < numWriters; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cs.SetKeyInCollection(fmt.Sprintf("collection%d", j%1000), fmt.Sprintf("key%d", j), fmt.Sprintf("value%d", j))
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
}
