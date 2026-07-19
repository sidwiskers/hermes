package hermes

import (
	"context"
	"time"

	runtimecore "github.com/sidwiskers/hermes/internal/runtime"
)

// PollOptions configures long polling and transient-failure backoff. Zero
// values use production defaults: 100 updates, a 50-second Telegram timeout,
// and exponential backoff from 250 milliseconds to 8 seconds.
type PollOptions struct {
	// Offset is the first update identifier Telegram should return.
	Offset int64
	// Limit is clamped to Telegram's maximum of 100.
	Limit int
	// Timeout is Telegram's long-poll timeout in seconds.
	Timeout int
	// AllowedUpdates limits update types at Telegram's edge.
	AllowedUpdates []string
	// MinBackoff and MaxBackoff bound retries after transient failures.
	MinBackoff time.Duration
	MaxBackoff time.Duration
}

// Poll receives updates with the supplied long-poll and retry settings. It
// applies bounded backpressure and drains active handlers before returning.
func (b *Bot) Poll(ctx context.Context, options PollOptions) error {
	if b == nil || b.Client == nil {
		return ErrClientRequired
	}
	return runtimecore.Poll(ctx, b.Client, b.queue, b.Wait, runtimecore.PollOptions{
		Offset:         options.Offset,
		Limit:          options.Limit,
		Timeout:        options.Timeout,
		AllowedUpdates: options.AllowedUpdates,
		MinBackoff:     options.MinBackoff,
		MaxBackoff:     options.MaxBackoff,
	})
}
