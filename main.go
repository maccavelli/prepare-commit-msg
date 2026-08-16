package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/maccavelli/prepare-commit-msg/internal/config"
	"github.com/maccavelli/prepare-commit-msg/internal/fsutil"
	"github.com/maccavelli/prepare-commit-msg/internal/git"
	"github.com/maccavelli/prepare-commit-msg/internal/selfupdate"
	"github.com/maccavelli/prepare-commit-msg/internal/ui"

	"github.com/maccavelli/mcplib/llmprovider"
)

var generateWithRetry = llmprovider.GenerateWithRetry
var osGetenv = os.Getenv

// Version is overwritten by build flags during the compilation process.
var Version = "4.3.2"

const (
	// AppTitle is the name of the application used in help text and version output.
	AppTitle = "prepare-commit-msg"

	// minMessageRunes is the minimum acceptable cleaned AI message length.
	minMessageRunes = 5

	// gitGatherTimeout bounds the git subprocess used for staged diffs.
	gitGatherTimeout = 30 * time.Second
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s configure [flags]              - run setup wizard or non-interactive configure\n", AppTitle)
	fmt.Fprintf(os.Stderr, "  %s update [flags]                 - check for and apply updates from GitHub\n", AppTitle)
	fmt.Fprintf(os.Stderr, "  %s version                        - show version\n", AppTitle)
	fmt.Fprintf(os.Stderr, "  %s help                           - show this help\n", AppTitle)
	fmt.Fprintf(os.Stderr, "  %s <commit_msg_file> [source] [sha] - run as git prepare-commit-msg hook\n", AppTitle)
	fmt.Fprintf(os.Stderr, "\nConfigure flags:\n")
	fmt.Fprintf(os.Stderr, "  --provider string         gemini|openai|claude\n")
	fmt.Fprintf(os.Stderr, "  --model string            primary model name\n")
	fmt.Fprintf(os.Stderr, "  --api-key string          API key (or use provider env var)\n")
	fmt.Fprintf(os.Stderr, "  --fallback string         fallback model (repeatable, max %d)\n", config.MaxFallbacks)
	fmt.Fprintf(os.Stderr, "  --timeout-seconds int     LLM overall timeout (default %d)\n", config.DefaultTimeoutSeconds)
	fmt.Fprintf(os.Stderr, "  --max-diff-bytes int      max staged diff bytes (default %d)\n", config.DefaultMaxDiffBytes)
	fmt.Fprintf(os.Stderr, "  --retry-count int         retries per model (default %d)\n", config.DefaultRetryCount)
	fmt.Fprintf(os.Stderr, "  --retry-delay-seconds int base retry delay (default %d)\n", config.DefaultRetryDelaySeconds)
	fmt.Fprintf(os.Stderr, "  --no-env                  do not read API keys from the environment\n")
	fmt.Fprintf(os.Stderr, "  --yes                     non-interactive configure (no prompts)\n")
	fmt.Fprintf(os.Stderr, "\nUpdate flags:\n")
	fmt.Fprintf(os.Stderr, "  --check                   check if update is available without applying\n")
	fmt.Fprintf(os.Stderr, "  --force                   reinstall or force overwrite current binary\n")
	fmt.Fprintf(os.Stderr, "  --version string          target specific release tag (e.g. v4.4.0)\n")
	fmt.Fprintf(os.Stderr, "  --yes, -y                 non-interactive update\n")
}

var osExit = os.Exit

// softFail prints a user-friendly message and exits 0 so the commit is not blocked.
func softFail(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s: %s\n", AppTitle, msg)
	fmt.Fprintf(os.Stderr, "%s: commit editor left unchanged — type a message manually or run: %s configure\n", AppTitle, AppTitle)
	osExit(0)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		osExit(1)
	}

	switch args[0] {
	case "version", "--version", "-V":
		fmt.Printf("%s version %s\n", AppTitle, strings.TrimPrefix(Version, "v"))
		return
	case "help", "--help", "-h":
		printUsage()
		return
	case "configure":
		if err := runConfigure(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			osExit(1)
		}
		return
	case "update":
		if err := runUpdate(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			osExit(1)
		}
		return
	}

	// Git hook logic
	runHook(args)
}

func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	check := fs.Bool("check", false, "check for updates without applying")
	force := fs.Bool("force", false, "reinstall current version or force overwrite")
	targetVersion := fs.String("version", "", "target specific version tag (e.g. v4.4.0)")
	yes := fs.Bool("yes", false, "non-interactive update")
	fs.BoolVar(yes, "y", false, "non-interactive update (shorthand)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, err := selfupdate.Run(ctx, selfupdate.Options{
		CurrentVersion: Version,
		TargetVersion:  *targetVersion,
		CheckOnly:      *check,
		Force:          *force,
		Output:         os.Stdout,
	})
	return err
}

func runConfigure(args []string) error {
	fs := flag.NewFlagSet("configure", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	provider := fs.String("provider", "", "LLM provider: gemini, openai, or claude")
	model := fs.String("model", "", "primary model name")
	apiKey := fs.String("api-key", "", "API key (optional if env var or config exists)")
	timeout := fs.Int("timeout-seconds", 0, "LLM timeout in seconds")
	maxDiff := fs.Int("max-diff-bytes", 0, "max diff bytes for the prompt")
	retryCount := fs.Int("retry-count", 0, "retries per model")
	retryDelay := fs.Int("retry-delay-seconds", 0, "base retry delay seconds")
	noEnv := fs.Bool("no-env", false, "do not read API keys from the environment")
	yes := fs.Bool("yes", false, "non-interactive mode")

	var fallbacks multiFlag
	fs.Var(&fallbacks, "fallback", "fallback model (repeatable, max 3)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	conf, err := config.Load()
	if err != nil {
		return err
	}

	opts := ui.SetupOptions{
		Provider:          *provider,
		APIKey:            *apiKey,
		Model:             *model,
		Fallbacks:         config.ClampFallbacks([]string(fallbacks)),
		TimeoutSeconds:    *timeout,
		MaxDiffBytes:      *maxDiff,
		RetryCount:        *retryCount,
		RetryDelaySeconds: *retryDelay,
		NoEnv:             *noEnv,
		Yes:               *yes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return ui.RunSetupWithOptions(ctx, conf, opts, os.Stdin)
}

// multiFlag collects repeated --fallback values.
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	for part := range strings.SplitSeq(v, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*m = append(*m, part)
		}
	}
	return nil
}

func runHook(args []string) {
	commitMsgFile := args[0]

	if _, err := os.Stat(commitMsgFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Unknown command or file does not exist: '%s'\n\n", commitMsgFile)
		printUsage()
		osExit(1)
	}

	commitSource := ""
	if len(args) > 1 {
		commitSource = args[1]
	}

	if shouldSkipSource(commitSource) {
		return
	}

	if !git.IsCommitMsgEmpty(commitMsgFile) {
		return
	}

	conf, err := config.Load()
	if err != nil {
		softFail("could not load config: %v", err)
	}

	gitCtx, gitCancel := context.WithTimeout(context.Background(), gitGatherTimeout)
	defer gitCancel()

	info, err := git.GatherInfo(gitCtx, conf.MaxDiffBytes)
	if err != nil {
		softFail("could not gather git info: %v", err)
	}
	if info == nil || len(info.Files) == 0 {
		// Nothing staged — leave the editor alone.
		return
	}

	if err := runAnalyzer(commitMsgFile, conf, info); err != nil {
		softFail("could not generate a message: %v", err)
	}
}

// shouldSkipSource reports whether AI generation should be skipped for this
// git prepare-commit-msg source. Empty and "template" are allowed when the
// message file is comment-only; message/merge/squash/commit are always skipped.
func shouldSkipSource(source string) bool {
	if source == "" || source == "template" {
		return false
	}
	switch source {
	case "message", "merge", "squash", "commit":
		return true
	default:
		// Unknown sources: be conservative and skip.
		return true
	}
}

// runAnalyzer orchestrates the commit message generation process.
func runAnalyzer(file string, conf *config.Config, info *git.Info) error {
	if info == nil || len(info.Files) == 0 {
		return nil
	}

	pc, err := conf.GetActive()
	if err != nil {
		return err
	}

	apiKey := config.ResolveAPIKey(pc, conf.ActiveProvider, true, osGetenv)
	if err := config.ValidateActive(conf.ActiveProvider, pc, apiKey); err != nil {
		return err
	}

	timeout := time.Duration(conf.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = time.Duration(config.DefaultTimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	prompt := buildPrompt(info)
	fallbacks := config.ClampFallbacks(pc.FallbackModels)
	modelsToTry := slices.Concat([]string{pc.Model}, fallbacks)

	var msg string
	var loopErr error

	for _, m := range modelsToTry {
		fmt.Fprintf(os.Stderr, "%s: generating via %s (%s)...\n", AppTitle, conf.ActiveProvider, m)

		provider, err := llmprovider.NewProvider(conf.ActiveProvider, apiKey, m)
		if err != nil {
			loopErr = fmt.Errorf("failed to init %s: %w", conf.ActiveProvider, err)
			fmt.Fprintf(os.Stderr, "%s: warning: %v\n", AppTitle, loopErr)
			continue
		}

		msg, loopErr = generateWithRetry(ctx, provider, prompt, conf.RetryCount, time.Duration(conf.RetryDelaySeconds)*time.Second)
		if loopErr == nil {
			break
		}
		// Non-retryable auth: do not burn through all fallbacks with the same key.
		if errors.Is(loopErr, llmprovider.ErrAuthFailure) {
			return fmt.Errorf("authentication failed for %s: %w", conf.ActiveProvider, loopErr)
		}
		fmt.Fprintf(os.Stderr, "%s: warning: model %s failed: %v\n", AppTitle, m, loopErr)
	}

	if loopErr != nil {
		return fmt.Errorf("all models for %s failed, last error: %w", conf.ActiveProvider, loopErr)
	}

	msg = cleanLLMOutput(msg)
	if utf8.RuneCountInString(msg) < minMessageRunes {
		return fmt.Errorf("AI message too short or empty after cleaning")
	}

	return writeMessage(file, msg, info)
}

// cleanLLMOutput strips conversational filler and markdown fences.
func cleanLLMOutput(out string) string {
	out = strings.TrimSpace(normalizeNewlines(out))
	if out == "" {
		return ""
	}

	var sb strings.Builder
	sb.Grow(len(out))
	scanner := bufio.NewScanner(strings.NewReader(out))
	// Allow long commit bodies.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	firstLine := true

	for scanner.Scan() {
		line := scanner.Text()
		tl := strings.TrimSpace(line)

		if strings.HasPrefix(tl, "```") {
			continue
		}

		if strings.HasPrefix(tl, "Based on") ||
			strings.HasPrefix(tl, "Generate") ||
			strings.HasPrefix(tl, "Here is") ||
			tl == "---" {
			continue
		}
		if tl == "" && firstLine {
			continue
		}
		if !firstLine {
			sb.WriteString("\n")
		}
		sb.WriteString(line)
		firstLine = false
	}

	return strings.TrimSpace(sb.String())
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// buildPrompt constructs the LLM prompt from staged change metadata.
func buildPrompt(info *git.Info) string {
	var sb strings.Builder
	sb.Grow(500 + len(info.Diff) + len(info.Stats))

	sb.WriteString("Generate a conventional commit message for these changes.\n\n")
	sb.WriteString("IMPORTANT: Return ONLY the commit message. Do not include markdown code fences, conversational filler, or introductory text.\n\n")
	sb.WriteString("FILES CHANGED:\n")
	for _, f := range info.Files {
		sb.WriteString(f)
		sb.WriteString("\n")
	}
	sb.WriteString("\nSTATS: ")
	sb.WriteString(info.Stats)
	sb.WriteString(fmt.Sprintf(" (+%d, -%d)\n\n", info.Additions, info.Deletions))
	sb.WriteString("DIFF:\n")
	sb.WriteString(info.Diff)
	sb.WriteString("\n\n---\n")
	sb.WriteString("Format: type(scope): brief description (max 72 chars)\n")
	sb.WriteString("Body: Concise bullet points of technical changes.")

	return sb.String()
}

// writeMessage writes the generated commit message atomically, preserving
// existing file content (typically Git's comment template) after the message.
func writeMessage(path, msg string, info *git.Info) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing commit message file: %w", err)
	}

	return fsutil.ReplaceFileAtomic(path, func(tmp *os.File) error {
		w := bufio.NewWriter(tmp)
		if _, err := fmt.Fprintln(w, strings.TrimSpace(msg)); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "# AI-generated (%d files: +%d -%d)\n", len(info.Files), info.Additions, info.Deletions); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "# %s\n\n", info.Stats); err != nil {
			return err
		}
		if len(existing) > 0 {
			if _, err := w.Write(existing); err != nil {
				return err
			}
		}
		return w.Flush()
	})
}
