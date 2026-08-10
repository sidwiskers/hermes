# Hermes Fleet

Hermes Fleet is an optional way to run several independent Hermes bots in one
Go process. It gives a multi-bot application one lifecycle, one shared HTTP
client, and—when using webhooks—one HTTP server. Each bot still owns its token,
routes, middleware, sessions, finite-state machines, concurrency bound, and
error handling.

Fleet is useful when several Hermes bots belong in one deployment and process
overhead matters, such as a small virtual machine. It is not a daemon, a
container runtime, or a requirement for normal bots. Existing applications can
continue to call `bot.Run(ctx)` without importing `fleet` or executing any
Fleet code.

## Polling example

Create bots through the Fleet when they should share its HTTP client and
default update-concurrency setting:

```go
host := fleet.New(fleet.WithBotMaxConcurrentUpdates(32))

alerts := host.NewBot(os.Getenv("ALERTS_BOT_TOKEN"))
alerts.Command("start", alertsStart)

support := host.NewBot(os.Getenv("SUPPORT_BOT_TOKEN"))
support.Command("start", supportStart)

if err := host.Mount("alerts", alerts); err != nil {
    log.Fatal(err)
}
if err := host.Mount("support", support); err != nil {
    log.Fatal(err)
}

if err := host.Run(ctx); err != nil {
    log.Fatal(err)
}
```

Polling is the default. Per-bot options remain available:

```go
err := host.Mount("alerts", alerts, fleet.WithPolling(hermes.PollOptions{
    Timeout:        50,
    Limit:          100,
    AllowedUpdates: []string{"message", "callback_query"},
}))
```

An existing bot constructed with `hermes.New` can also be mounted. It keeps
its own HTTP client and construction options:

```go
bot := hermes.New(token, hermes.WithMaxConcurrentUpdates(16))
err := host.Mount("existing", bot)
```

## One webhook server

Webhook bots can share one hardened HTTP server while retaining separate exact
paths, secrets, and dispatchers:

```go
host := fleet.New(fleet.WithWebhookAddress(":8080"))

first := host.NewBot(firstToken)
second := host.NewBot(secondToken)

err := host.Mount("first", first, fleet.WithWebhook(
    "/telegram/first",
    hermes.WebhookOptions{Secret: firstSecret},
))
if err != nil {
    log.Fatal(err)
}
err = host.Mount("second", second, fleet.WithWebhookReplies(
    "/telegram/second",
    hermes.WebhookOptions{Secret: secondSecret},
))
if err != nil {
    log.Fatal(err)
}

if err := host.Run(ctx); err != nil {
    log.Fatal(err)
}
```

Fleet rejects duplicate or malformed paths before startup. Routing is an exact
map lookup, so `/telegram/first/` does not match `/telegram/first`. Queued and
synchronous reply webhooks preserve the same authentication, request bounds,
backpressure, retry, and graceful-drain behavior as the ordinary Hermes
webhook helpers. TLS termination remains the application's reverse-proxy
responsibility.

Polling and webhook bots may be mixed in one Fleet.

## Failure behavior and status

The default `fleet.IsolateFailures` policy contains preparation or update-source
failure to the affected bot. Healthy polling bots continue if one bot cannot
prepare, and healthy polling bots remain active if the shared webhook listener
fails. A shared listener failure necessarily affects every webhook mounted on
that listener.

Failures go to a panic-contained handler and are also retained in status:

```go
host := fleet.New(fleet.WithErrorHandler(func(failure *fleet.Failure) {
    logger.Error("bot stopped",
        "bot", failure.Bot,
        "phase", failure.Phase,
        "error", failure.Err,
    )
}))

for _, status := range host.Status() {
    fmt.Println(status.Name, status.Mode, status.State, status.LastError)
}
```

Use `fleet.WithFailurePolicy(fleet.StopAllOnFailure)` when the bots form one
logical application and any source failure should stop the deployment. Context
cancellation stops intake and waits for active handlers to drain under either
policy. Mounting and a second `Run` are rejected while a Fleet is active.

## Resource evidence

The checked-in Linux harness compares one five-bot Fleet process with five
one-bot processes. On the recorded Go 1.26.5 amd64 host, the median of five
steady-idle webhook samples was:

| Layout | RSS | PSS | Goroutines | File descriptors |
| --- | ---: | ---: | ---: | ---: |
| One Fleet process, five bots | 9,379,840 B | 5,432,320 B | 3 | 8 |
| Five one-bot processes | 46,665,728 B | 17,264,640 B | 15 | 40 |
| Measured saving | 79.9% | 68.5% | 12 | 32 |

Run the same experiment on a target host with:

```bash
go run ./cmd/hermesfleetbench -bots 5 -samples 5
```

The complete machine-readable record is in
[`benchmarks/results/2026-08-10-go1.26.5-fleet-5-idle.json`](../benchmarks/results/2026-08-10-go1.26.5-fleet-5-idle.json).
These are synthetic idle-process measurements, not production throughput or
latency claims. Application state, caches, and handler workloads can dominate
real memory use.

## Boundaries

Fleet deliberately does not merge bot state or routing. It does not add
cross-bot message passing, persistence, deployment supervision, or a global
handler limit. Configure each bot's bounded dispatcher for the host's capacity.

The efficiency gain comes with the normal same-process boundary: an operating
system kill, out-of-memory condition, or fatal application failure affects all
bots in that Fleet. Hermes contains ordinary handler panics and Fleet isolates
normal source failures by default, but it cannot provide OS process isolation.
Keep separate processes when independent failure domains matter more than total
resource use.
