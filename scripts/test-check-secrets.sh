#!/usr/bin/env bash
set -euo pipefail

check="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-secrets.sh"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT

repository="$test_root/repository"
mkdir -p "$repository"
git -C "$repository" init -q
printf 'clean fixture\n' > "$repository/clean.txt"
git -C "$repository" add clean.txt

if ! (cd "$repository" && "$check") >/dev/null 2>&1; then
	printf 'expected a clean repository to pass the credential scan\n' >&2
	exit 1
fi

secret="12345678:$(printf 'A%.0s' {1..32})"
printf '%s\n' "$secret" > "$repository/suspect.txt"
git -C "$repository" add suspect.txt

output="$test_root/output"
if (cd "$repository" && "$check") >"$output" 2>&1; then
	printf 'expected credential-shaped material to fail the scan\n' >&2
	exit 1
fi
if ! grep -q 'suspect.txt' "$output"; then
	printf 'expected the scan failure to identify the affected file\n' >&2
	exit 1
fi
if grep -qF "$secret" "$output"; then
	printf 'credential scan output must not disclose matched material\n' >&2
	exit 1
fi

outside="$test_root/outside"
mkdir -p "$outside"
if (cd "$outside" && "$check") >"$output" 2>&1; then
	printf 'expected the credential scan to fail closed outside a Git worktree\n' >&2
	exit 1
fi
if ! grep -q 'not inside a Git worktree' "$output"; then
	printf 'expected a clear failure when repository metadata is unavailable\n' >&2
	exit 1
fi

printf 'credential-pattern scan policy tests passed\n'
