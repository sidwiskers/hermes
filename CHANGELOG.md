# Changelog

## 1.0.0 - 2026-07-19

### Hermes identity and release polish

- Renamed the canonical module and root package to `github.com/sidwiskers/hermes` and `hermes`.
- Renamed the network-free recorder package to `testkit` and the profiling command to `hermesbench`.
- Standardized user-facing error prefixes and the default user agent on the Hermes name.
- Documented the pre-1.0 compatibility policy and explicit release gates.
- Published the canonical `framework` package so root aliases link to complete
  `Context`, `Router`, `Group`, filter, and middleware documentation.
- Added externally compiled package examples plus complete polling, webhook,
  upload, inline-mode, payment, and standalone-client programs.
- Standardized nil raw-call failures on `ErrClientRequired`.
- Added macOS and Windows CI, cross-architecture release compilation, and a
  single automated release-gate command.
- Added first-class Telegram test-environment endpoints and an opt-in live
  conformance suite for identity, message lifecycle, streamed uploads, polling
  cancellation, and webhook configuration.
- Corrected test-environment file downloads and added a live byte-for-byte
  upload, `getFile`, and streamed-download conformance check.
- Added bounded synchronous webhook acknowledgement, direct Bot API replies,
  shared global concurrency limits, response-size limits, and retry-preserving
  handler error semantics.
- Added optional typed `session`, `fsm`, `dedupe`, `ratelimit`, and `observe`
  packages without adding a dependency to the core module.
- Added an owned-lifecycle soak harness with JSON runtime, latency, overload,
  allocation, heap, goroutine, and drain evidence.
- Added a deterministic exported-API manifest plus credential, dependency, and
  official Go vulnerability release gates.
- Raised the minimum toolchain to Go 1.26.5 after the vulnerability gate found
  reachable standard-library flaws in the original Go 1.26.0 release.
- Expanded current competitor coverage to Telego, go-tg, and Telebot v4 beta,
  and excluded adapter setup from every timed benchmark.
- Recorded credentialed production polling, streamed file round trips, a
  bounded live period, public HTTPS webhook delivery, synchronous replies,
  Telegram redelivery after handler error and contained panic, secret/path/body
  rejection, and duplicate suppression.
- Added a build-tagged, dependency-free live webhook probe plus a checksum-
  verified Cloudflare Quick Tunnel runner for repeatable release conformance.

### Bot API 10.2 completion

- Completed typed entry points for all 185 official Bot API 10.2 methods while retaining raw JSON and multipart escape hatches.
- Added a deterministic official-schema inventory and source audit covering 937
  parameters, 362 objects, 1,838 object fields, 26 unions, and 187 variants.
- Added a strict zero-gap source gate; method and parameter gaps are fixed at
  zero and any schema regression fails CI.
- Generated every missing Bot API object, field, union root, and concrete union
  variant, with aliases in the low-level client and root facade.
- Extended schema gates to requiredness, Go wire types, JSON optionality, and
  optional-object nilability; every checked category is at zero gaps.
- Replaced permissive inline-result and Passport-error structs with sealed
  concrete unions that inject discriminators automatically. Raw compatibility
  forms remain available as `InlineQueryResultRaw` and
  `PassportElementErrorRaw`.
- Renamed menu-button discriminator constants to `MenuButtonTypeCommands`,
  `MenuButtonTypeWebApp`, and `MenuButtonTypeDefault`, and renamed the
  force-reply constructor to `NewForceReply` so the official concrete type can
  use the `ForceReply` name.
- Added Rich Messages with all 21 input block variants, strict format validation, draft support, and nested streamed uploads.
- Added invoices, Stars, paid media, gifts, forum topics, business accounts, managed bots, stories, checklists, games, Passport errors, suggested posts, and full sticker-set management.
- Added typed inline, Web App, guest, and prepared-message results, profile/admin settings, chat boosts, business rights, and bulk forwarding/copying.
- Added live-photo sending and media editing with attachment/upload parity validation.
- Replaced multipart attachment inspection through a decoded `any` tree with a single-pass scan of marshaled JSON and changed method validation to byte-wise ASCII checks.
- Added contract, validation, decode, upload, and micro-benchmark coverage for the new surfaces.

### Round 5

- Made raw update preservation opt-in for polling, webhooks, and standalone decoders, removing the default full-payload copy.
- Added allocation-free handler dispatch through pooled borrowed contexts, plus `Context.Clone` and a pooling opt-out.
- Batched startup route registration automatically and compile the immutable snapshot once on first dispatch, eliminating quadratic startup rebuilds.
- Added bounded transport-buffer pooling and precomputed method URL prefixes, reducing JSON call latency and allocations.
- Added direct framework/runtime tests, malformed-input coverage, queue/backpressure tests, and raw-mode tests.
- Added fuzz targets for update decoding, command parsing, callback codecs, API envelopes, and method validation.
- Added a reproducible benchmark package, profiling command, benchmark scripts, coverage gates, fuzz smoke tests, and CI performance smoke tests.

### Round 4

- Added typed photo, live-photo, video, audio, and document media groups with 2–10 item validation and streamed multi-file attachments.
- Added current poll schemas, closed typed poll-media unions, streamed poll-media uploads, poll creation/stopping, optional-boolean helpers, and all supported dice.
- Added typed message reactions, reaction removal, and incoming reaction-count decoding.
- Added member banning, restriction, promotion, permissions, administrator titles, member tags, and sender-chat moderation.
- Added administrator/member lookup and join-request actions.
- Added normal and subscription invite-link creation, editing, export, and revocation.
- Added chat photo, title, description, pin, and sticker-set management.
- Added context helpers for reactions, polls, dice, moderation, join requests, and pins.
- Added attachment/upload parity validation plus multipart, schema, administration, and context-action contract tests.

### Structure

- Restructured the flat root implementation into `api`, `types`, `framework`, and `internal/runtime` packages.
- Preserved a small root-facade syntax through type re-exports and the embedded low-level client.
- Made `api.New` available for applications that want the library without the framework.
- Isolated bounded dispatch, polling, and webhook lifecycle code from public declarations.


### Round 2

- Expanded incoming update, message, media, query, payment, community, and subscription types.
- Added ordered filter routes and composable filter helpers.
- Added nested route groups with group-local middleware.
- Added typed callback codecs and 64-byte validation.
- Added inline, reply, remove, and force-reply keyboard types and helpers.
- Added all 12 Bot API 10.2 ephemeral-capable send methods.
- Added streamed uploads for photo, animation, audio, document, sticker, video, video note, and voice.
- Added automatic context media, edit, caption, keyboard, delete, callback-answer, and chat-action helpers.
- Added forwarding, copying, bulk deletion, command scopes, command retrieval/deletion, and common chat methods.
- Added file metadata, safe streamed downloads, webhook status, and certificate uploads.
- Added structured logging and configurable panic reporting middleware.
- Added the `testkit` in-memory request recorder.
- Split large API/type implementations into focused files.

### Foundation

- Standard-library-only Telegram Bot API transport.
- Generic raw JSON and streamed multipart calls.
- Typed Telegram API and HTTP errors.
- Lock-free command and callback routing snapshots.
- Middleware, panic recovery, and handler deadlines.
- Bounded long-polling and webhook dispatch.
- Graceful webhook shutdown and secret-token verification.
- Typed `sendMessage`, `sendPhoto`, callback, command, and webhook methods.
- Bot API 10.2 ephemeral commands, sends, replies, edits, and deletion.
- Unit, race, backpressure, upload, and routing benchmark coverage.
