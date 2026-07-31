> **Mirror notice:** This repository is a one-way published export of a
> privately hosted project. History is squashed into sync snapshots, and pull
> requests cannot be merged here directly — open an issue instead. Changes
> land in the private source and are re-exported.

# prepare-commit-msg

AI-powered Git `prepare-commit-msg` hook written in Go. When you run a normal
`git commit` (empty editor / comment-only template), it gathers the **staged**
diff and asks an LLM (Gemini, OpenAI, or Claude) for a conventional commit
message.

Hook failures **never block the commit** (soft-fail): a short message is printed
to stderr and the editor is left unchanged so you can type a message manually.

## Features

- Conventional commit messages from staged diffs (numstat + unified patch)
- Providers: **gemini**, **openai**, **claude** (via split `mcplib.git/llmprovider`)
- **Curated model menus** (max 6): stable production text models only — not the raw
  Gemini catalog of embeddings/TTS/Live/image/preview IDs
- Primary model + up to **3** fallback models
- Interactive `configure` wizard **and** equivalent CLI flags
- Environment API keys (`GEMINI_API_KEY`, `OPENAI_API_KEY`, `CLAUDE_API_KEY`)
  - Configure: detect and prompt (default **use**)
  - Hook: config key first, then env (no prompt)
- Diff size limit, overall timeout, retries with backoff
- Cross-platform: Linux, macOS, Windows (see install notes)

## Build

```bash
cd ~/gitrepos/go/prepare-commit-msg
make build          # local OS/arch → dist/
make build-all      # linux/darwin/windows (amd64 + arm64 variants)
make test
make lint
```

Binaries land in `dist/` as `prepare-commit-msg-<goos>-<goarch>[.exe]`.

## Configure

### Interactive wizard

```bash
./dist/prepare-commit-msg-linux-amd64 configure
```

Walks through provider, API key (env / existing / prompt), model discovery,
fallback multi-select (max 3), and operational settings (timeout, max diff,
retries).

### Non-interactive flags

```bash
prepare-commit-msg configure --yes \
  --provider gemini \
  --model gemini-2.5-flash \
  --fallback gemini-2.0-flash \
  --fallback gemini-2.0-flash-lite \
  --timeout-seconds 120 \
  --max-diff-bytes 32000 \
  --retry-count 3 \
  --retry-delay-seconds 3
```

| Flag | Meaning |
|------|---------|
| `--provider` | `gemini`, `openai`, or `claude` |
| `--model` | Primary model |
| `--api-key` | API key (else env or existing config) |
| `--fallback` | Fallback model (repeatable, max 3) |
| `--timeout-seconds` | Overall LLM timeout |
| `--max-diff-bytes` | Truncate staged diff in the prompt |
| `--retry-count` / `--retry-delay-seconds` | Per-model retries |
| `--no-env` | Do not read keys from the environment |
| `--yes` | Non-interactive (no prompts) |

## Config file

Path (via `os.UserConfigDir`):

| OS | Typical path |
|----|----------------|
| Linux | `~/.config/prepare-commit-msg/config.json` (or `$XDG_CONFIG_HOME/...`) |
| macOS | `~/Library/Application Support/prepare-commit-msg/config.json` |
| Windows | `%AppData%\prepare-commit-msg\config.json` |

Legacy locations are still read and migrated:

- `~/.config/prepare-commit-msg/config.json`
- `~/.config/prepare-commit-msg-embedded/config.json`

File mode is `0600` where the OS honors Unix permissions.

## Install as a Git hook

### Linux / macOS

```bash
make install
# copies to ~/.global-git-hooks/prepare-commit-msg

# Option A: global hooks path
git config --global core.hooksPath ~/.global-git-hooks

# Option B: per-repo symlink
ln -sf /path/to/dist/prepare-commit-msg-linux-amd64 .git/hooks/prepare-commit-msg
chmod +x .git/hooks/prepare-commit-msg
```

### Windows

1. Build: `make windows-amd64` (or build on Windows with `make build`).
2. Create a hooks directory, e.g. `%USERPROFILE%\.global-git-hooks`.
3. Copy the binary as `prepare-commit-msg.exe` **or** a no-extension PE
   named `prepare-commit-msg` (Git for Windows accepts both in many setups).
   A tiny `prepare-commit-msg` shell wrapper that `exec`s the `.exe` also works
   under Git Bash.
4. Point Git at the directory:

```bat
git config --global core.hooksPath %USERPROFILE%\.global-git-hooks
```

Ensure **Git for Windows** is on `PATH` so the hook can run `git diff`.

> `make install` is intended for Unix shells (`cp`, `chmod`, `$HOME`).

## Hook behavior

Git invokes:

```text
prepare-commit-msg <COMMIT_EDITMSG> [source] [sha]
```

| `source` | AI generation |
|----------|----------------|
| *(empty)* plain `git commit` | Yes, if message is empty / comments-only |
| `template` | Yes, if message is empty / comments-only |
| `message` (`-m`) | Skipped |
| `merge` / `squash` | Skipped |
| `commit` (amend / `-c` / `-C`) | Skipped |

On LLM/config/git errors the hook **exits 0** after printing:

```text
prepare-commit-msg: could not generate a message (...)
prepare-commit-msg: commit editor left unchanged — type a message manually or run: prepare-commit-msg configure
```

## Usage (CLI)

```text
prepare-commit-msg configure [flags]
prepare-commit-msg version
prepare-commit-msg help
prepare-commit-msg <commit_msg_file> [source] [sha]
```

## Development

```bash
make fmt
make vet
make test
make lint
```

Module: `github.com/maccavelli/prepare-commit-msg` (Go 1.26.5), depends on split private module
`github.com/maccavelli/mcplib`. Vendor directory is
intentionally disabled for this repository.

## License / context

Part of the split Go repository fleet formerly maintained under `saxsmith-global-context`.
