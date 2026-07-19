package dedupe

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrCapacity reports that a bounded Memory store cannot retain another ID.
var ErrCapacity = errors.New("hermes/dedupe: memory store capacity reached")

// MemoryConfig configures an in-process claim store.
type MemoryConfig struct {
	// MaxEntries bounds retained update IDs. The default is 1,000,000. Zero
	// explicitly disables the bound.
	MaxEntries int
	// Shards controls lock striping. Values are rounded up to a power of two.
	// The default is 64 and the maximum is 256.
	Shards int
}

type memoryShard struct {
	mu     sync.Mutex
	claims map[Key]time.Time
}

// Memory is a sharded atomic claim store with explicit expiry cleanup.
type Memory struct {
	shards     []memoryShard
	maxEntries int64
	entries    atomic.Int64
	capacityMu sync.Mutex
	now        func() time.Time
}

// NewMemory creates a configured in-process claim store.
func NewMemory(options ...MemoryConfig) *Memory {
	config := MemoryConfig{MaxEntries: 1_000_000, Shards: 64}
	if len(options) != 0 {
		config = options[0]
		if config.Shards == 0 {
			config.Shards = 64
		}
	}
	if config.MaxEntries < 0 {
		config.MaxEntries = 0
	}
	if config.Shards < 1 {
		config.Shards = 1
	}
	if config.Shards > 256 {
		config.Shards = 256
	}
	config.Shards = nextPowerOfTwo(config.Shards)
	shards := make([]memoryShard, config.Shards)
	for index := range shards {
		shards[index].claims = make(map[Key]time.Time)
	}
	return &Memory{
		shards:     shards,
		maxEntries: int64(config.MaxEntries),
		now:        time.Now,
	}
}

// Claim implements Store.
func (m *Memory) Claim(ctx context.Context, key Key, ttl time.Duration) (bool, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return false, err
		}
	}
	if ttl <= 0 {
		return false, ErrInvalidTTL
	}
	if m == nil || len(m.shards) == 0 {
		return false, ErrStoreRequired
	}
	now := m.now()
	shard := m.shard(key)
	shard.mu.Lock()
	if expiresAt, exists := shard.claims[key]; exists {
		if now.Before(expiresAt) {
			shard.mu.Unlock()
			return false, nil
		}
		shard.claims[key] = now.Add(ttl)
		shard.mu.Unlock()
		return true, nil
	}

	m.capacityMu.Lock()
	defer m.capacityMu.Unlock()
	if m.maxEntries > 0 && m.entries.Load() >= m.maxEntries {
		shard.mu.Unlock()
		return false, ErrCapacity
	}
	shard.claims[key] = now.Add(ttl)
	m.entries.Add(1)
	shard.mu.Unlock()
	return true, nil
}

// Release implements Store.
func (m *Memory) Release(ctx context.Context, key Key) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if m == nil || len(m.shards) == 0 {
		return nil
	}
	shard := m.shard(key)
	shard.mu.Lock()
	if _, exists := shard.claims[key]; exists {
		delete(shard.claims, key)
		m.entries.Add(-1)
	}
	shard.mu.Unlock()
	return nil
}

// Len returns retained claims, including expired claims not yet swept.
func (m *Memory) Len() int {
	if m == nil {
		return 0
	}
	return int(m.entries.Load())
}

// Sweep removes expired claims and returns the number removed.
func (m *Memory) Sweep() int {
	if m == nil {
		return 0
	}
	now := m.now()
	removed := 0
	for index := range m.shards {
		shard := &m.shards[index]
		shardRemoved := 0
		shard.mu.Lock()
		for key, expiresAt := range shard.claims {
			if !now.Before(expiresAt) {
				delete(shard.claims, key)
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

func (m *Memory) shard(key Key) *memoryShard {
	return &m.shards[hashKey(key)&uint64(len(m.shards)-1)]
}

func hashKey(key Key) uint64 {
	const (
		offset = uint64(14695981039346656037)
		prime  = uint64(1099511628211)
	)
	hash := offset
	for index := 0; index < len(key.Namespace); index++ {
		hash ^= uint64(key.Namespace[index])
		hash *= prime
	}
	value := uint64(key.UpdateID)
	for index := 0; index < 8; index++ {
		hash ^= value & 0xff
		hash *= prime
		value >>= 8
	}
	return hash
}

func nextPowerOfTwo(value int) int {
	result := 1
	for result < value {
		result <<= 1
	}
	return result
}
