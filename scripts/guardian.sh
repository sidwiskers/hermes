#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go run ./internal/cmd/hermes-guardian prepare "$@"
go run ./internal/cmd/hermes-guardian validate
