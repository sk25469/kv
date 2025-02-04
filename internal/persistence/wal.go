package wal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sk25469/kv/utils"
)

type Operation string

const (
	SET              Operation = utils.SET
	DELETE           Operation = utils.DEL
	DEFAULT_LOG_DIR            = "/var/log/"
	DEFAULT_LOG_FILE           = "kvstore"
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
	file     *os.File
	mu       sync.Mutex
	sequence uint64
}

func NewFileWAL(path string, nodeID string) (*FileWAL, error) {
	// Set proper permissions for log file
	logFileName := createDynamicLogFileName(nodeID)
	logPath := filepath.Join(DEFAULT_LOG_DIR, logFileName)
	if _, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &FileWAL{
		file: file,
	}, nil
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

	if _, err := w.file.Write(append(data, '\n')); err != nil {
		return err
	}

	return w.file.Sync()
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
	return w.file.Close()
}

func createDynamicLogFileName(nodeID string) string {
	// trim the nodeID to 8 characters
	if len(nodeID) > 6 {
		nodeID = nodeID[:6]
	}
	logFileName := fmt.Sprintf("%v-%v.log", DEFAULT_LOG_FILE, nodeID)
	return logFileName
}
