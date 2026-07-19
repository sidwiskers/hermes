# Hermes performance laboratory

This package contains implementation-neutral workloads for measuring the parts
of a Telegram framework that materially affect real bots:

- update JSON decoding, with and without raw preservation;
- exact route dispatch at 1 and 1,000 registered routes;
- ten middleware layers;
- complete low-level JSON request/response processing;
- streamed 1 MiB multipart uploads.

Run a stable local sample:

```bash
./scripts/bench.sh
```

For machine-readable output:

```bash
go test -run '^$' -bench . -benchmem -count 10 ./benchmarks > bench.txt
```

Comparisons with other libraries must use the same fixture, Go version,
`GOMAXPROCS`, CPU governor, and benchmark count. Results from different machines
must not be compared directly. Pinned competitor adapters live under
`competitors`. Use `scripts/bench-competitors.sh` for sequential comparisons;
do not compare individual runs launched concurrently.

The current checked-in baseline and its limitations are summarized in
[`docs/performance.md`](../docs/performance.md). Raw output belongs in `results`;
never replace it with only a hand-curated table.
