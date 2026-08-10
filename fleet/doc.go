// Package fleet hosts several Hermes bots inside one Go process.
//
// A Fleet keeps every bot's token, router, middleware, state, and update
// dispatcher independent while coordinating their update sources and shutdown.
// Bots created through Fleet.NewBot share one HTTP client. Webhook bots may
// additionally share one hardened HTTP server through exact, private paths.
//
// Fleet is optional and uses only the Go standard library. Ordinary Hermes
// applications continue to call Bot.Run without importing this package or
// paying any Fleet runtime cost.
package fleet
