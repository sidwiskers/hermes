import copy
import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import main
import protocol
import repair


def document(body="Returns True on success.", extra="", title="sendMessage"):
    return f'''<h4><a name="sendmessage"></a>{title}</h4><p>{body}</p>
    <table><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr>
    <tr><td>text</td><td>String</td><td>Yes</td><td>Message text</td></tr>{extra}</table>
    <h4><a name="update"></a>Update</h4><p>An update.</p>'''


def diff(kind="unchanged"):
    return {"classification": kind, "objects": {"added": [], "changed": []}}


class ProtocolTests(unittest.TestCase):
    def test_ignores_html_style_and_whitespace(self):
        a = protocol.snapshot(document())
        b = protocol.snapshot(document().replace("Returns True", "Returns  <b>True</b>"))
        self.assertEqual(a, b)

    def test_return_contract_detected_without_version_change(self):
        a = protocol.snapshot(document())
        b = protocol.snapshot(document("Returns a Message on success."))
        self.assertIn("sendmessage", protocol.semantic_changes(a, b, diff())[0])

    def test_description_limit_detected(self):
        a = protocol.snapshot(document())
        b = protocol.snapshot(document().replace("Message text", "Message text, 1–200 characters"))
        self.assertTrue(protocol.semantic_changes(a, b, diff()))

    def test_link_destination_detected(self):
        a = protocol.snapshot(document('<a href="tg://first">link</a>'))
        b = protocol.snapshot(document('<a href="tg://second">link</a>'))
        self.assertTrue(protocol.semantic_changes(a, b, diff()))

    def test_missing_baseline_needs_review(self):
        self.assertTrue(protocol.semantic_changes(None, protocol.snapshot(document()), diff()))

    def test_incomplete_document_fails(self):
        with self.assertRaises(ValueError):
            protocol.snapshot("<html>Temporary error</html>")

    def test_duplicate_heading_fails(self):
        with self.assertRaises(ValueError):
            protocol.snapshot(document() + '<h4><a name="update"></a>Update</h4>')

    def test_new_optional_object_row_is_mechanical(self):
        a = protocol.snapshot(document())
        b = copy.deepcopy(a)
        b["sections"]["update"]["rows"].append(["new_field", "String", "Optional. Value"])
        d = diff("mechanical")
        d["objects"]["changed"] = ["Update"]
        self.assertEqual(protocol.semantic_changes(a, b, d), [])

    def test_changed_existing_row_requires_review(self):
        a = protocol.snapshot(document())
        b = copy.deepcopy(a)
        b["sections"]["sendmessage"]["rows"][1][-1] = "Different rules"
        d = diff("mechanical")
        d["objects"]["changed"] = ["sendMessage"]
        self.assertTrue(protocol.semantic_changes(a, b, d))

    def test_unknown_new_section_requires_review(self):
        a = protocol.snapshot(document())
        b = copy.deepcopy(a)
        b["sections"]["new-rule"] = {"title": "Rule", "body": "Changed", "rows": []}
        self.assertTrue(protocol.semantic_changes(a, b, diff("mechanical")))

    def test_source_guard_rejects_rollback_and_schema_loss(self):
        before = {"version": "10.3", "methods": [{"name": str(i)} for i in range(100)],
                  "objects": [{"name": "X"}], "unions": [], "source": "https://core.telegram.org/bots/api"}
        after = copy.deepcopy(before)
        protocol.validate_candidate(before, after)
        after["version"] = "10.2"
        with self.assertRaises(ValueError):
            protocol.validate_candidate(before, after)
        after["version"] = "10.4"
        after["methods"] = after["methods"][:89]
        with self.assertRaises(ValueError):
            protocol.validate_candidate(before, after)

    def test_identity_ignores_dictionary_order(self):
        self.assertEqual(protocol.digest({"a": 1, "b": 2}), protocol.digest({"b": 2, "a": 1}))


class RepairTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / "api").mkdir()
        (self.root / "api/method.go").write_text("package api\n// old\n")

    def test_exact_edit_and_new_regression_test(self):
        repair.apply_edits(self.root, [{"path": "api/method.go", "old": "// old", "new": "// new"},
            {"path": "api/new_test.go", "content": "package api\n"}])
        self.assertIn("// new", (self.root / "api/method.go").read_text())

    def test_batch_failure_is_atomic(self):
        with self.assertRaises(ValueError):
            repair.apply_edits(self.root, [{"path": "api/method.go", "old": "old", "new": "new"},
                {"path": "api/method.go", "old": "missing", "new": "new"}])
        self.assertIn("old", (self.root / "api/method.go").read_text())

    def test_agent_can_correct_its_own_new_tests(self):
        name = "api/new_test.go"
        repair.apply_edits(self.root, [{"path": name, "content": "package api\n// first"}], set())
        repair.apply_edits(self.root, [{"path": name, "old": "first", "new": "fixed"}], set())
        self.assertIn("fixed", (self.root / name).read_text())

    def test_existing_tests_and_control_files_are_protected(self):
        for name in ("api/existing_test.go", "scripts/verify.sh", ".github/workflows/ci.yml",
                     "spec/bot-api.json", "go.mod", "internal/cmd/botapi-audit/main.go",
                     "types/zz_botapi_generated.go"):
            target = self.root / name
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text("old")
            with self.subTest(name=name), self.assertRaises(ValueError):
                repair.apply_edits(self.root, [{"path": name, "old": "old", "new": "new"}])

    def test_path_traversal_and_symlinks_rejected(self):
        for name in ("../escape.go", "/tmp/escape.go", ".git/config"):
            with self.subTest(name=name), self.assertRaises(ValueError):
                repair.safe_file(self.root, name)
        (self.root / "api/link.go").symlink_to(self.root / "api/method.go")
        with self.assertRaises(ValueError):
            repair.apply_edits(self.root, [{"path": "api/link.go", "old": "old", "new": "new"}])

    def test_ambiguous_replacement_rejected(self):
        (self.root / "api/method.go").write_text("same same")
        with self.assertRaises(ValueError):
            repair.apply_edits(self.root, [{"path": "api/method.go", "old": "same", "new": "new"}])

    def test_provider_config_and_missing_keys(self):
        good = [{"url": "https://provider.example/v1/chat/completions", "model": "code", "key_env": "GUARDIAN_KEY_1"}]
        with patch.dict(os.environ, {"GUARDIAN_PROVIDERS": json.dumps(good)}, clear=True):
            self.assertEqual(repair.providers(), [])
        with patch.dict(os.environ, {"GUARDIAN_PROVIDERS": json.dumps(good), "GUARDIAN_KEY_1": "example"}, clear=True):
            self.assertEqual(len(repair.providers()), 1)
        for url in ("http://example.com", "https://user:password@example.com", "https://example.com?key=x"):
            bad = [dict(good[0], url=url)]
            with patch.dict(os.environ, {"GUARDIAN_PROVIDERS": json.dumps(bad)}, clear=True), self.assertRaises(ValueError):
                repair.providers()

    def test_provider_cannot_redirect_authorization(self):
        with self.assertRaises(ValueError):
            repair.NoRedirect().redirect_request(None, None, None, None, None)

    def test_fallback_and_bounded_loop(self):
        with patch.object(repair, "providers", return_value=[{"name": "first"}, {"name": "second"}]), \
                patch.object(repair, "completion", side_effect=[OSError("private error"), {"action": "done", "summary": "done"}]) as call:
            result = repair.repair(self.root, {}, lambda: {}, max_turns=2)
            self.assertEqual(result["status"], "proposed")
            self.assertEqual(call.call_count, 2)
        with patch.object(repair, "providers", return_value=[{}]), \
                patch.object(repair, "completion", return_value={"action": "read", "path": "api/method.go"}) as call:
            result = repair.repair(self.root, {}, lambda: {}, max_turns=2)
            self.assertEqual(result["status"], "exhausted")
            self.assertEqual(call.call_count, 2)

    def test_unavailable_errors_do_not_echo_provider_response(self):
        with patch.object(repair, "providers", return_value=[{}]), \
                patch.object(repair, "completion", side_effect=OSError("PRIVATE_KEY")):
            result = repair.repair(self.root, {}, lambda: {})
            self.assertNotIn("PRIVATE_KEY", json.dumps(result))


class ControlPlaneTests(unittest.TestCase):
    def test_paused_health_never_calls_github(self):
        with tempfile.TemporaryDirectory() as tmp, patch.object(main, "OUT", Path(tmp)), \
                patch.dict(os.environ, {"GUARDIAN_MODE": "paused"}), \
                patch.object(main, "api") as call, patch.object(main, "finish"):
            main.health(None)
            call.assert_not_called()

    def test_new_check_discards_stale_publication_bundle(self):
        import argparse
        with tempfile.TemporaryDirectory() as tmp, patch.object(main, "OUT", Path(tmp)), \
                patch.dict(os.environ, {"GUARDIAN_MODE": "paused"}), \
                patch.object(main, "run", return_value="a" * 40), patch.object(main, "finish"):
            (Path(tmp) / "bundle.json").write_text('{"stale":true}')
            main.prepare(argparse.Namespace(source=None, retry=False))
            self.assertFalse((Path(tmp) / "bundle.json").exists())

    def test_subprocess_environment_excludes_credentials(self):
        with patch.dict(os.environ, {"GH_TOKEN": "private", "GUARDIAN_KEY_1": "private", "ACTIONS_RUNTIME_TOKEN": "private"}), \
                patch.object(subprocess, "run", return_value=subprocess.CompletedProcess([], 0, "ok")) as call:
            main.run(["git", "status"])
            env = call.call_args.kwargs["env"]
            for key in ("GH_TOKEN", "GUARDIAN_KEY_1", "ACTIONS_RUNTIME_TOKEN"):
                self.assertNotIn(key, env)

    def test_sandbox_does_not_inherit_host_credentials_or_writable_source(self):
        with tempfile.TemporaryDirectory() as tmp, patch.object(main, "OUT", Path(tmp)), \
                patch.object(main, "validation_image", return_value="hermes-guardian:1.26.6"), \
                patch.object(subprocess, "run", return_value=subprocess.CompletedProcess([], 0)) as call:
            main.sandbox("go test ./...")
            args = call.call_args_list[0].args[0]
            self.assertIn("--network=none", args)
            self.assertIn("--read-only", args)
            self.assertIn("--cap-drop=ALL", args)
            self.assertIn(str(main.WORK) + ":/repo:ro", args)
            self.assertNotIn("GH_TOKEN", " ".join(args))

    def test_sandbox_timeout_removes_container(self):
        with tempfile.TemporaryDirectory() as tmp, patch.object(main, "OUT", Path(tmp)), \
                patch.object(main, "validation_image", return_value="hermes-guardian:1.26.6"), \
                patch.object(subprocess, "run", side_effect=[subprocess.TimeoutExpired("docker", 1),
                                                            subprocess.CompletedProcess([], 0)]) as call:
            result = main.sandbox("go test ./...", timeout=1)
            self.assertFalse(result["passed"])
            self.assertEqual(call.call_args.args[0][:3], ["docker", "rm", "-f"])

    def test_published_tag_never_moves(self):
        with patch.object(main, "api", return_value={"object": {"sha": "a" * 40}}) as call:
            with self.assertRaises(ValueError):
                main.publish_release("/repos/a/b", "v1.4.0", "b" * 40, "notes")
            self.assertEqual(call.call_count, 1)

    def test_release_retry_is_idempotent(self):
        with patch.object(main, "api", side_effect=[{"object": {"sha": "a" * 40}}, {"id": 1}]) as call:
            main.publish_release("/repos/a/b", "v1.4.0", "a" * 40, "notes")
            self.assertEqual(call.call_count, 2)

    def test_partial_release_recovers_existing_tag(self):
        with patch.object(main, "api", side_effect=[{"object": {"sha": "a" * 40}}, None, {"id": 1}]) as call:
            main.publish_release("/repos/a/b", "v1.4.0", "a" * 40, "notes")
            self.assertEqual(call.call_args.args[1], "POST")
            self.assertTrue(call.call_args.args[0].endswith("/releases"))

    def test_publication_rejects_concurrent_main_change(self):
        bundle = {"report": {"status": "ready", "branch": "automation/bot-api-10.4-" + "a" * 12,
                             "base": "a" * 40}, "changes": []}
        with patch.object(main, "read_json", return_value=bundle), \
                patch.object(main, "repo_route", return_value="/repos/a/b"), \
                patch.object(main, "api", return_value={"object": {"sha": "b" * 40}}) as call:
            with self.assertRaises(ValueError):
                main.publish(None)
            self.assertEqual(call.call_count, 1)

    def test_publication_rejects_concurrent_repair_change(self):
        bundle = {"report": {"status": "ready", "branch": "automation/bot-api-10.4-" + "a" * 12,
                             "base": "a" * 40, "expected_head": "b" * 40}, "changes": []}
        with patch.object(main, "read_json", return_value=bundle), \
                patch.object(main, "repo_route", return_value="/repos/a/b"), \
                patch.object(main, "api", side_effect=[{"object": {"sha": "a" * 40}},
                                                       {"object": {"sha": "c" * 40}}]) as call:
            with self.assertRaises(ValueError):
                main.publish(None)
            self.assertEqual(call.call_count, 2)

    def test_scope_check_preserves_existing_tests(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            def git(*args):
                return subprocess.run(["git", *args], cwd=root, check=True, capture_output=True, text=True).stdout
            git("init")
            git("config", "user.name", "Test")
            git("config", "user.email", "test@example.test")
            (root / "api").mkdir()
            (root / "api/existing_test.go").write_text("original")
            git("add", ".")
            git("commit", "-m", "baseline")
            base = git("rev-parse", "HEAD").strip()
            (root / "api/existing_test.go").write_text("weakened")
            with self.assertRaises(ValueError):
                main.scope_check(root, base)


if __name__ == "__main__":
    unittest.main()
