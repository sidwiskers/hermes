package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxPooledBufferCapacity = 64 << 10

var transportBufferPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func acquireTransportBuffer() *bytes.Buffer {
	buffer := transportBufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	return buffer
}

func releaseTransportBuffer(buffer *bytes.Buffer) {
	if buffer == nil {
		return
	}
	if buffer.Cap() > maxPooledBufferCapacity {
		return
	}
	buffer.Reset()
	transportBufferPool.Put(buffer)
}

type apiEnvelope struct {
	OK          bool                `json:"ok"`
	Result      json.RawMessage     `json:"result,omitempty"`
	ErrorCode   int                 `json:"error_code,omitempty"`
	Description string              `json:"description,omitempty"`
	Parameters  *ResponseParameters `json:"parameters,omitempty"`
}

// Call invokes any Bot API method with a JSON request and decodes its result.
// A nil result discards a successful result.
func (b *Client) Call(ctx context.Context, method string, params any, result any) error {
	if b == nil {
		return ErrClientRequired
	}
	raw, err := b.callJSON(ctx, method, params)
	if err != nil {
		return err
	}
	return decodeResult(method, raw, result)
}

func (b *Client) callJSON(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if b == nil {
		return nil, ErrClientRequired
	}
	if b.token == "" {
		return nil, ErrTokenRequired
	}
	if !validMethod(method) {
		return nil, ErrInvalidMethod
	}

	var body io.Reader
	var payload *bytes.Buffer
	if params != nil {
		payload = acquireTransportBuffer()
		defer releaseTransportBuffer(payload)
		encoder := json.NewEncoder(payload)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(params); err != nil {
			return nil, fmt.Errorf("hermes: encode %s request: %w", method, err)
		}
		body = payload
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, b.methodURL(method), body)
	if err != nil {
		return nil, b.transportError(method, "create request", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", b.userAgent)

	return b.doResult(request, method, CallJSON)
}

// Call invokes any Bot API method and decodes its result as T.
func Call[T any](ctx context.Context, client *Client, method string, params any) (T, error) {
	var result T
	if client == nil {
		return result, ErrClientRequired
	}
	err := client.Call(ctx, method, params, &result)
	return result, err
}

// Upload describes one streamed multipart file. The caller owns Reader.
type Upload struct {
	Field  string
	Name   string
	Reader io.Reader
}

// InputFile is Telegram's multipart file abstraction. Upload is retained as
// the concise Hermes name; both identifiers describe the same streamed value.
type InputFile = Upload

// CallMultipart invokes any Bot API method with streamed multipart uploads.
// The caller retains ownership of every Upload.Reader.
func (b *Client) CallMultipart(
	ctx context.Context,
	method string,
	fields map[string]string,
	uploads []Upload,
	result any,
) error {
	if b == nil {
		return ErrClientRequired
	}
	if b.token == "" {
		return ErrTokenRequired
	}
	if !validMethod(method) {
		return ErrInvalidMethod
	}
	if err := validateMultipartInputs(fields, uploads); err != nil {
		return err
	}

	reader, writer := io.Pipe()
	form := multipart.NewWriter(writer)

	go func() {
		var writeErr error
		defer func() {
			if value := recover(); value != nil {
				writeErr = fmt.Errorf("hermes: multipart writer panicked: %v", value)
			}
			if writeErr == nil {
				writeErr = form.Close()
			}
			_ = writer.CloseWithError(writeErr)
		}()

		for key, value := range fields {
			if writeErr = form.WriteField(key, value); writeErr != nil {
				return
			}
		}
		for _, upload := range uploads {
			name := upload.Name
			if name == "" {
				name = "file"
			}

			header := make(textproto.MIMEHeader)
			header.Set(
				"Content-Disposition",
				`form-data; name="`+escapeQuotes(upload.Field)+`"; filename="`+escapeQuotes(filepath.Base(name))+`"`,
			)
			header.Set("Content-Type", "application/octet-stream")

			part, err := form.CreatePart(header)
			if err != nil {
				writeErr = err
				return
			}
			if _, err = io.Copy(part, upload.Reader); err != nil {
				writeErr = err
				return
			}
		}
	}()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, b.methodURL(method), reader)
	if err != nil {
		_ = reader.Close()
		return b.transportError(method, "create multipart request", err)
	}
	defer reader.Close()
	request.Header.Set("Content-Type", form.FormDataContentType())
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", b.userAgent)

	raw, err := b.doResult(request, method, CallMultipart)
	if err != nil {
		return err
	}
	return decodeResult(method, raw, result)
}

func (b *Client) doResult(request *http.Request, method string, kind CallKind) (raw json.RawMessage, resultErr error) {
	if b.observer != nil {
		event := CallEvent{Method: method, Kind: kind}
		started := time.Now()
		observedContext := startObserver(b.observer, request.Context(), event)
		request = request.WithContext(observedContext)
		defer func() {
			finishObserver(b.observer, observedContext, event, CallResult{
				Duration: time.Since(started),
				Err:      resultErr,
			})
		}()
	}

	response, err := b.client.Do(request)
	if err != nil {
		return nil, b.transportError(method, "request", err)
	}
	defer response.Body.Close()

	limit := b.responseLimit
	buffer := acquireTransportBuffer()
	defer releaseTransportBuffer(buffer)
	if _, err := buffer.ReadFrom(io.LimitReader(response.Body, limit+1)); err != nil {
		return nil, b.transportError(method, "read response", err)
	}
	data := buffer.Bytes()
	if int64(len(data)) > limit {
		return nil, ErrResponseTooLarge
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, &HTTPError{
			StatusCode: response.StatusCode,
			Status:     redactToken(response.Status, b.token),
			Body:       b.compactBody(data),
		}
	}

	if !envelope.OK {
		return nil, &APIError{
			Code:        envelope.ErrorCode,
			Description: redactToken(envelope.Description, b.token),
			Parameters:  envelope.Parameters,
		}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, &HTTPError{
			StatusCode: response.StatusCode,
			Status:     redactToken(response.Status, b.token),
			Body:       b.compactBody(data),
		}
	}
	if len(envelope.Result) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Result), []byte("null")) {
		return nil, fmt.Errorf("hermes: %s: %w", method, ErrResultMissing)
	}
	return envelope.Result, nil
}

func decodeResult(method string, raw json.RawMessage, result any) error {
	if result == nil || len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil
	}
	if err := json.Unmarshal(raw, result); err != nil {
		return fmt.Errorf("hermes: decode %s result: %w", method, err)
	}
	return nil
}

func (b *Client) methodURL(method string) string {
	return b.methodPrefix + method
}

func validMethod(method string) bool {
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

func (b *Client) compactBody(data []byte) string {
	const max = 512
	value := redactToken(strings.TrimSpace(string(data)), b.token)
	if len(value) > max {
		return value[:max] + "…"
	}
	return value
}

func redactToken(value, token string) string {
	if token == "" {
		return value
	}
	return strings.ReplaceAll(value, token, "<redacted>")
}

func validateMultipartInputs(fields map[string]string, uploads []Upload) error {
	for field := range fields {
		if !validMultipartHeaderValue(field) {
			return fmt.Errorf("hermes: invalid multipart field name")
		}
	}
	for _, upload := range uploads {
		if !validMultipartHeaderValue(upload.Field) || upload.Reader == nil {
			return fmt.Errorf("hermes: invalid multipart upload")
		}
		if upload.Name != "" && !validMultipartHeaderValue(filepath.Base(upload.Name)) {
			return fmt.Errorf("hermes: invalid multipart file name")
		}
	}
	return nil
}

func validMultipartHeaderValue(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 0x20 || value[index] == 0x7f {
			return false
		}
	}
	return true
}

func escapeQuotes(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

// MultipartJSON encodes a structured multipart field as JSON.
func MultipartJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MultipartInt formats an integer multipart field.
func MultipartInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

func (b *Client) transportError(method, operation string, err error) error {
	if err == nil {
		return nil
	}
	for {
		var urlErr *url.Error
		if !errors.As(err, &urlErr) || urlErr.Err == nil {
			break
		}
		err = urlErr.Err
	}
	return &TransportError{
		Method:    method,
		Operation: operation,
		Err:       err,
		token:     b.token,
	}
}
