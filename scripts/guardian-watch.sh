#!/usr/bin/env bash
# Run from an independent scheduler; no checkout, Go, Docker or AI is needed.
set -euo pipefail
: "${GITHUB_REPOSITORY:?Set GITHUB_REPOSITORY to owner/repository}"
: "${GUARDIAN_WATCH_TOKEN:?Set an Actions read/write token for this repository}"
[[ "$GITHUB_REPOSITORY" =~ ^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$ ]]
# Keep the token out of argv/process listings and never enable shell tracing.
set +x
[[ "$GUARDIAN_WATCH_TOKEN" =~ ^[a-zA-Z0-9_]+$ ]]
curl --fail --silent --show-error --connect-timeout 15 --max-time 60 \
  --retry 2 --retry-all-errors --config - <<EOF_CURL
url = "https://api.github.com/repos/${GITHUB_REPOSITORY}/actions/workflows/guardian-health.yml/dispatches"
request = "POST"
header = "Authorization: Bearer ${GUARDIAN_WATCH_TOKEN}"
header = "Accept: application/vnd.github+json"
header = "Content-Type: application/json"
data = "{\"ref\":\"main\"}"
EOF_CURL
