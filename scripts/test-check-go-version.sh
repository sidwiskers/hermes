#!/usr/bin/env bash
set -euo pipefail

check="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-go-version.sh"

expect_accept() {
	local version="$1"
	if ! HERMES_GO_VERSION_SILENT=1 "$check" "$version" >/dev/null 2>&1; then
		printf 'expected release toolchain %s to be accepted\n' "$version" >&2
		exit 1
	fi
}

expect_reject() {
	local version="$1"
	if HERMES_GO_VERSION_SILENT=1 "$check" "$version" >/dev/null 2>&1; then
		printf 'expected release toolchain %s to be rejected\n' "$version" >&2
		exit 1
	fi
}

for version in go1.26.5 go1.26.6 go1.27.0 go1.30.1 go2.0.0; do
	expect_accept "$version"
done

for version in go1.25.99 go1.26.0 go1.26.4 go1.27rc1 go1.27.0-rc.1 devel-go1.27 unknown; do
	expect_reject "$version"
done

printf 'release toolchain version policy tests passed\n'
