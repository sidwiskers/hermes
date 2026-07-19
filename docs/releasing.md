# Releasing Hermes

Hermes releases are evidence-driven. Passing unit tests is necessary, but it
does not prove live Telegram behavior or production reliability by itself.

## Automated release gate

Run from a clean worktree with stable Go 1.26.5 or newer:

```bash
./scripts/release-check.sh
```

The command verifies:

- zero static Bot API schema gaps and deterministic generated sources;
- deterministic exported-API drift detection and credential-pattern scanning;
- formatting, vet, shuffled tests, race tests, and package coverage floors;
- official Go vulnerability analysis and a complete runtime dependency inventory;
- pinned competitor adapters and the routing benchmark smoke test;
- every runnable example;
- compilation of the opt-in live Telegram conformance suite;
- builds for Linux arm64, Windows amd64, and macOS amd64/arm64;
- a clean worktree before and after validation.

`RELEASE_ALLOW_DIRTY=1` exists only for developing the release checks. Its
output is not release evidence.

## Live Telegram conformance

Use a dedicated bot and chat created inside Telegram's separate test DC. The
suite is excluded from ordinary tests and creates then deletes messages:

```bash
HERMES_TEST_BOT_TOKEN='...' \
HERMES_TEST_CHAT_ID='...' \
go test -tags=integration -count=1 ./integration
```

This covers identity, send/edit/delete, streamed upload, and polling
cancellation through Telegram's `/test/` Bot API method endpoint. It also
round-trips an uploaded file byte-for-byte through `getFile` and the test file
endpoint. Results and outstanding behavioral probes are recorded in
[`conformance.md`](conformance.md).

For a dedicated disposable bot created in Telegram's production DC, opt into
the standard endpoint explicitly and pin the expected username as a wrong-bot
guard:

```bash
HERMES_TEST_PRODUCTION=true \
HERMES_TEST_BOT_USERNAME='example_bot' \
HERMES_TEST_BOT_TOKEN='...' \
HERMES_TEST_CHAT_ID='...' \
go test -tags=integration -count=1 ./integration
```

Never run the suite against a bot or chat that contains valuable state.

An inbound polling probe is enabled by sending a unique text message to the
test bot, then supplying that exact value as `HERMES_TEST_EXPECT_TEXT`.
Webhook configuration is additionally enabled when both of these are supplied:

```bash
HERMES_TEST_WEBHOOK_URL='https://example.com/telegram' \
HERMES_TEST_WEBHOOK_SECRET='long-random-secret' \
HERMES_TEST_BOT_TOKEN='...' \
HERMES_TEST_CHAT_ID='...' \
go test -tags=integration -count=1 ./integration
```

The suite restores message and webhook state on failure where Telegram permits
it. Never use a production bot token or production chat for this gate. Archive
the command output with the release evidence; compilation alone is not a live
conformance result.

`HERMES_TEST_KEEP_WEBHOOK=true` leaves a successfully verified webhook active
for the delivery probes. It is intentionally opt-in; delete the webhook before
stopping the public receiver or ending the release session.

`HERMES_TEST_FLOOD_WAIT=true` enables a bounded 64-request production probe
using ephemeral `sendChatAction` calls. It creates no messages and requires an
observed Telegram 429 response with a positive `retry_after`.

## Manual stable-release gates

Before a 1.0 tag, all of the following require recorded evidence:

- refresh the official Telegram schema and review the generated diff;
- exercise polling, webhook authentication, uploads, retry-after handling, and
  graceful shutdown with a dedicated live test bot;
- run a production soak and review latency, allocation, error, retry, queue
  saturation, and shutdown telemetry;
- review every exported name and the migration notes from the last pre-1.0 tag;
- review token redaction, dependency inventory, license, security policy, and
  changelog;
- reproduce and archive the competitor benchmark and stripped binary results.

## Synthetic soak harness

Run the checked-in queued-webhook/runtime harness before and during a candidate
release review:

```bash
go run ./cmd/hermessoak -duration 30m -concurrency 128 -workers 64
```

It emits a JSON record containing accepted and controlled-overload counts,
unexpected statuses, latency buckets, update metrics, allocations, heap,
goroutine counts, peak concurrency, and drain state. The build-tagged live soak
can be enabled with `HERMES_TEST_SOAK_DURATION` in the Telegram test suite.
Synthetic success is not a substitute for a sustained deployment: archive the
report together with host limits, traffic shape, downstream latency, and live
error/retry telemetry.

The checked-in Go 1.26.5 harness record processed 12,736,273 requests over two
minutes with injected handler delay: 1,865,724 were accepted, 10,870,549 were
rejected through controlled overload, no unexpected status or handler failure
occurred, concurrency peaked exactly at the configured 64, all work drained,
and goroutines returned to their starting count. See the
[`raw JSON report`](../benchmarks/results/2026-07-19-go1.26.5-soak-2m.json).
This is high-load regression evidence for overload and drain invariants. The
credentialed production-period result is recorded in `conformance.md`; neither
bounded release run replaces a long-horizon deployment canary.

Do not describe a release as universally fastest from a single fixture. Publish
the workload, hardware, Go version, raw results, and any cells another library
wins, as required by [`performance.md`](performance.md).

## Tagging

After automated and manual gates pass:

1. Replace the `Unreleased` changelog heading with the version and date.
2. Confirm `go.mod`, the schema version, examples, and README agree.
3. Run `./scripts/release-check.sh` once more from the exact release commit.
4. Create an immutable semantic-version tag from that exact commit and publish
   release notes linking the archived evidence.
5. Prefer a signed annotated tag when a maintainer signing identity is
   configured. Never move or replace a tag after it is public.
