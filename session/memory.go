package session

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrCapacity reports that a bounded Memory store cannot accept a new key.
var ErrCapacity = errors.New("hermes/session: memory store capacity reached")

// MemoryConfig configures an in-process session store.
type MemoryConfig struct {
	// TTL expires values after this duration since their most recent Save.
	// Zero disables expiration.
	TTL time.Duration
	// MaxEntries bounds retained keys. Zero is unbounded.
	MaxEntries int
	// Shards controls lock striping. Values are rounded up to a power of two.
	// The default is 32 and the maximum is 256.
	Shards int
}

type memoryEntry[T any] struct {
	value     T
	expiresAt time.Time
}

type memoryShard[T any] struct {
	mu      sync.RWMutex
	entries map[Key]memoryEntry[T]
}

// Memory is a sharded, bounded, lazily expiring Store. It creates no
// background goroutines; call Sweep periodically or run Cleanup when eager
// expiry is desired.
type Memory[T any] struct {
	shards     []memoryShard[T]
	ttl        time.Duration
	maxEntries int64
	entries    atomic.Int64
	capacityMu sync.Mutex
	now        func() time.Time
}

// NewMemory creates a store with the supplied TTL and otherwise default
// configuration. A zero TTL keeps values until Delete.
func NewMemory[T any](ttl time.Duration) *Memory[T] {
	return NewMemoryWithConfig[T](MemoryConfig{TTL: ttl})
}

// NewMemoryWithConfig creates a configured in-process store.
func NewMemoryWithConfig[T any](config MemoryConfig) *Memory[T] {
	shardCount := config.Shards
	if shardCount <= 0 {
		shardCount = 32
	}
	if shardCount > 256 {
		shardCount = 256
	}
	shardCount = nextPowerOfTwo(shardCount)
	shards := make([]memoryShard[T], shardCount)
	for index := range shards {
		shards[index].entries = make(map[Key]memoryEntry[T])
	}
	maxEntries := config.MaxEntries
	if maxEntries < 0 {
		maxEntries = 0
	}
	return &Memory[T]{
		shards:     shards,
		ttl:        max(config.TTL, 0),
		maxEntries: int64(maxEntries),
		now:        time.Now,
	}
}

// Load implements Store.
func (m *Memory[T]) Load(ctx context.Context, key Key) (T, bool, error) {
	var zero T
	if err := contextError(ctx); err != nil {
		return zero, false, err
	}
	if m == nil || len(m.shards) == 0 {
		return zero, false, nil
	}
	shard := m.shard(key)
	shard.mu.RLock()
	entry, ok := shard.entries[key]
	shard.mu.RUnlock()
	if !ok {
		return zero, false, nil
	}
	if !entry.expiresAt.IsZero() && !m.now().Before(entry.expiresAt) {
		shard.mu.Lock()
		current, exists := shard.entries[key]
		if exists && current.expiresAt.Equal(entry.expiresAt) {
			delete(shard.entries, key)
			m.entries.Add(-1)
		}
		shard.mu.Unlock()
		return zero, false, nil
	}
	return entry.value, true, nil
}

// Save implements Store.
func (m *Memory[T]) Save(ctx context.Context, key Key, value T) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if m == nil || len(m.shards) == 0 {
		return ErrStoreRequired
	}
	entry := memoryEntry[T]{value: value}
	if m.ttl > 0 {
		entry.expiresAt = m.now().Add(m.ttl)
	}

	shard := m.shard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	if _, exists := shard.entries[key]; exists {
		shard.entries[key] = entry
		return nil
	}

	m.capacityMu.Lock()
	defer m.capacityMu.Unlock()
	if m.maxEntries > 0 && m.entries.Load() >= m.maxEntries {
		return ErrCapacity
	}
	shard.entries[key] = entry
	m.entries.Add(1)
	return nil
}

// Delete implements Store.
func (m *Memory[T]) Delete(ctx context.Context, key Key) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if m == nil || len(m.shards) == 0 {
		return nil
	}
	shard := m.shard(key)
	shard.mu.Lock()
	if _, ok := shard.entries[key]; ok {
		delete(shard.entries, key)
		m.entries.Add(-1)
	}
	shard.mu.Unlock()
	return nil
}

// Len returns the number of retained, not-yet-swept entries. Expired values
// disappear immediately when loaded and on the next Sweep otherwise.
func (m *Memory[T]) Len() int {
	if m == nil {
		return 0
	}
	return int(m.entries.Load())
}

// Sweep removes all currently expired entries and returns the count removed.
func (m *Memory[T]) Sweep() int {
	if m == nil || m.ttl <= 0 {
		return 0
	}
	now := m.now()
	removed := 0
	for index := range m.shards {
		shard := &m.shards[index]
		shardRemoved := 0
		shard.mu.Lock()
		for key, entry := range shard.entries {
			if !entry.expiresAt.IsZero() && !now.Before(entry.expiresAt) {
				delete(shard.entries, key)
				shardRemoved++
			}
		}
		if shardRemoved != 0 {
			m.entries.Add(-int64(shardRemoved))
			removed += shardRemoved
		}
		shard.mu.Unlock()
	}
	return removed
}

// Cleanup sweeps at interval until ctx ends. It blocks and creates no hidden
// goroutine, so callers explicitly own its lifecycle.
func (m *Memory[T]) Cleanup(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return errors.New("hermes/session: cleanup interval must be positive")
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			m.Sweep()
		}
	}
}

func (m *Memory[T]) shard(key Key) *memoryShard[T] {
	return &m.shards[hashKey(key)&uint64(len(m.shards)-1)]
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func nextPowerOfTwo(value int) int {
	result := 1
	for result < value {
		result <<= 1
	}
	return result
}
