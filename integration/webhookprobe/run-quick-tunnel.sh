#!/usr/bin/env bash
set -Eeuo pipefail

readonly cloudflared_version="2026.7.2"
readonly release_base="https://github.com/cloudflare/cloudflared/releases/download/${cloudflared_version}"

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"

case "$(uname -m)" in
	x86_64)
		readonly probe="./hermes-webhook-linux-amd64"
		readonly cloudflared_asset="cloudflared-linux-amd64"
		readonly cloudflared_sha256="ec905ea7b7e327ff8abdde8cb64697a2152de74dbcdbf6aec9db8364eb3886cd"
		;;
	aarch64 | arm64)
		readonly probe="./hermes-webhook-linux-arm64"
		readonly cloudflared_asset="cloudflared-linux-arm64"
		readonly cloudflared_sha256="405df476437e027fc6d18729a5a77155c0a33a6082aeee60a799a688f3052e66"
		;;
	*)
		echo "unsupported architecture: $(uname -m)" >&2
		exit 1
		;;
esac

if [[ ! -x "$probe" ]]; then
	echo "missing probe binary: $probe" >&2
	exit 1
fi

cloudflared="./${cloudflared_asset}"
if [[ ! -x "$cloudflared" ]]; then
	temporary="${cloudflared}.download"
	echo "Downloading cloudflared ${cloudflared_version}..."
	curl --fail --location --proto '=https' --tlsv1.2 \
		--output "$temporary" "${release_base}/${cloudflared_asset}"
	actual_sha256="$(sha256sum "$temporary" | cut -d' ' -f1)"
	if [[ "$actual_sha256" != "$cloudflared_sha256" ]]; then
		echo "cloudflared checksum mismatch" >&2
		exit 1
	fi
	mv -- "$temporary" "$cloudflared"
	chmod 700 "$cloudflared"
fi

read -rsp "Disposable Telegram bot token: " bot_token
echo
if [[ -z "$bot_token" ]]; then
	echo "bot token is required" >&2
	exit 1
fi

webhook_secret="HermesProbe_$(od -An -N24 -tx1 /dev/urandom | tr -d ' \n')"
echo "Webhook secret: ${webhook_secret}"
echo "Keep this terminal open and send Codex the secret and trycloudflare URL."

HERMES_TEST_BOT_TOKEN="$bot_token" \
	HERMES_TEST_WEBHOOK_SECRET="$webhook_secret" \
	"$probe" &
probe_pid=$!
unset bot_token

cleanup() {
	kill "$probe_pid" 2>/dev/null || true
	wait "$probe_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

for _ in {1..100}; do
	if curl --silent --show-error --output /dev/null --max-time 0.2 \
		http://127.0.0.1:8080/telegram 2>/dev/null; then
		break
	fi
	if ! kill -0 "$probe_pid" 2>/dev/null; then
		wait "$probe_pid"
		exit 1
	fi
	sleep 0.1
done

"$cloudflared" tunnel --no-autoupdate --url http://127.0.0.1:8080
