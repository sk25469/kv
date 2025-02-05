package storage

type StorageType string

const (
	InMemory StorageType = "memory"
	FileBase StorageType = "file"
)

type StorageStructure string

const (
	HashMap   StorageStructure = "hashmap"
	BPlusTree StorageStructure = "bplustree"
)
