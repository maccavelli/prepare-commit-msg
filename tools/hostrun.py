#!/usr/bin/env python3
"""What one host runs for `make preflight-hosts`: a command, on a clone of HEAD.

Defined by docs/decisions/0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md
D4 and A1. tools/hosts.py reads this file's source, replaces its last two lines
with a call that carries the parameters, and sends the result to the host's
`python3 -` on stdin. It is never imported on the host, and the clone it makes
does not exist until it runs, so it uses only the standard library. Python 3.12
or later (shutil.rmtree's onexc).

params: {"label": str, "bundle_b64": str, "command": [str, ...]}

It writes the bundle into a new temporary directory, clones it, prints
`hostrun: <label>: python <version>, work in <path>`, runs the command in the
clone with every output line flushed, and removes the directory, whatever
happened. Exit: the command's status; 125 when the clone cannot be made; 127
when the command cannot be started.
"""

from __future__ import annotations

import base64
import json
import os
import shutil
import stat
import subprocess
import sys
import tempfile
from pathlib import Path

SETUP_FAILED = 125
CANNOT_RUN = 127


def _force(func, path, _exc) -> None:
    """rmtree's onexc: make a read-only file, such as a git object on
    Windows, writable, and try again."""
    os.chmod(path, stat.S_IWRITE)
    func(path)


def remove(path: Path) -> None:
    """Remove path and everything under it, read-only files included."""
    if path.exists():
        shutil.rmtree(path, onexc=_force)


def say(line: str) -> None:
    print(line, flush=True)


def main(params: dict) -> int:
    # A Windows pipe defaults to the ANSI code page; the output is UTF-8.
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    label, command = params["label"], params["command"]
    tmp = Path(tempfile.mkdtemp(prefix="pcm-hosts-"))
    try:
        bundle = tmp / "head.bundle"
        bundle.write_bytes(base64.b64decode(params["bundle_b64"]))
        work = tmp / "work"
        clone = subprocess.run(["git", "clone", "-q", str(bundle), str(work)],
                               capture_output=True, text=True, errors="replace", check=False)
        if clone.returncode != 0:
            say(f"hostrun: setup failed: git clone exit {clone.returncode}: {clone.stderr.strip()}")
            return SETUP_FAILED
        say(f"hostrun: {label}: python {sys.version.split()[0]}, work in {work}")
        # A bare name reaches only a .exe on Windows; resolve it on this host's PATH.
        exe = shutil.which(command[0]) or command[0]
        try:
            proc = subprocess.Popen([exe, *command[1:]], cwd=work, stdin=subprocess.DEVNULL,
                                    stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                    text=True, encoding="utf-8", errors="replace")
        except OSError as err:
            say(f"hostrun: cannot run {command[0]}: {err}")
            return CANNOT_RUN
        assert proc.stdout is not None
        for line in proc.stdout:
            say(line.rstrip("\r\n"))
        return proc.wait()
    finally:
        remove(tmp)


if __name__ == "__main__":
    sys.exit(main(json.load(sys.stdin)))
