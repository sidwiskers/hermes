#!/usr/bin/env bash
set -euo pipefail

snapshot="$(mktemp)"
generated="$(mktemp)"
trap 'rm -f "$snapshot" "$generated"' EXIT

curl --fail --location --silent --show-error \
	--connect-timeout 15 --max-time 120 --retry 3 --retry-all-errors \
	https://core.telegram.org/bots/api \
  --output "$snapshot"

go run ./internal/cmd/botapi-schema \
  -source "$snapshot" \
  -output "$generated"

mv "$generated" spec/bot-api.json

go run ./internal/cmd/botapi-models

printf 'updated Bot API schema, generated models, and facade aliases\n'
