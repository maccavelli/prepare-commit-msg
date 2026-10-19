#!/usr/bin/env python3
"""Run CI's jobs for HEAD on the test hosts: `make preflight-hosts`.

Defined by docs/decisions/0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md
D4 and A1. It refuses a dirty tree, because CI tests the pushed commit. It
bundles HEAD and the tags, and sends each host one program on stdin:
tools/hostrun.py's source, with the bundle, the command and the label. Every
host runs in parallel. Each host's output is redacted and written to its own
log; this prints one line per host, `<label>: exit=N (Ns)`, then the failing
hosts' last lines.

The hosts are named only in a config file outside the repository:
--config PATH, else $XDG_CONFIG_HOME/prepare-commit-msg/hosts.json, else
~/.config/prepare-commit-msg/hosts.json. tools/hosts.example.json shows its
shape, with placeholders.

Exit: 0 every host passed; 1 a host failed or could not be reached; 2 a
config, tree or argument error.
"""

from __future__ import annotations

import argparse
import base64
import json
import os
import re
import subprocess
import sys
import tempfile
import threading
import time
from collections import deque
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass
from pathlib import Path
from typing import TextIO

REPO = Path(__file__).resolve().parent.parent
RUNNER = Path(__file__).with_name("hostrun.py")

#: hostrun.py's last lines, which program() replaces with the call.
FOOTER = 'if __name__ == "__main__":\n    sys.exit(main(json.load(sys.stdin)))\n'

KINDS = ("linux", "windows", "local")

#: Each kind's CI job (0017-MADR D4's table).
DEFAULT_COMMANDS = {
    "linux": ("make", "verify"),
    "windows": ("go", "test", "./..."),
    "local": ("go", "test", "./..."),
}

LABEL = re.compile(r"[A-Za-z0-9._-]+")

#: The ssh prefix; tests replace it with a fake.
SSH = ("ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=5", "-T")

#: ssh's own lines, which say nothing of the run.
NOISE = re.compile(r"post-quantum|store now, decrypt later|pq\.html|Pseudo-terminal", re.IGNORECASE)

#: Home directories, whose last part is an account name.
HOMES = (
    (re.compile(r"/home/[^/\s]+"), "/home/<user>"),
    (re.compile(r"/Users/[^/\s]+"), "/Users/<user>"),
    (re.compile(r"/c/Users/[^/\s]+", re.IGNORECASE), "/c/Users/<user>"),
    (re.compile(r"([A-Za-z]:)([\\/]+)Users([\\/]+)[^\\/\s]+", re.IGNORECASE), r"\1\2Users\3<user>"),
)

TAIL = 20
CANNOT_START = 127


class UsageError(Exception):
    """A config, tree or argument error: exit 2."""


@dataclass(frozen=True)
class Host:
    label: str
    kind: str
    ssh: str | None
    wsl: str | None
    command: tuple[str, ...]


@dataclass(frozen=True)
class Result:
    label: str
    exit: int
    seconds: float
    tail: tuple[str, ...]


def default_config(env: dict[str, str] | None = None) -> Path:
    env = os.environ if env is None else env
    base = env.get("XDG_CONFIG_HOME") or str(Path.home() / ".config")
    return Path(base) / "prepare-commit-msg" / "hosts.json"


def load_config(path: Path) -> tuple[list[Host], list[re.Pattern[str]]]:
    """The hosts and the identifier patterns, or UsageError naming the
    path and the field."""
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        raise UsageError(f"no config at {path}: copy tools/hosts.example.json there, outside the repository") from None
    except (OSError, ValueError) as err:
        raise UsageError(f"{path}: {err}") from None
    if not isinstance(data, dict) or not isinstance(data.get("hosts"), list) or not data["hosts"]:
        raise UsageError(f"{path}: hosts: a non-empty list is required")
    hosts, seen = [], set()
    for i, entry in enumerate(data["hosts"]):
        where = f"{path}: hosts[{i}]"
        if not isinstance(entry, dict):
            raise UsageError(f"{where}: an object is required")
        unknown = set(entry) - {"label", "kind", "ssh", "wsl", "command"}
        if unknown:
            raise UsageError(f"{where}: unknown field {sorted(unknown)[0]!r}")
        label, kind = entry.get("label"), entry.get("kind")
        if not isinstance(label, str) or not LABEL.fullmatch(label):
            raise UsageError(f"{where}.label: letters, digits, '.', '_' or '-' are required")
        if label in seen:
            raise UsageError(f"{where}.label: {label!r} is used twice")
        seen.add(label)
        if kind not in KINDS:
            raise UsageError(f"{where}.kind: one of {', '.join(KINDS)} is required")
        ssh, wsl = entry.get("ssh"), entry.get("wsl")
        if kind == "local":
            if ssh is not None or wsl is not None:
                raise UsageError(f"{where}: a local host takes no ssh or wsl")
        elif not isinstance(ssh, str) or not ssh:
            raise UsageError(f"{where}.ssh: an ssh destination is required")
        if wsl is not None and (kind != "linux" or not isinstance(wsl, str) or not wsl):
            raise UsageError(f"{where}.wsl: a distro name, on a linux host only")
        command = entry.get("command", list(DEFAULT_COMMANDS[kind]))
        if not isinstance(command, list) or not command or not all(isinstance(a, str) and a for a in command):
            raise UsageError(f"{where}.command: a non-empty list of strings is required")
        hosts.append(Host(label, kind, ssh, wsl, tuple(command)))
    identifiers = data.get("identifiers", [])
    if not isinstance(identifiers, list) or not all(isinstance(s, str) and s for s in identifiers):
        raise UsageError(f"{path}: identifiers: a list of non-empty strings is required")
    return hosts, [re.compile(re.escape(s), re.IGNORECASE) for s in identifiers]


def transport(host: Host, ssh: tuple[str, ...] = SSH) -> list[str]:
    """The argument list that starts `python3 -` for host (0017-MADR A1).
    A WSL distro is entered through its Windows host's connection."""
    if host.kind == "local":
        return [sys.executable, "-"]
    assert host.ssh is not None
    if host.wsl:
        return [*ssh, host.ssh, "wsl", "-d", host.wsl, "--", "bash", "-lc", "'python3 -'"]
    return [*ssh, host.ssh, "python3", "-"]


def git(repo: Path, *args: str) -> subprocess.CompletedProcess[bytes]:
    return subprocess.run(["git", "-C", str(repo), *args], capture_output=True, check=False)


def head_bundle(repo: Path) -> bytes:
    """A bundle of HEAD and every tag, or UsageError if the tree is dirty."""
    status = git(repo, "status", "--porcelain")
    if status.returncode != 0:
        raise UsageError(f"git status: {status.stderr.decode(errors='replace').strip()}")
    if status.stdout.strip():
        raise UsageError("commit first: preflight-hosts tests HEAD, and the tree has changes")
    with tempfile.TemporaryDirectory(prefix="pcm-hosts-") as d:
        out = Path(d) / "head.bundle"
        made = git(repo, "bundle", "create", str(out), "HEAD", "--tags")
        if made.returncode != 0:
            raise UsageError(f"git bundle: {made.stderr.decode(errors='replace').strip()}")
        return out.read_bytes()


def program(label: str, bundle: bytes, command: tuple[str, ...]) -> str:
    """hostrun.py's source, with its call carrying the parameters."""
    source = RUNNER.read_text(encoding="utf-8")
    if not source.endswith(FOOTER):
        raise UsageError(f"{RUNNER.name} does not end with the call program() replaces")
    params = {"label": label, "bundle_b64": base64.b64encode(bundle).decode("ascii"), "command": list(command)}
    return source[: -len(FOOTER)] + f"sys.exit(main(json.loads({json.dumps(params)!r})))\n"


def redact(line: str, identifiers: list[re.Pattern[str]] = ()) -> str | None:
    """line with home paths and identifiers replaced, or None for ssh's
    noise."""
    if NOISE.search(line):
        return None
    for pattern, repl in HOMES:
        line = pattern.sub(repl, line)
    for pattern in identifiers:
        line = pattern.sub("<redacted>", line)
    return line


def run_host(host: Host, prog: str, log: Path, identifiers: list[re.Pattern[str]],
             ssh: tuple[str, ...] = SSH) -> Result:
    """Run prog on host; its redacted output goes to log."""
    start = time.monotonic()
    tail: deque[str] = deque(maxlen=TAIL)
    with log.open("w", encoding="utf-8") as out:
        try:
            proc = subprocess.Popen(transport(host, ssh), stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                    stderr=subprocess.STDOUT, text=True, encoding="utf-8", errors="replace")
        except OSError as err:
            line = f"preflight-hosts: cannot start the transport: {err}"
            out.write(line + "\n")
            return Result(host.label, CANNOT_START, time.monotonic() - start, (line,))
        assert proc.stdin is not None and proc.stdout is not None

        def feed() -> None:
            # A host that cannot be reached closes the pipe before reading.
            try:
                proc.stdin.write(prog)
                proc.stdin.close()
            except OSError:
                pass

        writer = threading.Thread(target=feed, daemon=True)
        writer.start()
        for raw in proc.stdout:
            line = redact(raw.rstrip("\r\n"), identifiers)
            if line is None:
                continue
            out.write(line + "\n")
            out.flush()
            tail.append(line)
        code = proc.wait()
        writer.join()
    return Result(host.label, code, time.monotonic() - start, tuple(tail))


def main(argv: list[str] | None = None, *, repo: Path = REPO, ssh: tuple[str, ...] = SSH,
         out: TextIO | None = None) -> int:
    out = sys.stdout if out is None else out
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--config", type=Path, help="the host config (default: %(default)s)",
                        default=default_config())
    parser.add_argument("--only", action="append", metavar="LABEL", help="run only this host; repeatable")
    try:
        args = parser.parse_args(argv)
    except SystemExit as err:
        return 0 if err.code == 0 else 2
    try:
        hosts, identifiers = load_config(args.config)
        if args.only:
            unknown = sorted(set(args.only) - {h.label for h in hosts})
            if unknown:
                raise UsageError(f"--only: no host labelled {unknown[0]!r} in {args.config}")
            hosts = [h for h in hosts if h.label in args.only]
        bundle = head_bundle(repo)
        progs = {h.label: program(h.label, bundle, h.command) for h in hosts}
    except UsageError as err:
        print(f"preflight-hosts: {err}", file=sys.stderr)
        return 2
    run_dir = Path(tempfile.mkdtemp(prefix="pcm-hosts-"))
    with ThreadPoolExecutor(max_workers=len(hosts)) as pool:
        futures = [pool.submit(run_host, h, progs[h.label], run_dir / f"{h.label}.log", identifiers, ssh)
                   for h in hosts]
        results = [f.result() for f in futures]
    for r in results:
        print(f"{r.label}: exit={r.exit} ({r.seconds:.0f}s)", file=out)
    for r in results:
        if r.exit != 0:
            print(f"--- {r.label}: last {len(r.tail)} lines", file=out)
            for line in r.tail:
                print(f"    {line}", file=out)
    passed = sum(r.exit == 0 for r in results)
    print(f"preflight-hosts: {passed}/{len(results)} hosts passed; logs in {run_dir}", file=out)
    return 0 if passed == len(results) else 1


if __name__ == "__main__":
    sys.exit(main())
