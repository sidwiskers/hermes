package dedupe

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sidwiskers/hermes/framework"
)

var (
	// ErrStoreRequired reports a Manager constructed without a Store.
	ErrStoreRequired = errors.New("hermes/dedupe: store is required")
	// ErrInvalidTTL reports a non-positive retention duration.
	ErrInvalidTTL = errors.New("hermes/dedupe: retention TTL must be positive")
)

// Key identifies one bot update. Namespace must distinguish bots when a store
// is shared because Telegram update IDs are unique only within a bot.
type Key struct {
	Namespace string
	UpdateID  int64
}

// Store atomically claims update IDs. Claim returns true only for the caller
// that created or replaced an expired claim. Implementations must be safe for
// concurrent use.
type Store interface {
	Claim(context.Context, Key, time.Duration) (bool, error)
	Release(context.Context, Key) error
}

// DuplicateFunc handles an already claimed update. nil consumes duplicates
// successfully without invoking the downstream handler.
type DuplicateFunc func(*framework.Context) error

type config struct {
	namespace   string
	ttl         time.Duration
	onDuplicate DuplicateFunc
}

// Option configures a Manager.
type Option func(*config)

// WithNamespace identifies the bot or update stream in a shared store.
func WithNamespace(namespace string) Option {
	return func(config *config) { config.namespace = namespace }
}

// WithTTL sets successful-claim retention. The default is 24 hours.
func WithTTL(ttl time.Duration) Option { return func(config *config) { config.ttl = ttl } }

// WithDuplicate installs a duplicate handler.
func WithDuplicate(handler DuplicateFunc) Option {
	return func(config *config) { config.onDuplicate = handler }
}

// Manager provides duplicate suppression middleware.
type Manager struct {
	store       Store
	namespace   string
	ttl         time.Duration
	onDuplicate DuplicateFunc
}

// New creates a duplicate manager.
func New(store Store, options ...Option) *Manager {
	config := config{ttl: 24 * time.Hour}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return &Manager{
		store:       store,
		namespace:   config.namespace,
		ttl:         config.ttl,
		onDuplicate: config.onDuplicate,
	}
}

// Middleware atomically claims each positive update ID. A downstream error
// releases the claim so a synchronous webhook retry can process it again.
func (m *Manager) Middleware() framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(c *framework.Context) (resultErr error) {
			if next == nil {
				return nil
			}
			if m == nil || m.store == nil {
				return ErrStoreRequired
			}
			if m.ttl <= 0 {
				return ErrInvalidTTL
			}
			if c == nil || c.Update == nil || c.Update.UpdateID <= 0 {
				return next(c)
			}
			ctx := c.Context
			if ctx == nil {
				ctx = context.Background()
			}
			key := Key{Namespace: m.namespace, UpdateID: c.Update.UpdateID}
			claimed, err := m.store.Claim(ctx, key, m.ttl)
			if err != nil {
				return fmt.Errorf("hermes/dedupe: claim: %w", err)
			}
			if !claimed {
				if m.onDuplicate != nil {
					return m.onDuplicate(c)
				}
				return nil
			}

			defer func() {
				if value := recover(); value != nil {
					_ = m.store.Release(context.WithoutCancel(ctx), key)
					panic(value)
				}
				if resultErr != nil {
					releaseErr := m.store.Release(context.WithoutCancel(ctx), key)
					if releaseErr != nil {
						resultErr = errors.Join(resultErr, fmt.Errorf("hermes/dedupe: release: %w", releaseErr))
					}
				}
			}()
			return next(c)
		}
	}
}
