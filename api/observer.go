package api

import (
	"context"
	"time"
)

// CallKind identifies the Bot API transport form observed by Observer.
type CallKind string

const (
	// CallJSON is an ordinary JSON Bot API request.
	CallJSON CallKind = "json"
	// CallMultipart is a streamed multipart Bot API request.
	CallMultipart CallKind = "multipart"
	// CallDownload is a streamed Telegram file download.
	CallDownload CallKind = "download"
)

// CallEvent contains non-secret metadata for one outbound operation. It never
// includes the bot token, request URL, parameters, or file path.
type CallEvent struct {
	Method string
	Kind   CallKind
}

// CallResult describes a completed outbound operation.
type CallResult struct {
	Duration time.Duration
	Err      error
}

// Observer receives outbound Bot API lifecycle events. StartCall may return a
// derived context for trace propagation. Hermes contains observer panics and
// continues the operation with the last valid context.
type Observer interface {
	StartCall(context.Context, CallEvent) context.Context
	FinishCall(context.Context, CallEvent, CallResult)
}

func startObserver(observer Observer, ctx context.Context, event CallEvent) (result context.Context) {
	result = ctx
	if observer == nil {
		return result
	}
	defer func() { _ = recover() }()
	if observed := observer.StartCall(ctx, event); observed != nil {
		result = observed
	}
	return result
}

func finishObserver(observer Observer, ctx context.Context, event CallEvent, result CallResult) {
	if observer == nil {
		return
	}
	defer func() { _ = recover() }()
	observer.FinishCall(ctx, event, result)
}
