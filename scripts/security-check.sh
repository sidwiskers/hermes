#!/usr/bin/env bash
set -euo pipefail

./scripts/check-secrets.sh

module_count="$(go list -m all | wc -l | tr -d ' ')"
printf 'dependency inventory: %s module(s), including Hermes\n' "$module_count"
go list -m all

version="${GOVULNCHECK_VERSION:-v1.6.0}"
tool_dir="$(mktemp -d)"
trap 'rm -rf "$tool_dir"' EXIT

printf 'installing govulncheck %s into an isolated temporary directory\n' "$version"
GOBIN="$tool_dir" go install "golang.org/x/vuln/cmd/govulncheck@${version}"
"$tool_dir/govulncheck" ./...
