#!/usr/bin/env python3
"""Check the exact staged snapshot before Go changes are committed.

Every run verifies that github.com/maccavelli/mcplib resolves from GitHub
(through the module proxy and checksum database) at the version go.mod pins:
no replace directive, no go.work redirect, no pseudo-version, no checksum
bypass, and module-cache bytes that match go.sum. Local mcplib sources on the
host are never consulted.

When Go sources or module files are staged (or passed as arguments), the
snapshot must also pass gofmt, golangci-lint (fmt and run) and govulncheck.

Usage: scripts/go-precheck.py [file ...]
"""

from __future__ import annotations

import fnmatch
import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path

MCPLIB = "github.com/maccavelli/mcplib"
LINT_CONFIG = ".golangci.yml"
MODULE_FILES = {"go.mod", "go.sum"}
PSEUDO_VERSION = re.compile(r"(^|[-.])(0\.)?\d{14}-[0-9a-f]{12}(\+incompatible)?$")
EXE_SUFFIX = ".exe" if os.name == "nt" else ""


class PrecheckError(Exception):
    """A gate failed; the message explains why."""


def run(cmd, cwd=None, env=None, capture=False):
    result = subprocess.run(
        [str(part) for part in cmd],
        cwd=cwd,
        env=env,
        text=True,
        capture_output=capture,
    )
    if result.returncode != 0:
        detail = (result.stderr or result.stdout or "").strip() if capture else ""
        raise PrecheckError(
            f"command failed (exit {result.returncode}): {' '.join(map(str, cmd))}"
            + (f"\n{detail}" if detail else "")
        )
    return result.stdout if capture else ""


def go_json(args, cwd, env):
    return json.loads(run(["go", *args], cwd=cwd, env=env, capture=True))


def go_environment():
    env = dict(os.environ)
    # A go.work file or an inherited -mod=mod could silently swap in the local
    # mcplib checkout or rewrite go.mod; the snapshot must build as committed.
    env["GOWORK"] = "off"
    env["GOFLAGS"] = "-mod=readonly"
    return env


def repo_root():
    return Path(run(["git", "rev-parse", "--show-toplevel"], capture=True).strip())


def is_target(name):
    return name.endswith(".go") or name in MODULE_FILES


def requested_targets(root, argv):
    if argv:
        names = []
        for arg in argv:
            path = Path(arg)
            if not path.is_absolute():
                path = Path.cwd() / path
            names.append(Path(os.path.relpath(path.resolve(), root.resolve())).as_posix())
    else:
        output = run(
            ["git", "-C", root, "diff", "--cached", "--name-only", "--diff-filter=ACM", "-z"],
            capture=True,
        )
        names = [name for name in output.split("\0") if name]
    return [name for name in names if is_target(name)]


def snapshot_index(root, destination):
    destination.mkdir(parents=True)
    run([
        "git", "-C", root, "-c", "core.autocrlf=false",
        "checkout-index", "--all", f"--prefix={destination.as_posix()}/",
    ])


def matches_module_patterns(patterns, module):
    """Mirror golang.org/x/mod/module.MatchPrefixPatterns."""
    for pattern in filter(None, (p.strip() for p in patterns.split(","))):
        depth = pattern.count("/") + 1
        prefix = "/".join(module.split("/")[:depth])
        if fnmatch.fnmatchcase(prefix, pattern):
            return True
    return False


def gosum_hash(snapshot, suffix):
    gosum = snapshot / "go.sum"
    lines = gosum.read_text(encoding="utf-8").splitlines() if gosum.is_file() else []
    for line in lines:
        fields = line.split()
        if len(fields) == 3 and fields[0] == MCPLIB and fields[1] == suffix:
            return fields[2]
    return None


def check_mcplib(snapshot, env):
    modfile = go_json(["mod", "edit", "-json"], snapshot, env)
    required = [r for r in modfile.get("Require") or [] if r["Path"] == MCPLIB]
    if not required:
        raise PrecheckError(f"go.mod does not require {MCPLIB}")
    version = required[0]["Version"]
    if PSEUDO_VERSION.search(version):
        raise PrecheckError(f"{MCPLIB} {version} is a pseudo-version; pin a released tag")

    for replace in modfile.get("Replace") or []:
        if replace["Old"]["Path"] == MCPLIB:
            new = replace["New"]
            target = f"{new['Path']} {new.get('Version', '')}".strip()
            raise PrecheckError(
                f"go.mod replaces {MCPLIB} with {target}; the dependency must resolve from GitHub"
            )

    settings = go_json(
        ["env", "-json", "GOPRIVATE", "GONOSUMDB", "GONOSUMCHECK", "GOINSECURE",
         "GOSUMDB", "GOPROXY", "GOMODCACHE"],
        snapshot, env,
    )
    if settings["GOSUMDB"] == "off":
        raise PrecheckError("GOSUMDB=off disables checksum verification for mcplib")
    if settings["GOPROXY"] == "off":
        raise PrecheckError("GOPROXY=off prevents resolving mcplib from GitHub")
    for key in ("GOPRIVATE", "GONOSUMDB", "GONOSUMCHECK", "GOINSECURE"):
        if matches_module_patterns(settings[key], MCPLIB):
            raise PrecheckError(f"{key}={settings[key]} exempts {MCPLIB} from checksum verification")

    zip_hash = gosum_hash(snapshot, version)
    mod_hash = gosum_hash(snapshot, f"{version}/go.mod")
    if not zip_hash or not mod_hash:
        raise PrecheckError(f"go.sum is missing {MCPLIB} {version} checksums")

    downloaded = go_json(["mod", "download", "-json", MCPLIB], snapshot, env)
    if downloaded.get("Error"):
        raise PrecheckError(f"downloading {MCPLIB}: {downloaded['Error']}")
    if downloaded.get("Version") != version or downloaded.get("Sum") != zip_hash:
        raise PrecheckError(
            f"{MCPLIB} download {downloaded.get('Version')} {downloaded.get('Sum')} "
            f"does not match go.sum {version} {zip_hash}"
        )

    resolved = go_json(["list", "-m", "-json", MCPLIB], snapshot, env)
    cache = os.path.normcase(os.path.realpath(settings["GOMODCACHE"]))
    source = os.path.normcase(os.path.realpath(resolved.get("Dir", "")))
    if resolved.get("Replace") or resolved.get("Version") != version:
        raise PrecheckError(f"{MCPLIB} does not resolve to {version} without replacement")
    if os.path.commonpath([cache, source]) != cache:
        raise PrecheckError(f"{MCPLIB} resolves to {resolved.get('Dir')}, outside the module cache")

    run(["go", "mod", "verify"], cwd=snapshot, env=env)
    run(["go", "mod", "tidy", "-diff"], cwd=snapshot, env=env)
    print(f"go-precheck: {MCPLIB} {version} resolved from GitHub ({zip_hash})", flush=True)


def pinned_tools(root):
    """Read tool pins from bootstrap-tools.sh so versions live in one place."""
    text = (root / "scripts" / "bootstrap-tools.sh").read_text(encoding="utf-8")
    variables = dict(re.findall(r'^(\w+)="([^"]*)"$', text, flags=re.MULTILINE))

    def expand(value):
        return re.sub(r"\$(\w+)", lambda m: variables.get(m.group(1), ""), value)

    tools = {}
    calls = re.findall(r'install_tool\s*\\\s*"([^"]+)"\s*\\\s*"([^"]+)"\s*\\\s*"([^"]+)"\s*\\\s*"([^"]+)"', text)
    for name, module, version, expected in calls:
        tools[name] = (module, expand(version), expand(expected))
    return tools, variables.get("GO_VERSION")


def ensure_tools(root, env, names):
    tools, go_version = pinned_tools(root)
    actual = run(["go", "env", "GOVERSION"], cwd=root, env=env, capture=True).strip()
    if go_version and actual != go_version:
        raise PrecheckError(f"expected {go_version}, got {actual}")

    bin_dir = root / ".tools" / "bin"
    paths = {}
    for name in names:
        module, version, expected = tools[name]
        binary = bin_dir / f"{name}{EXE_SUFFIX}"
        output = ""
        if binary.is_file():
            flag = "version" if name == "golangci-lint" else "-version"
            probe = subprocess.run([str(binary), flag], capture_output=True, text=True)
            output = probe.stdout + probe.stderr
        if expected not in output:
            print(f"installing {module}@{version}", flush=True)
            bin_dir.mkdir(parents=True, exist_ok=True)
            run(["go", "install", f"{module}@{version}"], cwd=root, env={**env, "GOBIN": str(bin_dir)})
        paths[name] = binary
    return paths


def check_sources(root, snapshot, env, targets):
    go_files = [snapshot / name for name in targets if name.endswith(".go") and (snapshot / name).is_file()]
    if go_files:
        unformatted = run(["gofmt", "-l", *go_files], env=env, capture=True).strip()
        if unformatted:
            raise PrecheckError(f"gofmt found unformatted staged files:\n{unformatted}")

    tools = ensure_tools(root, env, ["golangci-lint", "govulncheck"])
    run([tools["golangci-lint"], "fmt", "--diff", "-c", LINT_CONFIG], cwd=snapshot, env=env)
    run([tools["golangci-lint"], "run", "-c", LINT_CONFIG, "./..."], cwd=snapshot, env=env)
    run([tools["govulncheck"], "./..."], cwd=snapshot, env=env)


def main(argv):
    root = repo_root()
    env = go_environment()
    targets = requested_targets(root, argv)

    with tempfile.TemporaryDirectory(prefix="prepare-commit-msg-staged-", ignore_cleanup_errors=True) as temp:
        snapshot = Path(temp) / root.name
        snapshot_index(root, snapshot)
        check_mcplib(snapshot, env)
        if not targets:
            return 0
        check_sources(root, snapshot, env, targets)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main(sys.argv[1:]))
    except PrecheckError as error:
        print(f"go-precheck: {error}", file=sys.stderr)
        sys.exit(1)
