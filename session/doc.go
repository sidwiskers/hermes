// Package session provides typed, concurrency-safe update sessions with
// pluggable storage.
//
// A Manager is installed as ordinary Hermes middleware. It serializes updates
// for the same session key, loads the value before the handler, and commits a
// changed value after a successful handler. Applications that need a remote or
// distributed store implement Store; Memory is a sharded in-process
// implementation with optional TTL and capacity bounds.
//
// The package is optional and uses only the Go standard library. Importing the
// root Hermes package does not link it into a bot.
package session
