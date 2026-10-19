#!/usr/bin/env python3
"""The maintainer tools' tests: each fails on a planted break, and passes clean.

Defined by docs/decisions/0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md
D3. Run by `make tools-test`, which `make verify` runs.
"""
from __future__ import annotations

import json
import sys
import tempfile
import threading
import time
import unittest
import urllib.error
import urllib.request
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import fakeprovider  # noqa: E402

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


if __name__ == "__main__":
    unittest.main()
