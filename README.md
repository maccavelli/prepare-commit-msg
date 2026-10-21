# prepare-commit-msg

An intelligent, zero-friction Git `prepare-commit-msg` hook written in Go. It inspects your
**staged** changes and uses fast LLMs from Google Gemini, OpenAI, Anthropic Claude, xAI Grok, Kilo
Gateway (Kilo Code), OpenCode, Hugging Face, Together AI, or a local **Ollama** instance. It
automatically generates clean, structured [Conventional
Commit](https://www.conventionalcommits.org/) messages.

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
- [Changes with go-llmprovider-sdk v1.3.2](#changes-with-go-llmprovider-sdk-v132)
- [Changes with go-selfupdate-lib v1.13.0](#changes-with-go-selfupdate-lib-v1130)
- [Changes with go-llmprovider-sdk v1.12.0 and go-selfupdate-lib v1.15.2](#changes-with-go-llmprovider-sdk-v1120-and-go-selfupdate-lib-v1152)
- [Changes with go-llmprovider-sdk v1.15.0](#changes-with-go-llmprovider-sdk-v1150)
- [Changes with Go 1.27.3, go-llmprovider-sdk v1.21.0 and go-selfupdate-lib v1.17.2](#changes-with-go-1273-go-llmprovider-sdk-v1210-and-go-selfupdate-lib-v1172)
- [CLI Reference](#cli-reference)
- [Developer Experience](#developer-experience)
- [Documentation](#documentation)

---

## Why This Exists

Crafting meaningful commit messages that adhere to conventional commit standards can interrupt developer flow. `prepare-commit-msg` automates this process without taking away control:

- **Consistent Commit Quality:** Generates standardized `type(scope): description` headers with concise bullet points detailing technical changes and file statistics.
- **Zero-Friction Workflow:** Runs automatically during `git commit`. Your editor opens with the message pre-filled and ready for review, edit, or approval.
- **Privacy & Flexibility:** Run fully offline with local **Ollama** models,
  route through developer gateways (**Kilo Code**, **OpenCode**, **Hugging
  Face**), or use direct cloud APIs (**Gemini**, **OpenAI**, **Claude**,
  **Grok**).
- **Fail-Safe Reliability:** If an LLM call times out, authentication fails,
  or the network is down, the hook **never blocks your commit** (soft-fail).
  It leaves Git's standard editor untouched so you can type manually.

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
- **Curated Fast Models:** Prioritizes low-latency, high-throughput models (e.g., `gemini-3.5-flash-lite`, `gpt-6-luna`, `claude-haiku-5-5`, `kilo-auto/free`) over slow reasoning tiers.
- **Multi-Model Fallback Chain:** Automatically tries up to 3 configured fallback models if the primary model is rate-limited or unavailable.
- **Soft-Fail Guarantee:** Any runtime error prints a diagnostic warning to `stderr` and exits with code `0`. Git continues normally and opens your editor.

---

## Supported Providers & Models

`prepare-commit-msg` supports 10 distinct LLM backends through [go-llmprovider-sdk](https://github.com/maccavelli/go-llmprovider-sdk):

| Provider ID | Provider Name | Type / Endpoint | Default / Curated Models | Notes |
| :--- | :--- | :--- | :--- | :--- |
| **`gemini`** | Google Gemini | Cloud API | `gemini-3.5-flash-lite` (recommended), `gemini-3.8-flash`, `gemini-3.6-flash` | Frontier coding intelligence, sub-second latency. |
| **`openai`** | OpenAI | Cloud API | `gpt-6-luna` (recommended), `gpt-5.6-luna`, `gpt-4.1-mini`, `gpt-5.6-terra`, `gpt-6.1-sol`, `gpt-6-astra` | Responses API with structured output. |
| **`claude`** | Anthropic Claude | Cloud API | `claude-haiku-5-5` (recommended), `claude-sonnet-5-5`, `claude-sonnet-5`, `claude-opus-5-5`, `claude-opus-4-8` | Messages API with low latency. |
| **`grok`** | xAI Grok | Cloud API | `grok-4.6` (recommended), `grok-4.5`, `grok-4.7`, `grok-4.3`, `grok-build-0.1` | High-speed xAI Responses API. |
| **`kilo`** | Kilo Gateway | Developer Gateway<br>`https://api.kilo.ai/api/gateway` | `deepseek/deepseek-v4.1-flash` (recommended), `z-ai/glm-5.3-flash`, `google/gemini-3.6-flash`, `minimax/minimax-m3`, `nvidia/nemotron-3-ultra-550b-a55b`, `google/gemini-3.8-flash` | API behind the **Kilo Code** agent; free models & managed tiers. |
| **`ollama`** | Ollama | **Local / Offline**<br>`http://localhost:11434` | Dynamically discovered from local installation (e.g. `qwen2.5-coder`, `llama3.2`, `mistral`) | **Runs 100% locally on your machine.** No API key required. |
| **`opencode-zen`** | OpenCode Zen | Gateway<br>`https://opencode.ai/zen/v1` | `claude-haiku-5-5` (recommended), `gpt-6-luna`, `deepseek-v4.1-flash`, `qwen3.8-flash`, `glm-5.3-flash`, `deepseek-v4-flash` | Pay-as-you-go gateway; multi-protocol dispatch. |
| **`opencode-go`** | OpenCode Go | Gateway<br>`https://opencode.ai/zen/go/v1` | `claude-haiku-5-5` (recommended), `mimo-v2.6-flash`, `gpt-6-luna`, `space-bunny`, `qwen3.8-flash`, `glm-5.3-flash` | Subscription gateway. |
| **`huggingface`** | Hugging Face | Router Proxy<br>`https://router.huggingface.co/v1` | `deepseek-ai/DeepSeek-V4-Flash-0731` (recommended), `zai-org/GLM-5.3-Flash`, `tencent/Hy3`, `deepseek-ai/DeepSeek-V4.1-Flash`, `google/gemma-4-26B-A4B-it`, `google/gemma-4-31B-it` | Routing proxy across 18 partner inference backends. |
| **`together`** | Together AI | Cloud API<br>`https://api.together.ai/v1` | `zai-org/GLM-5.3-Flash` (recommended), `deepseek-ai/DeepSeek-V4-Flash-0731`, `deepseek-ai/DeepSeek-V4.1-Flash`, `MiniMaxAI/MiniMax-M3`, `openai/gpt-oss-120b`, `Qwen/Qwen3.6-Plus` | Pay-as-you-go. |

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

#### With the installer (recommended)

From `v1.8.0`, every release carries `install.sh` and `install.ps1`. Each downloads that release
over HTTPS, checks every file against its `SHA256SUMS`, installs the binary for your platform, and
runs `prepare-commit-msg identity` before keeping it. The previous binary stays as
`prepare-commit-msg.prev` (`prepare-commit-msg.exe.prev` on Windows). Install directly into the
global hooks directory from step 2:

```bash
curl -fsSL https://github.com/maccavelli/prepare-commit-msg/releases/latest/download/install.sh | sh -s -- --dir ~/.global-git-hooks
```

```powershell
& ([scriptblock]::Create((irm https://github.com/maccavelli/prepare-commit-msg/releases/latest/download/install.ps1))) -InstallDir "$env:USERPROFILE\.global-git-hooks" -NoPathUpdate
```

- **Name the directory.** Without `--dir` (`-InstallDir` on Windows), the
  installer uses `~/.local/bin`
  (`%LOCALAPPDATA%\Programs\prepare-commit-msg`), which Git does not run
  hooks from. `PREPARE_COMMIT_MSG_INSTALL_DIR` sets it too.
- **The installer never changes your Git configuration.** A global hooks
  directory can hold other hooks, so pointing Git at it is step 2's, and
  yours.
- `--version vX.Y.Z` installs another release; `--verify-attestation` also
  checks each download's build attestation with a logged-in `gh`;
  `--uninstall` removes the binary and its `.prev`. On Windows the same are
  `-Version`, `-VerifyAttestation`, and `-Uninstall`.

#### Pre-built binaries (manual)

Or download the executable for your OS and architecture from [Releases](https://github.com/maccavelli/prepare-commit-msg/releases), and check it against the release's `SHA256SUMS`:

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

Point Git at the global hooks directory, so all current and future repositories use the hook. The installer above has already put the binary there; with a manual download or a build, place it first:

```bash
# 1. Create a global hooks directory (the installer creates it)
mkdir -p ~/.global-git-hooks

# 2. Place the binary into the directory (the installer has done this)
cp dist/prepare-commit-msg-$(go env GOOS)-$(go env GOARCH) ~/.global-git-hooks/prepare-commit-msg
chmod +x ~/.global-git-hooks/prepare-commit-msg

# 3. Point Git at the global hooks directory
git config --global core.hooksPath ~/.global-git-hooks
```

If `core.hooksPath` already names another directory, put the binary there instead (`--dir` that directory) rather than replacing the setting, so the hooks it holds keep running.

#### Option B: Per-Repository Hook

Symlink or copy the binary into a single repository's `.git/hooks/` directory:

```bash
ln -sf /path/to/prepare-commit-msg .git/hooks/prepare-commit-msg
chmod +x .git/hooks/prepare-commit-msg
```

#### Windows Installation

1. Install with `install.ps1` as above, or place `prepare-commit-msg.exe` in `%USERPROFILE%\.global-git-hooks\prepare-commit-msg.exe` yourself.
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

Select your preferred provider (`gemini`, `openai`, `claude`, `grok`, `opencode-zen`, `opencode-go`,
`huggingface`, `kilo`, `together`, `ollama`), detect or enter credentials, and pick primary and
fallback models.

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

For OpenAI, the wizard also offers **Use your ChatGPT plan with this app**:
OpenAI's Sign in with ChatGPT for agents, under the name
`prepare-commit-msg`. The first time, it asks you to confirm OpenAI's notice
that this app's requests count toward your plan's usage, and names the page
where you manage that usage. The session is saved in the `oauth/` directory,
as every sign-in is.

OpenAI's terms for using a plan this way, and how this hook keeps each:

- **Requests from your own machine, for your own activity.** The hook runs
  on your machine, only when you run `git commit`, to write that commit's
  message.
- **Consent before any background use.** The hook runs nothing in the
  background.
- **The plan used only for the app you connected.** The session is used
  only by this hook, and sent only to OpenAI.
- **No pooling, sharing, transfer or resale.** The session stays in your own
  configuration directory; do not share that directory.
- **No charge.** The hook is free.

---

### Non-Interactive CLI Flags

Configure settings directly or in automated scripts using `--yes`:

```bash
# Example 1: Google Gemini (default cloud)
prepare-commit-msg configure --yes \
  --provider gemini \
  --model gemini-3.5-flash-lite \
  --fallback gemini-3.8-flash \
  --fallback gemini-3.6-flash

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
  --model claude-haiku-5-5 \
  --fallback claude-sonnet-5-5
```

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--provider` | `string` | `gemini` | `gemini`, `openai`, `claude`, `grok`, `opencode-zen`, `opencode-go`, `huggingface`, `kilo`, `together`, `ollama` |
| `--model` | `string` | *Provider default* | Primary model name |
| `--api-key` | `string` | `""` | Provider API key (uses existing config or env if omitted; ignored by `ollama`) |
| `--base-url` | `string` | *Saved `base_url`, else the provider's* | Endpoint, for `opencode-zen`, `opencode-go`, `huggingface`, `kilo`, `together` and `ollama`. `http` or `https` with a host; no user name, password, query or fragment; plain `http` to a remote host only for `ollama`. Without `--yes`, it is the wizard's suggested endpoint. |
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
      "model": "gemini-3.5-flash-lite",
      "fallback_models": ["gemini-3.8-flash", "gemini-3.6-flash"]
    },
    "kilo": {
      "api_key": "",
      "model": "kilo-auto/free",
      "fallback_models": ["kilo-auto/small"]
    },
    "ollama": {
      "api_key": "",
      "model": "qwen2.5-coder:7b",
      "fallback_models": ["llama3.2:latest"],
      "base_url": "http://192.0.2.10:11434"
    },
    "openai": {
      "api_key": "",
      "model": "gpt-6-luna",
      "fallback_models": ["gpt-5.6-luna"]
    },
    "claude": {
      "api_key": "",
      "model": "claude-haiku-5-5",
      "fallback_models": ["claude-sonnet-5-5"]
    }
  },
  "timeout_seconds": 120,
  "max_diff_bytes": 32000,
  "retry_count": 3,
  "retry_delay_seconds": 3
}
```

`base_url` is the endpoint a provider's requests go to. It is present only
when `configure` saved an endpoint other than the provider's default (see
[Supported Providers & Models](#supported-providers--models)), from the
endpoint you typed in the wizard or from `--base-url`. Only `opencode-zen`,
`opencode-go`, `huggingface`, `kilo`, `together` and `ollama` take one.

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
   # JSON: 1, Scripts: 2
   #
   # Please enter the commit message for your changes. Lines starting
   # with '#' will be ignored, and an empty message aborts the commit.
   ```

3. Edit the message or save and close the editor to complete the commit.

While it waits for the provider, the hook writes
`prepare-commit-msg: generating via <provider> (<model>)...` to stderr. A
request that fails in a way that can succeed later, such as a rate limit, an
unavailable service or a failure to connect, is retried up to `retry_count`
times, and before each wait the hook writes one line:

```text
prepare-commit-msg: openai: retrying in 2.4s (attempt 1 of 4): rate limited
```

The last part is the kind of failure: `rate limited`, `unavailable`,
`cannot reach`, or `error`. The provider's own message is not shown.

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
| `prepare-commit-msg update --json` | Write JSON Lines to stdout, ending with one `{"kind":"result",…}` object whose `result` carries `"schema_version":4`. |
| `prepare-commit-msg update --channel rc` | Follow a prerelease channel. |

Progress and the confirmation prompt go to **stderr**; stdout carries only `--json` output. The exit
status is `0` when up to date, declined, or installed; `10` when an update is available with
`--check`; and `1` on an error before the new binary is in place. An error after that, such as
failure to release the update lock, is reported as a `warning:` line on stderr, and the run still
exits `0`.

If an update fails and cannot put the old binary back, the backup remains beside it as
`.<name>.selfupdate-kept-<n>`, where `<name>` is the binary's file name. No later update removes it,
so delete it once the hook works. If something else replaces the binary while an update runs, the
update fails instead of overwriting it.

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
  is no longer logged (from `v1.10.0` it is again: see
  [Daily Workflow & Hook Behavior](#daily-workflow--hook-behavior)).
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
  model. This holds from `v1.8.0`, not `v1.6.0`: see
  [the next section](#changes-with-go-llmprovider-sdk-v132).
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

## Changes with go-llmprovider-sdk v1.3.2

Release `v1.8.0` moves from go-llmprovider-sdk `v1.2.1` to `v1.3.2`
([0011-MADR](docs/decisions/0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md)),
and from go-selfupdate-lib `v1.9.0` to `v1.13.0`, whose changes are in the
next section. What you may notice:

- **A rejected Gemini key stops at once.** The hook reports `authentication
  failed for gemini` and asks no fallback model. The note above promised
  this from `v1.6.0`, but Gemini sends that refusal in a form the SDK reads
  only from `v1.3.2`, so `v1.6.0` and `v1.7.0` still asked every fallback
  with the same key.
- **A malformed key fails at once.** A key with a control character in it is
  refused when the provider is set up, with a `failed to init` warning,
  instead of being sent and retried.
- **An answer cut off while the model was still thinking** is reported as
  `incomplete`. The next model is tried, as before.
- **Sign-in sessions lock through the operating system.** Refreshing an
  OAuth sign-in takes an OS lock on `<provider>.oslock` in the OAuth
  directory, which is released when the process ends, even if it is
  killed. A hook from before `v1.8.0` does not see that lock: update every
  copy of the hook that shares the directory.
- **`configure`** stops when you cancel it, even when the live listing
  recommends no model, and a masked key whose entry fails is never saved in
  part.
- **`update`** keeps a backup it could not restore, and does not overwrite a
  binary replaced while it ran (see [Self-Update](#self-update)). Under
  `--json`, a failed run's result names the product and the current
  version, and its `exit_code` matches the exit status.

---

## Changes with go-selfupdate-lib v1.13.0

Release `v1.8.0` is built with Go 1.27.2 and requires go-selfupdate-lib
`v1.13.0`
([0012-MADR](docs/decisions/0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md)).
What you may notice:

- **Built with Go 1.27.2.** Releases up to `v1.7.0` were built with Go
  1.27.1, whose standard library has published advisories that 1.27.2
  fixes. Run `prepare-commit-msg update` to take `v1.8.0`.
- **`update --json`'s result is schema 4.** It adds `rolled_back` (the
  previous binary was put back after the new one was installed),
  `probes_skipped` and `replaced_before_stop`; no key is renamed or
  removed.
- **An interrupted update keeps the previous binary.** If an `update` is
  killed between replacing the binary and finishing, the next `update`
  keeps the old copy beside the hook as
  `.prepare-commit-msg.selfupdate-kept-<n>`, instead of removing it.
  Delete it once the hook works.

---

## Changes with go-llmprovider-sdk v1.12.0 and go-selfupdate-lib v1.15.2

Release `v1.9.0` moves from go-llmprovider-sdk `v1.3.2` to `v1.12.0`, and
from go-selfupdate-lib `v1.13.0` to `v1.15.2`
([0014-MADR](docs/decisions/0014-MADR-adopt-go-llmprovider-sdk-v1-11-0-and-go-selfupdate-lib-v1-15-2.md)).
What you may notice:

- **Every provider can be configured.** `configure --provider` and the help
  text name all 10 providers, where `v1.8.0`'s named only `gemini`, `openai`,
  `claude` and `grok`. The next save of `config.json` adds an entry for
  each.
- **Ollama needs no key.** The hook runs a saved Ollama configuration
  without one, where `v1.8.0` stopped with `no API key for provider
  "ollama"`, and `configure --yes --provider ollama` no longer fails with
  `API key required`.
- **Your endpoint is kept.** An endpoint you type in the wizard, or pass
  with the new `--base-url`, is saved as `base_url` when it differs from the
  provider's default, and the hook sends its requests there (see
  [Configuration File](#configuration-file)). `--base-url` refuses a user
  name or password, a query and a fragment, and refuses plain `http` to a
  remote host for any provider that needs a key.
- **With Ollama before `v0.30.0`, name the models.** `configure --yes`
  takes the models in the order Ollama lists them, and Ollama lists embedding
  models too, which cannot write a commit message. Pass `--model` and
  `--fallback`, or choose them in the interactive wizard. From Ollama
  `v0.30.0`, which reports what each model can do, `v1.10.0` leaves those
  models out: see
  [Changes with go-llmprovider-sdk v1.15.0](#changes-with-go-llmprovider-sdk-v1150).
- **`configure --yes` stops on a refused key.** When the provider's model
  listing refuses the key, it fails with `<provider> refused the API key`
  and saves nothing. Any other listing failure still falls back to the
  curated models.
- **The model menu still shows six models,** the first six of the
  provider's curated list. The lists themselves changed: see
  [Supported Providers & Models](#supported-providers--models).
- **Typed operational settings are saved.** The timeout, diff size and
  retry values you enter in the wizard are kept, also when its answers are
  piped in. The prompt reads `Timeout seconds (default 120):` where it read
  `Timeout seconds [120]:`.
- **Requests name this program.** Each request's `User-Agent` begins
  `prepare-commit-msg/<version>`, where it named the SDK.
- **OpenAI's default model writes a prompt cache you pay for.** OpenAI
  caches prompts automatically on GPT-5.6 and later, and charges 1.25 times
  the input price to write the cache; nothing the hook sends turns that
  off. `gpt-6-luna`, now OpenAI's default, is one of those models, where
  `v1.8.0`'s `gpt-4.1-mini` is not. Each commit's prompt is new, so the
  cache is written and not read back. Choose `gpt-4.1-mini` to avoid the
  charge.
- **From the SDK:**
  - an HTTP 403 means the account may not use that model; it no longer
    renews a sign-in, and the next model is tried;
  - a refused Grok key stops at once, as a refused Gemini key does from
    `v1.8.0`;
  - a request that fails after it was sent is retried at most once;
  - a rate limit (HTTP 429) waits for the time the service names;
  - the wizard asks again for a typed key the service refuses, up to three
    times;
  - model listings are cached for 10 minutes.
- **`update`** behaves as in `v1.8.0`: go-selfupdate-lib `v1.15.2` changes
  nothing it uses.

---

## Changes with go-llmprovider-sdk v1.15.0

Release `v1.10.0` moves from go-llmprovider-sdk `v1.12.0` to `v1.15.0`
([0015-MADR](docs/decisions/0015-MADR-adopt-go-llmprovider-sdk-v1-15-0.md)).
go-selfupdate-lib stays at `v1.15.2`. What you may notice:

- **Use your ChatGPT plan.** For OpenAI, the wizard offers **Use your
  ChatGPT plan with this app**: see [Interactive Wizard](#interactive-wizard)
  for the notice and the terms. `configure --yes` stays API-key-only.
- **Retries are shown.** Before each retry's wait, the hook writes one line
  to stderr: see
  [Daily Workflow & Hook Behavior](#daily-workflow--hook-behavior).
- **Model listings:**
  - within one `configure` run, a provider's listing is fetched once and
    reused; if a later listing in that run fails for a passing reason, the
    earlier one is used;
  - with Ollama `v0.30.0` or later, models that cannot write text, such as
    embedding models, are left out, so `configure --yes` no longer picks one;
  - a Kilo listing with nothing else to recommend now offers the `kilo-auto`
    tiers, where `v1.9.0` fell back to the curated models.
- **Curated models changed** for `kilo`, `huggingface`, `together`,
  `opencode-zen` and `opencode-go`: see
  [Supported Providers & Models](#supported-providers--models).
- **Going back to `v1.9.0`.** A plan session saved by `v1.10.0` cannot be
  refreshed by `v1.9.0`: run `configure` and sign in again, or use an API
  key.
- **`update`** behaves as in `v1.9.0`.

---

## Changes with Go 1.27.3, go-llmprovider-sdk v1.21.0 and go-selfupdate-lib v1.17.2

The next release moves to Go 1.27.3, from go-llmprovider-sdk `v1.15.0` to
`v1.21.0`, and from go-selfupdate-lib `v1.15.2` to `v1.17.2`
([0018-MADR](docs/decisions/0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md)).
What you may notice:

- **Built with Go 1.27.3,** which fixes GO-2026-6809 in `crypto/x509`.
  `v1.10.0` was built with Go 1.27.2, which has that advisory. Run
  `prepare-commit-msg update` to take the release.
- **Retries.** When a provider's reply has started and the provider then
  reports a failure inside it, the hook retries once, whatever
  `retry_count` is. Other failures are retried as before.
- **Wizard messages:**
  - a refused key, a model-listing warning or a Kilo profile notice shows
    the error's class and HTTP status, such as
    `authentication failed, HTTP 401`, and not the service's own message,
    which can echo a key;
  - a failed browser or device sign-in, an expired device code and an
    unreachable Ollama server name the kind of failure first.
- **The installer.** When `install.sh --uninstall` or
  `install.ps1 -Uninstall` cannot remove a copy, for example on Windows
  while the program is running, it prints `could not remove <path>`, removes the
  rest and exits 3. Before, `install.ps1` reported the copy removed and
  exited 0. The usage lists a new `--force` (`-Force`). It applies only to
  a `before_uninstall` hook, and this app has none.
- **`update`** behaves as in `v1.10.0`.

---

## CLI Reference

```text
Usage:
  prepare-commit-msg configure [flags]                - run setup wizard or non-interactive configure
  prepare-commit-msg update [flags]                   - check for and apply updates from GitHub
  prepare-commit-msg version                          - show binary version, as vX.Y.Z (release) <revision>
  prepare-commit-msg identity                         - print the build identity alone, as the release checks run it
  prepare-commit-msg help                             - display help message
  prepare-commit-msg <commit_msg_file> [source] [sha] - run as git prepare-commit-msg hook
```

---

## Developer Experience

Brief guide for contributors and local development.

### Toolchain & Quality Gates

The repository uses pinned developer tools in `.tools/bin` to ensure identical results across local machines and CI. Building from source needs Go 1.27.3.

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
| `make preflight` | Run every CI gate that runs off GitHub: records, Markdown, then `make verify` |
| `make preflight-hosts` | Run CI's jobs for `HEAD` on the test hosts ([`tools/README.md`](tools/README.md)) |
| `make pre-add-check` | Check the staged Go snapshot, or `FILES=...` |
| `make clean` | Clean build outputs from `dist/` |

### Maintainer Release

A release is a pushed `v*` tag on a commit whose CI has passed on `main`:

```bash
git tag -a v1.2.3 -m "v1.2.3"
git push origin v1.2.3
```

Once the quality contract and native tests pass on Linux, macOS, and Windows, the tag's CI run calls
go-selfupdate-lib's build workflow. It builds the six binaries in `selfupdate-release.json` with a
fixed recipe, checks each binary's build information, runs `prepare-commit-msg identity` on the five
platforms GitHub hosts, requires `v1.2.3 (release)`, and writes `SHA256SUMS`. Its publish workflow
then publishes the GitHub Release. On a non-tag push, the same build runs as a rehearsal stamped
`rehearsal-<commit> (local)`, and publishes nothing
([0012-MADR](docs/decisions/0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md)).

Assets include cross-compiled binaries, `SHA256SUMS`, `install.sh`, `install.ps1`, and build-provenance attestations. The attestation's signer is go-selfupdate-lib's publish workflow, so name it:

```bash
sha256sum --check --ignore-missing SHA256SUMS
gh attestation verify prepare-commit-msg-linux-amd64 --repo maccavelli/prepare-commit-msg \
  --signer-workflow maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml
```

---

## Documentation

Start with the [documentation index](docs/README.md) for decisions, plans,
reports, and task-oriented navigation.

- To understand the current system, read
  [the architecture](docs/architecture.md).
- To run verification, manage repository settings, or perform a release, use
  [the CI/CD operations guide](docs/guides/cicd-operations.md).
- To understand the workspace and local-hook policy, read
  [MADR 0013](docs/decisions/0013-MADR-align-repository-with-workspace-scaffolding.md).

---

## License

Part of the split Go repository fleet.
