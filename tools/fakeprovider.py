#!/usr/bin/env python3
"""fakeprovider: a local OpenAI-compatible API that answers the first generations with an error.

Defined by docs/decisions/0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md
D2, for the live retry check of a release (0015-PLAN Phase 5 step 3). It serves,
on 127.0.0.1 only:

  GET  .../models            one model
  POST .../chat/completions  a chat completion
  POST .../responses         a Responses answer

The first --fail generations are answered with --status and a Retry-After
of --retry-after seconds; every later one with --message, in the shape of the
API asked. Each request is logged to --log as "METHOD PATH -> STATUS", and
nothing else: no header and no body, so a key the hook sends is never written.

Usage:
  tools/fakeprovider.py [--fail N] [--status CODE] [--retry-after S]
                        [--message TEXT] [--port-file FILE] [--log FILE]
                        [--max-requests N]

The port is chosen by the system and written to --port-file (and printed).
The server stops on an interrupt, or after --max-requests requests.

Exit codes: 0 stopped, 2 a bad argument.
"""
from __future__ import annotations

import argparse
import json
import sys
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

MODEL = "fake-model"
DEFAULT_MESSAGE = "chore: add live retry check file"


def chat_completion(message: str) -> dict:
    """A chat completion holding message, as the Chat Completions API sends it."""
    return {
        "id": "chatcmpl-fake", "object": "chat.completion", "created": 0, "model": MODEL,
        "choices": [{"index": 0, "finish_reason": "stop",
                     "message": {"role": "assistant", "content": message}}],
        "usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
    }


def responses_answer(message: str) -> dict:
    """A Responses answer holding message, as the Responses API sends it."""
    return {
        "id": "resp-fake", "object": "response", "status": "completed", "model": MODEL,
        "output": [{"type": "message", "id": "msg-fake", "role": "assistant", "status": "completed",
                    "content": [{"type": "output_text", "text": message, "annotations": []}]}],
        "usage": {"input_tokens": 1, "output_tokens": 1, "total_tokens": 2},
    }


MODELS = {"object": "list", "data": [{"id": MODEL, "object": "model", "owned_by": "fake"}]}
ANSWERS = {"/chat/completions": chat_completion, "/responses": responses_answer}


class FakeProvider(ThreadingHTTPServer):
    """The server, with its settings and its counters."""

    daemon_threads = True

    def __init__(self, fail: int, status: int, retry_after: int, message: str,
                 log: Path | None, max_requests: int | None) -> None:
        super().__init__(("127.0.0.1", 0), Handler)
        self.fail, self.status, self.retry_after, self.message = fail, status, retry_after, message
        self.log, self.max_requests = log, max_requests
        self.lock = threading.Lock()
        self.generations = 0
        self.requests = 0

    @property
    def port(self) -> int:
        return self.server_address[1]

    def next_generation(self) -> int:
        """The number of this generation, the first being 1."""
        with self.lock:
            self.generations += 1
            return self.generations

    def record(self, method: str, path: str, status: int) -> None:
        """Logs one request, and stops the server after max_requests."""
        with self.lock:
            self.requests += 1
            done = self.max_requests is not None and self.requests >= self.max_requests
            if self.log is not None:
                with self.log.open("a", encoding="utf-8") as f:
                    f.write(f"{method} {path} -> {status}\n")
        if done:
            threading.Thread(target=self.shutdown, daemon=True).start()


class Handler(BaseHTTPRequestHandler):
    server: FakeProvider

    def log_message(self, *_args: object) -> None:
        """No request line on stderr: the log file is the record."""

    def answer(self, status: int, body: dict, headers: tuple[tuple[str, str], ...] = ()) -> None:
        data = json.dumps(body).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        for name, value in headers:
            self.send_header(name, value)
        self.end_headers()
        self.wfile.write(data)
        self.server.record(self.command, self.path, status)

    def do_GET(self) -> None:
        if self.path.rstrip("/").endswith("/models"):
            self.answer(200, MODELS)
        else:
            self.answer(404, {"error": {"message": "not found"}})

    def do_POST(self) -> None:
        self.rfile.read(int(self.headers.get("Content-Length") or 0))
        endpoint = next((e for e in ANSWERS if self.path.rstrip("/").endswith(e)), None)
        if endpoint is None:
            self.answer(404, {"error": {"message": "not found"}})
            return
        s = self.server
        if s.next_generation() <= s.fail:
            self.answer(s.status, {"error": {"message": "fake provider: failing as asked", "type": "rate_limit"}},
                        (("Retry-After", str(s.retry_after)),))
            return
        self.answer(200, ANSWERS[endpoint](s.message))


def serve(fail: int = 1, status: int = 429, retry_after: int = 1, message: str = DEFAULT_MESSAGE,
          log: Path | None = None, max_requests: int | None = None) -> FakeProvider:
    """A running fake provider on 127.0.0.1, on a port the system chose.
    The caller stops it with shutdown() and server_close()."""
    server = FakeProvider(fail, status, retry_after, message, log, max_requests)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    return server


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--fail", type=int, default=1, help="generations to fail first (default 1)")
    parser.add_argument("--status", type=int, default=429, help="their status (default 429)")
    parser.add_argument("--retry-after", type=int, default=1, help="their Retry-After, in seconds (default 1)")
    parser.add_argument("--message", default=DEFAULT_MESSAGE, help="the answer after them")
    parser.add_argument("--port-file", type=Path, help="where to write the port")
    parser.add_argument("--log", type=Path, help="where to log each request")
    parser.add_argument("--max-requests", type=int, help="stop after this many requests")
    try:
        args = parser.parse_args(argv)
    except SystemExit as exit_:
        return 0 if exit_.code == 0 else 2
    if args.fail < 0 or args.retry_after < 0 or not 400 <= args.status <= 599 or (
            args.max_requests is not None and args.max_requests < 1):
        print("fakeprovider: --fail and --retry-after are at least 0, --status is 400 to 599, "
              "--max-requests is at least 1", file=sys.stderr)
        return 2
    server = FakeProvider(args.fail, args.status, args.retry_after, args.message, args.log, args.max_requests)
    if args.port_file is not None:
        args.port_file.write_text(str(server.port), encoding="utf-8")
    print(f"fakeprovider: listening on 127.0.0.1:{server.port}", flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()
    return 0


if __name__ == "__main__":
    sys.exit(main())
