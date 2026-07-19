# Live conformance evidence

Hermes records credentialed Telegram results separately from deterministic
unit, race, fuzz, and schema tests. Bot tokens, user identifiers, chat
identifiers, file identifiers, and webhook secrets are never retained.

## Telegram test environment — July 19, 2026

The suite ran from commit `425aef0` plus the test-file endpoint correction,
using Go 1.26.0 and a disposable bot created inside Telegram's separate test
DC.

Passed:

- bot authentication and identity through `getMe`;
- `sendMessage`, `editMessageText`, and `deleteMessage` lifecycle;
- streamed multipart `sendDocument` upload;
- `getFile` followed by an exact byte-for-byte streamed download;
- human-originated text update delivery through long polling and framework
  routing;
- polling request cancellation and handler draining.

The file round trip exposed a previously untested distinction in Telegram's
test environment: file downloads require the test segment in the file URL as
well as method calls. `WithTestEnvironment` now selects both prefixes once at
construction, retaining the branch-free request path.

## Telegram production environment — July 19, 2026

The Go 1.26.5 release lineage derived from commit `f8e2370` was exercised
with a dedicated disposable production bot and private chat. Credentials and
Telegram identifiers were discarded rather than written to the repository.

Production polling and transport passed:

- `getMe` identity with an expected-username guard;
- send, edit, delete, streamed upload, `getFile`, and byte-exact download;
- a human-originated text update through long polling and framework routing;
- polling cancellation and handler draining;
- a one-minute credentialed polling soak with zero non-cancellation errors,
  zero failed or panicked updates, zero work left in flight, and heap and
  goroutine snapshots recorded by the checked-in harness.

Production webhook delivery passed through a temporary Cloudflare Quick
Tunnel terminating at the checked-in Hermes probe:

- Telegram accepted `setWebhook` with a secret token and delivered a real
  command to the public HTTPS callback;
- Hermes returned a synchronous `sendMessage` response that Telegram executed;
- a deliberate handler error returned HTTP 500, released the deduplication
  claim, and succeeded when Telegram redelivered the same update;
- a deliberate panic was contained and reported, returned HTTP 500, released
  the claim, and succeeded on redelivery;
- missing and incorrect secrets returned 401, an unsupported method returned
  405, an incorrect path returned 404, and malformed JSON returned 400;
- a valid synthetic webhook returned the expected direct-response JSON with
  HTTP 200, while the same update ID repeated immediately returned an empty
  HTTP 200 after duplicate suppression.

The bounded live rate probe sent 16 simultaneous silent messages. Telegram
accepted the short burst, and `deleteMessages` removed all 16 in one call; no
live 429 was produced. Hermes's exact 429 envelope, `retry_after` exposure, and
polling delay behavior remain covered by deterministic transport and runtime
tests. The probe was not escalated into abusive traffic merely to force a
service-side limit.

The credentialed production period complements the checked-in two-minute
synthetic stress record: 12,736,273 requests, exact configured concurrency,
controlled overload, zero unexpected statuses, complete drain, and goroutines
returned to baseline. A long-horizon canary remains an operational practice,
not a substitute for these repeatable release gates.

The opt-in source is in `integration/live_test.go`. See `releasing.md` for the
environment variables and release procedure.
