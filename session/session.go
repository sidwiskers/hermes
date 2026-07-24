package session

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sidwiskers/hermes/framework"
)

var (
	// ErrStoreRequired reports a Manager constructed without a Store.
	ErrStoreRequired = errors.New("hermes/session: store is required")
	// ErrKeyRequired reports a Manager constructed without a KeyFunc.
	ErrKeyRequired = errors.New("hermes/session: key function is required")
	// ErrUnavailable reports session access outside the Manager middleware.
	ErrUnavailable = errors.New("hermes/session: value is unavailable; install the manager middleware")
)

// Key is a storage-safe session identity. Namespace lets several independent
// features share one Store without key collisions.
type Key struct {
	Namespace string
	ChatID    int64
	UserID    int64
}

// KeyFunc derives a session identity from an update. false skips session
// loading for that update without skipping the downstream handler.
type KeyFunc func(*framework.Context) (Key, bool)

// ByUser keeps one session per Telegram user.
func ByUser(c *framework.Context) (Key, bool) {
	if c == nil {
		return Key{}, false
	}
	user := c.Sender()
	if user == nil || user.ID == 0 {
		return Key{}, false
	}
	return Key{UserID: user.ID}, true
}

// ByChat keeps one session per Telegram chat.
func ByChat(c *framework.Context) (Key, bool) {
	if c == nil {
		return Key{}, false
	}
	chatID, ok := c.ChatID()
	return Key{ChatID: chatID}, ok
}

// ByChatUser keeps independent sessions for each user in each chat. Updates
// without a chat, such as inline queries, fall back to the sender identity.
func ByChatUser(c *framework.Context) (Key, bool) {
	if c == nil {
		return Key{}, false
	}
	var key Key
	if chatID, ok := c.ChatID(); ok {
		key.ChatID = chatID
	}
	if user := c.Sender(); user != nil {
		key.UserID = user.ID
	}
	return key, key.ChatID != 0 || key.UserID != 0
}

// Store persists typed session values. Implementations must be safe for
// concurrent use. A missing value is returned as (zero, false, nil).
type Store[T any] interface {
	Load(context.Context, Key) (T, bool, error)
	Save(context.Context, Key, T) error
	Delete(context.Context, Key) error
}

type config struct {
	namespace     string
	commitOnError bool
}

// Option configures a Manager.
type Option func(*config)

// WithNamespace sets a storage namespace. It overrides an empty namespace
// returned by the KeyFunc and prevents unrelated managers from colliding.
func WithNamespace(namespace string) Option {
	return func(config *config) { config.namespace = namespace }
}

// WithCommitOnError controls whether mutations are persisted when the
// downstream handler returns an error. The default is false, providing
// transaction-like handler semantics.
func WithCommitOnError(enabled bool) Option {
	return func(config *config) { config.commitOnError = enabled }
}

type mutation uint8

const (
	mutationNone mutation = iota
	mutationSave
	mutationDelete
)

type binding[T any] struct {
	value    T
	exists   bool
	mutation mutation
}

// The byte prevents distinct keys from being zero-size pointers, whose
// addresses are permitted to compare equal.
type contextKey struct{ _ byte }

type keyLock struct {
	mu   sync.Mutex
	refs int
}

type keyLockShard struct {
	mu    sync.Mutex
	locks map[Key]*keyLock
}

// keyedLocks serializes equal keys without making unrelated keys wait behind
// a fixed set of long-held striped locks. Shards protect only the short map
// lookup; each active key owns an independent mutex.
type keyedLocks struct {
	shards [64]keyLockShard
	pool   sync.Pool
}

type keyLockHandle struct {
	owner *keyedLocks
	shard *keyLockShard
	key   Key
	lock  *keyLock
}

func (l *keyedLocks) acquire(key Key) keyLockHandle {
	shard := &l.shards[hashKey(key)&uint64(len(l.shards)-1)]
	shard.mu.Lock()
	if shard.locks == nil {
		shard.locks = make(map[Key]*keyLock)
	}
	lock := shard.locks[key]
	if lock == nil {
		lock, _ = l.pool.Get().(*keyLock)
		if lock == nil {
			lock = new(keyLock)
		}
		lock.refs = 0
		shard.locks[key] = lock
	}
	lock.refs++
	shard.mu.Unlock()

	lock.mu.Lock()
	return keyLockHandle{owner: l, shard: shard, key: key, lock: lock}
}

func (h keyLockHandle) release() {
	h.lock.mu.Unlock()

	h.shard.mu.Lock()
	h.lock.refs--
	if h.lock.refs == 0 {
		delete(h.shard.locks, h.key)
		h.owner.pool.Put(h.lock)
	}
	h.shard.mu.Unlock()
}

// Manager binds one typed Store and key policy to Hermes middleware.
// Updates sharing a key are serialized across the complete downstream handler
// to prevent lost updates in read-modify-write workflows.
type Manager[T any] struct {
	store         Store[T]
	key           KeyFunc
	namespace     string
	commitOnError bool
	contextKey    *contextKey
	locks         keyedLocks
}

// New creates a typed session manager. ByChatUser is a practical default when
// key is nil.
func New[T any](store Store[T], key KeyFunc, options ...Option) *Manager[T] {
	if key == nil {
		key = ByChatUser
	}
	var config config
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return &Manager[T]{
		store:         store,
		key:           key,
		namespace:     config.namespace,
		commitOnError: config.commitOnError,
		contextKey:    new(contextKey),
	}
}

// Middleware loads and commits this manager's value around a handler. Updates
// for which the key function returns false pass through unchanged.
func (m *Manager[T]) Middleware() framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(c *framework.Context) error {
			if next == nil {
				return nil
			}
			if c == nil {
				return next(nil)
			}
			if m == nil || m.store == nil {
				return ErrStoreRequired
			}
			if m.key == nil {
				return ErrKeyRequired
			}
			key, ok := m.resolveKey(c)
			if !ok {
				return next(c)
			}

			lock := m.locks.acquire(key)
			defer lock.release()

			ctx := context.Background()
			if c != nil && c.Context != nil {
				ctx = c.Context
			}
			value, exists, err := m.store.Load(ctx, key)
			if err != nil {
				return fmt.Errorf("hermes/session: load: %w", err)
			}
			bound := &binding[T]{value: value, exists: exists}
			cloned := *c
			cloned.Context = context.WithValue(ctx, m.contextKey, bound)

			handlerErr := next(&cloned)
			if handlerErr != nil && !m.commitOnError {
				return handlerErr
			}
			commitErr := m.commit(ctx, key, bound)
			return errors.Join(handlerErr, commitErr)
		}
	}
}

// Key returns the session key for c after applying the manager namespace.
func (m *Manager[T]) Key(c *framework.Context) (Key, bool) {
	if m == nil || m.key == nil {
		return Key{}, false
	}
	return m.resolveKey(c)
}

// Get returns the value loaded by this manager's middleware.
func (m *Manager[T]) Get(c *framework.Context) (T, bool, error) {
	bound, err := m.binding(c)
	if err != nil {
		var zero T
		return zero, false, err
	}
	return bound.value, bound.exists && bound.mutation != mutationDelete, nil
}

// Set replaces the current session value. The store is updated after the
// downstream handler finishes according to the manager's commit policy.
func (m *Manager[T]) Set(c *framework.Context, value T) error {
	bound, err := m.binding(c)
	if err != nil {
		return err
	}
	bound.value = value
	bound.exists = true
	bound.mutation = mutationSave
	return nil
}

// Update applies a read-modify-write function to the current value. A missing
// session starts with T's zero value.
func (m *Manager[T]) Update(c *framework.Context, update func(*T) error) error {
	if update == nil {
		return nil
	}
	bound, err := m.binding(c)
	if err != nil {
		return err
	}
	value := bound.value
	if err := update(&value); err != nil {
		return err
	}
	bound.value = value
	bound.exists = true
	bound.mutation = mutationSave
	return nil
}

// Delete removes the current session after the handler completes.
func (m *Manager[T]) Delete(c *framework.Context) error {
	bound, err := m.binding(c)
	if err != nil {
		return err
	}
	var zero T
	bound.value = zero
	bound.exists = false
	bound.mutation = mutationDelete
	return nil
}

func (m *Manager[T]) resolveKey(c *framework.Context) (Key, bool) {
	key, ok := m.key(c)
	if ok && key.Namespace == "" {
		key.Namespace = m.namespace
	}
	return key, ok
}

func (m *Manager[T]) binding(c *framework.Context) (*binding[T], error) {
	if m == nil || m.contextKey == nil || c == nil || c.Context == nil {
		return nil, ErrUnavailable
	}
	bound, ok := c.Context.Value(m.contextKey).(*binding[T])
	if !ok || bound == nil {
		return nil, ErrUnavailable
	}
	return bound, nil
}

func (m *Manager[T]) commit(ctx context.Context, key Key, bound *binding[T]) error {
	var err error
	switch bound.mutation {
	case mutationSave:
		err = m.store.Save(ctx, key, bound.value)
	case mutationDelete:
		err = m.store.Delete(ctx, key)
	}
	if err != nil {
		return fmt.Errorf("hermes/session: commit: %w", err)
	}
	return nil
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
	for _, value := range [2]uint64{uint64(key.ChatID), uint64(key.UserID)} {
		for index := 0; index < 8; index++ {
			hash ^= value & 0xff
			hash *= prime
			value >>= 8
		}
	}
	return hash
}
