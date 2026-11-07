#!/usr/bin/env python3
"""The maintainer tools' tests: each fails on a planted break, and passes clean.

Defined by docs/decisions/0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md
D3; for hosts.py and hostrun.py,
docs/decisions/0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md
D4; and for prune_releases.py,
docs/decisions/0026-MADR-keep-five-stable-releases.md D5. Run by
`make tools-test`, which `make verify` runs.
"""
from __future__ import annotations

import io
import json
import re
import shutil
import subprocess
import sys
import tempfile
import threading
import time
import unittest
import urllib.error
import urllib.request
from contextlib import redirect_stderr
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import fakeprovider  # noqa: E402
import hosts  # noqa: E402
import prune_releases  # noqa: E402

MADE_UP_KEY = "made-up-key-0016-must-not-be-logged"


class FakeProviderTest(unittest.TestCase):
    """tools/fakeprovider.py, in-process."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        # Cleanups run last-in first-out, so this one, added first, runs after
        # every server a test starts has stopped and released the log.
        self.addCleanup(self.tmp.cleanup)
        self.log = Path(self.tmp.name) / "requests.log"

    def start(self, **kwargs: object) -> fakeprovider.FakeProvider:
        server = fakeprovider.serve(log=self.log, **kwargs)
        self.addCleanup(server.server_close)
        self.addCleanup(server.shutdown)
        return server

    def request(self, server: fakeprovider.FakeProvider, method: str, path: str,
                body: dict | None = None) -> tuple[int, dict, dict]:
        """The status, the headers and the decoded body of one request."""
        data = None if body is None else json.dumps(body).encode()
        req = urllib.request.Request(f"http://127.0.0.1:{server.port}{path}", data=data, method=method,
                                     headers={"Authorization": f"Bearer {MADE_UP_KEY}",
                                              "Content-Type": "application/json"})
        try:
            with urllib.request.urlopen(req, timeout=5) as resp:
                return resp.status, dict(resp.headers), json.loads(resp.read())
        except urllib.error.HTTPError as err:
            return err.code, dict(err.headers), json.loads(err.read())

    def test_models_lists_one_model(self) -> None:
        status, _, body = self.request(self.start(), "GET", "/v1/models")
        self.assertEqual(status, 200)
        self.assertEqual([m["id"] for m in body["data"]], [fakeprovider.MODEL])

    def test_chat_completions_fail_first_then_answer(self) -> None:
        server = self.start(fail=1, retry_after=1, message="feat: x")
        status, headers, _ = self.request(server, "POST", "/v1/chat/completions", {"messages": []})
        self.assertEqual(status, 429)
        self.assertEqual(headers.get("Retry-After"), "1")
        status, _, body = self.request(server, "POST", "/v1/chat/completions", {"messages": []})
        self.assertEqual(status, 200)
        self.assertEqual(body["object"], "chat.completion")
        self.assertEqual(body["choices"][0]["message"]["content"], "feat: x")

    def test_responses_answer_in_its_shape(self) -> None:
        server = self.start(fail=1, message="fix: y")
        self.assertEqual(self.request(server, "POST", "/v1/responses", {"input": []})[0], 429)
        status, _, body = self.request(server, "POST", "/v1/responses", {"input": []})
        self.assertEqual(status, 200)
        content = body["output"][0]["content"][0]
        self.assertEqual((content["type"], content["text"]), ("output_text", "fix: y"))
        self.assertNotIn("choices", body)

    def test_status_and_count_are_settable(self) -> None:
        server = self.start(fail=2, status=503)
        got = [self.request(server, "POST", "/v1/chat/completions", {})[0] for _ in range(3)]
        self.assertEqual(got, [503, 503, 200])

    def test_log_holds_method_path_status_and_no_header(self) -> None:
        server = self.start(fail=1)
        self.request(server, "GET", "/v1/models")
        self.request(server, "POST", "/v1/chat/completions", {"messages": []})
        self.request(server, "POST", "/v1/chat/completions", {"messages": []})
        log = self.log.read_text(encoding="utf-8")
        self.assertEqual(log.splitlines(), ["GET /v1/models -> 200", "POST /v1/chat/completions -> 429",
                                            "POST /v1/chat/completions -> 200"])
        self.assertNotIn(MADE_UP_KEY, log)
        self.assertNotIn("Authorization", log)

    def test_log_line_is_written_before_the_answer_arrives(self) -> None:
        # A slow log write must hold the answer back, not trail it
        # (0017-PLAN, Deviation 1).
        server = self.start()
        record = server.record

        def slow_record(*args: object) -> bool:
            time.sleep(0.3)
            return record(*args)

        server.record = slow_record
        self.assertEqual(self.request(server, "GET", "/v1/models")[0], 200)
        self.assertEqual(self.log.read_text(encoding="utf-8").splitlines(), ["GET /v1/models -> 200"])

    def test_max_requests_stops_the_server(self) -> None:
        server = fakeprovider.FakeProvider(0, 429, 1, "m", self.log, max_requests=1)
        self.addCleanup(server.server_close)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        self.assertEqual(self.request(server, "GET", "/v1/models")[0], 200)
        thread.join(timeout=5)
        self.assertFalse(thread.is_alive(), "the server kept serving after --max-requests")

    def test_binds_loopback_only(self) -> None:
        self.assertEqual(self.start().server_address[0], "127.0.0.1")

    def test_bad_argument_exits_two(self) -> None:
        self.assertEqual(fakeprovider.main(["--status", "200"]), 2)
        self.assertEqual(fakeprovider.main(["--fail", "-1"]), 2)
        self.assertEqual(fakeprovider.main(["--no-such-flag"]), 2)

    def test_main_writes_its_port_and_stops_after_max_requests(self) -> None:
        port_file = Path(self.tmp.name) / "port"
        thread = threading.Thread(
            target=fakeprovider.main, args=(["--port-file", str(port_file), "--max-requests", "1",
                                             "--log", str(self.log)],), daemon=True)
        thread.start()
        deadline = time.monotonic() + 5
        while not (port_file.exists() and port_file.read_text()) and time.monotonic() < deadline:
            time.sleep(0.05)
        port = int(port_file.read_text())
        with urllib.request.urlopen(f"http://127.0.0.1:{port}/v1/models", timeout=5) as resp:
            self.assertEqual(resp.status, 200)
        thread.join(timeout=5)
        self.assertFalse(thread.is_alive(), "main kept serving after --max-requests")


#: A fake ssh: records its argv, then runs the program on its stdin through
#: this Python, as the host's `python3 -` would.
FAKE_SSH = """import json, subprocess, sys
with open(sys.argv[1], "a", encoding="utf-8") as f:
    f.write(json.dumps(sys.argv[2:]) + "\\n")
sys.exit(subprocess.run([sys.executable, "-"], stdin=sys.stdin).returncode)
"""

#: A host that cannot be reached: exits as ssh does, without reading stdin.
UNREACHABLE_SSH = "import sys\nsys.exit(255)\n"

#: The command a local host runs in its clone: the tag is there, so
#: `git describe` works, and so is the committed file.
CHECK_CLONE = ("import pathlib, subprocess, sys\n"
               "d = subprocess.run(['git', 'describe', '--tags'], capture_output=True, text=True).stdout.strip()\n"
               "print('describe=' + d)\n"
               "sys.exit(0 if d == 'v0.0.1' and pathlib.Path('hello.txt').is_file() else 9)\n")


class HostsTest(unittest.TestCase):
    """tools/hosts.py, with tools/hostrun.py run for real through a local
    Python, against a throwaway repository."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.dir = Path(self.tmp.name)
        self.repo = self.dir / "repo"
        self.repo.mkdir()
        # No hooks: the global prepare-commit-msg hook must not run here.
        for args in (["init", "-q"], ["config", "user.name", "test"], ["config", "user.email", "test@example.invalid"],
                     ["config", "commit.gpgsign", "false"], ["config", "tag.gpgsign", "false"],
                     ["config", "core.hooksPath", str(self.dir / "no-hooks")]):
            self.git(*args)
        (self.repo / "hello.txt").write_text("hello\n", encoding="utf-8")
        self.git("add", "hello.txt")
        self.git("commit", "-q", "-m", "test: one commit")
        self.git("tag", "v0.0.1")
        self.record = self.dir / "ssh-argv.jsonl"

    def git(self, *args: str) -> None:
        subprocess.run(["git", "-C", str(self.repo), *args], check=True, capture_output=True)

    def config(self, hosts_list: list[dict], identifiers: list[str] | None = None) -> Path:
        path = self.dir / "hosts.json"
        path.write_text(json.dumps({"hosts": hosts_list, "identifiers": identifiers or []}), encoding="utf-8")
        return path

    def fake(self, source: str) -> tuple[str, ...]:
        script = self.dir / f"fake-{abs(hash(source))}.py"
        script.write_text(source, encoding="utf-8")
        return (sys.executable, str(script), str(self.record))

    def run_main(self, args: list[str], ssh: tuple[str, ...] = hosts.SSH) -> tuple[int, str, str]:
        out, err = io.StringIO(), io.StringIO()
        with redirect_stderr(err):
            code = hosts.main(args, repo=self.repo, ssh=ssh, out=out)
        # The tool keeps its logs for the person who ran it; a test does not.
        found = re.search(r"logs in (.+)$", out.getvalue(), re.MULTILINE)
        if found:
            self.addCleanup(shutil.rmtree, found.group(1), ignore_errors=True)
        return code, out.getvalue(), err.getvalue()

    def log_dir(self, output: str) -> Path:
        return Path(re.search(r"logs in (.+)$", output, re.MULTILINE).group(1))

    # --- config ---

    def assert_config_error(self, hosts_list: list[dict] | None, needle: str) -> None:
        path = self.dir / "absent.json" if hosts_list is None else self.config(hosts_list)
        code, _, err = self.run_main(["--config", str(path)])
        self.assertEqual(code, 2, err)
        self.assertIn(needle, err)

    def test_no_config_exits_two_and_names_the_path(self) -> None:
        self.assert_config_error(None, str(self.dir / "absent.json"))

    def test_bad_kind_exits_two(self) -> None:
        self.assert_config_error([{"label": "a", "kind": "bsd", "ssh": "h"}], "hosts[0].kind")

    def test_duplicate_label_exits_two(self) -> None:
        self.assert_config_error([{"label": "a", "kind": "local"}, {"label": "a", "kind": "local"}],
                                 "'a' is used twice")

    def test_wsl_on_a_windows_host_exits_two(self) -> None:
        self.assert_config_error([{"label": "w", "kind": "windows", "ssh": "h", "wsl": "Ubuntu-24.04"}],
                                 "hosts[0].wsl")

    def test_non_list_command_exits_two(self) -> None:
        self.assert_config_error([{"label": "a", "kind": "local", "command": "go test ./..."}],
                                 "hosts[0].command")

    def test_default_commands_are_ci_jobs(self) -> None:
        found, _ = hosts.load_config(self.config([{"label": "l", "kind": "linux", "ssh": "h"},
                                                  {"label": "w", "kind": "windows", "ssh": "h"},
                                                  {"label": "m", "kind": "local"}]))
        self.assertEqual([h.command for h in found],
                         [("make", "verify"), ("go", "test", "./..."), ("go", "test", "./...")])

    # --- transports (F4) ---

    def test_transports_are_exact_argument_lists(self) -> None:
        ssh = ["ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=5", "-T"]
        self.assertEqual(hosts.transport(hosts.Host("l", "linux", "dest", None, ())),
                         [*ssh, "dest", "python3", "-"])
        self.assertEqual(hosts.transport(hosts.Host("u", "linux", "win", "Ubuntu-24.04", ())),
                         [*ssh, "win", "wsl", "-d", "Ubuntu-24.04", "--", "bash", "-lc", "'python3 -'"])
        self.assertEqual(hosts.transport(hosts.Host("m", "local", None, None, ())), [sys.executable, "-"])

    # --- what is tested ---

    def test_dirty_tree_exits_two_and_starts_no_transport(self) -> None:
        (self.repo / "hello.txt").write_text("changed\n", encoding="utf-8")
        path = self.config([{"label": "l", "kind": "linux", "ssh": "dest"}])
        code, _, err = self.run_main(["--config", str(path)], ssh=self.fake(FAKE_SSH))
        self.assertEqual(code, 2)
        self.assertIn("commit first", err)
        self.assertFalse(self.record.exists(), "a transport started for a dirty tree")

    # --- redaction ---

    def test_redaction(self) -> None:
        ids = [re.compile(re.escape("secret-host"), re.IGNORECASE)]
        self.assertEqual(hosts.redact("/Users/alice/x /home/bob/y", ids), "/Users/<user>/x /home/<user>/y")
        self.assertEqual(hosts.redact(r"C:\Users\carol\AppData /c/Users/carol/go", ids),
                         r"C:\Users\<user>\AppData /c/Users/<user>/go")
        self.assertEqual(hosts.redact("ran on SECRET-HOST.example", ids), "ran on <redacted>.example")
        self.assertIsNone(hosts.redact("** WARNING: connection is not using a post-quantum key exchange", ids))

    # --- end to end, through a local Python ---

    def test_local_host_runs_on_a_clone_of_head_with_its_tags(self) -> None:
        path = self.config([{"label": "m", "kind": "local", "command": [sys.executable, "-c", CHECK_CLONE]}])
        code, out, err = self.run_main(["--config", str(path)])
        self.assertEqual(code, 0, out + err)
        self.assertIn("m: exit=0 (", out)
        self.assertIn("describe=v0.0.1", (self.log_dir(out) / "m.log").read_text(encoding="utf-8"))

    def test_failing_host_prints_its_tail_and_exits_one(self) -> None:
        path = self.config([{"label": "m", "kind": "local",
                             "command": [sys.executable, "-c", "print('boom-line'); raise SystemExit(3)"]}])
        code, out, _ = self.run_main(["--config", str(path)])
        self.assertEqual(code, 1)
        self.assertIn("m: exit=3 (", out)
        self.assertIn("    boom-line", out)
        self.assertIn("0/1 hosts passed", out)

    def test_ssh_host_gets_the_ssh_form_and_its_program(self) -> None:
        path = self.config([{"label": "l", "kind": "linux", "ssh": "dest",
                             "command": [sys.executable, "-c", CHECK_CLONE]}])
        code, out, err = self.run_main(["--config", str(path)], ssh=self.fake(FAKE_SSH))
        self.assertEqual(code, 0, out + err)
        argv = json.loads(self.record.read_text(encoding="utf-8").splitlines()[0])
        self.assertEqual(argv, ["dest", "python3", "-"])

    def test_unreachable_host_is_255_and_exits_one(self) -> None:
        path = self.config([{"label": "l", "kind": "linux", "ssh": "dest"}])
        code, out, _ = self.run_main(["--config", str(path)], ssh=self.fake(UNREACHABLE_SSH))
        self.assertEqual(code, 1)
        self.assertIn("l: exit=255 (", out)

    def test_only_selects_hosts_and_rejects_unknown_labels(self) -> None:
        path = self.config([{"label": "a", "kind": "local", "command": [sys.executable, "-c", "pass"]},
                            {"label": "b", "kind": "local", "command": [sys.executable, "-c", "raise SystemExit(1)"]}])
        code, out, _ = self.run_main(["--config", str(path), "--only", "a"])
        self.assertEqual(code, 0)
        self.assertNotIn("b: exit", out)
        self.assertEqual(self.run_main(["--config", str(path), "--only", "zz"])[0], 2)

    def test_clone_is_removed_afterwards(self) -> None:
        path = self.config([{"label": "m", "kind": "local", "command": [sys.executable, "-c", "pass"]}])
        code, out, _ = self.run_main(["--config", str(path)])
        self.assertEqual(code, 0)
        log = (self.log_dir(out) / "m.log").read_text(encoding="utf-8")
        work = Path(re.search(r"work in (.+)$", log, re.MULTILINE).group(1))
        self.assertTrue(work.name == "work" and work.parent.name.startswith("pcm-hosts-"), log)
        self.assertFalse(work.parent.exists(), f"{work.parent} was left behind")


#: A fake gh: records its argv, then answers from a state file the way
#: GitHub's API does, refusing a tag's deletion while its release exists.
FAKE_GH = """import json, sys
state_path, calls_path, args = sys.argv[1], sys.argv[2], sys.argv[3:]
with open(calls_path, "a", encoding="utf-8") as f:
    f.write(json.dumps(args) + "\\n")
with open(state_path, encoding="utf-8") as f:
    state = json.load(f)
delete = "--method" in args and args[args.index("--method") + 1] == "DELETE"
path = args[-1]
if path in state.get("fail", []):
    sys.exit("gh: HTTP 500: planted failure on " + path)
def save():
    with open(state_path, "w", encoding="utf-8") as f:
        json.dump(state, f)
rest = "/".join(path.split("?")[0].split("/")[3:])
sticky = path in state.get("sticky", [])
if not delete and rest == "releases":
    print(json.dumps([state["releases"]]))
elif not delete and rest == "releases/latest":
    if state.get("latest") not in [r["tag_name"] for r in state["releases"]]:
        sys.exit("gh: Not Found (HTTP 404)")
    print(json.dumps({"tag_name": state["latest"]}))
elif not delete and rest == "git/matching-refs/tags/":
    print(json.dumps([[{"ref": "refs/tags/" + t} for t in state["tags"]]]))
elif delete and rest.startswith("releases/"):
    if not sticky:
        state["releases"] = [r for r in state["releases"] if r["id"] != int(rest.split("/")[1])]
        save()
elif delete and rest.startswith("git/refs/tags/"):
    tag = rest[len("git/refs/tags/"):]
    if any(r["tag_name"] == tag for r in state["releases"]):
        sys.exit("gh: HTTP 422: Cannot delete a tag of an immutable release")
    if not sticky:
        state["tags"] = [t for t in state["tags"] if t != tag]
        save()
else:
    sys.exit("gh: the fake does not serve " + " ".join(args))
"""

#: The repository on 2026-11-07: thirteen stable releases, ids 1 to 13, and
#: 28 tags, the 15 oldest with no release.
LIVE_RELEASES = ("v1.2.0", "v1.3.0", "v1.4.0", "v1.5.0", "v1.6.0", "v1.7.0", "v1.8.0", "v1.9.0",
                 "v1.10.0", "v1.10.1", "v1.10.2", "v1.10.3", "v1.10.4")
LIVE_TAG_ONLY = ("v0.1.0", "v0.1.1", "v0.1.2", "v0.1.3", "v0.2.0", "v0.2.1", "v0.2.2", "v0.2.3",
                 "v1.0.0", "v1.0.2", "v1.1.0", "v1.1.1", "v1.1.2", "v1.1.3", "v1.1.4")
KEPT = LIVE_RELEASES[5:]


def live() -> dict:
    """LIVE, as the fake's state."""
    return {"releases": [{"id": i, "tag_name": t, "draft": False, "prerelease": False}
                         for i, t in enumerate(LIVE_RELEASES, 1)],
            "tags": list(LIVE_TAG_ONLY + LIVE_RELEASES), "latest": "v1.10.4"}


class PruneReleasesTest(unittest.TestCase):
    """tools/prune_releases.py, against a fake gh."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.dir = Path(self.tmp.name)
        self.state = self.dir / "state.json"
        self.calls = self.dir / "calls.jsonl"
        script = self.dir / "fake_gh.py"
        script.write_text(FAKE_GH, encoding="utf-8")
        self.gh = (sys.executable, str(script), str(self.state), str(self.calls))

    def plan(self, state: dict, keep: int = 8) -> prune_releases.Plan:
        releases = [prune_releases.Release(r["id"], r["tag_name"], r["draft"], r["prerelease"])
                    for r in state["releases"]]
        return prune_releases.plan(releases, state["tags"], state["latest"], keep)

    def run_main(self, state: dict | None, args: list[str]) -> tuple[int, str, str]:
        if state is not None:
            self.state.write_text(json.dumps(state), encoding="utf-8")
        out, err = io.StringIO(), io.StringIO()
        with redirect_stderr(err):
            code = prune_releases.main(["--repo", "o/r", *args], gh=self.gh, out=out, err=err)
        return code, out.getvalue(), err.getvalue()

    def read_calls(self) -> list[list[str]]:
        if not self.calls.exists():
            return []
        return [json.loads(line) for line in self.calls.read_text(encoding="utf-8").splitlines()]

    def deletes(self) -> list[str]:
        return [c[-1] for c in self.read_calls() if "DELETE" in c]

    def final(self) -> tuple[list[str], list[str]]:
        state = json.loads(self.state.read_text(encoding="utf-8"))
        return sorted(r["tag_name"] for r in state["releases"]), sorted(state["tags"])

    @staticmethod
    def add_release(state: dict, rid: int, tag: str, *, draft: bool = False, prerelease: bool = False,
                    with_tag: bool = True) -> None:
        state["releases"].append({"id": rid, "tag_name": tag, "draft": draft, "prerelease": prerelease})
        if with_tag:
            state["tags"].append(tag)

    # --- the plan ---

    def test_versions_sort_as_numbers(self) -> None:
        p = self.plan(live())
        self.assertEqual([r.tag for r in p.keep], list(reversed(KEPT)))

    def test_live_state_gives_the_expected_plan(self) -> None:
        p = self.plan(live())
        self.assertIsNone(p.refusal)
        self.assertEqual([r.tag for r in p.delete_releases], list(LIVE_RELEASES[:5]))
        self.assertEqual(p.delete_tags, list(LIVE_TAG_ONLY))
        self.assertEqual(p.report, [])
        self.assertEqual((len(p.delete_releases), p.tag_count()), (5, 20))

    def test_prerelease_below_newest_stable_is_deleted(self) -> None:
        state = live()
        for rid, tag in ((20, "v1.10.4-rc.1"), (21, "v1.10.0-rc.2"), (22, "v1.11.0-rc.1")):
            self.add_release(state, rid, tag, prerelease=True)
        p = self.plan(state)
        deleted = [r.tag for r in p.delete_releases]
        self.assertIn("v1.10.4-rc.1", deleted)
        self.assertIn("v1.10.0-rc.2", deleted)
        self.assertNotIn("v1.11.0-rc.1", deleted)
        self.assertTrue(any("v1.11.0-rc.1" in line for line in p.report), p.report)

    def test_draft_is_reported_never_deleted(self) -> None:
        state = live()
        self.add_release(state, 30, "v1.2.5", draft=True, with_tag=False)
        p = self.plan(state)
        self.assertNotIn("v1.2.5", [r.tag for r in p.delete_releases])
        self.assertIn("draft v1.2.5: drafts are never deleted", p.report)

    def test_tag_without_release_is_pruned_below_the_window_only(self) -> None:
        state = live()
        # v1.8.5 is inside the window, so the bound must be the oldest kept
        # release, not the newest (0026-PLAN, Deviation 1).
        state["tags"] += ["v1.6.5", "v1.8.5", "v1.10.5"]
        p = self.plan(state)
        self.assertIn("v1.6.5", p.delete_tags)
        for tag in ("v1.8.5", "v1.10.5"):
            self.assertNotIn(tag, p.delete_tags)
            self.assertTrue(any(f"tag {tag}" in line for line in p.report), (tag, p.report))
        self.assertEqual(p.keep[-1].tag, "v1.7.0")
        self.assertNotIn("v1.7.0", p.delete_tags)

    def test_other_tags_and_releases_are_left_alone(self) -> None:
        state = live()
        state["tags"] += ["nightly", "v1", "v1.2", "v01.2.3"]
        self.add_release(state, 40, "release-1")
        p = self.plan(state)
        touched = [r.tag for r in p.delete_releases] + p.delete_tags + [r.tag for r in p.keep]
        for name in ("nightly", "v1", "v1.2", "v01.2.3", "release-1"):
            self.assertNotIn(name, touched)
            self.assertTrue(any(f" {name}:" in line for line in p.report), (name, p.report))
        self.assertEqual([r.tag for r in p.delete_releases], list(LIVE_RELEASES[:5]))
        self.assertEqual(p.delete_tags, list(LIVE_TAG_ONLY))

    # --- applying it ---

    def test_apply_deletes_each_release_before_its_tag(self) -> None:
        code, out, err = self.run_main(live(), ["--apply"])
        self.assertEqual(code, 0, out + err)
        deletes = self.deletes()
        for rid, tag in enumerate(LIVE_RELEASES[:5], 1):
            self.assertLess(deletes.index(f"repos/o/r/releases/{rid}"),
                            deletes.index(f"repos/o/r/git/refs/tags/{tag}"))
        self.assertEqual(deletes[-15:], [f"repos/o/r/git/refs/tags/{t}" for t in LIVE_TAG_ONLY])
        self.assertEqual(len(deletes), 25)

    def test_refuses_with_fewer_stable_releases_than_keep(self) -> None:
        state = live()
        state["releases"] = state["releases"][6:]
        for args in ([], ["--apply"]):
            code, _, err = self.run_main(state, args)
            self.assertEqual(code, 2)
            self.assertIn("only 7 stable releases; --keep is 8", err)
        self.assertEqual(self.deletes(), [])

    def test_refuses_to_delete_the_latest_release(self) -> None:
        state = live()
        state["latest"] = "v1.2.0"
        code, _, err = self.run_main(state, ["--apply"])
        self.assertEqual(code, 2)
        self.assertIn("would delete the latest release, v1.2.0", err)
        self.assertEqual(self.deletes(), [])

    def test_refuses_when_a_listing_fails(self) -> None:
        for path in ("repos/o/r/releases?per_page=100", "repos/o/r/releases/latest",
                     "repos/o/r/git/matching-refs/tags/"):
            with self.subTest(path=path):
                state = live()
                state["fail"] = [path]
                code, _, err = self.run_main(state, ["--apply"])
                self.assertEqual(code, 2)
                self.assertIn("planted failure", err)
                self.assertEqual(self.deletes(), [])

    def test_dry_run_deletes_nothing(self) -> None:
        code, out, _ = self.run_main(live(), [])
        self.assertEqual(code, 1)
        self.assertIn("5 releases and 20 tags to delete", out)
        self.assertIn("dry run: nothing deleted; run again with --apply", out)
        self.assertEqual(self.deletes(), [])
        state = live()
        state["releases"] = state["releases"][5:]
        state["tags"] = list(KEPT)
        code, out, _ = self.run_main(state, [])
        self.assertEqual(code, 0, out)
        self.assertIn("nothing to prune", out)

    def test_apply_then_dry_run_is_clean(self) -> None:
        code, out, err = self.run_main(live(), ["--apply"])
        self.assertEqual(code, 0, out + err)
        self.assertIn("pruned: 5 releases and 20 tags", out)
        self.assertEqual(self.final(), (sorted(KEPT), sorted(KEPT)))
        code, out, _ = self.run_main(None, [])
        self.assertEqual(code, 0, out)
        self.assertIn("nothing to prune", out)

    def test_read_back_catches_a_deletion_that_did_not_hold(self) -> None:
        # v1.2.0's release has no tag ref, so nothing after its DELETE trips
        # over it: only the read-back can see that the deletion did not hold.
        state = live()
        state["tags"].remove("v1.2.0")
        state["sticky"] = ["repos/o/r/releases/1"]
        code, _, err = self.run_main(state, ["--apply"])
        self.assertEqual(code, 2)
        self.assertIn("deleted release v1.2.0 is still listed", err)

    def test_failed_deletion_stops_and_a_rerun_finishes(self) -> None:
        state = live()
        state["fail"] = ["repos/o/r/git/refs/tags/v1.4.0"]
        code, out, _ = self.run_main(state, ["--apply"])
        self.assertEqual(code, 2)
        self.assertEqual(self.read_calls()[-1][-1], "repos/o/r/git/refs/tags/v1.4.0")
        self.assertIn("deleted before the failure: release v1.2.0 (id 1), tag v1.2.0, release v1.3.0 (id 2), "
                      "tag v1.3.0, release v1.4.0 (id 3)", out)
        state = json.loads(self.state.read_text(encoding="utf-8"))
        del state["fail"]
        code, out, err = self.run_main(state, ["--apply"])
        self.assertEqual(code, 0, out + err)
        self.assertEqual(self.final(), (sorted(KEPT), sorted(KEPT)))

    # --- arguments ---

    def test_repository_from_origin_url(self) -> None:
        for url in ("https://github.com/OWNER/REPO", "https://github.com/OWNER/REPO.git\n",
                    "git@github.com:OWNER/REPO.git", "ssh://git@github.com/OWNER/REPO.git"):
            self.assertEqual(prune_releases.repo_from_url(url), "OWNER/REPO", url)
        for url in ("https://gitlab.example/o/r", "not-a-url"):
            with self.assertRaises(prune_releases.ToolError, msg=url):
                prune_releases.repo_from_url(url)

    def test_bad_arguments_exit_two(self) -> None:
        for args in (["--keep", "0"], ["--keep", "x"], ["--no-such-flag"]):
            self.assertEqual(self.run_main(live(), args)[0], 2, args)
        self.assertEqual(self.read_calls(), [])

    def test_keep_overrides_the_default(self) -> None:
        code, out, _ = self.run_main(live(), ["--keep", "5"])
        self.assertEqual(code, 1)
        for tag in LIVE_RELEASES[:8]:
            self.assertIn(f"delete  release {tag} and its tag", out)
        self.assertIn("8 releases and 23 tags to delete", out)
        code, _, err = self.run_main(live(), ["--keep", "14"])
        self.assertEqual(code, 2)
        self.assertIn("only 13 stable releases; --keep is 14", err)


if __name__ == "__main__":
    unittest.main()
