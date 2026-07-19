#!/usr/bin/env bash
set -euo pipefail

version="$(go env GOVERSION)"
patch="${version#go1.26.}"
if [[ "$version" != go1.26.* || ! "$patch" =~ ^[0-9]+$ || "$patch" -lt 5 ]]; then
	printf 'release validation requires patched Go 1.26.5 or newer in the 1.26 series; found %s\n' "$version" >&2
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
