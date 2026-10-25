#!/usr/bin/env python3
"""check_disclosure.py's tests: each fails on a planted break, passes clean.

Defined by docs/decisions/0022-MADR-align-docs-ci-and-disclosure-gates.md,
docs/decisions/0022-PLAN-align-docs-ci-and-disclosure-gates.md Phase 1 step 4,
and its Deviation 1 (the owner chose R-B). Run by `make tools-test`, which
`make verify` runs; the gate that runs them in CI is `make disclosure-check`,
which both workflows call.

Every payload composes its account segment and every path prefix from
check_disclosure's own prefix table at module level, so this file itself
carries no full home-path shape and stays scan-clean.
"""
from __future__ import annotations

import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import check_disclosure

MACOS_PREFIX, LINUX_PREFIX = (p[0] for p in check_disclosure.PREFIXES[:2])
WINDOWS_PREFIX, WINDOWS_FORWARD_PREFIX = (p[0] for p in check_disclosure.PREFIXES[2:])

ACCOUNT = b"jo"
PEER = b"pia"
CAROL = b"carol"

INPUT_MACOS = MACOS_PREFIX + ACCOUNT + b"/x\n"
INPUT_LINUX = b"at " + LINUX_PREFIX + ACCOUNT + b"/x here\n"
INPUT_WINDOWS = b"file=C:" + WINDOWS_PREFIX + ACCOUNT + b"\\x.json\n"
INPUT_WINDOWS_SHARE = b"see \\\\SERVER" + WINDOWS_PREFIX + ACCOUNT + b" too\n"
INPUT_WINDOWS_FORWARD = WINDOWS_FORWARD_PREFIX + ACCOUNT + b"/AppData\n"
INPUT_WINDOWS_UPPER = b"and C:" + WINDOWS_PREFIX.upper() + ACCOUNT.upper() + b" here\n"

PLACEHOLDER_INPUTS = (
    LINUX_PREFIX + b"<user>/notes.md\n",
    MACOS_PREFIX + b"<user>/Library/Lists\n",
    b"file=C:" + WINDOWS_PREFIX + b"<user>\\AppData\n",
    LINUX_PREFIX + b"~\n",
    LINUX_PREFIX + b"$HOME/x\n",
    MACOS_PREFIX + b"${HOME}/y\n",
    b"file=C:" + WINDOWS_PREFIX + b"%USERPROFILE%\\cfg\n",
    LINUX_PREFIX + b"$env:USERPROFILE/z\n",
)

README_CONVENTIONS = b"""\
%USERPROFILE%\\.global-git-hooks
%LOCALAPPDATA%\\Programs\\prepare-commit-msg
~/.config/prepare-commit-msg/config.json
$HOME/bin
placeholder /home/<user>/... shapes

all clean
"""

LONE_PATTERN_WORDS = b"layouts " + MACOS_PREFIX + b" and " + LINUX_PREFIX + b" below\n"
PROSE_REACH = b"reach " + MACOS_PREFIX + ACCOUNT + b" when needed\n"
LINE_NUMBER_INPUT = (b"clean\n\n" + WINDOWS_FORWARD_PREFIX + ACCOUNT + b"/app\n"
                     + MACOS_PREFIX + ACCOUNT + b" and " + MACOS_PREFIX + PEER
                     + b"\n")
BINARY_INPUT = b"\x00" + MACOS_PREFIX + ACCOUNT + b"/x\n"

DEMONSTRATION_SAME_LINE = LINUX_PREFIX + rb"[^/\s]+" + b'", "' + LINUX_PREFIX \
    + b"<user>\"\n"
DEMONSTRATION_KEYED_LINE = (b"file=C:" + WINDOWS_PREFIX + ACCOUNT + b"\\cfg, per"
                            b" the C:" + WINDOWS_PREFIX + b"<user> rule\n")
DEMONSTRATION_EVIDENCE = b"'" + MACOS_PREFIX + ACCOUNT + b"/x' != '" \
    + MACOS_PREFIX + b"<user>/x'\n"
DEMONSTRATION_ADJACENT = (b'    r"C:' + WINDOWS_PREFIX + CAROL
                          + b"\\\\AppData /c/Users/" + CAROL
                          + b"/go\", ids),\n"
                          b'    r"C:' + WINDOWS_PREFIX + b"<user>\\\\AppData"
                          b" /c/Users/<user>/go\")\n")
BOUNDARY_TEXT = (b"seen: <user> form\n"
                 b"no path on this line\n"
                 + MACOS_PREFIX + ACCOUNT + b"/x\n"
                 + b"nothing else\n"
                 + b"again <user>\n")

SECRET_DIR = b"secret-dir"
SECRET_INPUT = b"file=C:" + WINDOWS_PREFIX + ACCOUNT + b"\\" + SECRET_DIR + b"\\name\n"


class ScanBytesTest(unittest.TestCase):
    """check_disclosure.scan_bytes, over one file's bytes."""

    def findings(self, data: bytes) -> list[tuple[int, str]]:
        return check_disclosure.scan_bytes(data)

    def test_macos_linux_windows_real_paths(self) -> None:
        self.assertEqual(self.findings(INPUT_MACOS)[0], (1, "macOS home path"))
        self.assertEqual(self.findings(INPUT_LINUX)[0], (1, "Linux home path"))
        self.assertEqual(self.findings(INPUT_WINDOWS)[0], (1, "Windows home path"))
        self.assertEqual(self.findings(INPUT_WINDOWS_SHARE)[0],
                         (1, "Windows home path"))
        self.assertEqual(self.findings(INPUT_WINDOWS_FORWARD)[0],
                         (1, "Windows home path"))
        self.assertEqual(self.findings(INPUT_WINDOWS_UPPER)[0],
                         (1, "Windows home path"))

    def test_placeholder_segments_pass_clean(self) -> None:
        for line in PLACEHOLDER_INPUTS:
            self.assertEqual(self.findings(line), [], line)

    def test_placeholder_conventions_from_the_readme_pass(self) -> None:
        self.assertEqual(self.findings(README_CONVENTIONS), [])

    def test_prose_path_stops_at_whitespace(self) -> None:
        self.assertEqual(self.findings(LONE_PATTERN_WORDS), [])
        self.assertEqual(self.findings(PROSE_REACH), [(1, "macOS home path")])

    def test_lone_or_placeholder_only_segments_pass(self) -> None:
        self.assertEqual(self.findings(MACOS_PREFIX + b"/x " + LINUX_PREFIX
                                       + b"  \n"), [])

    def test_redaction_demonstrating_lines_pass_deviation_1_rb(self) -> None:
        self.assertEqual(self.findings(DEMONSTRATION_SAME_LINE), [])
        self.assertEqual(self.findings(DEMONSTRATION_KEYED_LINE), [])
        self.assertEqual(self.findings(DEMONSTRATION_EVIDENCE), [])
        self.assertEqual(self.findings(DEMONSTRATION_ADJACENT), [])

    def test_demonstration_coverage_stops_one_line_away(self) -> None:
        self.assertEqual(self.findings(BOUNDARY_TEXT), [(3, "macOS home path")])

    def test_line_numbers_and_one_finding_per_line(self) -> None:
        self.assertEqual(self.findings(LINE_NUMBER_INPUT),
                         [(3, "Windows home path"), (4, "macOS home path")])

    def test_binary_content_is_skipped_by_scan_files_not_here(self) -> None:
        self.assertNotEqual(self.findings(BINARY_INPUT), [])


class ScanFilesTest(unittest.TestCase):
    """check_disclosure.scan_files over a scratch tree; exit semantics."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)

    def write(self, name: str, data: bytes) -> Path:
        path = Path(self.tmp.name, name)
        path.write_bytes(data)
        return path

    def test_findings_are_reported_per_file_and_line(self) -> None:
        self.write("a.md", b"clean\n" + INPUT_MACOS)
        self.write("b.md", LINUX_PREFIX + PEER + b"/y\n")
        findings, scanned, unreadable = check_disclosure.scan_files(
            ["a.md", "b.md"], base=Path(self.tmp.name))
        self.assertEqual(findings, ["a.md:2: macOS home path",
                                    "b.md:1: Linux home path"])
        self.assertEqual((scanned, unreadable), (2, 0))

    def test_unreadable_tracked_input_is_a_tool_error(self) -> None:
        path = self.write("locked.md", INPUT_LINUX)
        path.chmod(0o000)
        self.addCleanup(path.chmod, 0o644)
        findings, _, unreadable = check_disclosure.scan_files(
            ["locked.md"], base=Path(self.tmp.name))
        self.assertEqual(unreadable, 1)
        self.assertEqual(findings,
                         ["locked.md: unreadable input (category: unreadable)"])

    def test_binary_content_is_skipped(self) -> None:
        self.write("blob.bin", BINARY_INPUT)
        findings, scanned, unreadable = check_disclosure.scan_files(
            ["blob.bin"], base=Path(self.tmp.name))
        self.assertEqual((findings, scanned, unreadable), ([], 1, 0))

    def test_no_content_in_findings(self) -> None:
        finds = check_disclosure.scan_bytes(SECRET_INPUT)
        rendered = "".join(f"{no}:{category}" for no, category in finds)
        self.assertEqual(finds, [(1, "Windows home path")])
        self.assertNotIn("secret-dir", rendered)
        self.assertNotIn(ACCOUNT.decode(), rendered)


if __name__ == "__main__":
    unittest.main()
