#!/usr/bin/env bash
set -euo pipefail

readonly minimum_major=1
readonly minimum_minor=26
readonly minimum_patch=5

version="${1:-$(go env GOVERSION)}"

# Release evidence must come from an official stable toolchain. Development,
# beta, and release-candidate builds deliberately do not match this grammar.
if [[ ! "$version" =~ ^go([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
	printf 'release validation requires a stable Go release; found %s\n' "$version" >&2
	exit 1
fi

major="${BASH_REMATCH[1]}"
minor="${BASH_REMATCH[2]}"
patch="${BASH_REMATCH[3]}"

if (( major < minimum_major ||
	(major == minimum_major && minor < minimum_minor) ||
	(major == minimum_major && minor == minimum_minor && patch < minimum_patch) )); then
	printf 'release validation requires Go 1.26.5 or a newer stable release; found %s\n' "$version" >&2
	exit 1
fi

if [[ "${HERMES_GO_VERSION_SILENT:-0}" != "1" ]]; then
	printf 'release toolchain accepted: %s\n' "$version"
fi
