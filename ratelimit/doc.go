// Package ratelimit provides bounded token-bucket middleware for Hermes.
//
// Limiters are safe for concurrent dispatch, create no background goroutines,
// and can be scoped per user, chat, or chat/user pair. Idle buckets are removed
// explicitly with Sweep, keeping lifecycle and scheduling under application
// control.
package ratelimit
