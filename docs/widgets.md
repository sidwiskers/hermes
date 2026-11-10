# Telegram HTML widgets

Hermes exposes Bot API 10.4's native HTML widget methods through the typed
client. Widgets are Telegram-rendered interactive message content; they are
not ordinary Web Apps and do not require the application to implement a
browser-based Telegram Web App launch flow.

In the current production environment, Telegram permits sending widgets only
to private chats. A sent widget is automatically protected and cannot be
forwarded; widget messages also cannot be copied with `copyMessage` or
`copyMessages`. The inline-query input-content models are available in Hermes,
but Telegram currently does not permit inline-query widget delivery in
production.

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

The `HTML` field is the widget document. Telegram limits the HTML to 32,768
UTF-8 bytes and to 4,096 bytes after zlib compression at level 9. The optional
`PublicData` and `ChatData` state values are each limited to 1,024 UTF-8 bytes.
The library validates these limits before sending. Keep untrusted user content
out of HTML unless it is safely escaped and validated by the application.

`PublicData` is the shared/public widget state, while `ChatData` is scoped to
the message. Set `AllowQueries` when the widget needs to send queries to the
bot. Use `IsShared` when copies of the widget should share the same public
state. Telegram assigns the resulting widget its identifier, which is returned
in the `Message.Widget` field.

## Handle widget queries

Widget and shared-widget events are decoded on `Update` and are also exposed
as `Context.WidgetQuery` and `Context.SharedWidgetQuery`:

```go
bot.On(hermes.UpdateIs(hermes.UpdateWidgetQuery), func(c *hermes.Context) error {
    if c.WidgetQuery == nil {
        return nil
    }
    // Answer promptly; the widget query expires after 15 seconds.
    return c.AnswerWidgetQuery("current-state")
})
```

`WidgetQuery.Type` is a typed union. Use `Kind()` to distinguish `data`,
`sign_data`, and `sent_transaction`; the matching variant is available in
`Data`, `SignData`, or `SentTransaction`. Unknown future variants are
retained in `Raw` and can be re-emitted without being discarded.

A widget request contains 1-64 bytes of data. Answer every received query within
15 seconds, including failed requests. The optional answer data is limited to
1,024 bytes. `AnswerWidgetQuery` sends the result only to the requesting
widget; it does not modify widget state.

For direct client use, call `AnswerWidgetQuery` to return a query result,
`EditMessageWidgetData` to change state for one widget message,
`EditSharedWidgetData` to change the public state of a shared widget, or
`SendSharedWidget` to place an already registered widget in another message.
Widget state fields use byte limits, not character limits. Passing an empty
string to either edit method clears the corresponding state; omitting a pointer
field in `EditMessageWidgetDataParams` leaves it unchanged.

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
