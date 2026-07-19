#!/usr/bin/env bash
set -euo pipefail

files="$(gofmt -l .)"
if [[ -n "$files" ]]; then
  printf 'These files are not gofmt-formatted:\n%s\n' "$files" >&2
  exit 1
fi

./scripts/check-secrets.sh
./scripts/check-generated.sh

go vet ./...
go test -shuffle=on -count=1 ./...
go test -race -shuffle=on -count=1 ./...

coverage_file="$(mktemp)"
coverage_log="$(mktemp)"
trap 'rm -f "$coverage_file" "$coverage_log"' EXIT
go test -coverprofile="$coverage_file" ./... | tee "$coverage_log"

assert_minimum() {
  local label="$1"
  local actual="$2"
  local minimum="$3"
  awk -v label="$label" -v actual="$actual" -v minimum="$minimum" 'BEGIN {
    if ((actual + 0) < (minimum + 0)) {
      printf "%s coverage %.1f%% is below required %.1f%%\n", label, actual, minimum > "/dev/stderr"
      exit 1
    }
  }'
}

package_coverage() {
  local package="$1"
  awk -v package="$package" '
    $1 == "ok" && $2 == package {
      for (i = 1; i <= NF; i++) {
        if ($i == "coverage:") {
          value = $(i + 1)
          gsub(/%/, "", value)
          print value
          exit
        }
      }
    }
  ' "$coverage_log"
}

total="$(go tool cover -func="$coverage_file" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')"
assert_minimum total "$total" "${MIN_COVERAGE:-39.0}"

while IFS='|' read -r package minimum; do
  actual="$(package_coverage "$package")"
  if [[ -z "$actual" ]]; then
    printf 'coverage result missing for %s\n' "$package" >&2
    exit 1
  fi
  assert_minimum "$package" "$actual" "$minimum"
done <<EOF_THRESHOLDS
github.com/sidwiskers/hermes|${MIN_ROOT_COVERAGE:-38.0}
github.com/sidwiskers/hermes/api|${MIN_API_COVERAGE:-46.0}
github.com/sidwiskers/hermes/dedupe|${MIN_DEDUPE_COVERAGE:-80.0}
github.com/sidwiskers/hermes/framework|${MIN_FRAMEWORK_COVERAGE:-30.0}
github.com/sidwiskers/hermes/fsm|${MIN_FSM_COVERAGE:-74.0}
github.com/sidwiskers/hermes/internal/runtime|${MIN_RUNTIME_COVERAGE:-66.0}
github.com/sidwiskers/hermes/observe|${MIN_OBSERVE_COVERAGE:-84.0}
github.com/sidwiskers/hermes/ratelimit|${MIN_RATELIMIT_COVERAGE:-82.0}
github.com/sidwiskers/hermes/session|${MIN_SESSION_COVERAGE:-80.0}
github.com/sidwiskers/hermes/testkit|${MIN_TESTKIT_COVERAGE:-54.0}
github.com/sidwiskers/hermes/types|${MIN_TYPES_COVERAGE:-15.0}
EOF_THRESHOLDS

printf 'coverage: total %s%%; package floors passed\n' "$total"
