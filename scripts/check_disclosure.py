#!/usr/bin/env python3
"""Scan tracked text for real user-home paths and print file, line, category.

Defined by docs/decisions/0022-MADR-align-docs-ci-and-disclosure-gates.md and
docs/decisions/0022-PLAN-align-docs-ci-and-disclosure-gates.md Phase 1 step 3.
CI runs it as `make disclosure-check` in the Docs workflow and the go job,
beside the owner's private pre-push disclosure guard, which keeps the private
deny patterns. This check is generic: it holds no hostname, account name or
machine path, and it prints no line content.

A finding is a user-home path written in the real layout with a concrete
(non-placeholder) account segment:

* macOS:    /Users/<name>
* Linux:    /home/<name>
* Windows:  \\Users\\<name> (covers C:\\Users\\<name>), and C:/Users/<name>;
  Windows forms are matched case-insensitively

A placeholder account segment is accepted, so the conventions AGENTS.md and
the records rely on pass:

* an environment substitution, such as $HOME, ${HOME}, %USERPROFILE%,
  $env:USERPROFILE
* a ~ segment

An angle-bracket token never decides here: a line carrying a full <…> token
demonstrates redaction in its own text — the redactor's rewriting table, its
tests, and the records' planted evidence look like that — and is accepted as
a whole, and so is the flagged line when the token sits on the line beside it
too, since a wrapped literal splits a demonstration across lines (0022-PLAN
Deviation 1, R-B, owner's choice 2026-10-25). A real home path within a line
of a <…> token passes it; the private pre-push guard owns the real names.
Markup reaching into the segment from either side is ignored first, and
markup without an account in it is accepted as no path occurrence.

One line is printed per file and line, `<path>:<line>: <category>`, then a
summary that names no content. A file with a NUL byte is read as binary and
skipped. A tracked file that cannot be read is a tool error, never a silent
skip. An account segment stops at the first separator or whitespace, since a
real one never spans either on its own.

Exit status: 0 no finding, 1 a finding, 2 a tool error, dominated by either
kind of problem.
"""
from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent

# (prefix bytes, matched case-insensitively?, category); the first prefix
# that yields a finding on a line wins.
PREFIXES = (
    (b"/Users/", False, "macOS home path"),
    (b"/home/", False, "Linux home path"),
    (b"\\users\\", True, "Windows home path"),
    (b"c:/users/", True, "Windows home path"),
)

SEPARATORS = (b"/", b"\\", b" ", b"\t")

# Markdown and prose markup that reaches into a segment from either side.
LEADING_MARKUP = b"([{`'\""
TRAILING_MARKUP = b")]}`'\".,;:!?"


def stripped_segment(segment: bytes) -> bytes:
    """One segment without markup reaching in from either side."""
    while segment[:1] in (b"(", b"[", b"{", b"`", b"'", b'"'):
        segment = segment[1:]
    while segment[-1:] in (b")", b"]", b"}", b"`", b"'", b'"',
                           b",", b".", b";", b":", b"!", b"?"):
        segment = segment[:-1]
    return segment


def is_placeholder(segment: bytes) -> bool:
    """Whether one home-path account segment is a documented placeholder.

    An angle-bracket token never reaches here on a live line: a line
    carrying a full <…> token is a redaction demonstration, accepted
    wholesale before this function decides. Markup without an account
    inside it is not a path occurrence, and is accepted too.
    """
    segment = stripped_segment(segment)
    if not segment or segment == b"~":
        return True
    if segment.startswith(b"$"):
        return True
    if segment.startswith(b"%") and segment.endswith(b"%"):
        return True
    return False


def segment_end(hay: bytes, start: int) -> int:
    """Where the account segment after a prefix stops in one line."""
    end = len(hay)
    for sep in SEPARATORS:
        stop = hay.find(sep, start)
        if stop != -1:
            end = min(end, stop)
    return end


def line_category(line: bytes) -> str | None:
    """The finding category for one line's own bytes, or None."""
    low = line.lower()
    for prefix, insensitive, name in PREFIXES:
        hay = low if insensitive else line
        at = hay.find(prefix)
        while at != -1:
            start = at + len(prefix)
            segment = line[start:segment_end(hay, start)]
            if segment and not is_placeholder(segment):
                return name
            at = hay.find(prefix, at + 1)
    return None


def scan_bytes(data: bytes) -> list[tuple[int, str]]:
    """One (line number, category) pair per affected line of one file."""
    lines = data.splitlines()
    raw = [(line_no, category) for line_no, line in enumerate(lines, 1)
           if (category := line_category(line)) is not None]

    def demonstrates_redaction(line_no: int) -> bool:
        for other in (line_no - 1, line_no, line_no + 1):
            if 1 <= other <= len(lines) and b"<" in lines[other - 1] \
                    and b">" in lines[other - 1]:
                return True
        return False

    return [(line_no, category) for line_no, category in raw
            if not demonstrates_redaction(line_no)]


def tracked_files() -> list[str]:
    """Every path git lists in the repository this file sits in."""
    head = subprocess.run(["git", "ls-files", "-z"], cwd=REPO,
                          stdout=subprocess.PIPE, check=True).stdout
    return [os.fsdecode(name) for name in head.split(b"\x00") if name]


def scan_files(paths: list[str] | None = None,
               base: Path = REPO) -> tuple[list[str], int, int]:
    """Finding lines, files scanned, and files that could not be read.

    Each path is read under `base` (the repository by default). One line per
    affected file and line, in the form `<path>:<line>: <category>`, never
    any line content; a file that cannot be read yields one line naming
    nothing but the path, under the category `unreadable`.
    """
    names = tracked_files() if paths is None else paths
    findings: list[str] = []
    unreadable = 0
    scanned = 0
    for name in names:
        path = Path(base, name)
        try:
            data = path.read_bytes()
        except OSError:
            unreadable += 1
            findings.append(f"{name}: unreadable input (category: unreadable)")
            continue
        scanned += 1
        if b"\x00" in data:
            continue
        seen: set[tuple[int, str]] = set()
        for line_no, category in scan_bytes(data):
            if (line_no, category) not in seen:
                seen.add((line_no, category))
                findings.append(f"{name}:{line_no}: {category}")
    return findings, scanned, unreadable


def main() -> int:
    findings, scanned, unreadable = scan_files()
    for line in findings:
        print(line)
    print(f"disclosure: {scanned} files scanned, {len(findings)} finding(s), "
          f"{unreadable} unreadable")
    if unreadable:
        return 2
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main())
