# Maintaining Hermes over time

Hermes maintenance is deterministic first. A language model may help implement
new semantics, but no named model, provider, API, or SDK is part of the update
pipeline. If every AI service disappears, Guardian still detects Telegram
changes, regenerates safe declarations, identifies every remaining typed gap,
and preserves an actionable branch for a maintainer.

## Hermes Guardian

The `Hermes Guardian` workflow runs daily and can also be started manually from
GitHub Actions. Its local entry point is:

```bash
./scripts/guardian.sh
```

The command:

1. downloads Telegram's official Bot API document with bounded retries;
2. parses it into a candidate deterministic manifest;
3. ignores source hash and release-date changes when the protocol is unchanged;
4. emits JSON and Markdown surface diffs;
5. updates the checked-in manifest and generated declarations when the protocol
   changed;
6. regenerates the exported API inventory;
7. runs the full typed parity audit without suppressing gaps; and
8. classifies the result as `unchanged`, `ready`, or `review`.

Local reports are written under `.guardian/` and are intentionally ignored by
Git. `diff.json` is the machine-readable protocol change, `diff.md` is the human
summary, and `audit.json` plus `audit.txt` describe the exact typed parity state.

An `unchanged` result does nothing. A `ready` result is limited to additive
object declarations, has zero static parity gaps, and must pass the complete
release gates before Guardian opens a draft pull request. Method changes, union
changes, removals, non-additive fields, version-only releases, and all remaining
audit gaps are classified as `review` even if generated code compiles. A review
result opens a draft pull request whose evidence names the hand-written work
still required. Guardian never
merges, tags, publishes a stable release, uses live Telegram credentials, or
changes protected `main` directly.

If downloading, parsing, generation, compilation, or GitHub publication fails,
the workflow creates or refreshes one infrastructure-failure issue with a link
to the run. A failure cannot silently turn into a release.

## Repair-agent contract

Any coding agent—hosted, local, commercial, open-source, or written in the
future—may repair a `review` branch. Give it the repository branch and these
inputs:

- `.guardian/diff.json` or the protocol diff in the draft pull request;
- the official `spec/bot-api.json` candidate;
- the complete `botapi-audit -json` report; and
- the failing CI and release-gate output.

The agent's output is ordinary Go source, tests, examples, and documentation on
the same branch. It must obey these non-negotiable rules:

- do not edit the official manifest to hide Telegram declarations;
- do not weaken, skip, or special-case the parity, compatibility, security,
  generated-source, race, coverage, or performance gates;
- preserve the zero-runtime-dependency core, streamed uploads, bounded memory,
  and reflection-free dispatch hot path;
- preserve existing exported behavior unless semantic versioning permits a
  documented change;
- add typed validation and wire fixtures for new semantics, not only structs;
- keep raw JSON and multipart escape hatches forward compatible; and
- never receive repository release credentials or live bot tokens as prompt
  text.

Completion is decided by deterministic evidence, not by the agent claiming that
the task is complete:

```bash
go run ./internal/cmd/botapi-audit
./scripts/check-generated.sh
./scripts/verify.sh
./scripts/security-check.sh
./scripts/release-check.sh
```

A second independent review—human or agent—is useful for novel Telegram
semantics, but the protected branch, reproducible checks, and live conformance
evidence remain the authority.

## Operating without GitHub Actions

Scheduled workflows on inactive public repositories can be disabled by the
hosting platform. Guardian therefore remains a normal repository script rather
than a GitHub-only service. Any cron system can clone the repository, install a
current stable Go release, and run `./scripts/guardian.sh`. The output contract
does not change, so another forge or automation service can publish the same
branch and report.

## One-time repository setting

For automatic draft pull requests, enable **Settings → Actions → General →
Workflow permissions → Allow GitHub Actions to create and approve pull
requests**. Guardian requests only repository contents, pull-request, and issue
write access. Branch protection and required CI remain in force. A pull request
created with the repository `GITHUB_TOKEN` may require a maintainer to approve
its CI run; that is an intentional final control, not an automation failure.
