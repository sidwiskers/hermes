#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

if [[ "${HERMES_GUARDIAN_ALLOW_DIRTY:-0}" != "1" ]] && [[ -n "$(git status --porcelain)" ]]; then
	printf 'guardian requires a clean worktree\n' >&2
	exit 1
fi

output="${HERMES_GUARDIAN_OUTPUT:-.guardian}"
mkdir -p "$output"
output="$(cd "$output" && pwd)"

snapshot="$output/bot-api.html"
candidate="$output/bot-api.json"
surface_status="$output/surface-status"

curl --fail --location --silent --show-error \
	--connect-timeout 15 --max-time 120 --retry 3 --retry-all-errors \
	https://core.telegram.org/bots/api \
	--output "$snapshot"

go run ./internal/cmd/botapi-schema \
	-source "$snapshot" \
	-output "$candidate"

go run ./internal/cmd/botapi-diff \
	-before spec/bot-api.json \
	-after "$candidate" \
	-format json \
	-status-file "$surface_status" \
	>"$output/diff.json"

go run ./internal/cmd/botapi-diff \
	-before spec/bot-api.json \
	-after "$candidate" \
	-format markdown \
	>"$output/diff.md"

classification="$(<"$surface_status")"
if [[ "$classification" == "unchanged" ]]; then
	printf 'unchanged\n' >"$output/status"
	printf 'Telegram Bot API surface is unchanged\n'
	exit 0
fi

cp "$candidate" spec/bot-api.json
go run ./internal/cmd/botapi-models
go run ./internal/cmd/api-surface -output spec/api-surface.txt

go run ./internal/cmd/botapi-audit \
	-json \
	-allow-gaps \
	-status-file "$output/parity-status" \
	>"$output/audit.json"
go run ./internal/cmd/botapi-audit -allow-gaps >"$output/audit.txt"

case "$classification:$(<"$output/parity-status")" in
	mechanical:complete)
		printf 'ready\n' >"$output/status"
		printf 'Telegram Bot API update is mechanically complete and ready for release gates\n'
		;;
	mechanical:gaps|review:complete|review:gaps)
		printf 'review\n' >"$output/status"
		printf 'Telegram Bot API update requires reviewed implementation; see %s\n' "$output/audit.txt"
		;;
	*)
		printf 'guardian produced an invalid change or parity classification\n' >&2
		exit 1
		;;
esac
