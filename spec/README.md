# Machine contracts

This directory contains versioned, machine-readable contracts used to prove
Hermes compatibility:

- `bot-api.json` is the normalized Telegram Bot API schema parsed from the
  official documentation.
- `api-surface.txt` is the deterministic inventory of exported Go declarations
  across every public Hermes package.

Both files are generated. Do not edit either file to conceal a schema or
compatibility difference.

Regenerate the Telegram schema and derived models with:

```bash
./scripts/update-bot-api-schema.sh
```

Regenerate the exported Go API inventory with:

```bash
go run ./internal/cmd/api-surface
```

Verify all generated contracts without changing the worktree with:

```bash
./scripts/check-generated.sh
```
