# Live webhook probe

This build-tagged command is release tooling, not an application template. It
exercises real Telegram webhook delivery, synchronous replies, deliberate
retryable handler errors, panic containment, duplicate-claim release, and
graceful shutdown with a dedicated disposable bot.

The release bundle contains static Linux amd64 and arm64 binaries. On a
disposable Linux host, run:

```bash
./run-quick-tunnel.sh
```

The script prompts for the bot token without echoing it, downloads the pinned
Cloudflare Tunnel binary for the detected architecture, verifies its official
SHA-256 checksum, starts the probe on loopback, and creates a temporary
`trycloudflare.com` HTTPS URL. It requires no domain, Cloudflare account,
inbound firewall rule, or root access.

Keep the process running until the release operator has deleted the Telegram
webhook. Press Ctrl-C to stop the tunnel and drain the probe. Quick Tunnels
have no uptime guarantee and must not be used for a deployed bot.
