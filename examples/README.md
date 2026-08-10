# Hermes examples

Every example is compiled in CI. Runtime examples are complete `main`
packages; the Lab example is a directly runnable Go test.

| Directory | Focus |
| --- | --- |
| `basic` | Long polling, commands, middleware, and graceful shutdown |
| `expanded` | Groups, typed callbacks, filters, polls, and moderation |
| `library` | Standalone low-level API client |
| `webhook` | Webhook registration, authentication, and server lifecycle |
| `webhook-reply` | Retry-safe synchronous acknowledgement and direct Bot API replies |
| `upload` | Constant-memory streamed file upload |
| `inline` | Typed inline-query results |
| `payments` | Telegram Stars invoices and pre-checkout approval |
| `stateful` | Typed sessions and finite-state conversations |
| `production` | Deduplication, rate limiting, observability, and owned cleanup |
| `lab` | Deterministic conversation, callback, and Bot API assertions without Telegram |

Export `BOT_TOKEN` before running an example. Individual examples document any
additional environment variables they require.
