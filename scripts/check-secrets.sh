#!/usr/bin/env bash
set -euo pipefail

if ! command -v git >/dev/null 2>&1; then
	printf 'credential-pattern scan failed: git is unavailable\n' >&2
	exit 2
fi

if ! repository="$(git rev-parse --show-toplevel 2>/dev/null)"; then
	printf 'credential-pattern scan failed: not inside a Git worktree\n' >&2
	exit 2
fi

status=0
matches="$(
	git -C "$repository" grep -l -I -E \
		-e '[0-9]{8,12}:[A-Za-z0-9_-]{30,}' \
		-e '-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----' \
		-e 'api\.telegram\.org/(file/)?bot[0-9]{8,12}:' \
		-- .
)" || status=$?

if ((status > 1)); then
	printf 'credential-pattern scan failed with status %d\n' "$status" >&2
	exit "$status"
fi

if [[ -n "$matches" ]]; then
	printf 'possible credential material found; inspect these files without publishing their contents:\n' >&2
	while IFS= read -r file; do
		printf '  %s\n' "$file" >&2
	done <<< "$matches"
	exit 1
fi

printf 'credential-pattern scan passed\n'
