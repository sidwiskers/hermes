#!/usr/bin/env bash
set -euo pipefail

generated="$(mktemp)"
api_aliases="$(mktemp)"
root_aliases="$(mktemp)"
api_surface="$(mktemp)"
trap 'rm -f "$generated" "$api_aliases" "$root_aliases" "$api_surface"' EXIT

go run ./internal/cmd/botapi-models \
	-output "$generated" \
	-api-output "$api_aliases" \
	-root-output "$root_aliases"

if ! cmp --silent types/zz_botapi_generated.go "$generated"; then
	printf 'types/zz_botapi_generated.go is stale; run go generate ./types\n' >&2
	exit 1
fi

if ! cmp --silent api/zz_botapi_aliases_generated.go "$api_aliases"; then
	printf 'api/zz_botapi_aliases_generated.go is stale; run go generate ./types\n' >&2
	exit 1
fi

if ! cmp --silent zz_botapi_aliases_generated.go "$root_aliases"; then
	printf 'zz_botapi_aliases_generated.go is stale; run go generate ./types\n' >&2
	exit 1
fi

go run ./internal/cmd/api-surface -output "$api_surface"

if ! cmp --silent spec/api-surface.txt "$api_surface"; then
	printf 'spec/api-surface.txt changed; regenerate it and review compatibility intentionally\n' >&2
	exit 1
fi
