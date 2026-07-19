# Bot API schema parity

Hermes measures compatibility against a deterministic inventory derived from
Telegram's official Bot API documentation. The checked-in Bot API 10.2
inventory contains:

| Surface | Official count | Missing in Hermes |
| --- | ---: | ---: |
| Methods | 185 | 0 |
| Method parameters | 937 | 0 |
| Concrete objects | 362 | 0 |
| Object fields | 1,838 | 0 |
| Union roots | 26 | 0 |
| Union variants | 187 | 0 |

The audit additionally verifies parameter and field requiredness, JSON
`omitempty` behavior, Go wire types, and nilability of optional nested objects.
All checked categories are currently at zero gaps.

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

CI requires the complete report to remain at zero and verifies that generated
models and facade aliases are byte-for-byte current.

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
