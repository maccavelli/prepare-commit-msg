#!/usr/bin/env python3
"""check_ci_filters.py's tests: each fixture plants one break, each passes clean.

Defined by docs/decisions/0022-MADR-align-docs-ci-and-disclosure-gates.md and
docs/decisions/0022-PLAN-align-docs-ci-and-disclosure-gates.md Phase 2 step 3,
with the plants 0019-PLAN's Phase 1 step 6 named for the superset of them this
checker replaces. Run by `make tools-test`, which `make verify` runs; the gate
that runs the checker in CI is `make workflow-lint`.

Fixtures are synthetic git repositories under a temporary directory; the
tracked file list comes from the index, so a `git add -A` is all they need.
The checkout pins in the templates are this repository's committed pins. The
fixture paths are globs and repo docs names, which no disclosure pattern
flags, so this file stays scan-clean.
"""
from __future__ import annotations

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import check_ci_filters

CHECKOUT = "actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.0"
SETUP_GO = "actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0"
OTHER_PIN = CHECKOUT.replace("3d3c42e5aac5ba805825da76410c181273ba90b1", "0" * 40)

IGNORE = "['docs/README.md', 'AGENTS.md']"
CLIPPED_IGNORE = "['docs/README.md']"
WIDE_IGNORE = "['README.md', 'docs/README.md', 'AGENTS.md']"
ARCH_IGNORE = "['docs/architecture.md', 'AGENTS.md']"
TWO_GATES = ("      - name: Check records\n"
             "        run: make check-records\n")

CI_SOURCE = """name: CI

on:
  push:
    branches: [main]
    tags:
      - "v*"
    paths-ignore: PUSH_IGNORE
  pull_request:
    paths-ignore: PR_IGNORE
  workflow_dispatch:

jobs:
  go:
    name: Go (quality contract)
    runs-on: ubuntu-24.04
    steps:
      - name: Checkout
        uses: CHECKOUT
      - name: Set up Go
        uses: SETUP_GO
      - name: Check records
        run: make check-records
      - name: Lint Markdown
        run: make markdownlint
      - name: Disclosure
        run: make disclosure-check
      - name: Contract
        run: make verify
"""

DOCS_SOURCE = """name: Docs

on:
  push:
    branches: [main]
    paths: DOCS_PATHS
  pull_request:
    paths: DOCS_PATHS

jobs:
  docs:
    name: Docs
    runs-on: ubuntu-24.04
    steps:
      - name: Checkout
        uses: CHECKOUT
      - name: Check records
        run: make check-records
      - name: Lint Markdown
        run: make markdownlint
      - name: Disclosure
        run: make disclosure-check
"""

GO_LINES = {
    "README-reading": 'package main\n\nimport "os"\n\nvar _, _ = os.ReadFile("README.md")\n',
    "embed": "package main\n\n//go:embed docs/architecture.md\nvar embedded string\n",
    "silence": "package main\n\nvar silent = 1\n",
}
TOOL_LINES = {
    "reading": ('#!/usr/bin/env python3\n"""Cites docs/decisions/0001-MADR-'
                'record-only.md in prose."""\nDOCS = "docs/README.md"\n'),
    "citing": ('#!/usr/bin/env python3\n"""Cites docs/decisions/0001-MADR-'
               'record-only.md in its docstring."""\n'),
}

READ_BROKEN = "ci filters: README.md is read by code, ignored by ci.yml, and " \
    "its reader does not run in Docs"
DOCS_READ_BROKEN = "ci filters: docs/README.md is read by code, ignored by " \
    "ci.yml, and its reader does not run in Docs"
EMBED_READ_BROKEN = "ci filters: docs/architecture.md is read by code, ignored " \
    "by ci.yml, and its reader does not run in Docs"


def ci_source(push_ignore: str = IGNORE, pr_ignore: str | None = None) -> str:
    text = CI_SOURCE.replace("PUSH_IGNORE", push_ignore)
    text = text.replace("PR_IGNORE", pr_ignore or push_ignore)
    return text.replace("CHECKOUT", CHECKOUT).replace("SETUP_GO", SETUP_GO)


def docs_source(paths: str = "**.md", checkout: str = CHECKOUT,
                without_disclosure: bool = False) -> str:
    """The docs.yml fixture; a bare glob becomes its one-entry flow list."""
    if not paths.startswith("["):
        paths = "['" + paths + "']"
    text = DOCS_SOURCE.replace("DOCS_PATHS", paths)
    if without_disclosure:
        text = text.replace(
            "      - name: Disclosure\n        run: make disclosure-check\n", "")
    return text.replace("CHECKOUT", checkout)


class CheckTest(unittest.TestCase):
    """check_ci_filters.check over a scratch repository."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)

    def write(self, rel: str, text: str) -> None:
        path = Path(self.root, rel)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text)

    def track_all(self) -> None:
        subprocess.run(["git", "init", "-q"], cwd=self.root, check=True)
        subprocess.run(["git", "add", "-A"], cwd=self.root, check=True)

    def check(self, files: dict[str, str]) -> list[str]:
        for rel, text in files.items():
            self.write(rel, text)
        self.track_all()
        return check_ci_filters.check(self.root).problems

    def good_files(self, push_ignore: str = IGNORE,
                   docs_paths: str = "**.md") -> dict[str, str]:
        return {
            ".github/workflows/ci.yml": ci_source(push_ignore),
            ".github/workflows/docs.yml": docs_source(docs_paths),
            "README.md": "product readme\n",
            "docs/README.md": "records index\n",
            "AGENTS.md": "agent rules\n",
            "fixture.go": GO_LINES["silence"],
        }

    def test_good_fixture_is_clean(self) -> None:
        problems = self.check(self.good_files())
        self.assertEqual(problems, [])

    def test_docstring_citation_is_not_a_read(self) -> None:
        self.write("tools/citing.py", TOOL_LINES["citing"])
        self.write("docs/decisions/0001-MADR-record-only.md", "record\n")
        self.track_all()
        self.assertEqual(check_ci_filters.tools_md_reads(self.root), set())

    def test_tool_constant_read_of_an_ignored_markdown_is_refused(self) -> None:
        self.write("tools/reader.py", TOOL_LINES["reading"])
        self.write("docs/README.md", "records index\n")
        self.track_all()
        reads = check_ci_filters.tools_md_reads(self.root)
        self.assertEqual(reads, {"docs/README.md"})

    def test_go_md_reads_resolve_package_relative_names(self) -> None:
        self.write("fixture.go", 'package main\n\nimport "os"\n\nvar _, _ = '
                   'os.ReadFile("docs/architecture.md")\n')
        self.track_all()
        reads = check_ci_filters.go_md_reads(self.root)
        self.assertEqual(reads, {"docs/architecture.md"})

    def test_event_lists_must_agree_within_a_workflow(self) -> None:
        files = self.good_files()
        files[".github/workflows/ci.yml"] = ci_source(IGNORE, CLIPPED_IGNORE)
        problems = self.check(files)
        self.assertEqual(problems, ["ci filters: push and pull_request ignore "
                                    "lists differ"])

    def test_markdown_reader_cannot_be_ignored(self) -> None:
        files = {
            ".github/workflows/ci.yml": ci_source(WIDE_IGNORE),
            ".github/workflows/docs.yml": docs_source("**.md"),
            "README.md": "product readme\n",
            "fixture.go": GO_LINES["README-reading"],
        }
        problems = self.check(files)
        self.assertIn(READ_BROKEN, problems)

    def test_tool_read_of_an_ignored_markdown_is_refused(self) -> None:
        files = {
            ".github/workflows/ci.yml": ci_source(IGNORE),
            ".github/workflows/docs.yml": docs_source("**.md"),
            "docs/README.md": "records index\n",
            "AGENTS.md": "agent rules\n",
            "tools/reader.py": TOOL_LINES["reading"],
        }
        problems = self.check(files)
        self.assertIn(DOCS_READ_BROKEN, problems)

    def test_embedded_markdown_cannot_be_ignored(self) -> None:
        files = {
            ".github/workflows/ci.yml": ci_source(ARCH_IGNORE),
            ".github/workflows/docs.yml": docs_source("**.md"),
            "docs/architecture.md": "system\n",
            "AGENTS.md": "agent rules\n",
            "fixture.go": GO_LINES["embed"],
        }
        problems = self.check(files)
        self.assertIn(EMBED_READ_BROKEN, problems)

    def test_reader_running_in_docs_allows_the_ignore(self) -> None:
        files = self.good_files(docs_paths="**.md")
        files[".github/workflows/docs.yml"] = docs_source(
            "**.md").replace(TWO_GATES, "      - name: Contract\n"
                             "        run: make verify\n")
        problems = self.check(files)
        self.assertTrue(all("read by code" not in p for p in problems), problems)

    def test_ignored_path_must_be_covered_by_docs(self) -> None:
        problems = self.check(self.good_files(docs_paths="docs/**.md"))
        self.assertEqual(
            [p for p in problems if "AGENTS.md" in p],
            ["ci filters: ignored path AGENTS.md is not covered by docs.yml"])

    def test_pattern_without_a_tracked_file_is_refused(self) -> None:
        problems = self.check(self.good_files(
            push_ignore="['docs/README.md', 'docs/**.mdx', 'AGENTS.md']",
            docs_paths="**.md"))
        self.assertEqual(
            [p for p in problems if "**.mdx" in p],
            ["ci filters: pattern 'docs/**.mdx' matches no tracked file"])

    def test_missing_docs_gate_is_refused(self) -> None:
        files = self.good_files()
        files[".github/workflows/docs.yml"] = docs_source(
            without_disclosure=True)
        problems = self.check(files)
        self.assertEqual([p for p in problems if "Docs lost" in p],
                         ["ci filters: Docs lost the shared doc gate(s) "
                          "make disclosure-check"])

    def test_differing_checkout_pin_is_refused(self) -> None:
        files = self.good_files()
        files[".github/workflows/docs.yml"] = docs_source(
            "**.md", checkout=OTHER_PIN)
        problems = self.check(files)
        self.assertTrue(any("Docs's use" in p for p in problems), problems)

    def test_docs_cannot_run_a_command_without_a_gate(self) -> None:
        files = self.good_files()
        files[".github/workflows/docs.yml"] = docs_source(
            "**.md") + "      - name: Extra\n        run: make test\n"
        problems = self.check(files)
        self.assertEqual([p for p in problems if "non-gate" in p],
                         ["ci filters: Docs runs non-gate command(s) make test"])

    def test_block_style_list_is_refused(self) -> None:
        files = self.good_files()
        files[".github/workflows/ci.yml"] = CI_SOURCE.replace(
            "    paths-ignore: PUSH_IGNORE",
            "    paths-ignore:\n      - 'AGENTS.md'").replace(
            "CHECKOUT", CHECKOUT).replace("SETUP_GO", SETUP_GO).replace(
            "PR_IGNORE", IGNORE)
        problems = self.check(files)
        self.assertTrue(any("exactly one one-line paths-ignore" in p
                            for p in problems), problems)

    def test_block_style_paths_in_docs_is_refused(self) -> None:
        files = self.good_files()
        files[".github/workflows/docs.yml"] = docs_source().replace(
            "    paths: ['**.md']", "    paths:\n      - '**.md'")
        problems = self.check(files)
        self.assertTrue(any("exactly one one-line filter list" in p
                            for p in problems), problems)

    def test_lost_tags_trigger_is_refused(self) -> None:
        files = self.good_files()
        files[".github/workflows/ci.yml"] = ci_source().replace(
            "    tags:\n      - \"v*\"\n", "")
        problems = self.check(files)
        self.assertEqual([p for p in problems if "tags" in p],
                         ["ci filters: .github/workflows/ci.yml's on.push lost "
                          "its tags trigger"])

    def test_lost_dispatch_is_refused(self) -> None:
        files = self.good_files()
        files[".github/workflows/ci.yml"] = ci_source().replace(
            "  workflow_dispatch:\n", "")
        problems = self.check(files)
        self.assertEqual([p for p in problems if "workflow_dispatch" in p],
                         ["ci filters: .github/workflows/ci.yml lost its "
                          "workflow_dispatch trigger"])

    def test_lost_go_job_is_refused(self) -> None:
        files = self.good_files()
        files[".github/workflows/ci.yml"] = ci_source().replace(
            "  go:\n    name: Go (quality contract)\n",
            "  gone:\n    name: Go (quality contract)\n")
        problems = self.check(files)
        self.assertEqual([p for p in problems if "go job" in p],
                         ["ci filters: .github/workflows/ci.yml lost its go job"])

    def test_workflow_files_must_exist(self) -> None:
        files = self.good_files()
        del files[".github/workflows/docs.yml"]
        problems = self.check(files)
        self.assertEqual(problems, ["ci filters: .github/workflows/docs.yml is "
                                    "missing or unreadable"])

    def test_glob_translation_follows_github_semantics(self) -> None:
        finder = check_ci_filters.glob_regex("**.md")
        self.assertTrue(all(finder.fullmatch(p)
                            for p in ("README.md", "docs/x.md", "a/b/c.md")))
        self.assertFalse(finder.fullmatch("docs/x.json"))
        self.assertFalse(check_ci_filters.glob_regex("docs/*.md")
                         .fullmatch("docs/a/x.md"))
        self.assertTrue(check_ci_filters.glob_regex("docs/x.m?")
                        .fullmatch("docs/x.md"))


if __name__ == "__main__":
    unittest.main()
