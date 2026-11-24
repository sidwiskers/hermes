# Telegram Bot API 10.5

Hermes tracks the official Bot API manifest and exposes the 10.5 additions
through typed models and methods.

## Change set

The official schema comparison against Bot API 10.4 contains one new method,
one new object, ten fields moved from `User` to `BotFullInfo`, one new
`WriteAccessAllowed` field, and the replacement of `text` with `blocks` in
the two expandable-quotation objects. No methods or union variants were
removed.

## Bot settings

`Client.SetMySettings` sends Telegram's `setMySettings` method. Its
`SetMySettingsParams` contains every optional setting from the 10.5 schema:
inline mode and placeholder, inline location, chosen-result feedback,
group membership and privacy, business connectivity, managed bots, guest
queries, join-request queries, private-chat topics, topic creation permissions,
topic sidebar, bot-to-bot communication, and callback-query mode.

Use pointer helpers for optional booleans. A nil pointer leaves the setting
unchanged; a pointer to false explicitly disables it:

```go
err := bot.SetMySettings(ctx, hermes.SetMySettingsParams{
    CanReadAllGroupMessages: hermes.Bool(false),
    CanCommunicateWithBots:  hermes.Bool(true),
    HasTopicSidebar:         hermes.Bool(true),
    CallbackQueryMode:       hermes.CallbackQueryModeAuto,
})
```

The inline feedback denominator accepts 1, 10, 100, or 1000; zero disables
chosen-inline-result feedback. Hermes validates that exact set locally.
Callback modes are:

- `default`: `message` is used, including an inaccessible-message object
  when Telegram cannot provide the full message.
- `auto`: Telegram uses `inline_message_id` instead of `message` when the
  originating bot-sent message is inaccessible; a just-deleted message may
  still arrive as an inaccessible `message`.
- `inline`: Telegram always uses `inline_message_id` instead of `message`
  for callback buttons in messages sent directly by the bot.

## Callback handler implications

When Telegram uses `inline_message_id`, the update contains no ordinary
message to derive a chat ID or message ID from. Hermes exposes the alternate
identifier using `Context.InlineMessageID()`. The context's `Edit`,
`EditCaption`, and `EditKeyboard` helpers use it automatically.

Do not assume `Context.Message`, `Context.ChatID()`, or
`Context.MessageID()` are available in these callbacks. Other actions that
need a chat and message pair (including delete, reaction and pin operations)
cannot be addressed with an inline message ID because the corresponding Bot
API methods do not accept one. `Sender()`, `Data()`, the callback query ID,
and answering the callback remain available.

## BotFullInfo and invited users

`Client.GetMeFullInfo` returns the complete official `*BotFullInfo` result,
including bot identity and configuration fields, active usernames, inline
query placeholder/location settings, and topic-sidebar state. Bot IDs use
`int64` on every architecture, matching the library’s other Telegram ID types.
The existing
`Client.GetMe` signature remains available as a deprecated compatibility
projection to `*User`; its ten historical bot-setting fields remain on
`User` with their original `omitempty` JSON tags and are populated by that
method. Zero-valued flags remain omitted for ordinary `User` values. New code
should prefer `GetMeFullInfo`.

A user joining through the bot owner's dedicated invite link is reported with
a `WriteAccessAllowed` service message. Its new `Inviter` field identifies
the inviting owner. Telegram's dedicated bot invite link can invite up to ten
users at once.

## Expandable quotations

`InputRichBlockExpandableBlockQuotation` now accepts nested
`[]InputRichBlock` and an optional credit. Received expandable quotations
decode into the existing generic `RichBlock` representation, whose
`Blocks` field preserves nested paragraphs, tables, media and other block
types. Hermes recursively validates nested input blocks, including button
actions and draft-only thinking blocks.

## Compatibility note

Hermes keeps the pre-10.5 `GetMe() (*User, error)` signature for source
compatibility and provides `GetMeFullInfo() (*BotFullInfo, error)` for the
official Telegram result. The legacy projection does not add bot-only fields
to JSON encoding of ordinary `User` values.
