#!/usr/bin/env python3
"""Repository maintenance control plane. Python standard library only."""
import argparse
import base64
import datetime as dt
import json
import os
import re
import shutil
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

from protocol import canonical, digest, semantic_changes, snapshot, validate_candidate
from repair import allowed, providers, repair, review

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / ".guardian"
WORK = OUT / "work"
STATE = ".github/guardian/update.json"
GENERATED = {"spec/bot-api.json", "spec/bot-api-semantics.json", "spec/api-surface.txt",
             "types/zz_botapi_generated.go", "api/zz_botapi_aliases_generated.go",
             "zz_botapi_aliases_generated.go", STATE, "CHANGELOG.md", "README.md",
             "docs/schema-parity.md", "api/client.go"}


def run(args, cwd=ROOT, timeout=120, check=True):
    env = {k: v for k, v in os.environ.items() if not (
        k.startswith(("GUARDIAN_KEY_", "ACTIONS_", "INPUT_")) or
        k in ("GH_TOKEN", "GITHUB_TOKEN", "GUARDIAN_PROVIDERS"))}
    env["GIT_TERMINAL_PROMPT"] = "0"
    result = subprocess.run(args, cwd=cwd, text=True, stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, timeout=timeout, env=env)
    if check and result.returncode:
        raise RuntimeError("Command failed: " + " ".join(args[:4]) + "\n" + result.stdout[-8000:])
    return result.stdout if check else result


def read_json(path):
    return json.loads(Path(path).read_text())


def write_json(path, value):
    Path(path).parent.mkdir(parents=True, exist_ok=True)
    Path(path).write_text(json.dumps(value, indent=2, ensure_ascii=False) + "\n")


def api(route, method="GET", body=None, missing=False):
    if not route.startswith("/repos/"):
        raise ValueError("repository API route required")
    headers = {"Accept": "application/vnd.github+json", "User-Agent": "hermes-guardian"}
    token = os.environ.get("GH_TOKEN")
    if token:
        headers["Authorization"] = "Bearer " + token
    request = urllib.request.Request("https://api.github.com" + route, method=method,
        headers=headers, data=None if body is None else json.dumps(body).encode())
    try:
        with urllib.request.urlopen(request, timeout=45) as response:
            data = response.read(12_000_001)
            if len(data) > 12_000_000:
                raise RuntimeError("GitHub response too large")
            return json.loads(data) if data else None
    except urllib.error.HTTPError as exc:
        if missing and exc.code == 404:
            return None
        # Deliberately exclude response bodies and authenticated request objects.
        raise RuntimeError(f"GitHub {method} request failed (HTTP {exc.code})") from None


def repo_route():
    repo = os.environ.get("GITHUB_REPOSITORY", "")
    if not re.fullmatch(r"[\w.-]+/[\w.-]+", repo):
        raise ValueError("set GITHUB_REPOSITORY to owner/repository")
    return "/repos/" + repo


def source_file(path=None):
    target = OUT / "bot-api.html"
    if path:
        shutil.copyfile(path, target)
        return target
    request = urllib.request.Request("https://core.telegram.org/bots/api",
                                     headers={"User-Agent": "hermes-guardian"})
    for attempt in range(3):
        try:
            with urllib.request.urlopen(request, timeout=45) as response:
                data = response.read(4_000_001)
            if len(data) > 4_000_000:
                raise ValueError("official document exceeds size limit")
            target.write_bytes(data)
            return target
        except (OSError, ValueError):
            if attempt == 2:
                raise RuntimeError("Cannot download official API; previous state is preserved") from None
            time.sleep(2 ** attempt)


def build_tools():
    for name in ("botapi-schema", "botapi-diff", "botapi-models", "botapi-audit", "api-surface"):
        run(["go", "build", "-o", str(OUT / name), "./internal/cmd/" + name], timeout=180)


def tool(name, *args, cwd=ROOT, check=True):
    return run([str(OUT / name), *map(str, args)], cwd=cwd, timeout=180, check=check)


def scope_check(root, base):
    names = run(["git", "diff", "--name-only", base], cwd=root).splitlines()
    names += run(["git", "ls-files", "--others", "--exclude-standard"], cwd=root).splitlines()
    for name in set(names):
        path = root / name
        existed = run(["git", "cat-file", "-e", base + ":" + name], cwd=root, check=False).returncode == 0
        if name not in GENERATED and not allowed(name, existed):
            raise ValueError("Candidate modified protected file: " + name)
        if not path.is_file() or path.is_symlink() or not path.resolve().is_relative_to(root.resolve()):
            raise ValueError("Candidate deleted or replaced a file: " + name)
        if path.stat().st_size > 2_000_000:
            raise ValueError("Candidate file too large: " + name)
    return sorted(set(names))


def validation_image(minimum=False):
    version = "1.25" if minimum else run(["go", "env", "GOVERSION"]).strip().removeprefix("go")
    if not re.fullmatch(r"\d+\.\d+(?:\.\d+)?", version):
        raise ValueError("validation requires a stable Go toolchain")
    image = "hermes-guardian:" + version if minimum else os.environ.get("GUARDIAN_IMAGE", "hermes-guardian:" + version)
    if not re.fullmatch(r"hermes-guardian:[0-9]+\.[0-9]+(?:\.[0-9]+)?", image):
        raise ValueError("invalid validation image")
    found = run(["docker", "image", "inspect", image], check=False)
    if found.returncode:
        run(["docker", "build", "--build-arg", "GO_VERSION=" + image.split(":")[1],
             "--tag", image, str(ROOT / "scripts/guardian")], timeout=600)
    return image


def sandbox(command, network=False, timeout=900, minimum=False):
    """Only fixed trusted commands enter the container; no inherited env/secrets."""
    image = validation_image(minimum)
    name = "hermes-guardian-" + str(os.getpid())
    args = ["docker", "run", "--rm", "--name", name, "--read-only", "--cap-drop=ALL",
            "--security-opt=no-new-privileges", "--pids-limit=512", "--cpus=2", "--memory=4g",
            "--network=" + ("bridge" if network else "none"),
            "--tmpfs", "/tmp:rw,exec,size=3g", "--user", "65534:65534",
            "-v", str(WORK) + ":/repo:ro", "-w", "/repo",
            "-e", "HOME=/tmp", "-e", "GOCACHE=/tmp/go-cache", "-e", "GOPATH=/tmp/go",
            "-e", "GOTOOLCHAIN=local", "-e", "GOFLAGS=-buildvcs=false",
            "-e", "RELEASE_ALLOW_DIRTY=1", "-e", "GOMAXPROCS=2",
            image, "bash", "-c", "git config --global --add safe.directory /repo && " + command]
    log = OUT / ("minimum-go.log" if minimum else "release.log" if network else "repair-check.log")
    try:
        with log.open("w") as output:
            result = subprocess.run(args, stdout=output, stderr=subprocess.STDOUT, timeout=timeout)
        text = log.read_text(errors="replace")
        return {"passed": result.returncode == 0, "output": text[-16000:]}
    except subprocess.TimeoutExpired:
        return {"passed": False, "output": "Validation exceeded its time limit"}
    finally:
        # A timed-out docker client can otherwise leave the container running.
        subprocess.run(["docker", "rm", "-f", name], stdout=subprocess.DEVNULL,
                       stderr=subprocess.DEVNULL, timeout=30)


def regenerate():
    tool("botapi-models", "-root", WORK)
    # Compilation does not execute application init functions or tests. The
    # generator binary is built from trusted main, never from the repair patch.
    tool("api-surface", "-output", "spec/api-surface.txt", cwd=WORK)
    audit = tool("botapi-audit", "-root", WORK, "-spec", WORK / "spec/bot-api.json",
                 "-json", "-allow-gaps")
    (OUT / "audit.json").write_text(audit)
    result = tool("botapi-audit", "-root", WORK, "-spec", WORK / "spec/bot-api.json", check=False)
    (OUT / "audit.txt").write_text(result.stdout)
    return result.returncode == 0


def summary(report):
    lines = ["## Hermes maintenance", "", "**" + report["status"].replace("_", " ") + "**", ""]
    for key in ("version", "release", "branch", "pr", "message"):
        if report.get(key):
            lines.append(f"- {key.capitalize()}: {report[key]}")
    if report.get("semantic_reasons"):
        lines += ["", "Documentation changes:"] + ["- " + x for x in report["semantic_reasons"][:30]]
    if report.get("repair"):
        lines += ["", "Repair: " + report["repair"].get("summary", "")]
    if report.get("review"):
        lines += ["", "Independent review: " + report["review"].get("summary", "")]
    lines += ["", "Reports include the official delta and exact validation result. "
              "Existing applications keep their installed Hermes version."]
    return "\n".join(lines) + "\n"


def finish(report):
    write_json(OUT / "report.json", report)
    (OUT / "summary.md").write_text(summary(report))
    target = os.environ.get("GITHUB_STEP_SUMMARY")
    if target:
        with open(target, "a") as output:
            output.write(summary(report))
    print(report["status"] + ": " + report.get("message", ""))


def prepare(args):
    # Never let an unchanged or failed run accidentally publish an older bundle.
    if OUT.exists():
        shutil.rmtree(OUT)
    OUT.mkdir(exist_ok=True)
    base = run(["git", "rev-parse", "HEAD"]).strip()
    report = {"format": 1, "base": base, "status": "checking", "created": dt.datetime.now(dt.timezone.utc).isoformat()}
    if os.environ.get("GUARDIAN_MODE", "pull-request") == "paused":
        report.update(status="paused", message="Maintenance is paused in repository variables.")
        finish(report)
        return
    build_tools()
    html = source_file(args.source)
    candidate_path = OUT / "bot-api.json"
    tool("botapi-schema", "-source", html, "-output", candidate_path)
    before, candidate = read_json(ROOT / "spec/bot-api.json"), read_json(candidate_path)
    validate_candidate(before, candidate)
    diff = json.loads(tool("botapi-diff", "-after", candidate_path, "-format", "json"))
    semantics = snapshot(html.read_text())
    previous = ROOT / "spec/bot-api-semantics.json"
    reasons = semantic_changes(read_json(previous) if previous.exists() else None, semantics, diff)
    write_json(OUT / "diff.json", diff)
    (OUT / "diff.md").write_text(tool("botapi-diff", "-after", candidate_path, "-format", "markdown"))
    report.update(version=candidate["version"], semantic_reasons=reasons)
    if not diff["changed"] and not reasons:
        report.update(status="unchanged", message="Hermes matches Telegram's API and documented behavior.")
        finish(report)
        return
    identity = digest({"schema": {k: v for k, v in candidate.items() if k not in ("source_sha256", "released")},
                       "semantics": semantics})
    branch = "automation/bot-api-" + candidate["version"] + "-" + identity[:12]
    report.update(branch=branch, identity=identity, mechanical=diff["classification"] == "mechanical" and not reasons)
    if WORK.exists():
        shutil.rmtree(WORK)
    run(["git", "clone", "--no-hardlinks", str(ROOT), str(WORK)])
    run(["git", "config", "user.name", "hermes-guardian[bot]"], cwd=WORK)
    run(["git", "config", "user.email", "hermes-guardian[bot]@users.noreply.github.com"], cwd=WORK)
    run(["git", "config", "core.hooksPath", "/dev/null"], cwd=WORK)
    expected = None
    previous_state = {}
    if os.environ.get("GITHUB_REPOSITORY"):
        remote = "https://github.com/" + os.environ["GITHUB_REPOSITORY"] + ".git"
        found = run(["git", "ls-remote", "--exit-code", remote, "refs/heads/" + branch], check=False)
        if found.returncode not in (0, 2):
            raise RuntimeError("Cannot check existing update branch; refusing to overwrite work")
        if found.returncode == 0:
            expected = found.stdout.split()[0]
            run(["git", "fetch", remote, "refs/heads/" + branch], cwd=WORK)
            run(["git", "checkout", "--detach", "FETCH_HEAD"], cwd=WORK)
            # Never execute branch-supplied maintenance scripts. Check the scope
            # before merging, generation or tests; preserve all human changes.
            ancestor = run(["git", "merge-base", base, "HEAD"], cwd=WORK).strip()
            existing_paths = scope_check(WORK, ancestor)
            if any(p.endswith(".go") and not Path(p).name.startswith("zz_")
                   and p != "api/client.go" for p in existing_paths):
                report["mechanical"] = False
            # User-agent metadata is only mechanical when the rest is identical.
            original_client = run(["git", "show", ancestor + ":api/client.go"], cwd=WORK)
            strip_version = lambda s: re.sub(r'hermes-go/\d+\.\d+\.\d+', 'hermes-go/VERSION', s)
            if strip_version(original_client) != strip_version((WORK / "api/client.go").read_text()):
                report["mechanical"] = False
            merged = run(["git", "merge", "--no-edit", base], cwd=WORK, check=False)
            if merged.returncode:
                report.update(status="needs_attention", message="The existing update conflicts with main; its branch is preserved.")
                finish(report)
                return
            scope_check(WORK, base)
            state_path = WORK / STATE
            if state_path.exists():
                previous_state = read_json(state_path)
    report["expected_head"] = expected
    if previous_state.get("identity") not in (None, identity):
        raise ValueError("update branch contains a different official source identity")
    if previous_state.get("identity") == identity:
        # Cosmetic HTML changes must not rewrite the same candidate every day.
        candidate["source_sha256"] = read_json(WORK / "spec/bot-api.json")["source_sha256"]
        write_json(candidate_path, candidate)
    if (previous_state.get("last_base") == base and not args.retry
            and previous_state.get("status") == "needs_attention"
            and (previous_state.get("repair_runs", 0) >= 2 or
                 os.environ.get("GUARDIAN_REPAIR", "false") != "true" or not providers())):
        report.update(status="needs_attention", repair=previous_state.get("repair"),
                      review=previous_state.get("review"),
                      message="The prepared update still needs attention. Its branch is preserved; use Retry repair to resume.")
        finish(report)
        return
    shutil.copyfile(candidate_path, WORK / "spec/bot-api.json")
    write_json(WORK / "spec/bot-api-semantics.json", semantics)
    parity = False
    try:
        parity = regenerate()
    except RuntimeError as exc:
        report["generation_error"] = str(exc)[-4000:]
    report["parity"] = parity
    repair_runs = previous_state.get("repair_runs", 0)
    if not isinstance(repair_runs, int) or not 0 <= repair_runs <= 100:
        raise ValueError("invalid persisted repair count")
    changed_sections = {k: v for k, v in semantics["sections"].items()
                        if k in {r.split(":")[0] for r in reasons} or
                        k in {n.lower() for category in ("methods", "objects", "unions")
                              for kind in ("added", "changed") for n in diff[category][kind]}}
    evidence = {"diff": diff, "semantic_reasons": reasons,
                "official_sections": changed_sections,
                "previous_review": previous_state.get("review"),
                "previous_validation": previous_state.get("validation"),
                "audit": read_json(OUT / "audit.json") if (OUT / "audit.json").exists() else report.get("generation_error")}
    if not report["mechanical"] or not parity:
        report["repair"] = previous_state.get("repair", {"status": "disabled", "summary": "Coding repair is disabled."})
        if os.environ.get("GUARDIAN_REPAIR", "false") == "true" and (repair_runs < 2 or args.retry):
            if not shutil.which("docker"):
                raise RuntimeError("Docker is required before running coding repair")
            validation_image()
            protected_tests = set(run(["git", "ls-tree", "-r", "--name-only", base]).splitlines())
            report["repair"] = repair(WORK, evidence, lambda: sandbox("go test -shuffle=on -count=1 ./...", timeout=300),
                                      protected_tests=protected_tests)
            if report["repair"]["status"] != "unconfigured":
                repair_runs += 1
            try:
                go_files = [str(WORK / n) for n in scope_check(WORK, base) if n.endswith(".go")]
                if go_files:
                    run(["gofmt", "-w", *go_files])
                report["parity"] = regenerate()
            except RuntimeError as exc:
                report["generation_error"] = str(exc)[-4000:]
                report["parity"] = False
            run(["git", "add", "--all"], cwd=WORK)
            patch = run(["git", "diff", "--cached", base, "--", "*.go", "docs/"], cwd=WORK)
            evidence["audit"] = read_json(OUT / "audit.json") if (OUT / "audit.json").exists() else None
            evidence["final_tests"] = sandbox("go test -shuffle=on -count=1 ./...", timeout=300)
            report["review"] = review(WORK, evidence, patch[:250000]) if len(patch) <= 250000 else {
                "approved": False, "summary": "Patch exceeds automated review budget"}
        elif repair_runs >= 2:
            report["message"] = "Repair reached its two-run budget. Use Retry repair after addressing the reported issue."
    report["repair_runs"] = repair_runs
    report["status"] = "prepared"
    report.setdefault("message", "Update prepared; validation decides readiness.")
    if "review" not in report and "review" in previous_state:
        # Reviews must be refreshed after code changes, never carried as approval.
        report["review"] = dict(previous_state["review"], approved=False)
    write_json(WORK / STATE, {k: report[k] for k in ("identity", "repair_runs", "repair", "review") if k in report})
    scope_check(WORK, base)
    finish(report)


def metadata(report):
    tags = run(["git", "tag", "--list", "v*"], cwd=ROOT).splitlines()
    versions = [tuple(map(int, t[1:].split("."))) for t in tags if re.fullmatch(r"v\d+\.\d+\.\d+", t)]
    if not versions:
        raise ValueError("no stable release tag found")
    major, minor, _ = max(versions)
    version = f"{major}.{minor+1}.0"
    report["release"] = "v" + version
    changelog = WORK / "CHANGELOG.md"
    body = changelog.read_text()
    if f"## {version} -" not in body:
        entry = (f"## {version} - {dt.datetime.now(dt.timezone.utc).date()}\n\n"
                 f"- Update the typed Telegram API and documentation baseline to Bot API {report['version']}.\n"
                 "- See the Guardian update PR for the official delta, compatibility review and validation evidence.\n\n")
        body = body.replace("# Changelog\n", "# Changelog\n\n" + entry, 1)
        if entry not in body:
            raise ValueError("unrecognized changelog format")
        changelog.write_text(body)
    client = WORK / "api/client.go"
    text, count = re.subn(r'hermes-go/\d+\.\d+\.\d+', 'hermes-go/' + version, client.read_text())
    if count != 1:
        raise ValueError("unrecognized user agent version")
    client.write_text(text)
    schema = read_json(WORK / "spec/bot-api.json")
    counts = [("Methods", len(schema["methods"])),
              ("Method parameters", sum(len(m["parameters"]) for m in schema["methods"])),
              ("Concrete objects", len(schema["objects"])),
              ("Object fields", sum(len(o["fields"]) for o in schema["objects"])),
              ("Union roots", len(schema["unions"])),
              ("Union variants", sum(len(u["variants"]) for u in schema["unions"]))]
    doc = WORK / "docs/schema-parity.md"
    text = re.sub(r"API \d+\.\d+ and contains:", "API " + report["version"] + " and contains:", doc.read_text())
    for label, count in counts:
        text = re.sub(r"\| " + re.escape(label) + r" \| [\d,]+ \| 0 \|",
                      f"| {label} | {count:,} | 0 |", text)
    doc.write_text(text)


def validate(args):
    report = read_json(OUT / "report.json")
    if report["status"] != "prepared":
        return
    try:
        if report["parity"]:
            metadata(report)
            regenerate()
        scope_check(WORK, report["base"])
        result = sandbox("./scripts/release-check.sh", network=True, timeout=2100)
        if result["passed"]:
            minimum = sandbox("./scripts/check-generated.sh && go vet ./... && go test -shuffle=on -count=1 ./...",
                              timeout=600, minimum=True)
            result = {"passed": minimum["passed"], "output": result["output"] + "\nGo 1.25 verification:\n" + minimum["output"][-4000:]}
        report["validation"] = result
        report["status"] = "ready" if result["passed"] and report["parity"] else "needs_attention"
        if report["status"] == "ready" and not report["mechanical"]:
            report["status"] = "reviewed" if report.get("review", {}).get("approved") else "needs_attention"
        report["message"] = {
            "ready": "Generated update passed the complete release gates.",
            "reviewed": "Repaired update passed release gates and a separate AI review; maintainer release approval is required.",
            "needs_attention": "Update is preserved as a draft. See validation and repair evidence for the remaining work."
        }[report["status"]]
    except (RuntimeError, ValueError, OSError) as exc:
        report.update(status="needs_attention", message=str(exc)[-6000:])
    state = {k: report[k] for k in ("identity", "repair_runs", "repair", "review", "status") if k in report}
    state["last_base"] = report["base"]
    if report["status"] == "needs_attention":
        state["validation"] = report.get("validation")
    write_json(WORK / STATE, state)
    paths = scope_check(WORK, report["base"])
    changes = [{"path": p, "content": (WORK / p).read_text(), "mode": "100644"} for p in paths]
    bundle = {"report": report, "changes": changes}
    if len(canonical(bundle).encode()) > 6_000_000:
        raise ValueError("update bundle exceeds publication limit")
    write_json(OUT / "bundle.json", bundle)
    finish(report)


def release_check(args):
    OUT.mkdir(exist_ok=True)
    if WORK.exists():
        shutil.rmtree(WORK)
    run(["git", "clone", "--no-hardlinks", str(ROOT), str(WORK)])
    sha = run(["git", "rev-parse", "HEAD"]).strip()
    changelog = (ROOT / "CHANGELOG.md").read_text()
    match = re.search(r"^## (\d+\.\d+\.\d+) -", changelog, re.MULTILINE)
    if not match:
        raise ValueError("no version in changelog")
    result = sandbox("./scripts/release-check.sh", network=True, timeout=2100)
    if result["passed"]:
        minimum = sandbox("./scripts/check-generated.sh && go vet ./... && go test -shuffle=on -count=1 ./...",
                          timeout=600, minimum=True)
        result = {"passed": minimum["passed"], "output": result["output"] + "\nGo 1.25 verification:\n" + minimum["output"][-4000:]}
    if not result["passed"]:
        raise RuntimeError("Release validation failed:\n" + result["output"])
    report = {"status": "release_verified", "base": sha, "release": "v" + match[1],
              "message": "Current main passed release validation.", "validation": result}
    write_json(OUT / "release.json", report)
    finish(report)


def sandbox_smoke(args):
    OUT.mkdir(exist_ok=True)
    if WORK.exists():
        shutil.rmtree(WORK)
    run(["git", "clone", "--no-hardlinks", str(ROOT), str(WORK)])
    result = sandbox("./scripts/check-generated.sh && "
                     "python3 -m unittest discover -s scripts/guardian -p 'test_*.py' && "
                     "go test ./internal/botapi ./api", timeout=600)
    print(result["output"])
    if not result["passed"]:
        raise RuntimeError("Read-only validation sandbox smoke test failed")


def release(args):
    report = read_json(OUT / "release.json")
    route = repo_route()
    if report.get("status") != "release_verified" or not report.get("validation", {}).get("passed"):
        raise ValueError("missing release validation")
    sha = api(route + "/git/ref/heads/main")["object"]["sha"]
    if sha != report["base"]:
        raise ValueError("main changed after release validation; rerun")
    publish_release(route, report["release"], sha, summary(report))
    finish(dict(report, status="released", message="Verified release published successfully."))


def publish(args):
    bundle = read_json(OUT / "bundle.json")
    report, changes = bundle["report"], bundle["changes"]
    route = repo_route()
    if report["status"] not in ("ready", "reviewed", "needs_attention"):
        raise ValueError("invalid candidate status")
    if not re.fullmatch(r"automation/bot-api-\d+\.\d+-[a-f0-9]{12}", report["branch"]):
        raise ValueError("invalid update branch")
    for change in changes:
        name = change["path"]
        if name not in GENERATED and not allowed(name):
            raise ValueError("publication includes a protected path")
        if change["mode"] != "100644" or not isinstance(change["content"], str):
            raise ValueError("publication must contain regular text files")
    main = api(route + "/git/ref/heads/main")["object"]["sha"]
    if main != report["base"]:
        raise ValueError("main changed during validation; rerun Guardian against the new main")
    existing = api(route + "/git/ref/heads/" + report["branch"], missing=True)
    expected = report.get("expected_head")
    if (existing["object"]["sha"] if existing else None) != expected:
        raise ValueError("update branch changed during validation; rerun to preserve that work")
    base_commit = api(route + "/git/commits/" + main)
    tree = api(route + "/git/trees", "POST", {"base_tree": base_commit["tree"]["sha"],
        "tree": [{"path": c["path"], "mode": c["mode"], "type": "blob", "content": c["content"]} for c in changes]})
    old_commit = api(route + "/git/commits/" + expected) if expected else None
    if old_commit and old_commit["tree"]["sha"] == tree["sha"]:
        commit = {"sha": expected}
    else:
        parents = [expected] if expected else [main]
        if expected and expected != main:
            parents.append(main)
        commit = api(route + "/git/commits", "POST", {
            "message": "Maintain Telegram Bot API " + report["version"], "tree": tree["sha"], "parents": parents})
        if existing:
            api(route + "/git/refs/heads/" + report["branch"], "PATCH", {"sha": commit["sha"], "force": False})
        else:
            api(route + "/git/refs", "POST", {"ref": "refs/heads/" + report["branch"], "sha": commit["sha"]})
    owner = os.environ["GITHUB_REPOSITORY"].split("/")[0]
    prs = api(route + "/pulls?state=open&head=" + owner + ":" + report["branch"] + "&base=main")
    body = summary(report) + "\n" + (OUT / "diff.md").read_text() + "\n### Validation\n\n```text\n" + report.get("validation", {}).get("output", "No complete validation")[-12000:] + "\n```\n"
    body += f"\nSource identity: `{report['identity']}`\n\nValidated base: `{main}`\n"
    data = {"title": "Update Hermes for Telegram Bot API " + report["version"], "body": body}
    if prs:
        pr = api(route + "/pulls/" + str(prs[0]["number"]), "PATCH", data)
    else:
        pr = api(route + "/pulls", "POST", dict(data, head=report["branch"], base="main", draft=report["status"] != "ready"))
    report.update(pr=pr["html_url"], head=commit["sha"])
    write_json(OUT / "report.json", report)
    # Existing PR drafts are never silently promoted. An owner may have deliberately
    # paused one. Repository branch protection is honored by the normal merge API.
    if (os.environ.get("GUARDIAN_MODE") == "automatic" and report["status"] == "ready"
            and report["mechanical"] and not pr["draft"]):
        latest = api(route + "/git/ref/heads/main")["object"]["sha"]
        if latest != main:
            raise ValueError("main moved before merge; rerun validation")
        try:
            merged = api(route + "/pulls/" + str(pr["number"]) + "/merge", "PUT",
                         {"sha": commit["sha"], "merge_method": "squash"})
        except RuntimeError as exc:
            if any(code in str(exc) for code in ("HTTP 405", "HTTP 409", "HTTP 422")):
                report.update(status="awaiting_merge", message="GitHub requires checks or approvals before merging. "
                              "The prepared PR is preserved; the next check will retry without replacing its commit.")
                finish(report)
                return
            raise
        if merged.get("merged") and os.environ.get("GUARDIAN_RELEASE") == "true":
            publish_release(route, report["release"], merged["sha"], body)
    finish(report)
    print(pr["html_url"])


def publish_release(route, tag, sha, notes):
    if not re.fullmatch(r"v\d+\.\d+\.\d+", tag) or not re.fullmatch(r"[a-f0-9]{40}", sha):
        raise ValueError("invalid release target")
    existing = api(route + "/git/ref/tags/" + tag, missing=True)
    target = existing["object"] if existing else None
    for _ in range(5):
        if not target or target.get("type") != "tag":
            break
        target = api(route + "/git/tags/" + target["sha"])["object"]
    if target and target["sha"] != sha:
        raise ValueError("release tag already exists at a different commit; never move published versions")
    if not existing:
        api(route + "/git/refs", "POST", {"ref": "refs/tags/" + tag, "sha": sha})
    published = api(route + "/releases/tags/" + tag, missing=True)
    if not published:
        api(route + "/releases", "POST", {"tag_name": tag, "target_commitish": sha,
            "name": "Hermes " + tag, "body": notes, "draft": False, "prerelease": False})
    elif published.get("draft") or published.get("prerelease"):
        api(route + "/releases/" + str(published["id"]), "PATCH", {"draft": False, "prerelease": False})


def health(args):
    OUT.mkdir(exist_ok=True)
    if os.environ.get("GUARDIAN_MODE") == "paused":
        finish({"status": "paused", "message": "Maintenance is intentionally paused; no recovery dispatch was made."})
        return
    route = repo_route()
    workflow = api(route + "/actions/workflows/guardian.yml")
    runs = api(route + "/actions/workflows/guardian.yml/runs?status=success&per_page=1")["workflow_runs"]
    now = dt.datetime.now(dt.timezone.utc)
    stale = not runs or now - dt.datetime.fromisoformat(runs[0]["created_at"].replace("Z", "+00:00")) > dt.timedelta(hours=48)
    report = {"status": "needs_attention" if stale else "healthy",
              "message": "Guardian has not completed successfully within 48 hours." if stale else "Guardian completed within the last 48 hours."}
    if workflow["state"] != "active" and os.environ.get("GUARDIAN_MODE") != "paused":
        api(route + "/actions/workflows/guardian.yml/enable", "PUT")
        report["message"] += " Re-enabled its schedule."
    if stale and os.environ.get("GUARDIAN_MODE") != "paused":
        api(route + "/actions/workflows/guardian.yml/dispatches", "POST", {"ref": "main"})
    OUT.mkdir(exist_ok=True)
    finish(report)
    if stale:
        status_issue(report)


def status_issue(report):
    route = repo_route()
    title = "Hermes maintenance status"
    # Update one persistent issue; avoid a new issue/comment on every daily run.
    issues = api(route + "/issues?state=open&creator=github-actions%5Bbot%5D&per_page=100")
    issue = next((i for i in issues if i["title"] == title and "pull_request" not in i), None)
    body = summary(report)
    run_id = os.environ.get("GITHUB_RUN_ID")
    if run_id:
        body += "\n[Open maintenance run](https://github.com/" + os.environ["GITHUB_REPOSITORY"] + "/actions/runs/" + run_id + ")\n"
    data = {"title": title, "body": body}
    if issue:
        api(route + "/issues/" + str(issue["number"]), "PATCH", data)
    else:
        api(route + "/issues", "POST", data)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("prepare", "validate", "publish", "health", "status", "release-check", "release", "sandbox-smoke"))
    parser.add_argument("--source", help="local official HTML snapshot for offline replay")
    parser.add_argument("--retry", action="store_true", help="owner-requested extra bounded repair attempt")
    args = parser.parse_args()
    try:
        if args.command == "status":
            status_issue(read_json(OUT / "report.json"))
        else:
            globals()[args.command.replace("-", "_")](args)
    except (RuntimeError, ValueError, KeyError, OSError, subprocess.TimeoutExpired) as exc:
        OUT.mkdir(exist_ok=True)
        report = read_json(OUT / "report.json") if (OUT / "report.json").exists() else {}
        report.update(status="needs_attention", message=str(exc)[:6000])
        finish(report)
        raise SystemExit(1)


if __name__ == "__main__":
    main()
