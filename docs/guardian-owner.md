# Managing Hermes without writing code

Guardian runs through repository workflows and does not depend on an ongoing AI conversation.
You do not need a ChatGPT subscription to run it. It needs a working scheduler
and build runner; optional coding repairs also need a separately configured model
service. An API account and its quota are separate from a ChatGPT subscription.

## What it can do by itself

Guardian checks Telegram's official API and its documented behavior daily. For
supported additive object changes, it generates the models, validates the library,
updates release metadata, and prepares an update PR. In automatic mode it can
merge these mechanical updates and publish their releases, subject to your
repository's normal merge rules.

New methods, changed return behavior, new union variants, and compatibility
changes need coding work. With repair enabled, Guardian attempts that work,
runs isolated tests, obtains a separate AI review, and prepares a PR with the
results. These updates require your release approval. An AI review is evidence,
not a guarantee that an unfamiliar Telegram feature is correct.

If a repair cannot pass the checks, Guardian preserves the branch and tells you
what failed. It does not release a partial implementation or weaken the checks.
There is no guarantee that arbitrary future API changes can always be solved
without a programmer. Running applications keep their installed version until
you upgrade and rebuild them.

## Start here

Guardian defaults to preparing pull requests for approval. Coding repair,
automatic merging, and automatic releases are off unless you enable them.

1. Open **Actions → Hermes Guardian → Run workflow**.
2. Keep branch **main** and action **check**, then select **Run workflow**.
3. Read the run's summary and the **Hermes maintenance status** issue. These show
   whether Hermes is up to date, an update is ready, or something needs attention.
4. Review any update PR before merging it. The [release guide](releasing.md)
   explains the checks and the separate publishing step.

You do not need model credentials to check for updates or generate supported
additions. Scheduled checks run daily while GitHub Actions is available and the
workflow is enabled.

To change the policy, open **Settings → Secrets and variables → Actions → Variables**:

| Variable | Default | Available controls |
| --- | --- | --- |
| `GUARDIAN_MODE` | `pull-request` | Keep every merge manual; choose `automatic` for verified generated updates, or `paused` to stop update preparation |
| `GUARDIAN_RELEASE` | `false` | Set to `true` to publish releases after automatic merges |
| `GUARDIAN_REPAIR` | `false` | Set to `true` after configuring a repair provider below |

Leave `GUARDIAN_MODE` unset or set it to `pull-request` to keep approval of every
change to main. Automatic releases require both `GUARDIAN_MODE=automatic` and
`GUARDIAN_RELEASE=true`. These settings do not authorize automatic merging of
coding repairs. The health check respects `paused`; the explicit
`release-current` action remains a separate owner-requested publishing action.

**GitHub permissions:** Actions must be allowed to create pull requests in
**Settings → Actions → General → Workflow permissions**. This repository setting
was already enabled when the system was implemented.

GitHub's built-in token may leave automation-created PR checks waiting for your
approval. For unattended checks, add an optional `GUARDIAN_GITHUB_TOKEN` secret
containing a fine-grained token for **this repository only**, with Contents and
Pull requests read/write permissions. It is used only in the publication job;
repair providers and candidate tests never receive it. Keep required checks and
branch protection enabled. Guardian will report a blocked merge rather than
bypassing them. An expiring token needs renewal; the status issue will report
publication failures.

## Enable coding repairs

Use a model service with an HTTPS chat-completions endpoint and a coding-capable
model. It may be hosted by any compatible provider or by you. The URL must be
the complete endpoint, including its `/chat/completions` path.

Under **Secrets**, add its API key as `GUARDIAN_KEY_1`. Under **Variables**, add:

- `GUARDIAN_REPAIR`: `true`
- `GUARDIAN_PROVIDERS`: the JSON below, replacing the URL and model with the
  values supplied by your provider. Never put the API key in this JSON.

```json
[
  {
    "url": "https://YOUR-PROVIDER/v1/chat/completions",
    "model": "YOUR-CODING-MODEL",
    "key_env": "GUARDIAN_KEY_1"
  }
]
```

You can add a second and third entry using `GUARDIAN_KEY_2` and `GUARDIAN_KEY_3`.
Guardian tries providers in order if one fails. The review starts with the last
configured available provider, in a fresh context. With only one provider, it
still uses a separate review conversation. It cannot automatically obtain free
model access, renew a subscription, or guarantee a provider's availability.

Each repair run has at most 24 action turns, three intermediate test runs, and
a 20-minute decision-loop budget (an already-started request/test may finish).
A provider response is capped at 8,000 output tokens. Provider fallback may make
up to three calls per turn. A source update gets at most **two unattended repair
runs**, persisted on its branch. The separate review can make up to three more
calls per run. These are upper bounds, not a spending estimate: use your
provider's account-level spending cap as well. An unconfigured service makes
no model calls.

## The three workflow actions

| Action | What you do | What Guardian does |
| --- | --- | --- |
| `check` | Run whenever you want a status/update check | Checks Telegram, prepares or resumes the update, and applies the configured policy |
| `retry-repair` | Use after restoring provider quota or when you want another attempt | Allows one more bounded repair run for the current update |
| `release-current` | Use after merging a reviewed update, or to finish an interrupted release | Revalidates current main and publishes the changelog version without moving an existing tag |

For a repaired PR, read its summary and validation result. If the update is
complete and you choose to accept it, mark the draft ready and merge through
GitHub. Then run `release-current`. There is no code to write. A draft with
failing checks should remain unmerged.

## Understanding status

| Status | Meaning / action |
| --- | --- |
| unchanged | Hermes matches the current API and documentation baseline |
| ready | Mechanical update passed release gates; automatic policy may merge it |
| awaiting merge | GitHub is waiting for required checks or approvals; open the linked PR |
| reviewed | Coding repair passed release gates and a separate AI review; review and merge its PR |
| needs attention | Open the linked run/PR; restore provider access, retry, or seek help for the reported unresolved change |
| paused | You paused maintenance through `GUARDIAN_MODE` |
| healthy | The maintenance workflow completed successfully within the last 48 hours; consult the update issue for code readiness |

Source files, diff, audit, and validation summaries are saved in the PR and its
run artifacts. Run artifacts expire after seven days to limit storage use; the
branch, PR, and persisted repair count remain available.

## Keeping the watcher alive

A second workflow, **Hermes maintenance health**, checks the maintenance
heartbeat daily. If no successful run exists within 48 hours, it dispatches a
new one and re-enables a disabled maintenance workflow unless you paused it.

Both workflows still depend on GitHub. They cannot recover themselves if both
schedules are disabled, the account cannot run Actions, or GitHub is down. For
independent monitoring, use a scheduler you already manage to run
`scripts/guardian-watch.sh` once a day. It dispatches the health workflow through
GitHub's API and needs only a repository-scoped token with Actions read/write.
It does not need Go, Docker, AI access, or a permanently running bot.

If GitHub Actions is unavailable, the maintenance commands can run on another
Linux runner with Git, Python 3.10+, a supported Go toolchain, and Docker. See
[the maintenance design](maintenance.md). Full validation is build work and
needs more resources than the lightweight watcher. A stopped external scheduler
also needs your attention; there is no scheduler that can guarantee its own
availability forever.
