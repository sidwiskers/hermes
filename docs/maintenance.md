# Maintaining Hermes over time

Hermes Guardian is a repository-owned maintenance pipeline, independent of any
ChatGPT session. The runtime library remains Go standard-library-only. Python,
Docker, model services and GitHub APIs are maintenance tools, never dependencies
of applications importing Hermes.

The [owner guide](guardian-owner.md) covers setup, controls, costs and recovery.
The default policy prepares PRs for approval. Coding repair, automatic merging,
and automatic releases each require explicit configuration.

## Detection and evidence

`python3 scripts/guardian/main.py prepare` downloads Telegram's official Bot API
with a bounded response size, timeout and retries. The trusted Go parser creates
its structural manifest. An independent HTML parser snapshots normalized
section descriptions, return-contract prose, field descriptions and link targets
in `spec/bot-api-semantics.json`. The snapshot also retains recent release
announcements so announcement-only changes can be referred for review.
Cosmetic HTML markup is ignored; changes to
existing documentation are conservatively referred for review. Wording changes
can therefore produce a review even when Telegram's actual behavior is unchanged.

Both inventories determine a stable source identity. Version rollback, duplicate
names, incomplete documents and large schema losses stop the update. A missing
semantic baseline requires review; it is never silently accepted. The checked-in
10.3 baseline was captured from the official document alongside this change.

## Generation and repair

The existing Go generator creates object declarations, field extensions and
facade aliases. Full static parity is audited. This generator does not invent
method implementations or behavior from prose. Non-mechanical changes and
remaining gaps can be handed to the optional provider-neutral repair controller.

Repair supports HTTPS chat-completions services, ordered provider fallback and
fresh-context review. It offers only repository reads, bounded text searches,
exact text edits, new files and a fixed isolated test command. It has no arbitrary
shell tool. The agent may change library/framework Go code, add regression tests,
and update documentation. Existing tests, official manifests, generated files,
module dependencies, automation and release policy are protected. Agent-created
tests can be corrected within the same update, but baseline tests cannot be
weakened. New behavior should cover JSON, multipart, zero/false values, response
decoding, unions and compatibility as applicable.

Provider credentials are kept in the controller. They are not sent as prompt
text or forwarded to subprocesses or candidate containers. HTTP redirects cannot
forward a provider authorization header to another host. Raw provider error
bodies are not logged. No provider is required for detection or mechanical
updates. Unavailable repair services leave an actionable update branch.

Each source identity has a bounded repair count persisted in
`.github/guardian/update.json` on its update branch. Subsequent attempts receive
previous review and validation feedback. Owners can explicitly request another
bounded attempt. AI claims alone cannot mark an update ready.

## Isolation and checks

Generators and orchestration run from trusted main. Candidate tests run in a
separate Docker container with a read-only repository and root filesystem,
no inherited credentials, dropped capabilities, no privilege escalation, and
CPU, memory, process, time and captured-log limits. Intermediate repair tests have no network.
Full release validation allows network access for public dependency and
vulnerability data downloads but receives no repair, publishing, or live-bot
credentials. No host home directory, Docker socket or credential cache is mounted.

The complete existing `release-check.sh` remains authoritative: typed parity,
generated-source checks, formatting, vet, shuffled tests, race tests, coverage
floors, security/dependency checks, examples, integration compilation, competitor
adapters, benchmark smoke tests and cross-platform builds. The benchmark smoke
check is not a statistical proof against all performance regressions, and the
integration compilation is not live Telegram conformance. Novel behavior still
needs appropriate independent conformance evidence and maintainer review.
After the full gate passes, a separate Go 1.25 container checks generated files,
runs vet, and executes the full test suite to verify the minimum supported Go
version.

The sandbox mounts candidate source read-only so tests cannot rewrite the patch
that is later published. The publication job starts on a fresh runner and reads
a bounded JSON bundle; it never executes candidate code. Automatic publication
accepts ordinary text files only and refuses protected-path changes.

## Branch lifecycle and publication

Update branches are named `automation/bot-api-VERSION-SOURCE_DIGEST`. Each source
snapshot has its own branch. Existing repair work is resumed and main is merged
without force-pushing. Conflicts preserve the remote branch and require attention.
A concurrent main or branch change invalidates publication. Identical trees reuse
the existing commit, allowing required CI to finish without daily commit churn.

Mechanical updates can be merged using the normal GitHub merge API, pinned to
the validated candidate head. Repository protection is never bypassed. Repaired
updates remain subject to maintainer approval even after passing tests and an AI
review. The explicit `release-current` action validates the exact current main
before publishing. Tags are never moved; a failed release can resume from an
already-created matching tag. The old single `automation/bot-api-update` branch,
if present, is left untouched for its maintainer to finish or retire.

For GitHub's built-in token, newly created or synchronized PR workflows may need
an owner approval. An optional dedicated token enables unattended PR checks.
See [GitHub's workflow triggering documentation](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/trigger-a-workflow).

## Health and owner controls

A persistent status issue and each run's summary explain outcomes. The separate
health workflow checks for a successful maintenance run within 48 hours and can
re-enable/dispatch the maintenance workflow. It cannot protect against both
GitHub schedules being disabled, quota exhaustion, or a platform outage; an
external watcher is supported for that reason. GitHub documents schedule delays
and automatic disabling after repository inactivity in its
[event reference](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#schedule).

## Running elsewhere and offline replay

The pipeline uses Git, Python 3.10+, a current supported stable Go toolchain, and
Docker. No Python packages are required. Supply `GITHUB_REPOSITORY=owner/repo`
when resuming or publishing remote branches. Supply `GH_TOKEN` only to the
publication/status command. `GUARDIAN_IMAGE` may select a matching `hermes-guardian:VERSION` image.
The trusted Dockerfile builds it from the official Go image and installs Python
and Git for the maintenance checks. GitHub selects the Go version installed by
setup-go; local validation builds a missing image automatically.

```bash
./scripts/guardian.sh
# Or keep repair credentials and publishing credentials in separate processes:
python3 scripts/guardian/main.py prepare
python3 scripts/guardian/main.py validate
python3 scripts/guardian/main.py publish
python3 scripts/guardian/main.py status
```

Do not run the publisher if no `.guardian/bundle.json` was produced (for example,
an unchanged API). The GitHub workflow handles this automatically. Publication
requires a complete cloned Git history and stable release tags for versioning.
Use a disposable clean checkout; local reports and candidates live in `.guardian/`.

For offline source replay:

```bash
python3 scripts/guardian/main.py prepare --source /path/to/official-bot-api.html
python3 -m unittest discover -s scripts/guardian -p 'test_*.py'
```

The legacy `scripts/guardian.sh` entry point now delegates to the same controller.
A successful no-change run produces no candidate bundle. Detection failures clear
stale bundles and preserve the last published branch. Reports are retained for
seven days in Actions and summarized durably in the update PR.
