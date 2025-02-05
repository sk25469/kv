package wal

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sk25469/kv/utils"
)

type Operation string

const (
	SET              Operation = utils.SET
	DELETE           Operation = utils.DEL
	DEFAULT_LOG_DIR            = "/var/lib/kvstore/"
	DEFAULT_LOG_FILE           = "wal.log"
)

const (
	FLUSH_INTERVAL    = 1 * time.Second
	COMPACT_INTERVAL  = 1 * time.Second // Adjust based on your needs
	COMPACT_THRESHOLD = 1024 * 1024 * 1 // 0.01MB threshold
)

type LogEntry struct {
	Operation Operation `json:"operation"`
	Key       string    `json:"key"`
	Value     string    `json:"value,omitempty"`
	Sequence  uint64    `json:"sequence"`
}

type WAL interface {
	AppendLog(entry LogEntry) error
	Recover() ([]LogEntry, error)
	Close() error
}

type FileWAL struct {
	file        *os.File
	mu          sync.Mutex
	sequence    uint64
	writeBuffer *bufio.Writer
	stopFlush   chan struct{}
	stopCompact chan struct{}
}

func NewFileWAL(path string) (*FileWAL, error) {

	// create folder if not exists
	if _, err := os.Stat(DEFAULT_LOG_DIR); os.IsNotExist(err) {
		err := os.Mkdir(DEFAULT_LOG_DIR, 0755)
		if err != nil {
			return nil, err
		}
	}

	// Set proper permissions for log file
	logPath := filepath.Join(DEFAULT_LOG_DIR, DEFAULT_LOG_FILE)

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	wal := &FileWAL{
		file:        file,
		writeBuffer: bufio.NewWriter(file),
		stopFlush:   make(chan struct{}),
	}

	// Start periodic flush
	go wal.periodicFlush()
	go wal.periodicCompact()

	return wal, nil
}

func (w *FileWAL) AppendLog(entry LogEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.sequence++
	entry.Sequence = w.sequence

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = w.writeBuffer.Write(append(data, '\n'))
	return err
}

func (w *FileWAL) Recover() ([]LogEntry, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var entries []LogEntry
	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, err
	}

	// TODO: decode each line individually
	decoder := json.NewDecoder(w.file)
	for {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			break
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (w *FileWAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Stop periodic routines
	close(w.stopFlush)
	close(w.stopCompact)

	// Final flush and sync
	w.writeBuffer.Flush()
	w.file.Sync()
	return w.file.Close()
}

func (w *FileWAL) periodicFlush() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.mu.Lock()
			w.writeBuffer.Flush()
			w.file.Sync()
			w.mu.Unlock()
		case <-w.stopFlush:
			return
		}
	}
}

func (w *FileWAL) compactWAL() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Flush current buffer
	w.writeBuffer.Flush()
	w.file.Sync()

	// Track latest sequence per key
	latestSequence := make(map[string]uint64)
	keyEntries := make(map[string]LogEntry)

	// Scan existing WAL
	scanner := bufio.NewScanner(w.file)
	for scanner.Scan() {
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		// Track only the latest sequence for each key
		if seq, exists := latestSequence[entry.Key]; !exists || entry.Sequence > seq {
			latestSequence[entry.Key] = entry.Sequence
			keyEntries[entry.Key] = entry
		}
	}

	// Create temp file
	tempPath := w.file.Name() + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	tempWriter := bufio.NewWriter(tempFile)

	// Write only latest entries
	for _, entry := range keyEntries {
		data, err := json.Marshal(entry)
		if err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return err
		}
		if _, err := tempWriter.Write(append(data, '\n')); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return err
		}
	}

	// Flush and rotate
	tempWriter.Flush()
	tempFile.Sync()
	tempFile.Close()

	oldPath := w.file.Name()
	w.file.Close()

	if err := os.Rename(tempPath, oldPath); err != nil {
		return err
	}

	// Reopen WAL
	file, err := os.OpenFile(oldPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	w.file = file
	w.writeBuffer = bufio.NewWriter(file)

	return nil
}

func (w *FileWAL) periodicCompact() {
	ticker := time.NewTicker(COMPACT_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check file size
			info, err := w.file.Stat()
			if err != nil {
				continue
			}

			// Compact if file size exceeds threshold
			if info.Size() > COMPACT_THRESHOLD {
				if err := w.compactWAL(); err != nil {
					log.Printf("WAL compaction failed: %v", err)
				}
			}
		case <-w.stopCompact:
			return
		}
	}
}
