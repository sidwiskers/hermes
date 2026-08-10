package fleet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/sidwiskers/hermes"
)

var (
	// ErrFleetRequired reports a method called on a nil Fleet.
	ErrFleetRequired = errors.New("hermes/fleet: fleet is required")
	// ErrBotRequired reports an attempt to mount a nil bot.
	ErrBotRequired = errors.New("hermes/fleet: bot is required")
	// ErrNameRequired reports a bot mounted without a name.
	ErrNameRequired = errors.New("hermes/fleet: bot name is required")
	// ErrDuplicateName reports two mounts with the same name.
	ErrDuplicateName = errors.New("hermes/fleet: bot name is already mounted")
	// ErrDuplicateBot reports one bot pointer mounted more than once.
	ErrDuplicateBot = errors.New("hermes/fleet: bot is already mounted")
	// ErrDuplicateWebhookPath reports two webhook bots using one path.
	ErrDuplicateWebhookPath = errors.New("hermes/fleet: webhook path is already mounted")
	// ErrWebhookPath reports a malformed webhook path.
	ErrWebhookPath = errors.New("hermes/fleet: invalid webhook path")
	// ErrWebhookAddressRequired reports webhook mounts without a server address.
	ErrWebhookAddressRequired = errors.New("hermes/fleet: webhook address is required")
	// ErrRunning reports mutation or a second Run while the Fleet is active.
	ErrRunning = errors.New("hermes/fleet: fleet is running")
	// ErrNoBots reports a Fleet run without mounted bots.
	ErrNoBots = errors.New("hermes/fleet: no bots are mounted")
	// ErrNoBotsRunning reports every mounted update source becoming unavailable.
	ErrNoBotsRunning = errors.New("hermes/fleet: no bot update sources are running")
	// ErrSourceStopped reports an update source returning without cancellation.
	ErrSourceStopped = errors.New("hermes/fleet: update source stopped unexpectedly")
)

// Mode identifies how a mounted bot receives updates.
type Mode string

const (
	// ModePolling uses Telegram long polling.
	ModePolling Mode = "polling"
	// ModeWebhook uses a queued webhook handler on the shared Fleet server.
	ModeWebhook Mode = "webhook"
	// ModeWebhookReplies uses synchronous retry-safe direct webhook replies.
	ModeWebhookReplies Mode = "webhook_replies"
)

// State is the lifecycle state of one mounted bot.
type State string

const (
	StateRegistered State = "registered"
	StatePreparing  State = "preparing"
	StateRunning    State = "running"
	StateStopped    State = "stopped"
	StateFailed     State = "failed"
)

// Phase identifies the Fleet lifecycle operation that failed.
type Phase string

const (
	PhasePrepare       Phase = "prepare"
	PhasePolling       Phase = "polling"
	PhaseWebhookServer Phase = "webhook_server"
	PhaseShutdown      Phase = "shutdown"
)

// FailurePolicy controls what happens when one bot source fails. Preparation
// and polling already retry transient Telegram failures internally.
type FailurePolicy uint8

const (
	// IsolateFailures keeps healthy bots running and reports the failed bot.
	IsolateFailures FailurePolicy = iota
	// StopAllOnFailure cancels every bot after the first isolated failure.
	StopAllOnFailure
)

// Failure describes a contained bot or shared-server failure.
type Failure struct {
	Bot   string
	Phase Phase
	Err   error
}

// Error implements error.
func (f *Failure) Error() string {
	if f == nil {
		return "<nil>"
	}
	if f.Bot == "" {
		return fmt.Sprintf("hermes/fleet: %s: %v", f.Phase, f.Err)
	}
	return fmt.Sprintf("hermes/fleet: bot %q %s: %v", f.Bot, f.Phase, f.Err)
}

// Unwrap exposes the underlying failure.
func (f *Failure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.Err
}

// ErrorHandler receives failures contained by IsolateFailures. It may be
// called concurrently for unrelated bots.
type ErrorHandler func(*Failure)

// Status is an immutable snapshot of one mounted bot.
type Status struct {
	Name      string
	Mode      Mode
	State     State
	LastError error
}

type config struct {
	httpClient       *http.Client
	webhookAddress   string
	botMaxConcurrent int
	failurePolicy    FailurePolicy
	errorHandler     ErrorHandler
}

// Option configures a Fleet.
type Option func(*config)

// WithHTTPClient sets the client shared by bots created with Fleet.NewBot.
// Fleet never mutates or closes the supplied client.
func WithHTTPClient(client *http.Client) Option {
	return func(config *config) {
		if client != nil {
			config.httpClient = client
		}
	}
}

// WithWebhookAddress sets the one listen address shared by webhook mounts.
// TLS termination is expected at a reverse proxy, as with Bot.ServeWebhook.
func WithWebhookAddress(address string) Option {
	return func(config *config) { config.webhookAddress = strings.TrimSpace(address) }
}

// WithBotMaxConcurrentUpdates overrides the Hermes per-bot update limit for
// bots created with Fleet.NewBot. Independently constructed bots retain their
// own setting.
func WithBotMaxConcurrentUpdates(limit int) Option {
	return func(config *config) {
		if limit > 0 {
			config.botMaxConcurrent = limit
		}
	}
}

// WithFailurePolicy selects isolated or fail-fast source behavior.
func WithFailurePolicy(policy FailurePolicy) Option {
	return func(config *config) {
		if policy == IsolateFailures || policy == StopAllOnFailure {
			config.failurePolicy = policy
		}
	}
}

// WithErrorHandler replaces the failure sink. Handler panics are contained.
func WithErrorHandler(handler ErrorHandler) Option {
	return func(config *config) {
		if handler != nil {
			config.errorHandler = handler
		}
	}
}

type mountConfig struct {
	mode        Mode
	poll        hermes.PollOptions
	path        string
	webhook     hermes.WebhookOptions
	directReply bool
}

// MountOption configures one mounted update source.
type MountOption func(*mountConfig)

// WithPolling selects long polling. Zero PollOptions use Hermes production
// defaults.
func WithPolling(options hermes.PollOptions) MountOption {
	return func(config *mountConfig) {
		config.mode = ModePolling
		config.poll = options
		config.path = ""
		config.directReply = false
	}
}

// WithWebhook selects queued webhook delivery on the shared exact path.
func WithWebhook(path string, options hermes.WebhookOptions) MountOption {
	return func(config *mountConfig) {
		config.mode = ModeWebhook
		config.path = path
		config.webhook = options
		config.directReply = false
	}
}

// WithWebhookReplies selects synchronous retry-safe webhook delivery with
// optional direct Bot API replies.
func WithWebhookReplies(path string, options hermes.WebhookOptions) MountOption {
	return func(config *mountConfig) {
		config.mode = ModeWebhookReplies
		config.path = path
		config.webhook = options
		config.directReply = true
	}
}

type mountedBot struct {
	name        string
	bot         *hermes.Bot
	mode        Mode
	poll        hermes.PollOptions
	path        string
	webhook     hermes.WebhookOptions
	directReply bool
	state       State
	lastError   error
}

// Fleet coordinates several independent Hermes bots in one process.
type Fleet struct {
	mu sync.Mutex

	httpClient       *http.Client
	webhookAddress   string
	botMaxConcurrent int
	failurePolicy    FailurePolicy
	errorHandler     ErrorHandler

	mounts   map[string]*mountedBot
	order    []string
	botNames map[*hermes.Bot]string
	paths    map[string]string
	running  bool
	address  string
}

// New creates an empty Fleet without starting goroutines or network I/O.
func New(options ...Option) *Fleet {
	configuration := config{
		httpClient:    &http.Client{},
		failurePolicy: IsolateFailures,
		errorHandler: func(failure *Failure) {
			if failure == nil {
				return
			}
			slog.Error("Hermes Fleet source failed",
				"bot", failure.Bot,
				"phase", failure.Phase,
				"error", failure.Err,
			)
		},
	}
	for _, option := range options {
		if option != nil {
			option(&configuration)
		}
	}
	return &Fleet{
		httpClient:       configuration.httpClient,
		webhookAddress:   configuration.webhookAddress,
		botMaxConcurrent: configuration.botMaxConcurrent,
		failurePolicy:    configuration.failurePolicy,
		errorHandler:     configuration.errorHandler,
		mounts:           make(map[string]*mountedBot),
		botNames:         make(map[*hermes.Bot]string),
		paths:            make(map[string]string),
	}
}

// NewBot creates an ordinary Hermes bot using the Fleet's shared HTTP client.
// All supplied Hermes options apply normally; Fleet-owned resource options are
// applied last. The bot is not mounted until Mount succeeds.
func (f *Fleet) NewBot(token string, options ...hermes.Option) *hermes.Bot {
	if f == nil {
		return hermes.New(token, options...)
	}
	f.mu.Lock()
	client := f.httpClient
	limit := f.botMaxConcurrent
	f.mu.Unlock()
	owned := make([]hermes.Option, 0, len(options)+2)
	owned = append(owned, options...)
	owned = append(owned, hermes.WithHTTPClient(client))
	if limit > 0 {
		owned = append(owned, hermes.WithMaxConcurrentUpdates(limit))
	}
	return hermes.New(token, owned...)
}

// Mount registers a named bot. Polling is the default update source. Mounts
// remain mutable until Run begins and may be extended after a completed Run.
func (f *Fleet) Mount(name string, bot *hermes.Bot, options ...MountOption) error {
	if f == nil {
		return ErrFleetRequired
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameRequired
	}
	if bot == nil || bot.Client == nil {
		return ErrBotRequired
	}
	mountOptions := mountConfig{mode: ModePolling}
	for _, option := range options {
		if option != nil {
			option(&mountOptions)
		}
	}
	if mountOptions.mode == ModeWebhook || mountOptions.mode == ModeWebhookReplies {
		if !validWebhookPath(mountOptions.path) {
			return ErrWebhookPath
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.running {
		return ErrRunning
	}
	if _, exists := f.mounts[name]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateName, name)
	}
	if mountedName, exists := f.botNames[bot]; exists {
		return fmt.Errorf("%w as %q", ErrDuplicateBot, mountedName)
	}
	if mountOptions.path != "" {
		if mountedName, exists := f.paths[mountOptions.path]; exists {
			return fmt.Errorf("%w by %q", ErrDuplicateWebhookPath, mountedName)
		}
	}
	mounted := &mountedBot{
		name:        name,
		bot:         bot,
		mode:        mountOptions.mode,
		poll:        mountOptions.poll,
		path:        mountOptions.path,
		webhook:     mountOptions.webhook,
		directReply: mountOptions.directReply,
		state:       StateRegistered,
	}
	f.mounts[name] = mounted
	f.order = append(f.order, name)
	f.botNames[bot] = name
	if mounted.path != "" {
		f.paths[mounted.path] = name
	}
	return nil
}

// Len returns the number of mounted bots.
func (f *Fleet) Len() int {
	if f == nil {
		return 0
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.mounts)
}

// Status returns stable registration-order lifecycle snapshots.
func (f *Fleet) Status() []Status {
	if f == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]Status, 0, len(f.order))
	for _, name := range f.order {
		mounted := f.mounts[name]
		result = append(result, Status{
			Name:      mounted.name,
			Mode:      mounted.mode,
			State:     mounted.state,
			LastError: mounted.lastError,
		})
	}
	return result
}

// WebhookAddress returns the active shared listener address. It is empty when
// no webhook server is running.
func (f *Fleet) WebhookAddress() string {
	if f == nil {
		return ""
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.address
}

type prepareResult struct {
	mounted *mountedBot
	err     error
}

type sourceEvent struct {
	mounts []*mountedBot
	phase  Phase
	err    error
}

// Run prepares and serves every mounted bot until ctx is canceled. By default
// a failed bot is isolated while healthy bots continue. Shared webhook-server
// failure affects webhook mounts but does not stop polling mounts unless
// StopAllOnFailure is configured.
func (f *Fleet) Run(ctx context.Context) error {
	if f == nil {
		return ErrFleetRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return nil
	}
	mounts, policy, webhookAddress, err := f.beginRun()
	if err != nil {
		return err
	}
	defer f.finishRun(mounts)

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	prepared, failures := f.prepare(runCtx, mounts)
	if policy == StopAllOnFailure && len(failures) != 0 {
		return errors.Join(failures...)
	}
	if len(prepared) == 0 {
		return errors.Join(append([]error{ErrNoBotsRunning}, failures...)...)
	}

	var polling []*mountedBot
	var webhooks []*mountedBot
	for _, mounted := range prepared {
		if mounted.mode == ModePolling {
			polling = append(polling, mounted)
		} else {
			webhooks = append(webhooks, mounted)
		}
	}

	var server *http.Server
	var listener net.Listener
	if len(webhooks) != 0 {
		server, listener, err = f.openWebhookServer(webhookAddress, webhooks)
		if err != nil {
			failure := &Failure{Phase: PhaseWebhookServer, Err: err}
			failures = append(failures, failure)
			for _, mounted := range webhooks {
				f.setFailure(mounted, err)
			}
			f.report(failure)
			webhooks = nil
			if policy == StopAllOnFailure || len(polling) == 0 {
				return errors.Join(append([]error{ErrNoBotsRunning}, failures...)...)
			}
		}
	}

	sourceCount := len(polling)
	if listener != nil {
		sourceCount++
	}
	if sourceCount == 0 {
		return errors.Join(append([]error{ErrNoBotsRunning}, failures...)...)
	}
	events := make(chan sourceEvent, sourceCount)
	var sources sync.WaitGroup
	for _, mounted := range polling {
		mounted := mounted
		f.setRunning(mounted)
		sources.Add(1)
		go func() {
			defer sources.Done()
			err := mounted.bot.Poll(runCtx, mounted.poll)
			events <- sourceEvent{mounts: []*mountedBot{mounted}, phase: PhasePolling, err: err}
		}()
	}
	if listener != nil {
		for _, mounted := range webhooks {
			f.setRunning(mounted)
		}
		f.setAddress(listener.Addr().String())
		sources.Add(1)
		go func() {
			defer sources.Done()
			err := server.Serve(listener)
			if errors.Is(err, http.ErrServerClosed) {
				err = nil
			}
			events <- sourceEvent{mounts: webhooks, phase: PhaseWebhookServer, err: err}
		}()
	}

	for sourceCount > 0 {
		select {
		case <-ctx.Done():
			cancel()
			shutdownErr := shutdownWebhookServer(server, webhooks)
			sources.Wait()
			if shutdownErr != nil {
				failure := &Failure{Phase: PhaseShutdown, Err: shutdownErr}
				f.report(failure)
				return failure
			}
			return nil
		case event := <-events:
			sourceCount--
			if runCtx.Err() != nil {
				continue
			}
			if event.phase == PhaseWebhookServer {
				f.setAddress("")
			}
			if event.err == nil {
				event.err = ErrSourceStopped
			}
			for _, mounted := range event.mounts {
				f.setFailure(mounted, event.err)
				failure := &Failure{Bot: mounted.name, Phase: event.phase, Err: event.err}
				failures = append(failures, failure)
				f.report(failure)
			}
			if policy == StopAllOnFailure {
				cancel()
				shutdownErr := shutdownWebhookServer(server, webhooks)
				sources.Wait()
				if shutdownErr != nil {
					failures = append(failures, &Failure{Phase: PhaseShutdown, Err: shutdownErr})
				}
				return errors.Join(failures...)
			}
			if sourceCount == 0 {
				cancel()
				shutdownErr := shutdownWebhookServer(server, webhooks)
				sources.Wait()
				if shutdownErr != nil {
					failures = append(failures, &Failure{Phase: PhaseShutdown, Err: shutdownErr})
				}
				return errors.Join(append([]error{ErrNoBotsRunning}, failures...)...)
			}
		}
	}
	return nil
}

func (f *Fleet) beginRun() ([]*mountedBot, FailurePolicy, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.running {
		return nil, f.failurePolicy, "", ErrRunning
	}
	if len(f.mounts) == 0 {
		return nil, f.failurePolicy, "", ErrNoBots
	}
	for _, name := range f.order {
		mounted := f.mounts[name]
		if mounted.mode != ModePolling && f.webhookAddress == "" {
			return nil, f.failurePolicy, "", ErrWebhookAddressRequired
		}
	}
	f.running = true
	f.address = ""
	mounts := make([]*mountedBot, 0, len(f.order))
	for _, name := range f.order {
		mounted := f.mounts[name]
		mounted.state = StatePreparing
		mounted.lastError = nil
		mounts = append(mounts, mounted)
	}
	return mounts, f.failurePolicy, f.webhookAddress, nil
}

func (f *Fleet) finishRun(mounts []*mountedBot) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.running = false
	f.address = ""
	for _, mounted := range mounts {
		if mounted.state == StatePreparing || mounted.state == StateRunning {
			mounted.state = StateStopped
		}
	}
}

func (f *Fleet) prepare(ctx context.Context, mounts []*mountedBot) ([]*mountedBot, []error) {
	results := make(chan prepareResult, len(mounts))
	for _, mounted := range mounts {
		mounted := mounted
		go func() { results <- prepareResult{mounted: mounted, err: mounted.bot.Prepare(ctx)} }()
	}
	prepared := make([]*mountedBot, 0, len(mounts))
	var failures []error
	for range mounts {
		result := <-results
		if result.err == nil {
			prepared = append(prepared, result.mounted)
			continue
		}
		f.setFailure(result.mounted, result.err)
		failure := &Failure{Bot: result.mounted.name, Phase: PhasePrepare, Err: result.err}
		failures = append(failures, failure)
		f.report(failure)
	}
	return prepared, failures
}

func (f *Fleet) setRunning(mounted *mountedBot) {
	f.mu.Lock()
	mounted.state = StateRunning
	mounted.lastError = nil
	f.mu.Unlock()
}

func (f *Fleet) setFailure(mounted *mountedBot, err error) {
	f.mu.Lock()
	mounted.state = StateFailed
	mounted.lastError = err
	f.mu.Unlock()
}

func (f *Fleet) setAddress(address string) {
	f.mu.Lock()
	f.address = address
	f.mu.Unlock()
}

func (f *Fleet) report(failure *Failure) {
	f.mu.Lock()
	handler := f.errorHandler
	f.mu.Unlock()
	if handler == nil {
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("Hermes Fleet error handler panicked", "panic", recovered)
		}
	}()
	handler(failure)
}

func validWebhookPath(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.ContainsAny(path, "{}?# \t\r\n")
}
