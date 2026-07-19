# Design

## Package architecture

The project is one versioned Go module with strict package boundaries:

### `types`

Telegram schemas, keyboards, and update helpers. It has no dependency on transport or framework code.

### `api`

The standalone low-level client. It owns JSON and multipart requests, typed Bot API methods, file transfers, Telegram errors, and raw calls. Developers can import this package directly without adopting the framework.

### Root `hermes`

A deliberately small facade. It combines `api.Client` with routing and runtime behavior, then re-exports common request and Telegram types so normal syntax remains `hermes.SendMessageParams`, `hermes.Context`, and `hermes.Update`.

### `framework`

Context helpers, immutable router snapshots, filters, groups, and middleware.
The root package aliases this package so ordinary applications keep the compact
`hermes.Context` syntax, while generated documentation has a public canonical
home and advanced integrations can construct an independent router.

### `internal/runtime`

Bounded dispatch, polling loops, retry policy, webhook parsing, HTTP server lifecycle, and graceful draining.

### `testkit`

An optional public package for deterministic API request recording without Telegram or network access.

This dependency direction is intentional:

```text
types <- api
  ^       ^
  |       |
framework            internal/runtime
          \           /
           root hermes
                ^
              testkit
```

No lower layer imports the root facade, so there are no package cycles and each layer can be tested independently. Runtime lifecycle remains internal; the framework package is public because its types are part of the root API.

## Routing

Route registration uses copy-on-write snapshots:

- startup writes are accumulated and compiled once on first dispatch;
- post-start writes are serialized and rebuild a complete immutable route table;
- reads use one atomic pointer load;
- exact commands and callbacks use map lookup;
- callback prefixes are pre-sorted by specificity;
- filtered routes preserve registration order;
- nested groups are flattened into route filters and middleware during registration;
- middleware chains are compiled during registration, not dispatch.

This favors the real workload: routing millions of updates after a small number of startup registrations.

## Concurrency

Update execution is bounded by a semaphore. The library never starts an unbounded number of handler goroutines.

- polling waits for capacity and therefore does not drop an update;
- webhooks return `503` with `Retry-After` when saturated, allowing Telegram to retry;
- graceful shutdown stops intake before waiting for active handlers;
- route registration and dispatch are race-safe;
- group configuration uses independent synchronization.

## Memory and I/O

- JSON requests and small response bodies reuse bounded pooled buffers.
- Single-file and album multipart requests use `io.Pipe` and stream readers.
- Multipart attachment planning validates missing, duplicate, and unreferenced fields before starting the request. It scans the already-marshaled JSON directly instead of building a second generic object tree.
- Downloads stream directly to callers.
- Telegram response bodies are capped.
- Update objects are not copied after dispatch; original JSON is copied only when raw preservation is explicitly enabled.
- Handler contexts are borrowed from a pool on the default hot path, producing zero routing allocations; `Context.Clone` provides a safe retained copy.
- No reflection is used in routing.

## Errors

Transport and API errors are distinct:

- `APIError` exposes Telegram error codes, flood waits, and migration IDs;
- `HTTPError` covers malformed non-Telegram responses;
- `TransportError` preserves cancellation while stripping token-bearing URLs;
- handler errors flow through one configurable error hook;
- `Recover` and `RecoverWith` convert panics into `PanicError` values with stacks.

## Compatibility

Unknown update fields can be retained in `Update.Raw` through `WithRawUpdates(true)` or `DecodeUpdate(..., true)`. The default fast mode avoids copying the complete payload. Less common nested objects use raw fields where fully typing them would add weight without immediate value. The raw method layer allows new methods and parameters before the typed surface expands.

The core follows semantic versioning. Public API additions are preferred over breaking replacements.
