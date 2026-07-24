# Bot API schema parity

Hermes measures compatibility against deterministic inventories derived from
Telegram's official Bot API documentation and the compiled Go packages. Each
release pins its audited Bot API version in `spec/bot-api.json` and its exported
Go declarations in `spec/api-surface.txt`. The Bot API manifest shipped with
Hermes 1.0.0 records Bot API 10.2 and contains:

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
models, facade aliases, and `spec/api-surface.txt` are byte-for-byte current.

The scheduled [Hermes Guardian](maintenance.md) parses the same source daily.
It ignores documentation-only HTML changes, classifies protocol changes into a
machine-readable diff, regenerates safe declarations, and opens a draft pull
request. If new Telegram semantics cannot be completed mechanically, Guardian
keeps the exact parity report in the pull request instead of guessing or
weakening this audit.

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
