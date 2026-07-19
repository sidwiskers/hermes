#!/usr/bin/env bash
set -euo pipefail

export GOMAXPROCS="${GOMAXPROCS:-1}"
count="${BENCH_COUNT:-10}"
time="${BENCH_TIME:-1s}"
pattern="${BENCH_PATTERN:-^(BenchmarkCompetitorDecodeUpdate|BenchmarkRouterExact1|BenchmarkRouterExact1000|BenchmarkRouterMiddleware10|BenchmarkAPICall)$}"

printf 'Hermes competitor benchmark\n'
printf 'date_utc: %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf 'go: %s\n' "$(go version)"
printf 'gomaxprocs: %s\n' "$GOMAXPROCS"
printf 'count: %s\n' "$count"
printf 'benchtime: %s\n' "$time"
if [[ -r /proc/cpuinfo ]]; then
	printf 'cpu: %s\n' "$(awk -F ': ' '/model name/ {print $2; exit}' /proc/cpuinfo)"
fi
printf 'go_env:\n'
go env -json GOOS GOARCH GOAMD64 CGO_ENABLED

printf '\n[hermes]\n'
go test -run '^$' -bench "$pattern" -benchmem -benchtime "$time" -count "$count" ./benchmarks

for module in tgbotapi telebot telebotv4 gotgbot gotelegrambot telego gotg; do
	printf '\n[%s]\n' "$module"
	(
		cd "benchmarks/competitors/$module"
		go test -run '^$' -bench "$pattern" -benchmem -benchtime "$time" -count "$count"
	)
done
