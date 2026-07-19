# Feature map

Hermes keeps the protocol client, routing framework, and optional production
tools separate. Applications pay only for packages they import.

## Protocol and transport

| Capability | Support |
| --- | --- |
| Telegram Bot API | Schema-audited typed surface for the version recorded in `spec/bot-api.json`; exact release counts are published in [`schema-parity.md`](schema-parity.md) |
| Forward compatibility | Raw JSON and multipart calls; optional raw update preservation |
| Uploads | Constant-memory streamed single and multi-file multipart uploads |
| Downloads | Validated paths, bounded errors, streamed readers and writers |
| Errors | Typed Telegram, HTTP, and transport errors with token redaction and `errors.Is`/`errors.As` support |
| Test environment | First-class Telegram test-DC method and file endpoints |
| Dependencies | Go standard library only in the root module |

## Framework

| Capability | Support |
| --- | --- |
| Routing | Exact commands/callbacks, longest callback prefix, ordered filters, groups, and fallback handlers |
| Middleware | Global and group-local compiled chains; recovery, deadlines, and structured logging |
| Update sources | Bounded long polling, queued webhooks, synchronous retry-safe webhooks, and custom sources through `Bot.Handle` |
| Direct webhook replies | One typed or raw Bot API method encoded in the webhook response with a response-size bound |
| Helpers | Common sends, replies, edits, media, albums, callbacks, polls, reactions, moderation, and chat actions |
| Callback data | Typed string, integer, and JSON codecs with Telegram's 64-byte limit enforced |
| Testing | Network-free JSON/multipart recorder and externally compiled examples |

## Optional production packages

| Package | Capability |
| --- | --- |
| `session` | Generic typed sessions, pluggable stores, per-key serialization, rollback-on-error, TTL, capacity bounds, and explicit cleanup |
| `fsm` | Generic finite-state conversations with ordered guards, transactional actions, wildcard transitions, filters, and handlers |
| `dedupe` | Atomic update claiming, release on error/panic, TTL, capacity bounds, and a store interface for distributed coordination |
| `ratelimit` | Sharded per-user/chat token buckets, retry estimates, bounded identity cardinality, and explicit idle cleanup |
| `observe` | Panic-contained trace hooks and fixed-cardinality lock-free metrics for updates and Bot API calls |

The in-memory stores are process-local by design. Multi-instance deployments
implement the small `session.Store` and `dedupe.Store` interfaces using their
transactional database or cache. Hermes does not force a Redis, SQL, telemetry,
or logging dependency into every bot.

## Deliberate boundaries

Hermes does not automatically retry arbitrary outbound Bot API calls because a
replayed send, payment, edit, or upload can duplicate side effects. It also
does not hide goroutines in optional stores. Applications explicitly own
cleanup, persistence, retry, localization, and deployment policy; the library
provides typed extension points and deterministic primitives for them.
