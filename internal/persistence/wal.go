package wal

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
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

const (
	OpSet    byte = 1
	OpDelete byte = 2
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

	data := encodeBinary(entry)
	_, err := w.writeBuffer.Write(data)
	return err
}

func (w *FileWAL) Recover() ([]LogEntry, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var entries []LogEntry
	for {
		entry, err := decodeBinary(w.file)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
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
	// scanner := bufio.NewScanner(w.file)
	for {
		entry, err := decodeBinary(w.file)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Update latest sequence
		if entry.Sequence > latestSequence[entry.Key] {
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
		data := encodeBinary(entry)
		if _, err := tempWriter.Write(data); err != nil {
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

func encodeBinary(entry LogEntry) []byte {
	// Calculate buffer size
	size := 1 + 4 + len(entry.Key) + 8 // op + keyLen + key + sequence
	if entry.Operation == SET {
		size += 4 + len(entry.Value) // valueLen + value
	}

	buf := make([]byte, 0, size)

	// Write operation
	if entry.Operation == SET {
		buf = append(buf, OpSet)
	} else {
		buf = append(buf, OpDelete)
	}

	// Write key length and key
	keyLen := uint32(len(entry.Key))
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, keyLen)
	buf = append(buf, lenBuf...)
	buf = append(buf, []byte(entry.Key)...)

	// Write value for SET operations
	if entry.Operation == SET {
		valueLen := uint32(len(entry.Value))
		binary.BigEndian.PutUint32(lenBuf, valueLen)
		buf = append(buf, lenBuf...)
		buf = append(buf, []byte(entry.Value)...)
	}

	// Write sequence
	seqBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(seqBuf, entry.Sequence)
	buf = append(buf, seqBuf...)

	return buf
}

func decodeBinary(r io.Reader) (LogEntry, error) {
	var entry LogEntry

	// Read operation
	opBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, opBuf); err != nil {
		return entry, err
	}

	switch opBuf[0] {
	case OpSet:
		entry.Operation = SET
	case OpDelete:
		entry.Operation = DELETE
	default:
		return entry, fmt.Errorf("invalid operation: %d", opBuf[0])
	}

	// Read key length
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return entry, err
	}
	keyLen := binary.BigEndian.Uint32(lenBuf)

	// Read key
	keyBuf := make([]byte, keyLen)
	if _, err := io.ReadFull(r, keyBuf); err != nil {
		return entry, err
	}
	entry.Key = string(keyBuf)

	// Read value for SET operations
	if entry.Operation == SET {
		if _, err := io.ReadFull(r, lenBuf); err != nil {
			return entry, err
		}
		valueLen := binary.BigEndian.Uint32(lenBuf)

		valueBuf := make([]byte, valueLen)
		if _, err := io.ReadFull(r, valueBuf); err != nil {
			return entry, err
		}
		entry.Value = string(valueBuf)
	}

	// Read sequence
	seqBuf := make([]byte, 8)
	if _, err := io.ReadFull(r, seqBuf); err != nil {
		return entry, err
	}
	entry.Sequence = binary.BigEndian.Uint64(seqBuf)

	return entry, nil
}
