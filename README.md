# prepare-commit-msg

An intelligent, zero-friction Git `prepare-commit-msg` hook written in Go. It inspects your **staged** changes and leverages fast LLMs—including Google Gemini, OpenAI, Anthropic Claude, xAI Grok, Kilo Gateway (Kilo Code), OpenCode, Hugging Face, Together AI, or a local **Ollama** instance—to automatically generate clean, structured [Conventional Commit](https://www.conventionalcommits.org/) messages.

---

## Table of Contents

- [Why This Exists](#why-this-exists)
- [How It Works & Design](#how-it-works--design)
- [Supported Providers & Models](#supported-providers--models)
- [Environment Variables Reference](#environment-variables-reference)
- [Quick Start & Installation](#quick-start--installation)
  - [1. Download or Build the Binary](#1-download-or-build-the-binary)
  - [2. Install as a Git Hook](#2-install-as-a-git-hook)
  - [3. Configure Provider & Model](#3-configure-provider--model)
- [Configuration](#configuration)
  - [Interactive Wizard](#interactive-wizard)
  - [Non-Interactive CLI Flags](#non-interactive-cli-flags)
  - [Configuration File](#configuration-file)
- [Daily Workflow & Hook Behavior](#daily-workflow--hook-behavior)
- [Self-Update](#self-update)
- [Changes from the mcplib Releases](#changes-from-the-mcplib-releases)
- [Changes with go-llmprovider-sdk v1.2.1](#changes-with-go-llmprovider-sdk-v121)
- [CLI Reference](#cli-reference)
- [Developer Experience](#developer-experience)

---

## Why This Exists

Crafting meaningful commit messages that adhere to conventional commit standards can interrupt developer flow. `prepare-commit-msg` automates this process without taking away control:

- **Consistent Commit Quality:** Generates standardized `type(scope): description` headers with concise bullet points detailing technical changes and file statistics.
- **Zero-Friction Workflow:** Runs automatically during `git commit`. Your editor opens with the message pre-filled and ready for review, edit, or approval.
- **Privacy & Flexibility:** Run fully offline with local **Ollama** models, route through developer gateways (**Kilo Code**, **OpenCode**, **Hugging Face**), or use direct cloud APIs (**Gemini**, **OpenAI**, **Claude**, **Grok**).
- **Fail-Safe Reliability:** If an LLM call times out, authentication fails, or network is down, the hook **never blocks your commit** (soft-fail). It leaves Git's standard editor untouched so you can type manually.

---

## How It Works & Design

```text
git commit
   │
   ▼
[prepare-commit-msg Hook]
   ├─ Check commit source (skips merges, squashes, amends, -m flags)
   ├─ Inspect staged changes (`git diff --cached`)
   ├─ Calculate file statistics & summarize additions/deletions
   ├─ Query primary LLM (Cloud / Gateway / Local) with auto-fallbacks & retries
   ├─ Clean output (strips markdown fences, conversational filler)
   └─ Atomically write commit message (preserves Git comments below)
   │
   ▼
Commit Editor Opens (Pre-populated & ready)
```

- **Staged Diff Analysis:** Runs `git diff --cached --numstat -p` with UTF-8 path preservation and diff size limits (`max_diff_bytes`) to keep prompt sizes fast and deterministic.
- **Curated Fast Models:** Prioritizes low-latency, high-throughput models (e.g., `gemini-3.7-flash`, `gpt-4.1-mini`, `claude-haiku-4-5`, `kilo-auto/free`) over slow reasoning tiers.
- **Multi-Model Fallback Chain:** Automatically tries up to 3 configured fallback models if the primary model is rate-limited or unavailable.
- **Soft-Fail Guarantee:** Any runtime error prints a diagnostic warning to `stderr` and exits with code `0`. Git continues normally and opens your editor.

---

## Supported Providers & Models

`prepare-commit-msg` supports 10 distinct LLM backends through [go-llmprovider-sdk](https://github.com/maccavelli/go-llmprovider-sdk):

| Provider ID | Provider Name | Type / Endpoint | Default / Curated Models | Notes |
| :--- | :--- | :--- | :--- | :--- |
| **`gemini`** | Google Gemini | Cloud API | `gemini-3.7-flash` (recommended), `gemini-3.6-flash`, `gemini-3.5-flash`, `gemini-3.5-flash-lite`, `gemini-2.5-flash`, `gemini-2.5-flash-lite` | Frontier coding intelligence, sub-second latency. |
| **`openai`** | OpenAI | Cloud API | `gpt-4.1-mini` (recommended), `gpt-4.1-nano`, `gpt-4o-mini`, `gpt-4.1`, `gpt-4o`, `o4-mini` | Responses API with structured output. |
| **`claude`** | Anthropic Claude | Cloud API | `claude-haiku-4-5` (recommended), `claude-sonnet-5`, `claude-sonnet-4-6`, `claude-opus-4-8` | Messages API with low latency. |
| **`grok`** | xAI Grok | Cloud API | `grok-4.6` (recommended), `grok-4.5`, `grok-3-mini-fast`, `grok-3-mini`, `grok-4`, `grok-4-fast-reasoning` | High-speed xAI Responses API. |
| **`kilo`** | Kilo Gateway | Developer Gateway<br>`https://api.kilo.ai/api/gateway` | `deepseek/deepseek-v4.1-flash` (recommended), `z-ai/glm-5.3-flash`, `google/gemini-3.8-flash`, `google/gemini-3.6-flash`, `meta/muse-spark-1.2`, `thinkingmachines/inkling` | API behind the **Kilo Code** agent; free models & managed tiers. |
| **`ollama`** | Ollama | **Local / Offline**<br>`http://localhost:11434` | Dynamically discovered from local installation (e.g. `qwen2.5-coder`, `llama3.2`, `mistral`) | **Runs 100% locally on your machine.** No API key required. |
| **`opencode-zen`** | OpenCode Zen | Gateway<br>`https://opencode.ai/zen/v1` | `deepseek-v4.1-flash` (recommended), `qwen3.8-flash`, `glm-5.3-flash`, `deepseek-v4-flash`, `gemini-3.5-flash-lite`, `gemini-3.8-flash` | Pay-as-you-go gateway; multi-protocol dispatch. |
| **`opencode-go`** | OpenCode Go | Gateway<br>`https://opencode.ai/zen/go/v1` | `mimo-v2.6-flash` (recommended), `qwen3.8-flash`, `glm-5.3-flash`, `gpt-6-luna`, `mimo-v2.6-pro`, `hy3` | Subscription gateway. |
| **`huggingface`** | Hugging Face | Router Proxy<br>`https://router.huggingface.co/v1` | `deepseek-ai/DeepSeek-V4-Flash-0731` (recommended), `zai-org/GLM-5.3-Flash`, `deepseek-ai/DeepSeek-V4.1-Flash`, `thinkingmachines/Inkling-Small`, `stepfun-ai/Step-3.7-Flash`, `stepfun-ai/Step-3.5-Flash` | Routing proxy across 18 partner inference backends. |
| **`together`** | Together AI | Cloud API<br>`https://api.together.ai/v1` | `deepseek-ai/DeepSeek-V4.1-Flash` (recommended), `zai-org/GLM-5.3`, `moonshotai/Kimi-K3`, `MiniMaxAI/MiniMax-M3`, `Qwen/Qwen3.6-Plus`, `openai/gpt-oss-120b` | Pay-as-you-go. |

---

## Environment Variables Reference

You can supply credentials and configure behavior using standard environment variables:

### LLM Provider Credentials

| Environment Variable | Target Provider | Description |
| :--- | :--- | :--- |
| **`GEMINI_API_KEY`** | `gemini` | Google AI Studio API key |
| **`OPENAI_API_KEY`** | `openai` | OpenAI platform API key |
| **`ANTHROPIC_API_KEY`** | `claude` | Anthropic Claude API key (`CLAUDE_API_KEY` is read when this is unset) |
| **`XAI_API_KEY`** | `grok` | xAI Grok API key |
| **`KILO_API_KEY`** | `kilo` | Kilo Gateway (Kilo Code) API key |
| **`OPENCODE_API_KEY`** | `opencode-zen`, `opencode-go` | OpenCode API key (serves both Zen and Go gateways) |
| **`HF_TOKEN`** | `huggingface` | Hugging Face user access token |
| **`TOGETHER_API_KEY`** | `together` | Together AI API key |

> **Note on Local Ollama:** The `ollama` provider requires **no credentials** and ignores API keys.

### GitHub Releases & Self-Update

| Environment Variable | Purpose | Description |
| :--- | :--- | :--- |
| **`GH_TOKEN`** / **`GITHUB_TOKEN`** | `prepare-commit-msg update` | Optional GitHub Personal Access Token used during self-update checks and binary downloads to prevent GitHub API rate-limiting (raises limit from 60 to 5,000 req/hr). |

### Precedence

1. **Config File Key (`api_key` in `config.json`):** Highest priority when non-empty.
2. **Environment Variable:** Used when config key is empty and `--no-env` is not specified.
3. **`--no-env` Flag:** Strictly disables environment variable lookup.

---

## Quick Start & Installation

### 1. Download or Build the Binary

#### Pre-built Binaries (GitHub Releases)
Download the latest executable for your OS and architecture from [Releases](https://github.com/maccavelli/prepare-commit-msg/releases):
- **Linux:** `prepare-commit-msg-linux-amd64` / `prepare-commit-msg-linux-arm64`
- **macOS:** `prepare-commit-msg-darwin-amd64` / `prepare-commit-msg-darwin-arm64`
- **Windows:** `prepare-commit-msg-windows-amd64.exe` / `prepare-commit-msg-windows-arm64.exe`

#### Or Build from Source
```bash
make build          # Builds local binary into dist/
make install        # Compiles and installs binary to ~/.global-git-hooks/prepare-commit-msg
```

---

### 2. Install as a Git Hook

#### Option A: Global Git Hook (Recommended for all repositories)
Configure Git to use a global hooks directory so all current and future repositories automatically use the hook:

```bash
# 1. Create a global hooks directory
mkdir -p ~/.global-git-hooks

# 2. Place binary into the directory
cp dist/prepare-commit-msg-$(go env GOOS)-$(go env GOARCH) ~/.global-git-hooks/prepare-commit-msg
chmod +x ~/.global-git-hooks/prepare-commit-msg

# 3. Point Git at the global hooks directory
git config --global core.hooksPath ~/.global-git-hooks
```

#### Option B: Per-Repository Hook
Symlink or copy the binary into a single repository's `.git/hooks/` directory:

```bash
ln -sf /path/to/prepare-commit-msg .git/hooks/prepare-commit-msg
chmod +x .git/hooks/prepare-commit-msg
```

#### Windows Installation
1. Place `prepare-commit-msg.exe` in `%USERPROFILE%\.global-git-hooks\prepare-commit-msg.exe`.
2. Configure Git:
   ```cmd
   git config --global core.hooksPath %USERPROFILE%\.global-git-hooks
   ```
   *(Git for Windows automatically searches for matching `.exe` or shell wrappers).*

---

### 3. Configure Provider & Model

Run the interactive setup wizard:

```bash
prepare-commit-msg configure
```

Select your preferred provider (`gemini`, `openai`, `claude`, `grok`, `kilo`, `ollama`, `opencode-zen`, `opencode-go`, `huggingface`), detect or enter credentials, and pick primary and fallback models.

---

## Configuration

### Interactive Wizard

```bash
prepare-commit-msg configure
```

The wizard guides you through:
1. Provider selection across all 10 supported cloud, gateway, and local backends.
2. Endpoint resolution (for Ollama or custom gateway endpoints).
3. API Key detection from environment or prompt entry; a saved sign-in is
   offered first.
4. Live model discovery from the provider API.
5. Fallback model multi-selection (up to 3 fallbacks), with your saved
   fallbacks preselected.
6. Operational settings (timeout, max diff size, retry count, retry delay).

For OpenAI and Grok, the interactive wizard can also use ChatGPT or Grok
subscription authentication. OAuth sessions are stored separately from the
main configuration in its private `oauth/` directory; access and refresh tokens
are never copied into `config.json`. Non-interactive `configure --yes` remains
API-key-only.

Choosing **Use the Codex CLI login** or **Use the Grok CLI login** stores only
the path to that CLI's `auth.json` (`vendor_auth_path` in `config.json`). The
file is read on every run and never refreshed here: the CLI keeps its own
login current. When it expires, sign in again with the CLI; there is no need
to re-run `configure`.

---

### Non-Interactive CLI Flags

Configure settings directly or in automated scripts using `--yes`:

```bash
# Example 1: Google Gemini (default cloud)
prepare-commit-msg configure --yes \
  --provider gemini \
  --model gemini-3.7-flash \
  --fallback gemini-3.6-flash \
  --fallback gemini-3.5-flash

# Example 2: Local Ollama (no API key needed)
prepare-commit-msg configure --yes \
  --provider ollama \
  --model qwen2.5-coder:7b

# Example 3: Kilo Gateway (Kilo Code)
prepare-commit-msg configure --yes \
  --provider kilo \
  --model kilo-auto/free \
  --fallback kilo-auto/small

# Example 4: Anthropic Claude
prepare-commit-msg configure --yes \
  --provider claude \
  --model claude-haiku-4-5 \
  --fallback claude-sonnet-5
```

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--provider` | `string` | `gemini` | `gemini`, `openai`, `claude`, `grok`, `kilo`, `ollama`, `opencode-zen`, `opencode-go`, `huggingface` |
| `--model` | `string` | *Provider default* | Primary model name |
| `--api-key` | `string` | `""` | Provider API key (uses existing config or env if omitted; ignored by `ollama`) |
| `--fallback` | `string` | *Curated fast models* | Fallback model (repeatable or comma-separated, max 3) |
| `--timeout-seconds`| `int` | `120` | Total LLM request timeout |
| `--max-diff-bytes` | `int` | `32000` (32 KB) | Maximum staged diff bytes sent in the prompt |
| `--retry-count` | `int` | `3` | Number of retry attempts per model |
| `--retry-delay-seconds` | `int` | `3` | Base delay between retries |
| `--no-env` | `bool` | `false` | Do not read API keys from environment variables |
| `--yes` | `bool` | `false` | Apply non-interactively without prompting |

---

### Configuration File

The configuration is saved as JSON with secure permissions (`0600`) at standard platform locations:

| OS | Config Path |
| :--- | :--- |
| **Linux** | `~/.config/prepare-commit-msg/config.json` (or `$XDG_CONFIG_HOME`) |
| **macOS** | `~/Library/Application Support/prepare-commit-msg/config.json` |
| **Windows** | `%AppData%\prepare-commit-msg\config.json` |

#### Example `config.json`

```json
{
  "active_provider": "gemini",
  "providers": {
    "gemini": {
      "api_key": "",
      "model": "gemini-3.7-flash",
      "fallback_models": ["gemini-3.6-flash", "gemini-3.5-flash"]
    },
    "kilo": {
      "api_key": "",
      "model": "kilo-auto/free",
      "fallback_models": ["kilo-auto/small"]
    },
    "ollama": {
      "api_key": "",
      "model": "qwen2.5-coder:7b",
      "fallback_models": ["llama3.2:latest"]
    },
    "openai": {
      "api_key": "",
      "model": "gpt-4.1-mini",
      "fallback_models": ["gpt-4o-mini"]
    },
    "claude": {
      "api_key": "",
      "model": "claude-haiku-4-5",
      "fallback_models": ["claude-sonnet-5"]
    }
  },
  "timeout_seconds": 120,
  "max_diff_bytes": 32000,
  "retry_count": 3,
  "retry_delay_seconds": 3
}
```

---

## Daily Workflow & Hook Behavior

Once installed, simply use Git as usual:

```bash
git add src/
git commit
```

1. The hook analyzes your staged changes and queries the configured provider.
2. Your default Git editor opens with the generated message:
   ```text
   feat(auth): add OAuth2 token refresh flow

   * Add refresh token expiration check in session manager
   * Implement backoff retry on 401 unauthorized responses
   * Update unit tests covering expired token renewal

   # AI-generated (3 files: +84 -12)
   # Go: 2, JSON: 1
   #
   # Please enter the commit message for your changes. Lines starting
   # with '#' will be ignored, and an empty message aborts the commit.
   ```
3. Edit the message or save and close the editor to complete the commit.

### Hook Trigger Matrix

Git invokes the hook as `prepare-commit-msg <file> [source] [sha]`. The hook evaluates whether to generate a message:

| Commit Command / Source | Hook Action | Rationale |
| :--- | :--- | :--- |
| `git commit` (empty source) | **Generates message** | Standard commit flow; populates empty editor buffer. |
| `git commit -t <template>` (`template`) | **Generates message** | Runs if template only contains comments or whitespace. |
| `git commit -m "..."` (`message`) | **Skipped** | Explicit message supplied by user. |
| `git commit -F <file>` (`message`) | **Skipped** | Explicit message supplied from file. |
| `git merge` (`merge`) | **Skipped** | Merge commit message generated by Git. |
| `git commit --amend` (`commit`) | **Skipped** | Preserves existing commit message. |
| `git commit` with no staged changes | **Skipped** | Nothing to analyze. |

---

## Self-Update

Keep your binary up-to-date with new LLM models and improvements:

```bash
prepare-commit-msg update [flags]
```

| Command | Purpose |
| :--- | :--- |
| `prepare-commit-msg update --check` | Check for newer releases without modifying binary (exits code `10` when update available). |
| `prepare-commit-msg update` | Interactive update check and prompt to apply. |
| `prepare-commit-msg update --yes` | Non-interactive in-place update. |
| `prepare-commit-msg update --version v1.2.0 --yes` | Pin or rollback to a specific release version. |
| `prepare-commit-msg update --force --yes` | Force reinstall / overwrite local or dev builds. |
| `prepare-commit-msg update --dry-run` | Download and verify the release; install nothing. |
| `prepare-commit-msg update --json` | Write JSON Lines to stdout, ending with one `{"kind":"result",…}` object whose `result` carries `"schema_version":2`. |
| `prepare-commit-msg update --channel rc` | Follow a prerelease channel. |

Progress and the confirmation prompt go to **stderr**; stdout carries only `--json` output. The exit status is `0` when up to date, declined or installed, `10` when an update is available (with `--check`), and `1` on an error before the new binary is in place. An error after that, such as failing to release the update lock, is reported as a `warning:` line on stderr, and the run still exits `0`.

> **Security Note:** Self-update only modifies regular binaries located within the user's home directory. System paths (`/usr`, Homebrew Cellar, Nix store) are rejected.

---

## Changes from the mcplib Releases

Releases before the move to go-llmprovider-sdk and go-selfupdate-lib
([0008-MADR](docs/decisions/0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md))
behaved differently in these ways:

- **`update`** writes progress and its prompt to stderr, not stdout, and a
  failure reads `update failed:` rather than `Update failed:`.
- **`version`** prints `vX.Y.Z (release) <revision>` rather than
  `X.Y.Z (release)`. Scripts that parse it must change.
- **Retries** cover rate limits (not an exhausted quota), an unavailable
  service, and a failure to connect. Other errors return at once, and a retry
  is no longer logged.
- **Error messages** begin `llmprovider:` rather than `llm:`.
- **Model-metadata overrides** are read from `LLMPROVIDER_*` variables, not
  `MCPLIB_*`.
- **Grok's OAuth issuer and client overrides** are no longer read during
  `configure`.
- **New:** the `together` provider, Kilo device logins with an organization,
  and generation on any provider's sign-in.

---

## Changes with go-llmprovider-sdk v1.2.1

Release `v1.6.0` moves from go-llmprovider-sdk `v1.0.0` to `v1.2.1`
([0009-MADR](docs/decisions/0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md)).
What you may notice:

- **A refused answer is never the message.** When a model refuses, the hook
  says so and tries the next model.
- **"Not permitted" tries the next model.** A model your account may not use,
  such as an OpenCode or Kilo HTTP 403, no longer ends the run. A rejected key
  or sign-in still does, without trying the fallbacks.
- **A rejected Gemini key stops at once,** instead of failing every fallback
  model.
- **Retries.** An OpenAI or Claude HTTP 409 is retried. A reply cut off after
  it began is retried once. A whole but empty reply is not sent again. No
  retry waits past the run's timeout.
- **`configure --yes`** falls back to the curated models when the live listing
  recommends none.
- **The interactive wizard** offers a saved sign-in first and preselects your
  saved fallbacks. It checks a typed endpoint URL, and asks before plain
  `http` to a remote host. It checks a pasted OpenAI credential, and masked
  entry handles pasting, Escape and Ctrl-C.
- **The OAuth directory** must be a real directory you own on Linux and
  macOS, and loses any group or other write permission. On Windows it is
  restricted to your account.

---

## CLI Reference

```text
Usage:
  prepare-commit-msg configure [flags]                - run setup wizard or non-interactive configure
  prepare-commit-msg update [flags]                   - check for and apply updates from GitHub
  prepare-commit-msg version                          - show binary version, as vX.Y.Z (release) <revision>
  prepare-commit-msg help                             - display help message
  prepare-commit-msg <commit_msg_file> [source] [sha] - run as git prepare-commit-msg hook
```

---

## Developer Experience

Brief guide for contributors and local development.

### Toolchain & Quality Gates

The repository uses pinned developer tools in `.tools/bin` to ensure identical results across local machines and CI. Building from source needs Go 1.27.1.

```bash
make tools          # Bootstrap pinned tools (golangci-lint, govulncheck, actionlint)
make verify         # Run complete quality contract: mod-check, fmt-check, lint, vet, test, coverage, vuln, workflow-lint, build-all
```

### Build & Test Targets

| Target | Description |
| :--- | :--- |
| `make build` | Compile static binary for current OS/architecture (`dist/`) |
| `make build-all` | Cross-compile for all 6 target platforms (Linux, macOS, Windows on AMD64 and ARM64) |
| `make test` | Run tests with race detector (`go test -race -v ./...`) |
| `make coverage` | Verify statement coverage meets minimum threshold (80.0%) |
| `make lint` | Run `golangci-lint` with fleet configuration |
| `make fmt` | Format source files and imports |
| `make hooks-install` | Install composable repository-local Git hooks |
| `make hooks-test` | Test hook composition against temporary repositories |
| `make clean` | Clean build outputs from `dist/` |

### Maintainer Release

A release is a pushed `v*` tag on a commit whose CI has passed on `main`:

```bash
git tag -a v1.2.3 -m "v1.2.3"
git push origin v1.2.3
```

The tag's CI run builds the six binaries and checks them with `make verify-release VERSION=v1.2.3`, which a maintainer can also run locally first. Once the native tests pass on Linux, macOS and Windows, go-selfupdate-lib's reusable workflow publishes the GitHub Release.

Assets include cross-compiled binaries, `SHA256SUMS`, and build-provenance attestations:
```bash
sha256sum --check SHA256SUMS
gh attestation verify prepare-commit-msg-linux-amd64 --repo maccavelli/prepare-commit-msg
```

---

## License

Part of the split Go repository fleet.
