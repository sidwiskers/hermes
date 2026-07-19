// Package dedupe provides atomic duplicate-update suppression middleware for
// Hermes.
//
// Telegram may deliver an update more than once, especially when a webhook
// response is lost or retried. Manager claims an update ID before handling it,
// releases failed claims for a later retry, and retains successful claims for a
// configurable TTL. Distributed applications can implement Store with their
// database's compare-and-set primitive.
package dedupe
