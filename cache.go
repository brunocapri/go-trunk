package trunk

import (
	"container/heap"
	"errors"
	"hash/fnv"
	"runtime"
	"sync"
	"time"
)

var (
	ErrNoEmptyKey       = errors.New("empty key is not allowed")
	ErrNegativeInterval = errors.New("interval cannot be negative")
)

type Cache[T any] struct {
	shards   []shard[T]
	interval time.Duration
}

type shard[T any] struct {
	mu         sync.RWMutex
	store      map[string]*cacheEntry[T]
	expiryHeap expiryHeap[T]
}

type cacheEntry[T any] struct {
	key       string
	createdAt time.Time
	value     T
	index     int
}

// interval: the time interval the cache will be evicted
//
// shards: the amount of cache shards. If < 0 one per CPU will be used.
//
// - more shards = potentially more performance
//
// - less shards = potentially less memory overhead
func NewCache[T any](interval time.Duration, shards int) (*Cache[T], error) {
	if interval < 0 {
		return nil, ErrNegativeInterval
	}

	if shards <= 0 {
		shards = runtime.NumCPU()
	}

	cache := &Cache[T]{
		shards:   make([]shard[T], shards),
		interval: interval,
	}
	for i := range cache.shards {
		cache.shards[i].store = make(map[string]*cacheEntry[T])
	}
	go cache.reapLoop()
	return cache, nil
}

func (c *Cache[T]) getShard(key string) *shard[T] {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return &c.shards[hash.Sum32()%uint32(len(c.shards))]
}

func (c *Cache[T]) Add(key string, value T) error {
	if key == "" {
		return ErrNoEmptyKey
	}
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	entry := &cacheEntry[T]{
		key:       key,
		createdAt: time.Now(),
		value:     value,
	}
	shard.store[key] = entry
	heap.Push(&shard.expiryHeap, entry)
	return nil
}

func (c *Cache[T]) Get(key string) (T, bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	entry, found := shard.store[key]
	shard.mu.RUnlock()

	if found && time.Since(entry.createdAt) <= c.interval {
		return entry.value, true
	}

	var zeroValue T
	return zeroValue, false
}

func (c *Cache[T]) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for range ticker.C {
		c.evict()
	}
}

func (c *Cache[T]) evict() {
	now := time.Now()
	for i := range c.shards {
		shard := &c.shards[i]
		shard.mu.Lock()
		for len(shard.expiryHeap) > 0 {
			if now.Before(shard.expiryHeap[0].createdAt.Add(c.interval)) {
				break
			}
			expiredEntry := heap.Pop(&shard.expiryHeap).(*cacheEntry[T])
			delete(shard.store, expiredEntry.key)
		}
		shard.mu.Unlock()
	}
}
