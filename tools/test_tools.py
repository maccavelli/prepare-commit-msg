#!/usr/bin/env python3
"""The maintainer tools' tests: each fails on a planted break, and passes clean.

Defined by docs/decisions/0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md
D3, and, for hosts.py and hostrun.py,
docs/decisions/0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md
D4. Run by `make tools-test`, which `make verify` runs.
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


if __name__ == "__main__":
    unittest.main()
