#!/usr/bin/env bash
set -euo pipefail

export GOMAXPROCS="${GOMAXPROCS:-1}"
count="${BENCH_COUNT:-5}"
time="${BENCH_TIME:-1s}"

patterns=(
  '^BenchmarkDecodeUpdate'
  '^BenchmarkWebhookDecode$'
  '^BenchmarkRouter'
  '^BenchmarkAPICall$'
  '^BenchmarkMultipartUpload1MiB$'
)

for pattern in "${patterns[@]}"; do
  go test -run '^$' -bench "$pattern" -benchmem -benchtime "$time" -count "$count" ./benchmarks
done
