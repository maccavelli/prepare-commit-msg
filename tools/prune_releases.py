#!/usr/bin/env python3
"""Keep the newest stable GitHub releases; delete older releases and tags.

Defined by docs/decisions/0026-MADR-keep-five-stable-releases.md D1 to D5.
It keeps the --keep newest stable releases (default 8). It deletes each older
stable release, and each prerelease at or below the newest stable release,
the release before its tag. It deletes each `v*` tag with no release below
the oldest kept release. Drafts, newer prereleases and tags, tags outside
`v*`, and tags that are not vMAJOR.MINOR.PATCH[-PRERELEASE] are reported,
never deleted. By default it prints the plan and deletes nothing; --apply
deletes, then reads the state back and checks it.

Usage: python3 tools/prune_releases.py [--repo OWNER/REPO] [--keep N] [--apply]

The repository defaults to this clone's `origin`. Every call goes through
`gh api`, logged in as an account that may delete releases and tags.

Exit: 0 nothing to prune, or --apply succeeded; 1 a dry run with something
to prune; 2 a refusal, a gh or git failure, or an argument error.
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import TextIO

REPO = Path(__file__).resolve().parent.parent

#: The gh prefix; tests replace it with a fake.
GH = ("gh",)

#: The window when --keep is not given (MADR D1, D5).
DEFAULT_KEEP = 8

VERSION = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)"
                     r"(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?")

GITHUB_URL = re.compile(r"(?:https://github\.com/|git@github\.com:|ssh://git@github\.com/)"
                        r"([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+?)(?:\.git)?/?")


class ToolError(Exception):
    """A gh or git failure, a bad read-back, or a bad origin URL: exit 2."""


@dataclass(frozen=True)
class Release:
    id: int
    tag: str
    draft: bool
    prerelease: bool


@dataclass
class Plan:
    latest: str | None
    tags: frozenset[str]
    keep: list[Release] = field(default_factory=list)
    delete_releases: list[Release] = field(default_factory=list)
    delete_tags: list[str] = field(default_factory=list)
    report: list[str] = field(default_factory=list)
    refusal: str | None = None

    def tag_count(self) -> int:
        """The tags --apply deletes: each deleted release's that exists, and
        each tag with no release."""
        return sum(r.tag in self.tags for r in self.delete_releases) + len(self.delete_tags)

    def empty(self) -> bool:
        return not self.delete_releases and not self.delete_tags


def parse_version(tag: str) -> tuple[int, int, int, str | None] | None:
    """(major, minor, patch, prerelease) for vMAJOR.MINOR.PATCH[-PRERELEASE],
    with no leading zeros; None for anything else."""
    m = VERSION.fullmatch(tag)
    if not m:
        return None
    return int(m[1]), int(m[2]), int(m[3]), m[4]


def lower_than(version: tuple[int, int, int, str | None], core: tuple[int, int, int]) -> bool:
    """Whether version is below the stable version core: a lower core, or the
    same core with a prerelease."""
    return version[:3] < core or (version[:3] == core and version[3] is not None)


def repo_from_url(url: str) -> str:
    """OWNER/REPO from a GitHub HTTPS, scp-form or ssh:// URL."""
    m = GITHUB_URL.fullmatch(url.strip())
    if not m:
        raise ToolError(f"not a GitHub repository URL: {url.strip()!r}; pass --repo OWNER/REPO")
    return f"{m[1]}/{m[2]}"


def plan(releases: list[Release], tags: list[str], latest: str | None, keep: int) -> Plan:
    """What to keep, delete and report (MADR D1 to D4). Pure."""
    result = Plan(latest=latest, tags=frozenset(tags))
    stable: list[tuple[tuple[int, int, int], Release]] = []
    pre: list[tuple[tuple[int, int, int], Release]] = []
    for r in releases:
        v = parse_version(r.tag)
        if v is None:
            result.report.append(f"release {r.tag}: not a vMAJOR.MINOR.PATCH tag; never touched")
        elif r.draft:
            result.report.append(f"draft {r.tag}: drafts are never deleted")
        elif r.prerelease or v[3] is not None:
            pre.append((v[:3], r))
        else:
            stable.append((v[:3], r))
    stable.sort(key=lambda item: item[0], reverse=True)
    if len(stable) < keep:
        result.refusal = f"only {len(stable)} stable releases; --keep is {keep}"
        return result
    result.keep = [r for _, r in stable[:keep]]
    result.delete_releases = [r for _, r in stable[keep:]]
    newest, oldest = stable[0][0], stable[keep - 1][0]
    for core, r in pre:
        if core <= newest:
            result.delete_releases.append(r)
        else:
            result.report.append(f"prerelease {r.tag}: newer than the newest stable release; kept")
    released = {r.tag for r in releases}
    for tag in tags:
        if tag in released:
            continue
        v = parse_version(tag)
        if v is None:
            result.report.append(f"tag {tag}: not a vMAJOR.MINOR.PATCH tag; never touched")
        elif lower_than(v, oldest):
            result.delete_tags.append(tag)
        else:
            result.report.append(f"tag {tag}: no release, not below the window; kept")
    result.delete_releases.sort(key=lambda r: version_key(r.tag))
    result.delete_tags.sort(key=version_key)
    if latest is not None and latest in {r.tag for r in result.delete_releases}:
        result.refusal = f"the plan would delete the latest release, {latest}"
    if result.refusal:
        result.delete_releases, result.delete_tags = [], []
    return result


def version_key(tag: str) -> tuple:
    """Semantic-version order for a tag parse_version accepts: a prerelease
    sorts before its stable version."""
    v = parse_version(tag)
    assert v is not None
    return (*v[:3], v[3] is None, v[3] or "")


def gh_json(gh: tuple[str, ...], *args: str) -> object:
    """gh's decoded JSON output, or ToolError with its stderr."""
    proc = subprocess.run([*gh, *args], capture_output=True, text=True)
    if proc.returncode != 0:
        raise ToolError(f"gh {' '.join(args)}: exit {proc.returncode}: {proc.stderr.strip()}")
    try:
        return json.loads(proc.stdout)
    except ValueError as err:
        raise ToolError(f"gh {' '.join(args)}: unreadable JSON: {err}") from None


def read_state(gh: tuple[str, ...], repo: str) -> tuple[list[Release], list[str], str]:
    """The releases, the tag names and the latest release's tag."""
    pages = gh_json(gh, "api", "--paginate", "--slurp", f"repos/{repo}/releases?per_page=100")
    latest = gh_json(gh, "api", f"repos/{repo}/releases/latest")
    ref_pages = gh_json(gh, "api", "--paginate", "--slurp", f"repos/{repo}/git/matching-refs/tags/")
    try:
        releases = [Release(int(r["id"]), str(r["tag_name"]), bool(r["draft"]), bool(r["prerelease"]))
                    for page in pages for r in page]
        tags = [str(ref["ref"]).removeprefix("refs/tags/") for page in ref_pages for ref in page]
        return releases, tags, str(latest["tag_name"])
    except (KeyError, TypeError, ValueError) as err:
        raise ToolError(f"unexpected gh api output: {err!r}") from None


def apply(gh: tuple[str, ...], repo: str, p: Plan, out: TextIO) -> None:
    """Delete each release, oldest first, then its tag; then each tag with no
    release. Stop at the first failure."""
    steps: list[tuple[str, str]] = []
    for r in p.delete_releases:
        steps.append((f"release {r.tag} (id {r.id})", f"repos/{repo}/releases/{r.id}"))
        if r.tag in p.tags:
            steps.append((f"tag {r.tag}", f"repos/{repo}/git/refs/tags/{r.tag}"))
    steps += [(f"tag {t}", f"repos/{repo}/git/refs/tags/{t}") for t in p.delete_tags]
    done: list[str] = []
    for what, path in steps:
        proc = subprocess.run([*gh, "api", "--method", "DELETE", path], capture_output=True, text=True)
        if proc.returncode != 0:
            print(f"failed  {what}: exit {proc.returncode}: {proc.stderr.strip()}", file=out)
            print(f"deleted before the failure: {', '.join(done) if done else 'nothing'}", file=out)
            raise ToolError(f"deleting {what} failed; run --apply again to finish")
        done.append(what)
        print(f"deleted {what}", file=out)


def read_back(gh: tuple[str, ...], repo: str, p: Plan, err: TextIO) -> None:
    """Check the live state matches the plan, or ToolError."""
    releases, tags, latest = read_state(gh, repo)
    listed = {r.tag for r in releases}
    problems = [f"kept release {r.tag} is not listed" for r in p.keep if r.tag not in listed]
    problems += [f"deleted release {r.tag} is still listed" for r in p.delete_releases if r.tag in listed]
    gone = [r.tag for r in p.delete_releases if r.tag in p.tags] + p.delete_tags
    problems += [f"deleted tag {t} is still listed" for t in gone if t in tags]
    if latest != p.latest:
        problems.append(f"the latest release is {latest}, not {p.latest}")
    for problem in problems:
        print(f"read-back: {problem}", file=err)
    if problems:
        raise ToolError(f"the read-back found {len(problems)} problem(s)")


def print_plan(repo: str, keep: int, p: Plan, out: TextIO) -> None:
    print(f"repository {repo}, keeping the {keep} newest stable releases", file=out)
    for r in p.keep:
        print(f"keep    release {r.tag}{' (latest)' if r.tag == p.latest else ''}", file=out)
    for r in p.delete_releases:
        print(f"delete  release {r.tag} {'and its tag' if r.tag in p.tags else '(it has no tag)'}", file=out)
    for t in p.delete_tags:
        print(f"delete  tag {t} (no release)", file=out)
    for line in p.report:
        print(f"report  {line}", file=out)
    print(f"{len(p.delete_releases)} releases and {p.tag_count()} tags to delete", file=out)


def positive(text: str) -> int:
    try:
        value = int(text)
    except ValueError:
        raise argparse.ArgumentTypeError(f"not an integer: {text!r}") from None
    if value < 1:
        raise argparse.ArgumentTypeError("must be at least 1")
    return value


def main(argv: list[str] | None = None, gh: tuple[str, ...] = GH, out: TextIO | None = None,
         err: TextIO | None = None) -> int:
    out = sys.stdout if out is None else out
    err = sys.stderr if err is None else err
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--repo", metavar="OWNER/REPO", help="default: this clone's origin")
    parser.add_argument("--keep", type=positive, default=DEFAULT_KEEP, metavar="N",
                        help="stable releases to keep (default: %(default)s)")
    parser.add_argument("--apply", action="store_true", help="delete; without it, print the plan only")
    try:
        args = parser.parse_args(argv)
    except SystemExit as exc:
        return 0 if exc.code == 0 else 2
    try:
        repo = args.repo
        if repo is None:
            proc = subprocess.run(["git", "-C", str(REPO), "remote", "get-url", "origin"],
                                  capture_output=True, text=True)
            if proc.returncode != 0:
                raise ToolError(f"git remote get-url origin: {proc.stderr.strip()}")
            repo = repo_from_url(proc.stdout)
        releases, tags, latest = read_state(gh, repo)
        p = plan(releases, tags, latest, args.keep)
        print_plan(repo, args.keep, p, out)
        if p.refusal:
            print(f"prune_releases: refused, nothing deleted: {p.refusal}", file=err)
            return 2
        if p.empty():
            print("nothing to prune", file=out)
            return 0
        if not args.apply:
            print("dry run: nothing deleted; run again with --apply", file=out)
            return 1
        apply(gh, repo, p, out)
        read_back(gh, repo, p, err)
    except ToolError as exc:
        print(f"prune_releases: {exc}", file=err)
        return 2
    print(f"pruned: {len(p.delete_releases)} releases and {p.tag_count()} tags", file=out)
    return 0


if __name__ == "__main__":
    sys.exit(main())
