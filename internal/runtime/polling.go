package runtime

import (
	"context"
	"errors"
	"time"

	"github.com/sidwiskers/hermes/api"
	telegram "github.com/sidwiskers/hermes/types"
)

type UpdateSource interface {
	GetUpdates(context.Context, api.GetUpdatesParams) ([]telegram.Update, error)
}

var (
	ErrUpdateSourceRequired = errors.New("hermes: polling update source is required")
	ErrDispatchRequired     = errors.New("hermes: polling dispatch function is required")
)

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

func (options PollOptions) normalized() PollOptions {
	if options.Limit <= 0 || options.Limit > 100 {
		options.Limit = 100
	}
	if options.Timeout <= 0 {
		options.Timeout = 50
	}
	if options.MinBackoff <= 0 {
		options.MinBackoff = 250 * time.Millisecond
	}
	if options.MaxBackoff <= 0 {
		options.MaxBackoff = 8 * time.Second
	}
	if options.MaxBackoff < options.MinBackoff {
		options.MaxBackoff = options.MinBackoff
	}
	return options
}

func Poll(
	ctx context.Context,
	source UpdateSource,
	dispatch func(context.Context, *telegram.Update, bool) bool,
	wait func(),
	options PollOptions,
) error {
	if source == nil {
		return ErrUpdateSourceRequired
	}
	if dispatch == nil {
		return ErrDispatchRequired
	}
	options = options.normalized()
	offset := options.Offset
	backoff := options.MinBackoff
	if wait != nil {
		defer wait()
	}

	for {
		updates, err := source.GetUpdates(ctx, api.GetUpdatesParams{
			Offset: offset, Limit: options.Limit, Timeout: options.Timeout,
			AllowedUpdates: options.AllowedUpdates,
		})
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) {
				return nil
			}
			delay := backoff
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.RetryAfter() > 0 {
				delay = retryAfterDuration(apiErr.RetryAfter())
			}
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil
			case <-timer.C:
			}
			backoff = nextBackoff(backoff, options.MaxBackoff)
			continue
		}

		backoff = options.MinBackoff
		for index := range updates {
			update := updates[index]
			if update.UpdateID < offset {
				continue
			}
			offset = update.UpdateID + 1
			if !dispatch(ctx, &update, true) {
				return nil
			}
		}
	}
}

func nextBackoff(current, maximum time.Duration) time.Duration {
	if current >= maximum || current > maximum/2 {
		return maximum
	}
	return current * 2
}

func retryAfterDuration(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	const maximum = time.Duration(1<<63 - 1)
	if int64(seconds) > int64(maximum/time.Second) {
		return maximum
	}
	return time.Duration(seconds) * time.Second
}
