# Roadmap

The foundation is designed to expand, not be replaced.

## Completed structured foundation

- One-module, multi-package architecture with a compact root facade and a
  documented public framework package.
- Standalone `api` client and dependency-free `types` schemas.
- Internal framework/runtime boundaries that prevent implementation leakage.
- Standard-library-only JSON and streamed multipart transport.
- Typed API, HTTP, and sanitized transport errors.
- Bounded polling and webhook dispatch with graceful shutdown.
- Lock-free command, callback, prefix, filter, and group route snapshots.
- Global and group middleware compiled during registration.
- Broad update decoding with raw forward compatibility.
- Typed callback-data codecs with the 64-byte limit enforced.
- Inline/reply keyboards and common context actions.
- All 12 Bot API 10.2 ephemeral-capable send methods.
- Complete ephemeral edit/delete lifecycle.
- Primary media uploads, file metadata, and streamed downloads.
- Deterministic `testkit` request recorder.
- Typed photo/live-photo/video/audio/document albums with streamed multi-file attachments and attachment parity checks.
- Polls, dice, reactions, moderation, permissions, members, invite links, and common chat administration.
- Typed entry points for all 185 methods in Bot API 10.2.
- Rich Messages and all 21 input block variants with nested streamed uploads.
- Payments, Stars, paid media, gifts, forum topics, inline mode, Web Apps, business accounts, managed bots, stories, checklists, games, Passport, and sticker sets.


## Completed performance and proof round

- Opt-in raw update preservation for polling, webhooks, and standalone decoding.
- Allocation-free exact routing through borrowed pooled contexts.
- Bounded transport buffer pooling and precomputed API method URL prefixes.
- Direct framework and runtime tests instead of relying only on facade tests.
- Fuzz targets for updates, command parsing, callback codecs, envelopes, and method validation.
- Reproducible decode, routing, middleware, API, and multipart benchmarks.
- CPU/heap profiling command and CI benchmark smoke tests.
- Race, coverage, fuzz-smoke, formatting, and vet regression gates.
- Deterministic Bot API 10.2 inventory and source audit covering methods,
  parameters, objects, fields, unions, and variants.
- Zero-gap generated object, field, union-root, and union-variant declarations,
  re-exported through `types`, `api`, and the root facade.
- Requiredness, Go wire-type, JSON optionality, and nested-object nilability
  parity gates.
- Sealed inline-result and Passport-error request unions with automatic
  discriminator encoding and raw forward-compatibility forms.
- Pinned, isolated adapters for four established Go Telegram libraries.
- Sequential decode, routing, middleware, and mocked API-call comparisons with
  checked-in raw results and stripped binary-size measurements.
- Pooled update envelopes that remove one allocation without sharing returned
  payload ownership.
- Panic containment at asynchronous handler and error-reporting boundaries.
- Duplicate-safe polling offsets, overflow-safe backoff, exact-path webhooks,
  forced shutdown fallback, and concurrency-safe dispatcher draining.
- Malformed-success rejection, token-redacted remote failures, bounded response
  bodies, multipart metadata validation, and upload-reader panic containment.
- Bounded synchronous webhooks whose handler result controls acknowledgement,
  with optional direct Bot API replies.
- Generic typed sessions with per-key serialization, rollback-on-error, TTL,
  capacity bounds, pluggable storage, and explicit cleanup.
- Generic finite-state conversations with guarded transactional transitions.
- Atomic duplicate suppression, bounded token-bucket rate limiting, trace
  lifecycle hooks, and fixed-cardinality lock-free metrics.
- Deterministic exported-API manifests, credential scanning, dependency
  inventory, and official Go vulnerability analysis in release gates.

## Next compatibility expansion

- Communities and subscriptions beyond incoming update decoding.
- Extend parity checks to uploads and request/response wire fixtures.
- Bot API release-diff automation for future Telegram changes.
- Profile and generate a standard-library decode fast path that preserves full
  schema parity while closing the remaining fixture latency gap.

## Optional ecosystem expansion

- Reusable menu/layout and localization helpers where they can remain typed and
  dependency-free.
- Maintained Redis and SQL adapters in separate modules, outside the
  zero-dependency root module.
- OpenTelemetry and Prometheus adapters built on the dependency-free observer
  contracts.

## Performance discipline

Every expansion must preserve:

- standard-library-only root core;
- bounded concurrency;
- no reflection in the dispatch path;
- no full-file multipart buffering;
- benchmark coverage for hot paths;
- race-test coverage for concurrent registration and dispatch.
