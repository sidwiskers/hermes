# Competitor adapter contract

Cross-library results belong here only when they can be reproduced. Each
competitor should live in its own nested Go module so its dependencies never
enter Hermes' zero-dependency module.

Every adapter must benchmark the same operations:

1. Decode `../testdata/update.json` into the library's normal update type.
2. Register one exact `/start` command and dispatch the decoded update.
3. Register 1,000 exact commands and dispatch `/start`.
4. Apply ten pass-through middleware layers and dispatch `/start`.
5. Encode one `sendMessage` request and decode the fixed successful response.

Rules:

- pin the competitor version in its `go.mod`;
- record optional JSON/HTTP backends in the benchmark name;
- exclude network latency;
- run `GOMAXPROCS=1`, the same Go version, `-count=10`, and `-benchmem`;
- commit raw output, CPU information, and `go env` output;
- do not normalize away allocations or unsupported behavior.

Pinned adapters currently cover:

| Adapter | Version | Decode | Exact routing | 1,000 routes | Middleware | API call |
| --- | --- | --- | --- | --- | --- | --- |
| go-telegram-bot-api | v5.5.1 | yes | unsupported | unsupported | unsupported | yes |
| Telebot | v3.3.8 | yes | yes | yes | yes | yes |
| gotgbot | v2.0.0-rc.35 | yes | yes | yes | unsupported | yes |
| go-telegram/bot | v1.22.0 | yes | yes | yes | yes | yes |
| Telego (go-json, net/http) | v1.11.0 | yes | yes | yes | yes | yes |
| go-tg | v0.18.0 | yes | yes | yes | yes | yes |
| Telebot v4 beta | v4.0.0-beta.10 | yes | yes | yes | yes | yes |

“Unsupported” means the pinned library does not provide that framework layer;
the adapter does not fabricate one. Run every supported workload sequentially:

```bash
GOMAXPROCS=1 BENCH_COUNT=10 BENCH_TIME=1s ./scripts/bench-competitors.sh > competitor-results.txt
```

Compile every isolated adapter with:

```bash
./scripts/verify-competitors.sh
```

Measure equivalent stripped minimal programs with:

```bash
./scripts/bench-binary-size.sh
```
