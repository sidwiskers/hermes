package runtime

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	telegram "github.com/sidwiskers/hermes/types"
)

const SecretHeader = "X-Telegram-Bot-Api-Secret-Token"

var (
	ErrQueueFull              = errors.New("hermes: update queue is full")
	ErrWebhookHandlerRequired = errors.New("hermes: webhook handler is required")
	ErrWebhookReplyInvalid    = errors.New("hermes: invalid webhook reply")
	ErrWebhookReplyTooLarge   = errors.New("hermes: webhook reply exceeds configured limit")
)

type WebhookOptions struct {
	// Secret is compared with Telegram's secret-token header when non-empty.
	Secret string
	// MaxBodyBytes bounds the request before JSON decoding. The default is 8 MiB.
	MaxBodyBytes int64
	// PreserveRawUpdate copies the accepted JSON into Update.Raw.
	PreserveRawUpdate bool
	// MaxResponseBytes bounds a direct webhook reply. The default is 8 MiB.
	MaxResponseBytes int64
}

// WebhookReply is a Bot API method call encoded directly in a synchronous
// webhook HTTP response.
type WebhookReply struct {
	Method string
	Params any
}

func WebhookHandler(
	options WebhookOptions,
	enqueue func(context.Context, *telegram.Update, bool) bool,
) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		update, ok := decodeWebhook(writer, request, options)
		if !ok {
			return
		}

		handlerContext := context.WithoutCancel(request.Context())
		if enqueue == nil || !enqueue(handlerContext, update, false) {
			writer.Header().Set("Retry-After", "1")
			http.Error(writer, ErrQueueFull.Error(), http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// WebhookReplyHandler validates and handles an update synchronously, allowing
// one Bot API method call to be returned in the HTTP response. Handler errors
// return 500 so Telegram can retry; ErrQueueFull returns 503 with Retry-After.
func WebhookReplyHandler(
	options WebhookOptions,
	handle func(context.Context, *telegram.Update) (WebhookReply, error),
) http.Handler {
	maxResponse := options.MaxResponseBytes
	if maxResponse <= 0 {
		maxResponse = 8 << 20
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		update, ok := decodeWebhook(writer, request, options)
		if !ok {
			return
		}
		if handle == nil {
			http.Error(writer, "handler unavailable", http.StatusInternalServerError)
			return
		}
		reply, err := handle(request.Context(), update)
		if err != nil {
			if errors.Is(err, ErrQueueFull) {
				writer.Header().Set("Retry-After", "1")
				http.Error(writer, ErrQueueFull.Error(), http.StatusServiceUnavailable)
				return
			}
			http.Error(writer, "webhook handler failed", http.StatusInternalServerError)
			return
		}
		if reply.Method == "" {
			writer.WriteHeader(http.StatusOK)
			return
		}
		data, err := encodeWebhookReply(reply)
		if err != nil {
			http.Error(writer, ErrWebhookReplyInvalid.Error(), http.StatusInternalServerError)
			return
		}
		if int64(len(data)) > maxResponse {
			http.Error(writer, ErrWebhookReplyTooLarge.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(data)
	})
}

func decodeWebhook(
	writer http.ResponseWriter,
	request *http.Request,
	options WebhookOptions,
) (*telegram.Update, bool) {
	if request.Method != http.MethodPost {
		writer.Header().Set("Allow", http.MethodPost)
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return nil, false
	}
	if options.Secret != "" && subtle.ConstantTimeCompare(
		[]byte(request.Header.Get(SecretHeader)), []byte(options.Secret),
	) != 1 {
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	maxBody := options.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 8 << 20
	}
	body := http.MaxBytesReader(writer, request.Body, maxBody)
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(writer, "update too large", http.StatusRequestEntityTooLarge)
			return nil, false
		}
		http.Error(writer, "invalid update", http.StatusBadRequest)
		return nil, false
	}
	if len(data) == 0 {
		http.Error(writer, "empty update", http.StatusBadRequest)
		return nil, false
	}
	update, err := telegram.DecodeUpdate(data, options.PreserveRawUpdate)
	if err != nil {
		http.Error(writer, "invalid update", http.StatusBadRequest)
		return nil, false
	}
	return &update, true
}

func encodeWebhookReply(reply WebhookReply) ([]byte, error) {
	if !validWebhookMethod(reply.Method) {
		return nil, ErrWebhookReplyInvalid
	}
	fields := make(map[string]json.RawMessage)
	if reply.Params != nil {
		data, err := json.Marshal(reply.Params)
		if err != nil {
			return nil, fmt.Errorf("%w: encode parameters", ErrWebhookReplyInvalid)
		}
		if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
			return nil, fmt.Errorf("%w: parameters must be an object", ErrWebhookReplyInvalid)
		}
	}
	if _, exists := fields["method"]; exists {
		return nil, fmt.Errorf("%w: parameters contain method", ErrWebhookReplyInvalid)
	}
	method, _ := json.Marshal(reply.Method)
	fields["method"] = method
	return json.Marshal(fields)
}

func validWebhookMethod(method string) bool {
	if method == "" || len(method) > 128 {
		return false
	}
	for index := 0; index < len(method); index++ {
		char := method[index]
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '_':
		default:
			return false
		}
	}
	return true
}

func ServeWebhook(
	ctx context.Context,
	address string,
	path string,
	handler http.Handler,
	wait func(),
) error {
	if handler == nil {
		return ErrWebhookHandlerRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if path == "" {
		path = "/telegram"
	}
	if !strings.HasPrefix(path, "/") || strings.ContainsAny(path, "{}?# \t\r\n") {
		return fmt.Errorf("hermes: invalid webhook path")
	}
	exactPath := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != path {
			http.NotFound(writer, request)
			return
		}
		handler.ServeHTTP(writer, request)
	})
	server := &http.Server{
		Addr: address, Handler: exactPath,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	result := make(chan error, 1)
	go func() { result <- server.ListenAndServe() }()

	select {
	case err := <-result:
		if wait != nil {
			wait()
		}
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownCtx)
		var closeErr error
		if shutdownErr != nil {
			closeErr = server.Close()
		}
		if wait != nil {
			wait()
		}
		return errors.Join(shutdownErr, closeErr)
	}
}
