#!/usr/bin/env python3
"""Check ci.yml's and docs.yml's path filters and their parity.

Defined by docs/decisions/0022-MADR-align-docs-ci-and-disclosure-gates.md and
docs/decisions/0022-PLAN-align-docs-ci-and-disclosure-gates.md Phase 2 step 2,
extending the filter checker of
docs/decisions/0019-PLAN-skip-full-ci-for-docs-only-pushes.md D3 whose Phase 1
this plan superseded. Run by `make workflow-lint`, which `make verify` runs.

With no arguments it checks the repository it sits in, and reports one line
per problem, then the summary `ci filters: N patterns, M reads checked,
K problems`; exit 1 iff problems, 2 if git itself failed.

What it refuses:

* a filter that is missing, block-shaped, or duplicated (the flow lists are
  one line each; anything else is malformed);
* push and pull_request event filters that differ within a workflow;
* a paths-ignore pattern that matches no tracked file, so a typo cannot pass
  unseen;
* a tracked path that the ignore set covers while Docs' paths do not;
* a path that code reads — Go string literals, //go:embed matches, non-doc
  string constants of tools/*.py, every tracked file under testdata/ — while
  the ignore set covers it, unless Docs' steps run the reader themselves;
* Docs running a command or action pin the go job does not run, and the go
  job or Docs losing one of the shared documentation gates, so the common
  contract cannot drift apart silently;
* ci.yml losing its tags trigger or workflow_dispatch: every tag and a
  dispatch keep the full workflow, and GitHub does not evaluate path filters
  for a tag push.
"""
from __future__ import annotations

import ast
import posixpath
import re
import subprocess
import sys
from dataclasses import dataclass, field
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent

CI = ".github/workflows/ci.yml"
DOCS = ".github/workflows/docs.yml"

# The go job's run that is the repository's quality contract rather than a
# documentation gate: with the 0022 MADR it stays go-job-only on every push,
# and no docs-only push runs it.
GO_CONTRACT_RUN = "make verify"

# Documentation gates the two workflows share, per 0022-MADR.
DOC_GATES = ("make check-records", "make markdownlint", "make disclosure-check")

EVENT_HEADER = re.compile(r"^  (push|pull_request|workflow_dispatch): *$")
JOB_HEADER = re.compile(r"^  ([a-z][A-Za-z0-9_-]*): *$")
FLOW_LIST = re.compile(r"^    (paths-ignore|paths): \[(.*)\]$")
TAGS_LINE = re.compile(r"^    tags: *$")
RUN_LINE = re.compile(r"^\s+run: (.+)$")
USES_LINE = re.compile(r"^\s+uses: (.+)$")
EMBED_LINE = re.compile(r"^[ \t]*//go:embed\s+(.+)$", re.MULTILINE)
GO_MD_STRING = re.compile(r"(?:\"([^\"\n]*\.md)\"|`([^`\n]*\.md)`)")
ON_HEADER = re.compile(r"^on: *$")


def tracked_files(repo: Path) -> list[str]:
    """Every path git lists for the repository, in index order."""
    listing = subprocess.run(["git", "ls-files", "-z"], cwd=repo,
                             stdout=subprocess.PIPE, check=True).stdout
    return [part.decode() for part in listing.split(b"\x00") if part]


def glob_regex(pattern: str) -> re.Pattern:
    """GitHub's path glob: ** is any characters, * any but /, ? exactly one."""
    out = ["^"]
    at = 0
    while at < len(pattern):
        char = pattern[at]
        if char == "*":
            out.append(".*" if pattern[at + 1:at + 2] == "*" else "[^/]*")
            at += 1 if pattern[at + 1:at + 2] == "*" else 0
        elif char == "?":
            out.append("[^/]")
        else:
            out.append(re.escape(char))
        at += 1
    return re.compile("".join(out) + "$")


def matching(tracked: list[str], pattern: str) -> list[str]:
    """Tracked paths one filter pattern matches."""
    finder = glob_regex(pattern)
    return [rel for rel in tracked if finder.fullmatch(rel)]


class Workflow:
    """The parsed pieces of one workflow file the parity check needs."""

    def __init__(self, repo: Path, rel: str) -> None:
        self.rel = rel
        self.lines = (Path(repo, rel)).read_text().splitlines()
        self.on_lines = self._on_lines()
        self.jobs: dict[str, list[str]] = {name: block
                                           for name, block in self._job_blocks()}

    def _on_lines(self) -> list[str]:
        """The on: mapping's lines, ending at the jobs: key.

        Event keys carry the same two-space indentation as job names, so the
        section boundary is the literal jobs: line, never an indent test.
        """
        for at, line in enumerate(self.lines):
            if ON_HEADER.match(line):
                end = next((n for n in range(at + 1, len(self.lines))
                            if self.lines[n].rstrip() == "jobs:"),
                           len(self.lines))
                return self.lines[at:end]
        return []

    def _job_blocks(self) -> list[tuple[str, list[str]]]:
        heads = [n for n, line in enumerate(self.lines)
                 if JOB_HEADER.match(line)
                 and n > self._jobs_marker()]
        blocks = []
        for position, start in enumerate(heads):
            to = heads[position + 1] if position + 1 < len(heads) else len(self.lines)
            name = JOB_HEADER.match(self.lines[start]).group(1)
            blocks.append((name, self.lines[start:to]))
        return blocks

    def _jobs_marker(self) -> int:
        """The line jobs: starts on, or one past the end when absent."""
        for at, line in enumerate(self.lines):
            if line.rstrip() == "jobs:":
                return at
        return len(self.lines)

    def event_block(self, event: str) -> list[str]:
        """The lines of one event block under on:, header included."""
        heads = [n for n, line in enumerate(self.on_lines)
                 if EVENT_HEADER.match(line)]
        wanted = [n for n in heads if self.on_lines[n] == f"  {event}:"]
        if not wanted:
            return []
        to = next((n for n in heads if n > wanted[0]), len(self.on_lines))
        return self.on_lines[wanted[0]:to]

    def flow_occurrences(self, block: list[str], key: str) -> list[list[str]]:
        """One-line flow lists under key, entry lists in order of appearance."""
        found = []
        for line in block:
            match = FLOW_LIST.match(line)
            if match and match.group(1) == key:
                found.append([entry.strip().strip("'\"")
                              for entry in match.group(2).split(",") if entry.strip()])
        return found

    def runs(self, block: list[str]) -> list[str]:
        return [RUN_LINE.match(line).group(1).strip() for line in block
                if RUN_LINE.match(line)]

    def uses(self) -> list[str]:
        return [USES_LINE.match(line).group(1).strip() for line in self.lines
                if USES_LINE.match(line)]

    def has_event(self, event: str) -> bool:
        return bool(self.event_block(event))


def resolve_read(reading: str, ref: str) -> str | None:
    """A reference resolved against the reading file's directory, in-repo.

    Git paths are posix on every platform, so the resolution stays in posix
    string space; a reference leaving the repository resolves to None.
    """
    ref = ref[2:] if ref.startswith("./") else ref
    resolved = posixpath.normpath(posixpath.join(posixpath.dirname(reading), ref))
    if not resolved or resolved.startswith("..") or resolved == ".":
        return None
    return resolved


def go_md_reads(repo: Path) -> set[str]:
    """Tracked .md paths Go sources name as string literals."""
    reads: set[str] = set()
    for rel in tracked_files(repo):
        if not rel.endswith(".go"):
            continue
        text = (Path(repo, rel)).read_text()
        for match in GO_MD_STRING.finditer(text):
            named = match.group(1) if match.group(1) is not None else match.group(2)
            resolved = resolve_read(rel, named)
            if resolved is not None:
                reads.add(resolved)
    return reads


def go_embed_paths(repo: Path) -> set[str]:
    """Tracked files any //go:embed pattern matches, per embedding package."""
    tracked = tracked_files(repo)
    paths: set[str] = set()
    for rel in tracked:
        if not rel.endswith(".go"):
            continue
        directory = posixpath.dirname(rel)
        text = (Path(repo, rel)).read_text()
        for match in EMBED_LINE.finditer(text):
            for pattern in match.group(1).split():
                finder = glob_regex(pattern)
                paths.update((posixpath.join(directory, other) if directory
                              else other) for other in tracked
                             if finder.fullmatch(other))
    return paths


def tools_md_reads(repo: Path) -> set[str]:
    """Tracked repo .md paths the string constants of tools/*.py name.

    A tool names repository paths from the repository root, so its constants
    resolve root-relative; that also makes a refusal conservative. Docstrings
    are citations, not reads; a constant's string value ending in .md
    elsewhere is a read. ast keeps the distinction exact.
    """
    reads: set[str] = set()
    tracked = set(tracked_files(repo))
    for rel in tracked:
        if not (rel.startswith("tools/") and rel.endswith(".py")):
            continue
        source = (Path(repo, rel)).read_text()
        tree = ast.parse(source, rel)
        doc_nodes = set()
        for node in ast.walk(tree):
            if isinstance(node, (ast.Module, ast.ClassDef, ast.FunctionDef,
                                 ast.AsyncFunctionDef)):
                body = node.body
                if body and isinstance(body[0], ast.Expr) \
                        and isinstance(body[0].value, ast.Constant) \
                        and isinstance(body[0].value.value, str):
                    doc_nodes.add(id(body[0].value))
        for node in ast.walk(tree):
            if isinstance(node, ast.Constant) and isinstance(node.value, str) \
                    and node.value.endswith(".md") and id(node.value) not in doc_nodes:
                reads.add(node.value)
    return reads & tracked


def read_set(repo: Path) -> tuple[set[str], set[str]]:
    """Markdown paths code reads, and every tracked testdata/ path.

    The Markdown set is what an ignored-path refusal must name; the testdata
    set joins the ignore refusal whole, so a fixture a pattern would cover
    cannot arrive unnoticed.
    """
    markdown = go_md_reads(repo) | go_embed_paths(repo) | tools_md_reads(repo)
    markdown = {rel for rel in markdown if rel.endswith(".md")} & set(tracked_files(repo))
    testdata = {rel for rel in tracked_files(repo) if rel.startswith("testdata/")}
    return markdown, testdata


def docs_run_commands(workflow: Workflow) -> list[str]:
    """docs.yml's step commands."""
    runs: list[str] = []
    for name, block in workflow.jobs.items():
        runs.extend(workflow.runs(block))
    return runs


@dataclass
class Result:
    """The check() outcome: problem lines and the summary's counts."""

    problems: list[str] = field(default_factory=list)
    patterns: int = 0
    reads: int = 0


def check(repo: Path) -> Result:
    """Every problem line over the two workflows, with the summary counts."""
    result = Result()
    problems: list[str] = result.problems
    workflows: dict[str, Workflow] = {}
    for rel in (CI, DOCS):
        try:
            workflows[rel] = Workflow(repo, rel)
        except (OSError, UnicodeDecodeError):
            problems.append(f"ci filters: {rel} is missing or unreadable")
    if problems:
        return result

    ci, docs = workflows[CI], workflows[DOCS]
    tracked = tracked_files(repo)
    if not ci.has_event("push") or not ci.has_event("pull_request"):
        problems.append(f"ci filters: {CI} lacks an on.push or on.pull_request block")
        return result
    if not docs.has_event("push") or not docs.has_event("pull_request"):
        problems.append(f"ci filters: {DOCS} lacks an on.push or on.pull_request block")
        return result

    ci_push_block, ci_pr_block = ci.event_block("push"), ci.event_block("pull_request")
    docs_push_block, docs_pr_block = docs.event_block("push"), docs.event_block("pull_request")

    for block, event in ((ci_push_block, "on.push"), (ci_pr_block, "on.pull_request")):
        if len(ci.flow_occurrences(block, "paths-ignore")) != 1:
            problems.append(f"ci filters: {CI}'s {event} needs exactly one one-line "
                            "paths-ignore list")

    ci_ignore_push = ci.flow_occurrences(ci_push_block, "paths-ignore")
    ci_ignore_pr = ci.flow_occurrences(ci_pr_block, "paths-ignore")
    docs_push = docs.flow_occurrences(docs_push_block, "paths")
    docs_pr = docs.flow_occurrences(docs_pr_block, "paths")
    result.patterns = len(ci_ignore_push[0]) if len(ci_ignore_push) == 1 else 0

    for event, occurrences in (("ci.yml's push", ci_ignore_push),
                               ("ci.yml's pull_request", ci_ignore_pr),
                               ("docs.yml's push", docs_push),
                               ("docs.yml's pull_request", docs_pr)):
        if len(occurrences) != 1:
            problems.append(f"ci filters: {event} must carry exactly one one-line "
                            "filter list")

    if len(ci_ignore_push) == 1 and len(ci_ignore_pr) == 1 \
            and ci_ignore_push[0] != ci_ignore_pr[0]:
        problems.append("ci filters: push and pull_request ignore lists differ")
    if len(docs_push) == 1 and len(docs_pr) == 1 and docs_push[0] != docs_pr[0]:
        problems.append("ci filters: docs.yml push and pull_request paths differ")

    if not any(TAGS_LINE.match(line) for line in ci_push_block):
        problems.append(f"ci filters: {CI}'s on.push lost its tags trigger")
    if not ci.has_event("workflow_dispatch"):
        problems.append(f"ci filters: {CI} lost its workflow_dispatch trigger")

    read_md, testdata = read_set(repo)
    result.reads = len(read_md | testdata)
    if len(ci_ignore_push) == 1:
        ignore_patterns = ci_ignore_push[0]
        docs_patterns = docs_push[0] if len(docs_push) == 1 else []
        for pattern in ignore_patterns:
            hits = matching(tracked, pattern)
            if not hits:
                problems.append(
                    f"ci filters: pattern {pattern!r} matches no tracked file")
                continue
            for path in hits:
                if not any(glob_regex(doc).fullmatch(path) for doc in docs_patterns):
                    problems.append(f"ci filters: ignored path {path} is not covered "
                                    "by docs.yml")
        docs_gates_go = any(GO_CONTRACT_RUN in run or "go test" in run
                            for run in docs_run_commands(docs))
        if not docs_gates_go:
            for read in sorted(read_md):
                if any(glob_regex(pattern).fullmatch(read)
                       for pattern in ignore_patterns):
                    problems.append(
                        f"ci filters: {read} is read by code, ignored by ci.yml, "
                        "and its reader does not run in Docs")
            for fixture in sorted(testdata):
                if any(glob_regex(pattern).fullmatch(fixture)
                       for pattern in ignore_patterns):
                    problems.append(
                        f"ci filters: tracked fixture {fixture} is ignored by ci.yml")

    if "go" not in ci.jobs:
        problems.append(f"ci filters: {CI} lost its go job")
        return result
    go_runs = set(ci.runs(ci.jobs["go"]))
    doc_runs = docs_run_commands(docs)
    go_missing = sorted(set(DOC_GATES) - {run for run in go_runs if run in DOC_GATES})
    if go_missing:
        problems.append("ci filters: the go job lost the shared doc gate(s) "
                        + ", ".join(go_missing))
    docs_missing = sorted(set(DOC_GATES) - set(doc_runs))
    if docs_missing:
        problems.append("ci filters: Docs lost the shared doc gate(s) "
                        + ", ".join(docs_missing))
    extra = sorted(set(doc_runs) - set(DOC_GATES))
    if extra:
        problems.append("ci filters: Docs runs non-gate command(s) "
                        + ", ".join(extra))
    for use in docs.uses():
        if use not in ci.uses():
            problems.append(f"ci filters: Docs's use {use!r} is absent from ci.yml")
    return result


def main() -> int:
    try:
        result = check(REPO)
    except subprocess.CalledProcessError as error:
        print(f"ci filters: git failed: {error}", file=sys.stderr)
        return 2
    print(f"ci filters: {result.patterns} patterns, {result.reads} reads checked, "
          f"{len(result.problems)} problems")
    for problem in result.problems:
        print(problem)
    return 1 if result.problems else 0


if __name__ == "__main__":
    sys.exit(main())
