# Reliability contract

Hermes treats cancellation, overload, malformed peers, and application panics
as ordinary operating conditions. The guarantees below are covered by tests
and the race detector.

## Update execution

- Polling and webhook handlers run through a bounded dispatcher.
- Polling applies backpressure; webhooks fail fast with `503 Service
  Unavailable` and `Retry-After: 1` when the dispatcher is full.
- `Wait` safely drains active work even when another goroutine is enqueueing.
- A panic from an asynchronously dispatched handler becomes a `PanicError`, is
  sent to the configured error handler, and does not terminate the process or
  stop later updates.
- A panic in the error handler itself is contained and logged.
- Direct `Bot.Handle` calls remain synchronous. Use `hermes.Recover()` when the
  caller wants panic-to-error conversion on that path.

Handler contexts are pooled by default. They are valid only until the handler
returns; `c.Clone()` creates a context value that can be retained.

## Polling

- Long-poll cancellation returns promptly and drains accepted handlers.
- Accepted polling handlers are detached from intake cancellation, matching
  queued webhook delivery; stopping intake does not abort work that was
  already admitted.
- Transient failures use bounded exponential backoff.
- Telegram `retry_after` values are honored without duration overflow.
- Stale or duplicate update IDs below the active offset are not dispatched
  again.
- If Telegram returns a non-empty batch entirely below the requested positive
  offset, polling treats it as Telegram's documented post-idle random
  `update_id` epoch and resumes instead of rejecting the new epoch forever.
- Invalid sources and dispatch functions return stable sentinel errors instead
  of panicking.

Hermes does not persist the polling offset. Applications that require seamless
process-level failover should persist their own idempotency or completion state.

## Webhooks

- Only `POST` is accepted.
- Secret-token comparison is constant-time when a secret is configured.
- Request bodies are bounded before decoding; empty, malformed, trailing, and
  oversized payloads are rejected.
- The serving path is exact rather than a wildcard pattern.
- Header, body-read, write, and idle timeouts bound slow or abandoned clients.
- Shutdown stops accepting requests, closes timed-out connections, and drains
  accepted update work.
- Queued dispatch receives a context detached from request cancellation, so a
  client disconnect after acceptance does not cancel the update handler.
- Synchronous reply mode shares the same global concurrency bound, returns 500
  when a handler fails or panics, and returns 503 with `Retry-After` on overload.
- Direct Bot API replies are encoded only after successful handling and have an
  independent response-size limit.

`WebhookHandler` acknowledges after queuing, before the application handler
finishes. Handler errors therefore cannot ask Telegram to redeliver that update.
Use `WebhookReplyHandler` or `ServeWebhookReplies` when the HTTP acknowledgement
must reflect handler success.

A webhook can still be delivered more than once if Telegram does not observe
the successful response. The optional `dedupe` middleware atomically claims
`update_id` values and releases failed or panicked claims. Multi-process bots
must back its `Store` interface with shared atomic storage; exactly-once domain
side effects still require application-level idempotency.

## Stateful middleware

- `session.Manager` serializes a complete handler transaction per session key,
  preventing in-process read-modify-write loss.
- Active session keys use independent pooled locks. Unrelated sessions never
  serialize because of lock-striping collisions, even while handlers are held
  open for remote I/O.
- Session mutations roll back on handler error by default. Store load/commit
  failures are wrapped and remain inspectable.
- `fsm.Machine` changes state only after its selected guard and action succeed;
  a later handler error participates in the session rollback policy.
- In-memory session, deduplication, and rate-limit stores are bounded when
  configured and create no hidden cleanup goroutines.
- Distributed safety is explicit: applications supply shared implementations of
  the small store interfaces when several bot processes consume one stream.

## Observability

`observe.Middleware` and `api.Observer` expose lifecycle hooks without exposing
tokens, parameters, request URLs, or file paths. Observer panics are contained.
`observe.Metrics` uses fixed-cardinality atomic counters—update IDs, users,
chats, commands, and API method names are not retained as metric labels.

## Bot API transport

- Bot tokens are removed from transport, proxy, response-body, and status errors.
- Response bodies have a configurable hard limit.
- Successful envelopes must include a non-null `result`; malformed successes
  never silently become zero Go values.
- Network causes remain available through `errors.Is` and `errors.As`.
- Multipart files stream through `io.Pipe`; file contents are not buffered in
  memory.
- Multipart field and file metadata reject control characters before a request
  starts, preventing MIME header injection.
- Panics from asynchronous upload readers become returned errors.

Hermes intentionally does not retry arbitrary API methods. Automatically
replaying a send, payment, edit, or upload can duplicate side effects. Callers
may safely retry operations according to their own idempotency policy and can
inspect `APIError.RetryAfter()` for Telegram rate limits.

An upload reader that blocks forever inside its own `Read` method cannot be
forcibly interrupted by Go's `io.Reader` interface. Long-running custom readers
should observe their own cancellation signal.

## Verification

The complete local gate is:

```bash
./scripts/verify.sh
```

It checks generated-source drift, Bot API schema parity, formatting, vet,
shuffled tests, the race detector, and package coverage floors. CI also runs
short fuzz passes for update decoding, transport envelopes, method names,
multipart metadata, command parsing, and callback codecs.
