package storage

import (
	"container/list"
	"sync"

	"github.com/sureshKrishna05/aegisq-framework/core/block"
)

// LRUCache implements a simple, thread-safe LRU cache for blocks
type LRUCache struct {
	maxItems int
	mu       sync.RWMutex
	ll       *list.List
	cache    map[uint64]*list.Element
}

type entry struct {
	key   uint64
	block *block.Block
}

func NewLRUCache(maxItems int) *LRUCache {
	return &LRUCache{
		maxItems: maxItems,
		ll:       list.New(),
		cache:    make(map[uint64]*list.Element),
	}
}

func (c *LRUCache) Add(height uint64, b *block.Block) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If it already exists, move to front and update
	if ele, hit := c.cache[height]; hit {
		c.ll.MoveToFront(ele)
		ele.Value.(*entry).block = b
		return
	}

	// Add new item
	ele := c.ll.PushFront(&entry{key: height, block: b})
	c.cache[height] = ele

	// Evict if over capacity
	if c.maxItems != 0 && c.ll.Len() > c.maxItems {
		c.RemoveOldest()
	}
}

func (c *LRUCache) Get(height uint64) (*block.Block, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, hit := c.cache[height]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).block, true
	}
	return nil, false
}

func (c *LRUCache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
	}
}

// CachedStore wraps any Store implementation (like Pebble) with an LRU Block Cache
type CachedStore struct {
	Store // Embed the interface
	cache *LRUCache
}

// NewCachedStore creates a wrapper around the underlying DB
func NewCachedStore(db Store, cacheSize int) *CachedStore {
	return &CachedStore{
		Store: db,
		cache: NewLRUCache(cacheSize),
	}
}

func (c *CachedStore) SaveBlock(b *block.Block) error {
	// First save to the underlying database natively
	if err := c.Store.SaveBlock(b); err != nil {
		return err
	}
	// Immediately cache the block since it was just written
	c.cache.Add(b.Index, b)
	return nil
}

func (c *CachedStore) GetBlock(height uint64) (*block.Block, error) {
	// 1. Try cache (O(1) Memory hit)
	if b, ok := c.cache.Get(height); ok {
		return b, nil
	}

	// 2. Cache miss -> Hit PebbleDB (Disk hit)
	b, err := c.Store.GetBlock(height)
	if err != nil {
		return nil, err
	}

	// 3. Pin to cache for future requests
	c.cache.Add(height, b)
	return b, nil
}

// Phase 4 Pass-Throughs
func (c *CachedStore) CreateSnapshot(destDir string) error {
	return c.Store.CreateSnapshot(destDir)
}

func (c *CachedStore) PruneOldBlocks(retain uint64) error {
	// Let the DB prune old blocks from disk
	err := c.Store.PruneOldBlocks(retain)
	if err != nil {
		return err
	}
	
	// Flush RAM cache to ensure no pruned blocks survive in memory
	c.cache.mu.Lock()
	c.cache.ll.Init()
	c.cache.cache = make(map[uint64]*list.Element)
	c.cache.mu.Unlock()
	return nil
}

func (c *CachedStore) CompactDB() error {
	return c.Store.CompactDB()
}
