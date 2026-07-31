package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Info holds the staged change metadata collected from git for prompt construction.
type Info struct {
	Files     []string
	Stats     string
	Additions int
	Deletions int
	Diff      string
}

// lookPath is a package variable to allow mocking in tests.
var lookPath = exec.LookPath

// GatherInfo collects staged file names, diff stats, and the unified diff from git.
// A missing git binary returns an error. An empty staged set returns empty Info and nil.
// Other git failures return an error (callers may soft-fail).
func GatherInfo(ctx context.Context, maxDiffBytes int) (*Info, error) {
	info := &Info{}

	gitBin, err := lookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git not found in PATH: %w (install Git and ensure it is on PATH)", err)
	}

	// -c core.quotepath=false: readable non-ASCII paths on all platforms.
	// stdout only: stderr warnings (CRLF tips, etc.) must not corrupt the parse.
	cmd := newGitCmdContext(ctx, gitBin,
		"-c", "core.quotepath=false",
		"diff", "--cached", "--numstat", "-p", "--unified=3",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Empty staged tree can still exit 0; non-zero is a real failure
		// (not a repo, index lock, etc.).
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git diff --cached failed: %s", msg)
	}

	raw := normalizeNewlines(stdout.String())
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return info, nil
	}

	var counts struct {
		yaml, json, tf, ci, script, other int
	}

	diffStartIndex := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			diffStartIndex = i
			break
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		// Binary files show "-" for add/del in numstat.
		add := parseNumstatCount(parts[0])
		del := parseNumstatCount(parts[1])
		f := strings.Join(parts[2:], " ")

		info.Additions += add
		info.Deletions += del
		info.Files = append(info.Files, f)

		lower := strings.ToLower(f)
		switch {
		case strings.HasSuffix(lower, ".yaml"), strings.HasSuffix(lower, ".yml"):
			counts.yaml++
		case strings.HasSuffix(lower, ".json"):
			counts.json++
		case strings.HasSuffix(lower, ".tf"), strings.HasSuffix(lower, ".tfvars"):
			counts.tf++
		case strings.Contains(lower, "gitlab-ci"), strings.Contains(lower, "jenkinsfile"):
			counts.ci++
		case strings.HasSuffix(lower, ".sh"), strings.HasSuffix(lower, ".py"),
			strings.HasSuffix(lower, ".js"), strings.HasSuffix(lower, ".go"),
			strings.HasSuffix(lower, ".ts"), strings.HasSuffix(lower, ".ps1"),
			strings.HasSuffix(lower, ".bat"), strings.HasSuffix(lower, ".cmd"):
			counts.script++
		default:
			counts.other++
		}
	}

	var stats []string
	if counts.yaml > 0 {
		stats = append(stats, fmt.Sprintf("YAML: %d", counts.yaml))
	}
	if counts.json > 0 {
		stats = append(stats, fmt.Sprintf("JSON: %d", counts.json))
	}
	if counts.tf > 0 {
		stats = append(stats, fmt.Sprintf("Terraform: %d", counts.tf))
	}
	if counts.ci > 0 {
		stats = append(stats, fmt.Sprintf("CI/CD: %d", counts.ci))
	}
	if counts.script > 0 {
		stats = append(stats, fmt.Sprintf("Scripts: %d", counts.script))
	}
	if counts.other > 0 {
		stats = append(stats, fmt.Sprintf("Other: %d", counts.other))
	}
	info.Stats = strings.Join(stats, ", ")

	if diffStartIndex != -1 {
		diffStr := strings.Join(lines[diffStartIndex:], "\n")
		info.Diff = TruncateUTF8(diffStr, maxDiffBytes)
	}

	return info, nil
}

func parseNumstatCount(s string) int {
	if s == "-" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// TruncateUTF8 shortens s to at most maxBytes without splitting a UTF-8 rune.
// Prefers cutting at the last newline before the limit. Appends a marker when truncated.
func TruncateUTF8(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	// Back up to rune boundary.
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	// Prefer line boundary for cleaner LLM prompts.
	if i := strings.LastIndex(s[:cut], "\n"); i > maxBytes/4 {
		cut = i
	}
	return s[:cut] + "\n\n[diff truncated]"
}

// normalizeNewlines converts CRLF/CR to LF.
func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// IsCommitMsgEmpty returns true if the commit message file contains only blank lines and comments.
func IsCommitMsgEmpty(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	text := normalizeNewlines(string(data))
	for line := range strings.SplitSeq(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return false
		}
	}
	return true
}
