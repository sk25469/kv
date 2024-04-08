package main

import (
	"fmt"
	"sync"
	"testing"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/internal/server"
)

func BenchmarkParseCommand(b *testing.B) {
	// Raw command string to parse
	rawCommand := "SET collection1 key1 value1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform command parsing
		server.ParseCommand(rawCommand)
	}
}

func BenchmarkExecuteCommand(b *testing.B) {
	// Initialize your key-value database
	cs := models.NewCollectionStore()
	ts := models.NewTransactionalKeyValueStore()

	// Raw command string to parse
	rawCommand := "SET collection1 key1 value1"
	cmd := server.ParseCommand(rawCommand)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform command execution
		server.ExecuteCommand(cmd, cs, ts, nil, nil)
	}
}

func BenchmarkBeginTransaction(b *testing.B) {
	// Initialize your key-value database
	ts := models.NewTransactionalKeyValueStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform begin transaction operation
		ts.BeginTransaction()
	}
}

func BenchmarkExecTransaction(b *testing.B) {
	// Initialize your key-value database
	ts := models.NewTransactionalKeyValueStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform exec transaction operation
		ts.ExecTransaction()
	}
}

func BenchmarkRollbackTransaction(b *testing.B) {
	// Initialize your key-value database
	ts := models.NewTransactionalKeyValueStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform rollback transaction operation
		ts.RollbackTransaction()
	}
}

func BenchmarkSetTransaction(b *testing.B) {
	// Initialize your key-value database
	ts := models.NewTransactionalKeyValueStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform set operation
		ts.Set(fmt.Sprintf("col%d", i), fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
}

func BenchmarkGetTransaction(b *testing.B) {
	// Initialize your key-value database
	ts := models.NewTransactionalKeyValueStore()

	// Preload the database with test data
	for i := 0; i < b.N; i++ {
		ts.Set(fmt.Sprintf("col%d", i), fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))

	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform get operation
		ts.Get(fmt.Sprintf("col%d", i), fmt.Sprintf("key%d", i))
	}
}

func BenchmarkKTransaction(b *testing.B) {
	// Initialize your key-value database
	ts := models.NewTransactionalKeyValueStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform set operation
		ts.Set(fmt.Sprintf("col%d", i), fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		ts.Get(fmt.Sprintf("col%d", i), fmt.Sprintf("key%d", i))
		ts.BeginTransaction()
		ts.ExecTransaction()
		ts.RollbackTransaction()
	}
}

func BenchmarkSet(b *testing.B) {
	// Initialize your key-value database
	cs := models.NewCollectionStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform set operation
		cs.SetKeyInCollection(fmt.Sprintf("collection%d", i%b.N), fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
}

func BenchmarkGet(b *testing.B) {
	// Initialize your key-value database
	cs := models.NewCollectionStore()

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
	cs := models.NewCollectionStore()

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
