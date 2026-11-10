# Changelog

## Unreleased

### Telegram Bot API 10.4

- Update the checked-in official schema to Bot API 10.4 (2026-11-09):
  193 methods, 982 method parameters, 396 objects, 1,983 object fields,
  29 union roots, and 204 union variants.
- Add typed native HTML widget methods for sending standalone and shared
  widgets, answering widget queries, and updating per-message or shared widget
  data. Widget message content is also supported by inline-query input content.
- Add discriminator-aware `ChatShareTheme`, `EmojiType`, and `WidgetQueryType`
  response unions, including round-trip preservation of unknown future
  variants. Expose widget and shared-widget updates through `Update`, including
  update classification and sender attribution.
- Add TON Wallet card and transfer models, wallet-related message and chat
  fields, chat-share theme models, and the `ton_address` rich-text discriminator.
- Support callback-scoped ephemeral sends without requiring a chat ID, while
  retaining explicit validation and multipart handling for all typed send
  paths. Make `receiver_user_id` optional as specified by Bot API 10.4.
- Add rich-text quote fields and formatted verification descriptions, plus
  lookup of unique gifts by name and deletion of business-account gifts.
- Include the new paid-message waiver service message, managed-bot suggested
  usernames, and `ton_address`/`bank_card_number` message-entity values in
  the generated protocol models.
- Add request/response, discriminator, unknown-variant, payload-limit, and
  ephemeral-send regression tests for the new protocol surface.
- Document production widget restrictions, query deadlines, and Telegram's
  UTF-8 byte-based limits. Premium bot username changes remain managed through
  Telegram's account tooling rather than a Bot API method.

## 1.4.0 - 2026-09-10
