# Hermes

[![CI](https://github.com/sidwiskers/hermes/actions/workflows/ci.yml/badge.svg)](https://github.com/sidwiskers/hermes/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sidwiskers/hermes.svg)](https://pkg.go.dev/github.com/sidwiskers/hermes)
[![Release](https://img.shields.io/github/v/release/sidwiskers/hermes)](https://github.com/sidwiskers/hermes/releases/latest)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Hermes is a fast, small Telegram Bot API library **and** framework for modern Go. The module and root package share the `hermes` name.

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/sidwiskers/hermes"
)

func main() {
	bot := hermes.New(os.Getenv("BOT_TOKEN"))

	bot.Use(hermes.Recover())

	bot.Command("start", func(c *hermes.Context) error {
		return c.Send("Hello.")
	})

	bot.Command("profile", func(c *hermes.Context) error {
		return c.Ephemeral("Only you can see this.", hermes.HTML)
	})

	log.Fatal(bot.Run(context.Background()))
}
```

The root facade also exposes the direct typed client, so applications can mix framework helpers and exact Bot API calls without a second client:

```go
message, err := bot.SendMessage(ctx, hermes.SendMessageParams{
    ChatID: chatID,
    Text:   "Hello",
})
```

And a permanent raw escape hatch for every current or future Bot API method:

```go
var result MyTelegramResult
err := bot.Call(ctx, "newTelegramMethod", params, &result)
```

## Package structure

The repository is one Go module with focused packages rather than one flat package:

```text
github.com/sidwiskers/hermes
├── . (package hermes)    # root facade: Bot, routing entry points, re-exports
├── api                   # standalone low-level Bot API client and typed methods
├── types                 # Telegram schemas and keyboard objects
├── framework             # public routing primitives behind the root facade
├── session               # optional typed sessions and pluggable storage
├── fsm                   # optional typed conversation state machines
├── dedupe                # optional atomic update claiming
├── ratelimit             # optional bounded token buckets
├── observe               # optional tracing hooks and fixed-cardinality metrics
├── internal/runtime      # bounded dispatch, polling, webhook lifecycle
└── testkit                # optional network-free test recorder
```

Most applications import only the root package. Low-level-only applications can skip the framework entirely:

```go
client := api.New(os.Getenv("BOT_TOKEN"))
message, err := client.SendMessage(ctx, api.SendMessageParams{
    ChatID: chatID,
    Text:   "Hello",
})
```

Most applications need only the root package. The public `framework` package
keeps canonical documentation for the aliased routing types and supports custom
integrations. The runtime engine remains internal and replaceable.

## Principles

- **Fast by construction:** allocation-free route dispatch, lock-free route reads, bounded concurrency, pooled small transport buffers, connection reuse, and streamed uploads.
- **Tiny core:** the root module uses only the Go standard library.
- **Easy syntax:** common actions are one clear method call.
- **No hidden magic:** explicit contexts, errors, middleware, and concurrency limits.
- **Expandable without rewrites:** transport, typed API, and framework layers are separate.
- **Schema-complete releases:** every supported Bot API version is audited against a checked-in official inventory; ephemeral messages and Rich Messages are native concepts rather than compatibility patches.

## Install

```bash
go get github.com/sidwiskers/hermes
```

Hermes supports Go 1.25 and every newer stable Go release. Use the latest patch
available in your chosen Go line: the application toolchain supplies the
standard library and therefore determines which standard-library security fixes
are present. The module recommends Go 1.26.5 to contributors without forcing
that toolchain on applications that import Hermes.

## Telegram test environment

Telegram provides a separate test environment for destructive integration
testing. Create the bot and test chat inside Telegram's test DC, then enable
the test method endpoint at construction time:

```go
bot := hermes.New(testToken, hermes.WithTestEnvironment(true))
```

The option selects Telegram's test method and file endpoints at construction
time and adds no per-request branch. It is also available as
`api.WithTestEnvironment`. The opt-in live suite and required environment
variables are documented in [`docs/releasing.md`](docs/releasing.md).

## Project status

Hermes 1.0.0 is the stable release. Its typed API is audited against the Bot API
version recorded in `spec/bot-api.json`, with zero known static parity gaps at
release time. It also provides permanent raw escape hatches, streamed
uploads/downloads, bounded update dispatch, race-tested routing, retry-safe
webhooks, and a standard-library-only runtime.

Telegram evolves independently of Hermes. Version-specific claims and exact
surface counts live in [`docs/schema-parity.md`](docs/schema-parity.md) and the
changelog; `Call` and `CallMultipart` provide day-zero access to newly released
methods while the next typed schema update is prepared. The deterministic
[`Hermes Guardian`](docs/maintenance.md) watches the official API, regenerates
safe additions, and opens evidence-backed draft updates without depending on a
specific AI provider.

The code is held to stable-v1 gates rather than treating “v1” as a first
iteration. Deterministic local, Telegram test-DC, credentialed production,
public webhook-delivery, bounded production-period, and synthetic stress
evidence are recorded rather than inferred from unit tests. See
[`docs/schema-parity.md`](docs/schema-parity.md) and
[`docs/compatibility.md`](docs/compatibility.md). Pinned cross-library results
and their limitations are recorded in [`docs/performance.md`](docs/performance.md).
Cancellation, overload, panic, retry, and shutdown guarantees are specified in
[`docs/reliability.md`](docs/reliability.md).
Credentialed Telegram test-DC and production results are tracked
in [`docs/conformance.md`](docs/conformance.md).

## Routes, filters, and groups

Exact commands and callbacks use map lookup. Ordered filters and nested groups are compiled into the same immutable route snapshot.

```go
bot.On(hermes.TextMessage, func(c *hermes.Context) error {
    return c.Send("I received text.")
})

private := bot.Group(hermes.PrivateChat)
private.Use(authMiddleware)

private.Command("settings", func(c *hermes.Context) error {
    return c.Send("Private settings")
})

admins := private.Group(hermes.FromUsers(1001, 1002))
admins.Command("panel", adminPanel)
```

Filters compose without reflection:

```go
bot.On(
    hermes.All(hermes.GroupChat, hermes.TextPrefix("hello")),
    greetGroup,
)
```

Included filters cover update types, text/captions, common media, chat kinds, users, chats, callback prefixes, and ephemeral messages. Handler contexts are borrowed from a pool by default; call `c.Clone()` before retaining one beyond the handler, or use `hermes.WithContextPooling(false)`.

## Typed callback data

Keep button generation and parsing together:

```go
userCallback := hermes.Int64Callback("user:")

button := userCallback.MustButton("Open", 42)

bot.CallbackPrefix(userCallback.Prefix, userCallback.Handler(
    func(c *hermes.Context, userID int64) error {
        return c.Edit(fmt.Sprintf("User %d", userID))
    },
))
```

The codec checks Telegram's 64-byte callback-data limit before sending.

## Easy context actions

```go
c.Send("Hello")
c.Reply("Replying to this message")
c.Edit("Updated text")
c.EditCaption("Updated caption")
c.EditKeyboard(&keyboard)
c.Delete()
c.Acknowledge()
c.Alert("Important")
c.ChatAction(hermes.ActionTyping)
```

For media IDs or URLs:

```go
c.Photo(photoID, "Caption", hermes.HTML, hermes.WithKeyboard(keyboard))
c.Document(documentID, "Report")
c.Video(videoID, "Trailer", hermes.Streaming)
c.Animation(animationID, "Animation")
c.Audio(audioID, "Song")
c.Voice(voiceID, "Voice note")
c.Sticker(stickerID)
```

Every helper has a corresponding typed low-level method for full parameter control.

## Ephemeral messages

Register a private group command:

```go
err := bot.SetMyCommands(ctx, hermes.SetMyCommandsParams{
    Commands: []hermes.BotCommand{{
        Command:     "profile",
        Description: "Show your private profile",
        IsEphemeral: true,
    }},
})
```

Respond privately from an ephemeral command or callback:

```go
bot.Command("profile", func(c *hermes.Context) error {
    return c.Ephemeral("Level 42 · Rank #7")
})

bot.Callback("private_stats", func(c *hermes.Context) error {
    return c.EphemeralPhoto(statsImageID, "Your stats")
})
```

The context automatically supplies the receiver, callback-query identifier, or incoming ephemeral-message reply identifier.

The typed API supports ephemeral delivery across every ephemeral-capable send
method in the checked-in schema:

- message, photo, animation, audio, document, sticker;
- video, video note, voice;
- contact, location, and venue.

It also includes the complete text/media/caption/keyboard edit lifecycle and deletion.

## Fast and raw update decoding

Normal polling and webhook decoding use the allocation-minimized mode and do not copy the complete JSON payload into every update:

```go
bot := hermes.New(token) // Update.Raw remains empty
```

Raw preservation is explicit when forward compatibility or diagnostics require it:

```go
bot := hermes.New(token, hermes.WithRawUpdates(true))
```

Standalone decoding is also available:

```go
update, err := hermes.DecodeUpdate(payload, true)
```

Raw mode copies the complete payload and therefore costs one additional allocation per update.

## Performance laboratory

`benchmarks`, `cmd/hermesbench`, and the benchmark scripts provide reproducible
decode, routing, middleware, transport, upload, binary-size, and pinned
cross-library workloads. Current results show zero-allocation constant-time
dispatch at 1,000 routes, while also preserving the cells where another library
is faster. See [`docs/performance.md`](docs/performance.md) for the raw evidence,
methodology, and limitations.

## Streaming uploads

New files are streamed from `io.Reader`; the full file is never buffered in memory:

```go
file, err := os.Open("video.mp4")
if err != nil {
    return err
}
defer file.Close()

message, err := bot.SendVideoUpload(
    ctx,
    hermes.SendVideoParams{
        ChatID:            chatID,
        Caption:           "Uploaded",
        SupportsStreaming: true,
    },
    "video.mp4",
    file,
)
```

Primary upload helpers are available for photo, animation, audio, document, sticker, video, video note, and voice. `CallMultipart` remains available for unusual multi-file requests.

## Albums and multi-file streaming

Albums use typed media items and can combine Telegram file IDs, URLs, and streamed attachments:

```go
items := []hermes.MediaGroupItem{
    hermes.InputMediaPhoto{Media: hermes.Attachment("cover"), Caption: "Cover"},
    hermes.InputMediaVideo{Media: hermes.Attachment("clip"), SupportsStreaming: true},
}

messages, err := bot.SendMediaGroupUpload(
    ctx,
    hermes.SendMediaGroupParams{ChatID: chatID, Media: items},
    hermes.NewUpload("cover", "cover.jpg", coverReader),
    hermes.NewUpload("clip", "clip.mp4", clipReader),
)
```

The client supports photo, live-photo, video, audio, and document albums. It validates Telegram's 2–10 item limit, homogeneous audio/document albums, complete live-photo components, attachment/upload parity, and media discriminators before transmission. Every attachment is streamed.

## Polls, dice, and reactions

Simple framework helpers remain short:

```go
c.Poll("Choose one", "Alpha", "Beta")
c.Dice(hermes.DiceEmoji)
c.React(hermes.EmojiReaction("🔥"), true)
```

The low-level API exposes complete parameter structs, including quiz answers, optional booleans that preserve an explicit `false`, strongly typed poll-description/option media, streamed poll-media uploads, revoting, country/member restrictions, and stopping polls. Typed reaction updates and reaction-count updates decode directly into `hermes.ReactionType` and `hermes.ReactionCount`.

## Moderation and chat administration

Round 4 adds typed methods for:

- banning, unbanning, restricting, and promoting members;
- administrator titles, member tags, permissions, and sender-chat bans;
- member/admin lookup and join-request approval or rejection;
- normal and paid-subscription invite links;
- message reactions and reaction removal;
- chat title, description, photo, pins, and sticker-set management.

Context-aware actions use the current update safely:

```go
bot.On(hermes.UpdateIs(hermes.UpdateChatJoinRequest), func(c *hermes.Context) error {
    return c.ApproveJoinRequest()
})

bot.Command("ban", func(c *hermes.Context) error {
    return c.BanSender(true)
})
```

For advanced moderation, use the corresponding typed `bot.RestrictChatMember`, `bot.PromoteChatMember`, or invite-link methods directly.

## File downloads

```go
file, err := bot.GetFile(ctx, fileID)
if err != nil {
    return err
}

_, err = bot.DownloadFile(ctx, file.FilePath, destination)
```

Downloads stream directly to an `io.Writer`. File paths are validated, and network errors never expose the token-bearing URL.

## Webhooks

```go
err := bot.ServeWebhook(
    ctx,
    ":8080",
    "/telegram",
    hermes.WebhookOptions{Secret: os.Getenv("WEBHOOK_SECRET")},
)
```

The webhook handler validates Telegram's secret header, limits request bodies, rejects trailing JSON, returns backpressure instead of dropping updates, and shuts down gracefully.

Use synchronous acknowledgement when handler success must control whether
Telegram retries the update. It also supports one direct Bot API response:

```go
bot.Command("start", func(c *hermes.Context) error {
    chatID, _ := c.ChatID()
    return c.RespondWebhook("sendMessage", hermes.SendMessageParams{
        ChatID: chatID,
        Text:   "Hello",
    })
})

err := bot.ServeWebhookReplies(ctx, ":8080", "/telegram", options)
```

`ServeWebhook` queues accepted work and acknowledges immediately for lower
HTTP latency. `ServeWebhookReplies` runs within the same global concurrency
bound and returns 500 on handler failure so Telegram can redeliver. Direct
replies save an outbound round trip, but Telegram does not return the called
method's result to the bot; use an ordinary typed API call when that result is
needed.

Webhook setup includes typed set/delete/status methods and streamed self-signed certificate upload.

## Long polling

```go
err := bot.Run(ctx)
```

Or configure it directly:

```go
err := bot.Poll(ctx, hermes.PollOptions{
    Timeout:        50,
    Limit:          100,
    AllowedUpdates: []string{"message", "callback_query"},
})
```

Polling uses bounded dispatch, exponential backoff, Telegram flood-wait values, and graceful shutdown.

## Middleware

```go
bot.Use(
    hermes.Recover(),
    hermes.Logger(slog.Default()),
    hermes.Timeout(10*time.Second),
)
```

Middleware chains are compiled when routes change, not for every update. Groups can carry their own middleware without adding another runtime router layer.

## Stateful and production middleware

Typed sessions and finite-state conversations live in optional packages:

```go
store := session.NewMemory[fsm.Snapshot[State, Profile]](24 * time.Hour)
sessions := session.New(store, session.ByChatUser,
    session.WithNamespace("profile"),
)
flow := fsm.New(sessions, StateIdle)
bot.Use(flow.Middleware())
```

Sessions serialize updates by key and commit only after successful handlers by
default, so a failed send rolls state changes back. The `dedupe`, `ratelimit`,
and `observe` packages add atomic update claims, bounded token buckets, trace
hooks, and lock-free aggregate metrics without changing or adding dependencies
to the root package. See the runnable
[`stateful`](examples/stateful) and [`production`](examples/production)
programs and the complete [`feature map`](docs/features.md).

## Testing without Telegram

The optional `testkit` package supplies an in-memory Bot API recorder:

```go
bot, recorder := testkit.New()
recorder.Respond(hermes.Message{MessageID: 9})

_, err := bot.SendMessage(ctx, hermes.SendMessageParams{
    ChatID: 1,
    Text:   "hello",
})

request, _ := recorder.Last()
fmt.Println(request.Method, request.JSON["text"])
```

It records JSON and multipart requests and can queue successful or Telegram-style error responses.

## Complete, versioned Bot API surface

The current structured foundation includes:

- broad update decoding with optional raw forward compatibility;
- message, callback, command, chat, webhook, file, edit/delete, forwarding, and chat-action methods;
- every ephemeral-capable send method in the checked-in schema;
- albums with typed media items and streamed multi-file attachments;
- polls, dice, reactions, members, moderation, permissions, invite links, and chat administration;
- Rich Messages with all 21 input blocks and streamed nested media;
- invoices, Stars, paid media, gifts, business accounts, stories, and managed bots;
- forum topics, checklists, suggested posts, games, Passport errors, and sticker-set lifecycle;
- inline, Web App, guest, and prepared-message results;
- common incoming media/query/payment/community/subscription objects;
- inline and reply keyboards;
- streamed primary media uploads and file downloads;
- filters, groups, typed callback codecs, middleware, and deterministic testing.

Each Hermes release pins and audits an official Bot API schema manifest. The
exact version, surface counts, and audit result are recorded in
[`docs/schema-parity.md`](docs/schema-parity.md). The raw `Call` and
`CallMultipart` layers remain permanent escape hatches for forward
compatibility.

See [`docs/design.md`](docs/design.md) and [`docs/roadmap.md`](docs/roadmap.md).
Runnable polling, webhook, upload, inline-mode, payment, and standalone-client
programs are indexed in [`examples`](examples/README.md).
The automated and manual release gates are defined in
[`docs/releasing.md`](docs/releasing.md).
