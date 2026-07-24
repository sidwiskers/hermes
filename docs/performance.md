# Performance discipline

Hermes treats performance claims as testable engineering statements. The
standard-library-only implementation is the reference; every comparison keeps
competitor dependencies in isolated nested modules.

## Reproducible workloads

The benchmark laboratory measures:

- update decoding from one shared JSON fixture;
- exact command routing with 1 and 1,000 registered routes;
- callback-prefix routing with 1,000 registered prefixes;
- ten pass-through middleware layers;
- complete `sendMessage` request and response processing with no network;
- stripped size of an equivalent minimal program;
- Hermes-specific raw decoding, batch decoding, startup, webhook, and streamed
  1 MiB upload paths.

Run the cross-library suite sequentially:

```bash
GOMAXPROCS=1 BENCH_COUNT=10 BENCH_TIME=1s ./scripts/bench-competitors.sh
```

Build the stripped minimal programs with:

```bash
./scripts/bench-binary-size.sh
```

## Checked-in expanded baseline

These are medians of ten one-second samples recorded on 2026-07-19 with Go
1.26.5, `GOMAXPROCS=1`, `GOAMD64=v1`, and an Intel Xeon Platinum 8370C. Adapter
setup is excluded from every timer. Times are nanoseconds per operation and
must not be compared with measurements from another machine or Go version.

| Workload | Hermes | tgbotapi 5.5 | Telebot 3.3 | Telebot 4 beta | gotgbot rc.35 | go-telegram/bot 1.22 | Telego 1.11 | go-tg 0.18 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Decode update | 9,880 | 9,748 | 9,744 | 10,648 | 14,797 | 9,651 | **6,302** | 10,822 |
| Exact route, 1 | 113 | unsupported | 1,119 | 1,189 | 736 | **39** | 11,773 | 313 |
| Exact route, 1,000 | **124** | unsupported | 1,159 | 1,228 | 194,404 | 13,720 | 10,933,916 | 85,448 |
| Ten middleware | **133** | unsupported | 1,611 | 1,792 | unsupported | 497 | 12,590 | 1,072 |
| Mocked API call | **10,737** | 11,066 | 16,320 | 16,785 | 36,530 | 25,450 | 22,853 | 10,758 |

Median allocations and bytes from the same runs:

| Workload | Hermes | tgbotapi 5.5 | Telebot 3.3 | Telebot 4 beta | gotgbot rc.35 | go-telegram/bot 1.22 | Telego 1.11 | go-tg 0.18 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Decode update | 18 / 1,904 B | 20 / 1,600 B | 20 / 1,936 B | 20 / 2,672 B | 25 / 3,328 B | 19 / 1,984 B | **9** / 4,048 B | 20 / 1,952 B |
| Exact route, 1 | **0 / 0 B** | unsupported | 5 / 672 B | 5 / 736 B | 7 / 328 B | **0 / 0 B** | 18 / 4,896 B | 3 / 72 B |
| Exact route, 1,000 | **0 / 0 B** | unsupported | 5 / 672 B | 5 / 736 B | 1,006 / 32,296 B | **0 / 0 B** | 13,006 / 4,704,317 B | 2 / 16,408 B |
| Ten middleware | **0 / 0 B** | unsupported | 15 / 832 B | 15 / 896 B | unsupported | 10 / 160 B | 18 / 4,896 B | 13 / 312 B |
| Mocked API call | **39 / 4,064 B** | 51 / 4,304 B | 58 / 5,768 B | 58 / 6,504 B | 97 / 8,312 B | 88 / 5,888 B | 42 / 5,931 B | 51 / 4,672 B |

The stripped minimal Hermes program is 4,223,241 bytes. Measured competitor
programs range from 5,681,417 bytes (Telego) to 6,693,129 bytes (Telebot v4),
making this Hermes fixture about 26–37% smaller.

The results do not support a universal-fastest claim. Telego's optional
third-party JSON backend wins this decode fixture and `go-telegram/bot` wins the
single-route microcase. Hermes and go-tg are effectively tied in the mocked
API-call timing on this host (10,737 versus 10,758 ns median), while Hermes uses
fewer allocations and bytes. Hermes leads the measured 1,000-route,
ten-middleware, API-call-allocation, and binary-size cells. It is the only
adapter in this matrix combining effectively constant-time 1,000-route dispatch
with zero hot-path allocations.

The full unedited records are in
[`benchmarks/results/2026-07-19-go1.26.5-linux-amd64-expanded.txt`](../benchmarks/results/2026-07-19-go1.26.5-linux-amd64-expanded.txt)
and
[`benchmarks/results/2026-07-19-go1.26.5-binary-size-linux-amd64-expanded.tsv`](../benchmarks/results/2026-07-19-go1.26.5-binary-size-linux-amd64-expanded.tsv).
The original five-library baseline remains archived in the same directory for
historical comparison.

The runtime harness also has a checked-in
[`two-minute Go 1.26.5 overload-and-drain record`](../benchmarks/results/2026-07-19-go1.26.5-soak-2m.json).
It is useful for detecting queue, concurrency, leak, and shutdown regressions;
it is not presented as production latency evidence because request generation
and the server run in the same process.

The current systems pass adds a
[`30-second 1,000-route record`](../benchmarks/results/2026-07-24-go1.26.5-soak-30s.json):
4,080,713 requests, 976,764 accepted updates, controlled overload for every
rejected request, zero unexpected responses or handler failures, exact
64-update peak concurrency, full drain, and goroutines returned to the starting
count.

## Design consequences

Exact commands and callbacks use immutable map-backed route snapshots. Route
definitions are accumulated and compiled on first dispatch, avoiding quadratic
snapshot rebuilding during startup. Handler contexts are borrowed from a
`sync.Pool`; call `c.Clone()` before retaining a context after its handler, or
disable pooling with `hermes.WithContextPooling(false)`.

Callback-prefix routes use a length-indexed immutable map. Dispatch checks at
most one map entry for each registered prefix length, longest first. Because
Telegram callback data is bounded, lookup work is bounded by callback-data
length rather than total route count, while preserving filtered-route fallback
within each prefix. A paired ten-sample run on the same host reduced the
1,000-prefix router median from 2,563.5 ns to 22.59 ns (113.5x) without adding
an allocation. The complete samples and environment are
[`checked in`](../benchmarks/results/2026-07-24-go1.26.5-callback-prefix.txt).

Polling dispatches pointers to decoded batch elements instead of copying every
update into a separately escaping value. In a paired 100-update batch
benchmark, that reduced the median from 9,025 ns and 101 allocations to 286.3
ns and two allocations. The webhook fast path now shares the bounded byte
decoder with raw-preserving mode; paired samples reduced allocated memory from
9,880 B / 41 allocations to 8,120 B / 37 allocations. The complete runtime
samples are [`checked in`](../benchmarks/results/2026-07-24-go1.26.5-runtime-hot-paths.txt).

FSM rule reads use immutable atomic snapshots instead of merging slices under
an `RWMutex` on every trigger. Paired samples reduced the lookup median from
166.7 ns and 128 B to 40.79 ns and zero allocation while registration remains
race-safe.

The update decoder similarly pools only its outer envelope, removing one heap
escape without sharing returned message data. Raw preservation remains disabled
by default. Enabling it copies the complete input and adds one allocation; use
`hermes.WithRawUpdates(true)` only when diagnostics or raw forward compatibility
require it.

## Profiling

CPU and heap profiles require no extra dependency:

```bash
go run ./cmd/hermesbench -workload decode -n 3000000 -cpuprofile decode.cpu -memprofile decode.mem
go tool pprof -http=:8080 decode.cpu
```

The benchmark contract, pinned dependency versions, and unsupported framework
layers are documented in [`benchmarks/competitors`](../benchmarks/competitors/README.md).

## Nested attachment validation

Bot API 10.2 includes methods whose files are referenced inside nested JSON
objects. Hermes scans the already-marshaled JSON for `attach://` references
instead of unmarshaling it into a second generic tree. The direct and legacy
implementations remain benchmarked in `api` to prevent regression.
