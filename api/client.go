package api

import (
	"net/http"
	"strings"
)

const (
	// DefaultBaseURL is Telegram's hosted Bot API origin.
	DefaultBaseURL = "https://api.telegram.org"
	// DefaultResponseLimit is the maximum accepted response size, 32 MiB.
	DefaultResponseLimit = 32 << 20
	// DefaultUserAgent is sent by clients that do not configure one.
	DefaultUserAgent = "hermes-go/1.0.0"
)

// Config contains construction-time settings for Client.
type Config struct {
	// HTTPClient performs requests. A nil value uses a fresh http.Client.
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
	// Observer receives outbound Bot API lifecycle events.
	Observer Observer
}

// Option mutates Client construction settings.
type Option func(*Config)

// Client performs typed and raw Telegram Bot API calls.
type Client struct {
	token              string
	methodPrefix       string
	filePrefix         string
	userAgent          string
	client             *http.Client
	responseLimit      int64
	preserveRawUpdates bool
	observer           Observer
}

// New creates an independent low-level Telegram Bot API client.
func New(token string, options ...Option) *Client {
	config := Config{
		HTTPClient:    &http.Client{},
		BaseURL:       DefaultBaseURL,
		UserAgent:     DefaultUserAgent,
		ResponseLimit: DefaultResponseLimit,
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{}
	}
	if value := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/"); value != "" {
		config.BaseURL = value
	} else {
		config.BaseURL = DefaultBaseURL
	}
	if value := strings.TrimSpace(config.UserAgent); value != "" {
		config.UserAgent = value
	} else {
		config.UserAgent = DefaultUserAgent
	}
	if config.ResponseLimit <= 0 {
		config.ResponseLimit = DefaultResponseLimit
	}

	trimmedToken := strings.TrimSpace(token)
	methodPrefix := config.BaseURL + "/bot" + trimmedToken + "/"
	filePrefix := config.BaseURL + "/file/bot" + trimmedToken + "/"
	if config.TestEnvironment {
		methodPrefix += "test/"
		filePrefix += "test/"
	}

	return &Client{
		token:              trimmedToken,
		methodPrefix:       methodPrefix,
		filePrefix:         filePrefix,
		userAgent:          config.UserAgent,
		client:             config.HTTPClient,
		responseLimit:      config.ResponseLimit,
		preserveRawUpdates: config.PreserveRawUpdates,
		observer:           config.Observer,
	}
}

// WithHTTPClient replaces the HTTP client used for every request.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(config *Config) {
		if httpClient != nil {
			config.HTTPClient = httpClient
		}
	}
}

// WithBaseURL replaces Telegram's API origin.
func WithBaseURL(baseURL string) Option {
	return func(config *Config) { config.BaseURL = baseURL }
}

// WithUserAgent replaces the request User-Agent.
func WithUserAgent(userAgent string) Option {
	return func(config *Config) { config.UserAgent = userAgent }
}

// WithResponseLimit bounds accepted response bodies. Non-positive values keep
// the default.
func WithResponseLimit(bytes int64) Option {
	return func(config *Config) {
		if bytes > 0 {
			config.ResponseLimit = bytes
		}
	}
}

// WithRawUpdates controls whether polling responses preserve each update's
// complete original JSON in Update.Raw. It is disabled by default to avoid an
// allocation and payload copy on every update.
func WithRawUpdates(enabled bool) Option {
	return func(config *Config) { config.PreserveRawUpdates = enabled }
}

// WithTestEnvironment routes method calls and file downloads to Telegram's
// separate Bot API test environment. It requires a bot token created inside
// Telegram's test DC.
func WithTestEnvironment(enabled bool) Option {
	return func(config *Config) { config.TestEnvironment = enabled }
}

// WithObserver installs an outbound Bot API lifecycle observer. Observer
// panics are contained and never interrupt a request.
func WithObserver(observer Observer) Option {
	return func(config *Config) { config.Observer = observer }
}

// RawUpdatesEnabled reports whether this client preserves polling update JSON.
func (c *Client) RawUpdatesEnabled() bool {
	return c != nil && c.preserveRawUpdates
}
