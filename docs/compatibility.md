# Compatibility and release policy

Hermes 1.0.0 is the first stable release. Its exported API follows semantic
versioning; incompatible changes require a new major version and every
deprecation must identify a direct migration path.

## Bot API support

- Each Hermes release targets the Bot API version recorded in the checked-in
  `spec/bot-api.json` manifest.
- Hermes 1.0.0 targets Bot API 10.2. Its release inventory covers all 185
  methods, 937 parameters, 362 objects, 1,838 object fields, 26 unions, and 187
  variants in that version.
- Source-audit tests require zero missing methods, parameters, objects, fields,
  union roots, or union variants. They also verify requiredness, Go wire types,
  JSON optionality, and optional-object nilability.
- `Client.Call` and `Client.CallMultipart` remain permanent escape hatches for new methods and parameters that arrive before a Hermes release.
- Unknown update fields can be preserved with `WithRawUpdates(true)` or `DecodeUpdate(..., true)`.

The current zero-gap static schema report and remaining behavioral conformance
work are published in [`schema-parity.md`](schema-parity.md). After Telegram
publishes a Bot API release, a Hermes typed update requires a regenerated
manifest, models, aliases, and reviewed source audit. Until then, the raw call
layers provide forward-compatible access without claiming typed support that
has not yet been audited.

## Go compatibility

The module declares Go 1.25.0 as its language and consumer compatibility floor.
CI compiles and tests with the latest Go 1.25 patch under `GOTOOLCHAIN=local`,
and independently runs the complete gates with the current stable Go release.
The `toolchain go1.26.5` line is a contributor recommendation; Go does not use a
dependency module's toolchain recommendation to select an application's
toolchain.

Applications should use the latest patch available in their chosen Go line.
The application toolchain supplies the standard library, so an old patch can
retain vulnerabilities even when Hermes itself is current. Stable-release
evidence must be produced by Go 1.26.5 or any newer stable release, including
future Go lines such as 1.27 and 1.28. Prerelease and development toolchains do
not qualify as release evidence.

## Semantic versioning

- Patch releases fix behavior without intentionally breaking public code.
- Minor releases may add methods, schema fields, filters, helpers, or optional packages.
- Deprecations must name the replacement and remain for at least one minor release once the project reaches 1.0.
- Major releases are reserved for unavoidable public API breaks.

## Migration from the development `tg` identity

The rename is intentionally completed before 1.0:

- change imports from `github.com/sidwiskers/tg` to `github.com/sidwiskers/hermes`;
- change the root qualifier from `tg` to `hermes` unless an explicit import alias is preferred;
- change `github.com/sidwiskers/tg/tgtest` to `github.com/sidwiskers/hermes/testkit`;
- change profiling invocations from `go run ./cmd/tgbench` to `go run ./cmd/hermesbench`.

The previous development module path is not retained as a compatibility shim. Publishing both identities would split module versions, documentation, and vulnerability metadata before the first stable release.

## Stable-release gates

Hermes 1.0.0 satisfied all of the following stable-release gates:

- [x] Zero missing methods, parameters, object types, object fields, union roots,
  and union variants in the automated Bot API schema audit.
- [x] Reproducible competitor benchmarks with pinned versions and checked-in raw results.
- [x] Complete end-to-end production webhook delivery, authentication,
  synchronous reply, handler-error retry, panic recovery, and duplicate
  suppression evidence. Telegram accepted the bounded live rate probe without
  returning 429; deterministic transport tests cover the exact rate-limit
  envelope and `retry_after` behavior. Credentials are intentionally not stored. See
  [`conformance.md`](conformance.md).
- [x] Runnable examples covering middleware, webhooks, uploads, inline mode, payments, and graceful shutdown.
- [x] Deterministic exported-API manifest reviewed by CI.
- [x] Standard-library-only dependency inventory, credential scan, and official
  Go vulnerability gate.
- [x] Typed sessions, finite-state conversations, deduplication, rate limiting,
  lifecycle observation, and bounded synchronous webhook acknowledgement.
- [x] At least one credentialed production period with performance and failure
  telemetry reviewed, complemented by the checked-in 12.7-million-request
  overload and drain stress record.
- [x] A final exported-API review and documented migration from the development
  `tg` identity.
