package hermes

import (
	"context"
	"log/slog"
	"net/http"
	stdruntime "runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sidwiskers/hermes/api"
	"github.com/sidwiskers/hermes/framework"
	runtimecore "github.com/sidwiskers/hermes/internal/runtime"
)

type errorHandlerBox struct {
	fn ErrorHandler
}

type usernameBox struct {
	value string
}

// ErrorHandler receives errors returned by asynchronously dispatched handlers.
// It may be replaced safely while the bot is running.
type ErrorHandler func(*Context, error)

// Config contains construction-time settings for the combined facade.
type Config struct {
	// HTTPClient performs Bot API requests. A nil value uses a fresh client.
	HTTPClient *http.Client
	// BaseURL is the Bot API origin without a trailing slash.
	BaseURL string
	// UserAgent is sent with every Bot API request.
	UserAgent string
	// ResponseLimit is the maximum accepted response body size in bytes.
	ResponseLimit int64
	// PreserveRawUpdates copies each decoded update's original JSON.
	PreserveRawUpdates bool
	// TestEnvironment uses Telegram's separate method and file test endpoints.
	TestEnvironment bool
	// APIObserver receives outbound Bot API lifecycle events.
	APIObserver api.Observer
	// ContextPooling reuses handler contexts and is enabled by default.
	ContextPooling bool
	// MaxConcurrentUpdates bounds simultaneously executing update handlers.
	MaxConcurrentUpdates int
	// BotUsername avoids a startup getMe request for command parsing.
	BotUsername string
	// ErrorHandler receives asynchronous handler errors.
	ErrorHandler ErrorHandler
}

// Option mutates Bot construction settings. Applications may compose custom
// options when useful.
type Option func(*Config)

// Bot combines the standalone low-level api.Client with the routing framework.
// API methods are promoted directly through the embedded client.
type Bot struct {
	*api.Client
	username   atomic.Pointer[usernameBox]
	usernameMu sync.Mutex

	router       *Router
	contextPool  *framework.ContextPool
	dispatcher   *runtimecore.Dispatcher
	errorHandler atomic.Pointer[errorHandlerBox]
}

// New constructs a bot without performing network I/O. The token is validated
// when the first Bot API request is made. Run discovers the bot username with
// getMe unless WithBotUsername supplied it at construction time.
func New(token string, options ...Option) *Bot {
	workers := stdruntime.GOMAXPROCS(0) * 8
	if workers < 8 {
		workers = 8
	}
	config := Config{
		HTTPClient:           &http.Client{},
		BaseURL:              api.DefaultBaseURL,
		UserAgent:            api.DefaultUserAgent,
		ResponseLimit:        api.DefaultResponseLimit,
		ContextPooling:       true,
		MaxConcurrentUpdates: workers,
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	if config.MaxConcurrentUpdates <= 0 {
		config.MaxConcurrentUpdates = workers
	}

	bot := &Bot{
		Client: api.New(token,
			api.WithHTTPClient(config.HTTPClient),
			api.WithBaseURL(config.BaseURL),
			api.WithUserAgent(config.UserAgent),
			api.WithResponseLimit(config.ResponseLimit),
			api.WithRawUpdates(config.PreserveRawUpdates),
			api.WithTestEnvironment(config.TestEnvironment),
			api.WithObserver(config.APIObserver),
		),
		router: NewRouter(),
	}
	bot.storeUsername(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(config.BotUsername)), "@"))
	if config.ContextPooling {
		bot.contextPool = framework.NewContextPool()
	}
	handler := config.ErrorHandler
	if handler == nil {
		handler = func(_ *Context, err error) {
			slog.Error("telegram update handler failed", "error", err)
		}
	}
	bot.errorHandler.Store(&errorHandlerBox{fn: handler})
	bot.dispatcher = runtimecore.NewDispatcher(config.MaxConcurrentUpdates, func(ctx context.Context, update *Update) {
		handlerCtx := bot.acquireContext(ctx, update)
		defer bot.releaseContext(handlerCtx)
		defer bot.recoverHandlerPanic(handlerCtx)
		bot.report(handlerCtx, bot.router.Handle(handlerCtx))
	})
	return bot
}

// WithHTTPClient replaces the client used for all Bot API requests.
// Hermes does not mutate the supplied client.
func WithHTTPClient(client *http.Client) Option {
	return func(config *Config) {
		if client != nil {
			config.HTTPClient = client
		}
	}
}

// WithBaseURL replaces Telegram's API origin. It is primarily useful for a
// local Bot API server, proxy, or test transport.
func WithBaseURL(baseURL string) Option {
	return func(config *Config) {
		if value := strings.TrimRight(strings.TrimSpace(baseURL), "/"); value != "" {
			config.BaseURL = value
		}
	}
}

// WithUserAgent replaces the User-Agent sent with Bot API requests.
func WithUserAgent(userAgent string) Option {
	return func(config *Config) {
		if value := strings.TrimSpace(userAgent); value != "" {
			config.UserAgent = value
		}
	}
}

// WithResponseLimit bounds a Bot API response body. Non-positive values leave
// the default limit unchanged.
func WithResponseLimit(bytes int64) Option {
	return func(config *Config) {
		if bytes > 0 {
			config.ResponseLimit = bytes
		}
	}
}

// WithRawUpdates preserves each incoming update's original JSON in Update.Raw.
// It is disabled by default because preservation adds an allocation and copies
// the complete payload. Enable it only for forward-compatibility or debugging.
func WithRawUpdates(enabled bool) Option {
	return func(config *Config) { config.PreserveRawUpdates = enabled }
}

// WithTestEnvironment routes method calls and file downloads to Telegram's
// separate Bot API test environment. It requires a bot token created inside
// Telegram's test DC.
func WithTestEnvironment(enabled bool) Option {
	return func(config *Config) { config.TestEnvironment = enabled }
}

// WithAPIObserver installs an outbound Bot API lifecycle observer. Observer
// panics are contained and never interrupt a request.
func WithAPIObserver(observer api.Observer) Option {
	return func(config *Config) { config.APIObserver = observer }
}

// WithContextPooling controls reuse of handler Context values. It is enabled
// by default for zero-allocation routing. Disable it only when application code
// intentionally retains *Context after a handler returns.
func WithContextPooling(enabled bool) Option {
	return func(config *Config) { config.ContextPooling = enabled }
}

// WithMaxConcurrentUpdates bounds simultaneously executing handlers.
// Polling applies backpressure; webhooks reject overload with HTTP 503 so
// Telegram can retry it.
func WithMaxConcurrentUpdates(limit int) Option {
	return func(config *Config) {
		if limit > 0 {
			config.MaxConcurrentUpdates = limit
		}
	}
}

// WithErrorHandler sets the asynchronous handler-error sink. A nil handler is
// ignored; the default logs with slog.Default().
func WithErrorHandler(handler ErrorHandler) Option {
	return func(config *Config) {
		if handler != nil {
			config.ErrorHandler = handler
		}
	}
}

// WithBotUsername avoids the startup getMe request used for addressed-command
// parsing. Both "name" and "@name" forms are accepted.
func WithBotUsername(username string) Option {
	return func(config *Config) { config.BotUsername = username }
}

// SetErrorHandler atomically replaces the asynchronous handler-error sink.
// Passing nil leaves the current handler unchanged.
func (b *Bot) SetErrorHandler(handler ErrorHandler) {
	if handler != nil {
		b.errorHandler.Store(&errorHandlerBox{fn: handler})
	}
}

// Command registers an exact slash-command handler.
func (b *Bot) Command(command string, handler Handler) { b.router.Command(command, handler) }

// Callback registers an exact callback-data handler.
func (b *Bot) Callback(data string, handler Handler) { b.router.Callback(data, handler) }

// CallbackPrefix registers a callback-data prefix handler. The longest
// matching prefix wins.
func (b *Bot) CallbackPrefix(prefix string, handler Handler) {
	b.router.CallbackPrefix(prefix, handler)
}

// On registers an ordered filtered route. The first matching route runs.
func (b *Bot) On(filter Filter, handler Handler) { b.router.On(filter, handler) }

// Group creates a route group with shared filters.
func (b *Bot) Group(filters ...Filter) *Group { return b.router.Group(filters...) }

// Use appends global middleware in declaration order.
func (b *Bot) Use(middleware ...Middleware) { b.router.Use(middleware...) }

// OnUpdate sets the fallback handler used when no command, callback, or
// filtered route matches.
func (b *Bot) OnUpdate(handler Handler) { b.router.OnUpdate(handler) }

// Handle routes one update synchronously and returns the handler error. It is
// useful for tests and custom update sources; Poll and ServeWebhook use the
// bounded asynchronous dispatcher instead.
func (b *Bot) Handle(ctx context.Context, update *Update) error {
	if b == nil || b.Client == nil {
		return ErrClientRequired
	}
	if update == nil {
		return nil
	}
	handlerCtx := b.acquireContext(ctx, update)
	defer b.releaseContext(handlerCtx)
	return b.router.Handle(handlerCtx)
}

func (b *Bot) acquireContext(ctx context.Context, update *Update) *Context {
	if b != nil && b.contextPool != nil {
		return b.contextPool.Acquire(ctx, b.Client, update, b.loadUsername())
	}
	return newContext(ctx, b, update)
}

func (b *Bot) releaseContext(value *Context) {
	if b != nil && b.contextPool != nil {
		b.contextPool.Release(value)
	}
}

func (b *Bot) report(handlerCtx *Context, err error) {
	if err == nil {
		return
	}
	defer func() {
		if value := recover(); value != nil {
			slog.Error("telegram error handler panicked", "panic", value, "stack", string(debug.Stack()))
		}
	}()
	if box := b.errorHandler.Load(); box != nil && box.fn != nil {
		box.fn(handlerCtx, err)
	}
}

func (b *Bot) recoverHandlerPanic(handlerCtx *Context) {
	if value := recover(); value != nil {
		b.report(handlerCtx, &framework.PanicError{Value: value, Stack: debug.Stack()})
	}
}

func (b *Bot) queue(ctx context.Context, update *Update, wait bool) bool {
	return b != nil && b.dispatcher != nil && b.dispatcher.Queue(ctx, update, wait)
}

// Wait blocks until every asynchronously dispatched handler has returned.
func (b *Bot) Wait() {
	if b != nil && b.dispatcher != nil {
		b.dispatcher.Wait()
	}
}

// Prepare resolves the bot username needed for addressed-command routing.
// Run and the ServeWebhook helpers call it automatically. Applications that
// mount WebhookHandler or WebhookReplyHandler in their own server should call
// Prepare during startup or provide WithBotUsername.
func (b *Bot) Prepare(ctx context.Context) error { return b.ensureUsername(ctx) }

// Run starts long polling with production-oriented defaults and drains active
// handlers before returning. Canceling ctx is a normal shutdown and returns
// nil once polling has begun.
func (b *Bot) Run(ctx context.Context) error {
	if b == nil || b.Client == nil {
		return ErrClientRequired
	}
	if err := b.ensureUsername(ctx); err != nil {
		return err
	}
	return b.Poll(ctx, PollOptions{
		Timeout: 50, Limit: 100,
		MinBackoff: 250 * time.Millisecond,
		MaxBackoff: 8 * time.Second,
	})
}

func (b *Bot) ensureUsername(ctx context.Context) error {
	if b == nil || b.Client == nil {
		return ErrClientRequired
	}
	if b.loadUsername() != "" {
		return nil
	}
	b.usernameMu.Lock()
	defer b.usernameMu.Unlock()
	if b.loadUsername() != "" {
		return nil
	}
	me, err := b.GetMe(ctx)
	if err != nil {
		return err
	}
	b.storeUsername(strings.ToLower(me.Username))
	return nil
}

func (b *Bot) loadUsername() string {
	if b == nil {
		return ""
	}
	if box := b.username.Load(); box != nil {
		return box.value
	}
	return ""
}

func (b *Bot) storeUsername(username string) {
	if b != nil {
		b.username.Store(&usernameBox{value: username})
	}
}
