#!/usr/bin/env bash
set -euo pipefail

matches="$(
	rg -l -I --hidden \
		-g '!.git/**' \
		-e '[0-9]{8,12}:[A-Za-z0-9_-]{30,}' \
		-e '-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----' \
		-e 'api\.telegram\.org/(file/)?bot[0-9]{8,12}:' \
		. || true
)"

if [[ -n "$matches" ]]; then
	printf 'possible credential material found; inspect these files without publishing their contents:\n' >&2
	while IFS= read -r file; do
		printf '  %s\n' "$file" >&2
	done <<< "$matches"
	exit 1
fi

printf 'credential-pattern scan passed\n'
