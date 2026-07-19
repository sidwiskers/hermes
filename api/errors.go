package api

import (
	"errors"
	"fmt"
)

var (
	ErrClientRequired   = errors.New("hermes: Bot API client is required")
	ErrTokenRequired    = errors.New("hermes: bot token is required")
	ErrInvalidMethod    = errors.New("hermes: invalid Bot API method")
	ErrResponseTooLarge = errors.New("hermes: Bot API response exceeds configured limit")
	ErrResultMissing    = errors.New("hermes: Bot API success response is missing result")
)

// TransportError is a sanitized network or request-construction error.
// It deliberately never includes the token-bearing request URL.
type TransportError struct {
	Method    string
	Operation string
	Err       error
	token     string
}

func (e *TransportError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err == nil {
		return fmt.Sprintf("hermes: %s %s failed", e.Method, e.Operation)
	}
	return redactToken(fmt.Sprintf("hermes: %s %s failed: %v", e.Method, e.Operation, e.Err), e.token)
}

func (e *TransportError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// APIError is an error returned by Telegram.
type APIError struct {
	Code        int
	Description string
	Parameters  *ResponseParameters
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code == 0 {
		return "hermes: " + e.Description
	}
	return fmt.Sprintf("hermes: Telegram API error %d: %s", e.Code, e.Description)
}

func (e *APIError) RetryAfter() int {
	if e == nil || e.Parameters == nil {
		return 0
	}
	return e.Parameters.RetryAfter
}

func (e *APIError) MigrateToChatID() int64 {
	if e == nil || e.Parameters == nil {
		return 0
	}
	return e.Parameters.MigrateToChatID
}

// HTTPError is returned when the endpoint does not return a valid Telegram
// response envelope.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Body == "" {
		return fmt.Sprintf("hermes: Bot API HTTP error: %s", e.Status)
	}
	return fmt.Sprintf("hermes: Bot API HTTP error: %s: %s", e.Status, e.Body)
}
