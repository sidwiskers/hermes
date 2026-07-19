// Package observe provides dependency-free lifecycle hooks and aggregate
// metrics for Hermes handlers and Bot API calls.
//
// Hooks can bridge Hermes to OpenTelemetry, Prometheus, expvar, or an internal
// monitoring system without forcing any of those dependencies into the core
// module. Metrics intentionally uses fixed-cardinality atomic counters.
package observe
