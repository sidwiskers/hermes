// Package testkit provides Hermes Lab and deterministic in-memory Bot API
// transports for testing applications without Telegram or network access.
//
// Lab supplies virtual users and chats, synchronous update delivery, automatic
// responses for common Bot API methods, retained message state, failure
// injection, duplicate-update replay, and readable request expectations. The
// lower-level Recorder remains available when tests need complete control over
// every response.
package testkit
