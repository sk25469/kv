package server

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	models "github.com/sk25469/kv/internal/model"
)

func WatchSnapshotAndUpdate(file string, cs *models.CollectionStore, ts *models.TransactionalKeyValueStore, kvServer *models.KVServer, ps *models.PubSub) {
	// Initialize the file watcher
	err := waitUntilFind(file)
	if err != nil {
		log.Fatalln(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	err = watcher.Add(file)
	if err != nil {
		log.Fatalln(err)
	}

	errCh := make(chan error)

	go handleFileEvent(watcher, file, errCh, cs, ts, kvServer, ps)
	<-errCh

}

func waitUntilFind(filename string) error {
	for {
		time.Sleep(1 * time.Second)
		_, err := os.Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return err
			}
		}
		break
	}
	return nil
}

func handleFileEvent(watcher *fsnotify.Watcher, file string, errCh chan error, cs *models.CollectionStore, ts *models.TransactionalKeyValueStore, kvServer *models.KVServer, ps *models.PubSub) {
	var lastPosition int64 = 0 // Keep track of the last read position
	absFilePath, _ := filepath.Abs(file)
	log.Printf("Absolute path being watched: %s", absFilePath)
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				file, err := os.Open(file)
				if err != nil {
					fmt.Println("Error opening filepath:", err)
					continue
				}

				// Seek to the last known position before reading new entries
				_, err = file.Seek(lastPosition, 0)
				if err != nil {
					fmt.Println("Error seeking file:", err)
					file.Close()
					continue
				}

				// Read new entries from the current position
				lastEntry := ""
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					newEntry := scanner.Text()
					// fmt.Println("New Entry:", newEntry)
					lastEntry = newEntry
					// Here, you would process/store the new entry as needed
				}

				// Update the last known position
				lastPosition, err = file.Seek(0, 1) // 1 means current position
				if err != nil {
					fmt.Println("Error updating last position:", err)
				}

				if err := scanner.Err(); err != nil {
					fmt.Println("Scanner Error:", err)
				}

				log.Printf("Last Entry: %s", lastEntry)
				result := ReplicateChanges(lastEntry, cs, ts, kvServer, ps)
				log.Printf("result for replication: %v -------- %v", lastEntry, result)

				file.Close()
			}
		case err := <-watcher.Errors:
			log.Printf("Error: %v", err)
			errCh <- err
		}
	}
}
