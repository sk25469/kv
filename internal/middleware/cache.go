package middleware

import (
	"container/list"
	"sync"
)

type CacheStrategy interface {
	Put(key string, value string)
	Get(key string) (string, bool)
	Remove(key string)
}

// LRU Cache Implementation
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

type entry struct {
	key   string
	value string
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

func (c *LRUCache) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		c.list.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	if c.list.Len() >= c.capacity {
		// Remove the least recently used item
		oldest := c.list.Back()
		if oldest != nil {
			delete(c.cache, oldest.Value.(*entry).key)
			c.list.Remove(oldest)
		}
	}

	// Add new item
	elem := c.list.PushFront(&entry{key, value})
	c.cache[key] = elem
}

func (c *LRUCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, exists := c.cache[key]; exists {
		c.list.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return "", false
}

func (c *LRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		delete(c.cache, key)
		c.list.Remove(elem)
	}
}

// LFU Cache Implementation
type LFUCache struct {
	capacity int
	cache    map[string]*lfuEntry
	freqMap  map[int]*list.List
	minFreq  int
	mu       sync.RWMutex
}

type lfuEntry struct {
	key      string
	value    string
	freq     int
	listElem *list.Element
}

func NewLFUCache(capacity int) *LFUCache {
	return &LFUCache{
		capacity: capacity,
		cache:    make(map[string]*lfuEntry),
		freqMap:  make(map[int]*list.List),
		minFreq:  0,
	}
}

// Add these methods to your existing LFUCache implementation

func (c *LFUCache) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update its value and frequency
	if entry, exists := c.cache[key]; exists {
		entry.value = value
		c.incrementFreq(entry)
		return
	}

	// If at capacity, remove least frequent item
	if len(c.cache) >= c.capacity {
		c.removeLFU()
	}

	// Add new item with frequency 1
	c.minFreq = 1
	entry := &lfuEntry{
		key:   key,
		value: value,
		freq:  1,
	}

	// Initialize frequency list if it doesn't exist
	if c.freqMap[1] == nil {
		c.freqMap[1] = list.New()
	}

	// Add to frequency list and cache
	entry.listElem = c.freqMap[1].PushFront(key)
	c.cache[key] = entry
}

func (c *LFUCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, exists := c.cache[key]; exists {
		c.incrementFreq(entry)
		return entry.value, true
	}
	return "", false
}

func (c *LFUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.cache[key]; exists {
		// Remove from frequency list
		c.freqMap[entry.freq].Remove(entry.listElem)
		if c.freqMap[entry.freq].Len() == 0 {
			delete(c.freqMap, entry.freq)
			if c.minFreq == entry.freq {
				c.minFreq = 0
			}
		}
		// Remove from cache
		delete(c.cache, key)
	}
}

// Helper method to increment frequency of an entry
func (c *LFUCache) incrementFreq(entry *lfuEntry) {
	// Remove from current frequency list
	c.freqMap[entry.freq].Remove(entry.listElem)
	if c.freqMap[entry.freq].Len() == 0 {
		delete(c.freqMap, entry.freq)
		if c.minFreq == entry.freq {
			c.minFreq++
		}
	}

	// Increment frequency
	entry.freq++

	// Create new frequency list if it doesn't exist
	if c.freqMap[entry.freq] == nil {
		c.freqMap[entry.freq] = list.New()
	}

	// Add to new frequency list
	entry.listElem = c.freqMap[entry.freq].PushFront(entry.key)
}

// Helper method to remove least frequently used item
func (c *LFUCache) removeLFU() {
	// Get least frequent list
	if list := c.freqMap[c.minFreq]; list != nil && list.Len() > 0 {
		// Remove least recently used item from least frequent list
		lruKey := list.Back().Value.(string)
		list.Remove(list.Back())
		delete(c.cache, lruKey)

		// Clean up empty frequency list
		if list.Len() == 0 {
			delete(c.freqMap, c.minFreq)
		}
	}
}

// Implement similar methods for LFU...
