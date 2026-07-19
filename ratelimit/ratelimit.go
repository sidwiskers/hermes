package ratelimit

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sidwiskers/hermes/framework"
	"github.com/sidwiskers/hermes/session"
)

const maxExactBurst = 1 << 53

var (
	// ErrInvalidLimit reports a non-positive event count or period.
	ErrInvalidLimit = errors.New("hermes/ratelimit: limit and period must be positive")
	// ErrLimited is wrapped by LimitError when a bucket has no token.
	ErrLimited = errors.New("hermes/ratelimit: rate limit exceeded")
	// ErrCapacity reports that the configured maximum number of keys has been
	// reached. Sweep idle keys or increase the explicit bound.
	ErrCapacity = errors.New("hermes/ratelimit: key capacity reached")
)

// Key is the comparable identity used by a limiter.
type Key = session.Key

// KeyFunc derives a limiter identity from an update.
type KeyFunc = session.KeyFunc

// ByUser limits each Telegram user independently.
func ByUser(c *framework.Context) (Key, bool) { return session.ByUser(c) }

// ByChat limits each Telegram chat independently.
func ByChat(c *framework.Context) (Key, bool) { return session.ByChat(c) }

// ByChatUser limits each user independently within each chat.
func ByChatUser(c *framework.Context) (Key, bool) { return session.ByChatUser(c) }

// Decision describes one token-bucket check.
type Decision struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

// LimitError reports the rejected key and minimum estimated wait.
type LimitError struct {
	Key        Key
	RetryAfter time.Duration
}

// Error implements error.
func (e *LimitError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("hermes/ratelimit: rate limit exceeded; retry after %s", e.RetryAfter)
}

// Unwrap lets errors.Is identify ErrLimited.
func (e *LimitError) Unwrap() error { return ErrLimited }

// RejectFunc handles a denied update. Returning nil treats the update as
// successfully consumed; returning an error forwards it to Hermes.
type RejectFunc func(*framework.Context, Decision) error

type config struct {
	burst      int
	maxKeys    int
	idleTTL    time.Duration
	namespace  string
	key        KeyFunc
	onRejected RejectFunc
}

// Option configures a Limiter.
type Option func(*config)

// WithBurst allows short bursts up to size. The default equals limit.
func WithBurst(size int) Option { return func(config *config) { config.burst = size } }

// WithMaxKeys bounds retained limiter identities. The default is 100,000.
// Zero explicitly removes the bound.
func WithMaxKeys(maximum int) Option { return func(config *config) { config.maxKeys = maximum } }

// WithIdleTTL makes buckets eligible for Sweep after they have not been
// checked for ttl. The default is ten minutes; zero disables expiry.
func WithIdleTTL(ttl time.Duration) Option { return func(config *config) { config.idleTTL = ttl } }

// WithKey replaces the default per-chat/user identity policy.
func WithKey(key KeyFunc) Option { return func(config *config) { config.key = key } }

// WithNamespace sets the namespace on keys whose KeyFunc leaves it empty.
func WithNamespace(namespace string) Option {
	return func(config *config) { config.namespace = namespace }
}

// WithRejected installs a custom denial handler.
func WithRejected(rejected RejectFunc) Option {
	return func(config *config) { config.onRejected = rejected }
}

type bucket struct {
	tokens float64
	last   time.Time
	seen   time.Time
}

type shard struct {
	mu      sync.Mutex
	buckets map[Key]bucket
}

// Limiter is a sharded token bucket. It creates no goroutines.
type Limiter struct {
	rate       float64
	burst      int
	maxKeys    int64
	idleTTL    time.Duration
	namespace  string
	key        KeyFunc
	onRejected RejectFunc
	shards     [64]shard
	entries    atomic.Int64
	capacityMu sync.Mutex
	now        func() time.Time
}

// New permits limit events per period with a default burst of limit.
func New(limit int, period time.Duration, options ...Option) (*Limiter, error) {
	if limit <= 0 || period <= 0 {
		return nil, ErrInvalidLimit
	}
	config := config{
		burst:   limit,
		maxKeys: 100_000,
		idleTTL: 10 * time.Minute,
		key:     ByChatUser,
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	if config.burst <= 0 || uint64(config.burst) > maxExactBurst {
		return nil, ErrInvalidLimit
	}
	if config.maxKeys < 0 || config.idleTTL < 0 {
		return nil, ErrInvalidLimit
	}
	if config.key == nil {
		config.key = ByChatUser
	}
	limiter := &Limiter{
		rate:       float64(limit) / period.Seconds(),
		burst:      config.burst,
		maxKeys:    int64(config.maxKeys),
		idleTTL:    config.idleTTL,
		namespace:  config.namespace,
		key:        config.key,
		onRejected: config.onRejected,
		now:        time.Now,
	}
	for index := range limiter.shards {
		limiter.shards[index].buckets = make(map[Key]bucket)
	}
	return limiter, nil
}

// Allow consumes one token for key when available.
func (l *Limiter) Allow(key Key) (Decision, error) {
	if l == nil || l.rate <= 0 || l.burst <= 0 {
		return Decision{}, ErrInvalidLimit
	}
	if key.Namespace == "" {
		key.Namespace = l.namespace
	}
	return l.allowAt(key, l.now())
}

// Middleware checks one token before invoking the downstream handler. Updates
// without a key pass through unchanged.
func (l *Limiter) Middleware() framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(c *framework.Context) error {
			if next == nil {
				return nil
			}
			if l == nil || l.key == nil {
				return ErrInvalidLimit
			}
			key, ok := l.key(c)
			if !ok {
				return next(c)
			}
			if key.Namespace == "" {
				key.Namespace = l.namespace
			}
			decision, err := l.Allow(key)
			if err != nil {
				return err
			}
			if decision.Allowed {
				return next(c)
			}
			if l.onRejected != nil {
				return l.onRejected(c, decision)
			}
			return &LimitError{Key: key, RetryAfter: decision.RetryAfter}
		}
	}
}

// Len returns the number of retained identities.
func (l *Limiter) Len() int {
	if l == nil {
		return 0
	}
	return int(l.entries.Load())
}

// Sweep removes buckets idle for at least the configured TTL.
func (l *Limiter) Sweep() int {
	if l == nil || l.idleTTL <= 0 {
		return 0
	}
	deadline := l.now().Add(-l.idleTTL)
	removed := 0
	for index := range l.shards {
		shard := &l.shards[index]
		shardRemoved := 0
		shard.mu.Lock()
		for key, bucket := range shard.buckets {
			if !bucket.seen.After(deadline) {
				delete(shard.buckets, key)
				shardRemoved++
			}
		}
		if shardRemoved != 0 {
			l.entries.Add(-int64(shardRemoved))
			removed += shardRemoved
		}
		shard.mu.Unlock()
	}
	return removed
}

func (l *Limiter) allowAt(key Key, now time.Time) (Decision, error) {
	shard := &l.shards[hashKey(key)&uint64(len(l.shards)-1)]
	shard.mu.Lock()
	current, exists := shard.buckets[key]
	if !exists {
		l.capacityMu.Lock()
		if l.maxKeys > 0 && l.entries.Load() >= l.maxKeys {
			l.capacityMu.Unlock()
			shard.mu.Unlock()
			return Decision{}, ErrCapacity
		}
		current = bucket{tokens: float64(l.burst), last: now, seen: now}
		shard.buckets[key] = current
		l.entries.Add(1)
		l.capacityMu.Unlock()
	}

	if elapsed := now.Sub(current.last); elapsed > 0 {
		current.tokens = min(float64(l.burst), current.tokens+elapsed.Seconds()*l.rate)
	}
	current.last = now
	current.seen = now
	decision := Decision{}
	if current.tokens >= 1 {
		current.tokens--
		decision.Allowed = true
		decision.Remaining = int(math.Floor(current.tokens))
	} else {
		decision.RetryAfter = retryDuration((1 - current.tokens) / l.rate)
	}
	shard.buckets[key] = current
	shard.mu.Unlock()
	return decision, nil
}

func retryDuration(seconds float64) time.Duration {
	nanoseconds := math.Ceil(seconds * float64(time.Second))
	if nanoseconds >= float64(math.MaxInt64) {
		return time.Duration(math.MaxInt64)
	}
	if nanoseconds <= 0 {
		return time.Nanosecond
	}
	return time.Duration(nanoseconds)
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
