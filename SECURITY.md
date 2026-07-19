# Security policy

## Supported versions

| Version | Supported |
| --- | --- |
| Latest `1.x` minor | Yes |
| Older `1.x` minors | By published advisory only |
| Pre-1.0 releases | No |

Security fixes are made on the default branch and released from the latest
supported minor line. A published advisory will state explicitly when support
is extended to an older minor release.

## Reporting a vulnerability

Do not open a public issue. Use GitHub's private vulnerability-reporting form:

<https://github.com/sidwiskers/hermes/security/advisories/new>

Include the affected version or commit, impact, a minimal reproduction, and any
suggested mitigation. Do not test against bots, accounts, or chats you do not
control. The project will acknowledge a report within 72 hours, keep the
reporter informed during investigation, and coordinate disclosure after a fix
is available.

Never include a bot token, webhook secret, or credential-bearing Telegram file
URL in an issue, fixture, panic, benchmark result, or log. If a credential is
exposed, revoke it with BotFather immediately; redaction does not make an
already exposed credential safe.

## Release security gates

Stable releases run the Go vulnerability scanner, enumerate runtime module
dependencies, scan the repository for common credential forms, test token
redaction, and verify bounded request bodies, constant-time webhook-secret
comparison, panic containment, and graceful overload behavior.
