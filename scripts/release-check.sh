#!/usr/bin/env bash
set -euo pipefail

version="$(go env GOVERSION)"
minimum_major=1
minimum_minor=26
minimum_patch=5
version_supported=false

if [[ "$version" =~ ^go([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
	major="${BASH_REMATCH[1]}"
	minor="${BASH_REMATCH[2]}"
	patch="${BASH_REMATCH[3]}"

	if (( major > minimum_major ||
		(major == minimum_major && minor > minimum_minor) ||
		(major == minimum_major && minor == minimum_minor && patch >= minimum_patch) )); then
		version_supported=true
	fi
fi

if [[ "$version_supported" != true ]]; then
	printf 'release validation requires stable Go 1.26.5 or newer; found %s\n' "$version" >&2
	exit 1
fi

if [[ "${RELEASE_ALLOW_DIRTY:-0}" != "1" ]] && [[ -n "$(git status --porcelain)" ]]; then
	printf 'release validation requires a clean worktree\n' >&2
	exit 1
fi

go run ./internal/cmd/botapi-audit
./scripts/verify.sh
./scripts/security-check.sh
./scripts/verify-competitors.sh
go test -count=1 ./examples/...
bash -n ./integration/webhookprobe/run-quick-tunnel.sh
go test -tags=integration -run '^$' ./integration/...
go test -run '^$' -bench '^BenchmarkRouterExact1$' -benchtime 100ms ./benchmarks

for target in windows/amd64 darwin/amd64 darwin/arm64 linux/arm64; do
	goos="${target%/*}"
	goarch="${target#*/}"
	printf 'compile: %s/%s\n' "$goos" "$goarch"
	GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build ./...
done

if [[ "${RELEASE_ALLOW_DIRTY:-0}" != "1" ]] && [[ -n "$(git status --porcelain)" ]]; then
	printf 'release checks changed the worktree\n' >&2
	exit 1
fi

printf 'automated release gates passed with %s\n' "$version"
