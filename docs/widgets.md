# Telegram HTML widgets

Hermes exposes Bot API 10.4's native HTML widget methods through the typed
client. Widgets are Telegram-rendered interactive message content; they are
not ordinary Web Apps and do not require the application to implement a
browser-based Telegram Web App launch flow.

## Send a widget

```go
message, err := bot.SendWidget(ctx, hermes.SendWidgetParams{
    ChatID:       chatID,
    HTML:         "<main><h1>Status</h1><button>Refresh</button></main>",
    Width:        480,
    Height:       320,
    PublicData:   "release=stable",
    ChatData:     "view=overview",
    AllowQueries: true,
})
if err != nil {
    return err
}
_ = message
```

The `HTML` field is the widget document. `PublicData` is the shared/public
widget state, while `ChatData` is scoped to the chat. Set `AllowQueries` when
the widget needs to send queries to the bot. Keep untrusted user content out of
HTML unless it is safely escaped and validated by the application.

## Handle widget queries

Widget and shared-widget events are decoded on `Update`:

```go
bot.On(hermes.UpdateIs(hermes.UpdateWidgetQuery), func(c *hermes.Context) error {
    query := c.Update.WidgetQuery
    if query == nil {
        return nil
    }

    return c.AnswerWidgetQuery(hermes.AnswerWidgetQueryParams{
        WidgetQueryID: query.ID,
        Data:          "current-state",
    })
})
```

`WidgetQuery.Type` is a typed union. Use `Kind()` to distinguish `data`,
`sign_data`, and `sent_transaction`; the matching variant is available in
`Data`, `SignData`, or `SentTransaction`. Unknown future variants are
retained in `Raw` and can be re-emitted without being discarded.

For direct client use, call:

- `AnswerWidgetQuery` to return a query result.
- `EditMessageWidgetData` to change state for one widget message.
- `EditSharedWidgetData` to change the public state of a shared widget.
- `SendSharedWidget` to place an already registered widget in another
  message.

The pointer-valued `PublicData` and `ChatData` arguments on
`EditMessageWidgetDataParams` distinguish an omitted update from an explicit
empty string, which clears that state.

## TON Wallet events

The widget query union also exposes Wallet interactions. Telegram may include a
wallet card in a data query or provide signing and transaction event payloads.
The received wallet data is modeled by `TonWalletCard`,
`TonWalletCardSticker`, `TonNft`, and `WidgetQueryWallet`. A wallet transfer
service message is exposed through `Message.TonWalletTransfer`.

Treat these as protocol data, not proof that a transaction is safe or final.
The application must independently validate transaction intent and follow
its own signing, custody, and authorization requirements.

## Shared themes and rich text

Chat share themes are decoded as `ChatShareTheme` with `Emoji` and
`UniqueGift` variants. Each emoji item uses the `EmojiType` union, allowing
standard emoji and custom emoji IDs. Rich-text TON address nodes are available
as `RichTextTonAddress`.

See the [Bot API schema parity guide](schema-parity.md) for the complete
10.4 protocol inventory and the regression gates.