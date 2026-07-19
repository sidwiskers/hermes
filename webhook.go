package hermes

import (
	"context"
	"net/http"
	"runtime/debug"

	"github.com/sidwiskers/hermes/framework"
	runtimecore "github.com/sidwiskers/hermes/internal/runtime"
)

// WebhookSecretHeader is the Telegram header carrying the configured webhook
// secret token.
const WebhookSecretHeader = runtimecore.SecretHeader

const webhookSecretHeader = WebhookSecretHeader

// WebhookOptions configures webhook authentication and request decoding.
type WebhookOptions struct {
	// Secret is compared with Telegram's secret-token header when non-empty.
	Secret string
	// MaxBodyBytes bounds the request before JSON decoding. The default is 8 MiB.
	MaxBodyBytes int64
	// PreserveRawUpdate copies accepted JSON into Update.Raw. WithRawUpdates on
	// Bot enables this automatically.
	PreserveRawUpdate bool
	// MaxResponseBytes bounds a direct response from WebhookReplyHandler. The
	// default is 8 MiB.
	MaxResponseBytes int64
}

// WebhookHandler returns an HTTP handler suitable for mounting in an existing
// server. It authenticates, validates, decodes, and queues one Telegram update
// per request. Under overload it returns HTTP 503 with Retry-After.
func (b *Bot) WebhookHandler(options WebhookOptions) http.Handler {
	if b != nil && b.Client != nil && b.Client.RawUpdatesEnabled() {
		options.PreserveRawUpdate = true
	}
	return runtimecore.WebhookHandler(runtimecore.WebhookOptions{
		Secret:            options.Secret,
		MaxBodyBytes:      options.MaxBodyBytes,
		PreserveRawUpdate: options.PreserveRawUpdate,
		MaxResponseBytes:  options.MaxResponseBytes,
	}, b.queue)
}

// WebhookReplyHandler returns a bounded synchronous handler that may send one
// Bot API method in its HTTP response through Context.RespondWebhook. This
// avoids a second HTTP round trip when Telegram's direct webhook-reply tradeoff
// is appropriate. Handler failures return 500; overload returns 503 so
// Telegram can retry. Call Prepare first or construct the bot with
// WithBotUsername when mounting this handler in an existing server.
func (b *Bot) WebhookReplyHandler(options WebhookOptions) http.Handler {
	if b != nil && b.Client != nil && b.Client.RawUpdatesEnabled() {
		options.PreserveRawUpdate = true
	}
	return runtimecore.WebhookReplyHandler(runtimecore.WebhookOptions{
		Secret:            options.Secret,
		MaxBodyBytes:      options.MaxBodyBytes,
		PreserveRawUpdate: options.PreserveRawUpdate,
		MaxResponseBytes:  options.MaxResponseBytes,
	}, b.handleWebhookReply)
}

// ServeWebhook runs a hardened HTTP server on address and serves the handler
// only at path. It stops intake and drains active handlers when ctx is
// canceled. TLS termination is expected to happen at a reverse proxy.
func (b *Bot) ServeWebhook(
	ctx context.Context,
	address string,
	path string,
	options WebhookOptions,
) error {
	if b == nil || b.Client == nil {
		return ErrClientRequired
	}
	if err := b.ensureUsername(ctx); err != nil {
		return err
	}
	return runtimecore.ServeWebhook(ctx, address, path, b.WebhookHandler(options), b.Wait)
}

// ServeWebhookReplies runs the hardened webhook server in synchronous direct-
// reply mode. It shares the bot's global update concurrency bound and drains
// active requests during shutdown.
func (b *Bot) ServeWebhookReplies(
	ctx context.Context,
	address string,
	path string,
	options WebhookOptions,
) error {
	if b == nil || b.Client == nil {
		return ErrClientRequired
	}
	if err := b.ensureUsername(ctx); err != nil {
		return err
	}
	return runtimecore.ServeWebhook(ctx, address, path, b.WebhookReplyHandler(options), b.Wait)
}

func (b *Bot) handleWebhookReply(
	ctx context.Context,
	update *Update,
) (reply runtimecore.WebhookReply, resultErr error) {
	if b == nil || b.Client == nil || b.dispatcher == nil {
		return reply, ErrClientRequired
	}
	release, ok := b.dispatcher.Reserve(ctx, false)
	if !ok {
		return reply, ErrQueueFull
	}
	defer release()

	handlerCtx := b.acquireContext(ctx, update)
	defer b.releaseContext(handlerCtx)
	defer func() {
		if value := recover(); value != nil {
			panicErr := &framework.PanicError{Value: value, Stack: debug.Stack()}
			b.report(handlerCtx, panicErr)
			resultErr = panicErr
		}
	}()

	if err := b.router.Handle(handlerCtx); err != nil {
		b.report(handlerCtx, err)
		return reply, err
	}
	if response, ok := handlerCtx.DirectWebhookResponse(); ok {
		reply.Method = response.Method
		reply.Params = response.Params
	}
	return reply, nil
}
