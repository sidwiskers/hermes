# Bot API schema parity

Hermes measures compatibility against deterministic inventories derived from
Telegram's official Bot API documentation and the compiled Go packages. Each
release pins its audited Bot API version in `spec/bot-api.json` and its exported
Go declarations in `spec/api-surface.txt`. The current manifest records Bot
API 10.4, released on 2026-11-09, and contains:

| Surface | Official count | Missing in Hermes |
| --- | ---: | ---: |
| Methods | 193 | 0 |
| Method parameters | 982 | 0 |
| Concrete objects | 396 | 0 |
| Object fields | 1,983 | 0 |
| Union roots | 29 | 0 |
| Union variants | 204 | 0 |

The audit additionally verifies parameter and field requiredness, JSON
`omitempty` behavior, Go wire types, and nilability of optional nested objects.
Every category must remain at zero gaps before a Bot API update is considered
complete.

## Source of truth

`spec/bot-api.json` is generated from <https://core.telegram.org/bots/api> by:

```bash
./scripts/update-bot-api-schema.sh
```

The manifest records the Bot API version, release date, source URL, source
document SHA-256, methods, parameters, objects, fields, unions, and variants.
The parser and checked-in counts are covered by tests.

Run the local source audit with:

```bash
go run ./internal/cmd/botapi-audit
go run ./internal/cmd/botapi-audit -json
```

The release gate (`scripts/release-check.sh`) requires the complete report to
remain at zero. CI also verifies that generated models, facade aliases, and
`spec/api-surface.txt` are byte-for-byte current.

The scheduled [Hermes Guardian](maintenance.md) parses the same source daily.
It ignores documentation-only HTML changes, classifies protocol changes into a
machine-readable diff, regenerates safe declarations, and opens a draft pull
request. If new Telegram semantics cannot be completed mechanically, Guardian
keeps the exact parity report in the pull request instead of guessing or
weakening this audit.

## Bot API 10.4: widgets and Wallet

The typed client exposes Telegram's native widget lifecycle:

- `SendWidget` sends an HTML widget with its dimensions, public/chat state,
  query permissions, shared mode, and ordinary message delivery options.
- `SendSharedWidget` sends a previously registered widget by `widget_id`.
- `AnswerWidgetQuery` returns data to a widget query.
- `EditMessageWidgetData` updates public or chat-scoped state for a widget
  message. Pointer fields preserve the distinction between leaving a value
  unchanged and explicitly clearing it.
- `EditSharedWidgetData` changes shared public state and its cache lifetime.

Widget query updates are available as `Update.WidgetQuery` and
`Update.SharedWidgetQuery`; `Update.Type()` distinguishes both, and
`Update.Sender()` resolves their user. The typed query union recognizes data,
wallet-signing, and sent-transaction events while preserving unknown future
variants as raw JSON.

TON Wallet data is represented by the `TonWalletCard`, `TonNft`,
`TonWalletCardSticker`, and `TonWalletTransfer` types. Chat share themes and
rich-text TON addresses have their own typed representations. These models
describe Telegram-provided data; applications remain responsible for their own
transaction authorization, verification, and key-handling policy.

Bot API 10.4 also permits ephemeral send targets to be specified by a callback
query without a `chat_id`. Hermes accepts this only when
`EphemeralMessageParameters.CallbackQueryID` is present and carries the
parameters consistently through JSON and multipart send paths. The older
`ReceiverUserID` and `CallbackQueryID` convenience fields remain available
for source compatibility and are translated into the nested
`ephemeral_message_parameters` object.

The generated types also include `PaidMessagesWaived` service messages,
the optional `suggested_username` on managed-bot keyboard requests, and the
new `ton_address` and `bank_card_number` message-entity type values.
The new private-bot-topic link form uses the existing message-thread/topic
identifiers; it does not require another Bot API method.

## Completion strategy

1. Use marker interfaces and discriminator-aware encoding for request unions,
   so invalid combinations fail at compile time whenever Go can express the
   constraint.
2. Add discriminator-aware decoding and unknown-variant preservation for
   response unions, so new Telegram variants remain inspectable before Hermes
   updates.
3. Extend conformance proof beyond static shape to upload behavior and
   request/response fixture verification.

Schema declarations are generated from the factual manifest. Validators,
transport behavior, ergonomic constructors, and compatibility adapters remain
reviewed Go code. This keeps future Bot API updates mechanical without moving
reflection or generated machinery into runtime hot paths.
