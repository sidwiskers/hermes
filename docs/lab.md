# Hermes Lab

Hermes Lab is a small, deterministic Telegram environment for Go tests. It
runs the real Hermes router, middleware, context helpers, sessions, and FSMs,
but replaces Telegram's network API with an in-process model.

Importing Lab is optional:

```go
import "github.com/sidwiskers/hermes/testkit"
```

Applications that do not import `testkit` pay no binary-size or runtime cost.
Lab itself uses only the Go standard library and Hermes.

## A complete conversation test

```go
func TestSettings(t *testing.T) {
    lab := testkit.NewLab(t)

    lab.Bot.Command("settings", func(c *hermes.Context) error {
        return c.Send("Settings", hermes.WithKeyboard(
            hermes.Keyboard(hermes.Row(hermes.Button("Save", "save"))),
        ))
    })
    lab.Bot.Callback("save", func(c *hermes.Context) error {
        if err := c.Acknowledge(); err != nil {
            return err
        }
        return c.Edit("Saved")
    })

    alice := lab.PrivateUser(42, "alice")
    alice.Command("settings").Want(testkit.Sent("Settings"))
    alice.Callback("save").Want(
        testkit.Acknowledged(),
        testkit.Edited("Saved"),
    )
}
```

`NewLab` returns an ordinary `*hermes.Bot` connected to the virtual API. Bot
options are accepted, but Lab always applies its non-network transport and bot
identity last so a test cannot accidentally reach Telegram.

## Virtual actors

Create an actor in a private chat or supergroup:

```go
alice := lab.PrivateUser(42, "alice")
bob := lab.GroupUser(-100123, 84, "bob")
```

An actor can send a command, plain text, or any custom message shape:

```go
alice.Command("search", "fast", "bots")
alice.Text("hello")
alice.Message(hermes.Message{
    Photo: []hermes.PhotoSize{{FileID: "photo-id"}},
})
```

`Callback` presses a button on the most recent virtual bot message in that
chat. `CallbackOn` targets a specific retained message. Missing update fields
such as sender, chat, message ID, update ID, and date are filled with stable
values. Explicit fields are preserved.

## Step expectations

Each actor action returns a `Step`: one input update, its handler result, and
only the Bot API calls caused by that update. Built-in expectations cover the
common conversation assertions:

```go
alice.Command("start").Want(
    testkit.Called("sendChatAction"),
    testkit.Sent("Ready"),
)

alice.Callback("close").Want(
    testkit.Acknowledged(),
    testkit.Deleted(),
)
```

Available expectations include `Called`, `Sent`, `Edited`, `Answered`,
`Acknowledged`, `Deleted`, `Uploaded`, and `NoCalls`. `Matching` handles exact
application-specific assertions while keeping access to the decoded request:

```go
step.Want(testkit.Matching("sendMessage", func(request testkit.Request) bool {
    return request.JSON["disable_notification"] == true
}))
```

`Step.Requests` and `Lab.Requests` remain available when direct inspection is
clearer. JSON requests are decoded into `Request.JSON`; streamed multipart
fields and files appear in `Request.Form` and `Request.Files`.

## Stateful Bot API model

Common send methods automatically return valid message objects. Lab retains
those outgoing messages, applies text/caption/keyboard edits, and removes
deleted messages. This makes callbacks and multi-step conversations work
without queuing boilerplate responses:

```go
messages := alice.Messages()
last, ok := lab.API.LastMessage(alice.Chat().ID)
```

Incoming and outgoing message IDs share a deterministic per-chat sequence,
matching Telegram's chat-local message identity. Lab is safe for concurrent
actors, and each step retains only its own outbound requests even when handlers
run at the same time.

## Failure injection

Failures are queued by Bot API method, so an unrelated call cannot consume the
scenario:

```go
lab.API.RateLimitNext("sendMessage", 1500*time.Millisecond)
alice.Command("start").WantAPIError(429)

networkDown := errors.New("network unavailable")
lab.API.TransportErrorNext("sendMessage", networkDown)
alice.Command("start").WantError(networkDown)

lab.API.FailNext("editMessageText", 400, "message cannot be edited", nil)
alice.Callback("save").WantAPIError(400)
```

`RespondNext` supplies an exact success result for any typed or raw method that
the common model does not emulate:

```go
lab.API.RespondNext("getChatMemberCount", 3)
```

## Retry and isolation tests

`Step.Replay` delivers the exact same update again, including its update ID.
It is intended for duplicate suppression and webhook-redelivery tests:

```go
first := alice.Text("charge")
first.Replay().Want(testkit.NoCalls())
```

`Lab.Reset` clears generated IDs, virtual messages, queued failures, and
recorded calls while retaining routes and actors. Application-owned stores,
including session and deduplication stores, are intentionally not reset.

## Boundary

Lab proves application behavior deterministically; it is not a replacement for
Telegram conformance testing. Unknown Bot API methods return a successful
boolean by default, so tests that need another result shape should call
`RespondNext`. Use Telegram's separate test environment for final integration
coverage involving Telegram-side permissions, limits, or delivery behavior.
