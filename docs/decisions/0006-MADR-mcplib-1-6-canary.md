---
status: proposed
date: 2026-09-28
decision-makers: Project Owner
consulted: mcplib (the library this repository consumes)
informed: mcp-server-magictools, mcp-server-magicdev (the consumers that follow the canary)
---
# Canary mcplib v1.6.0-rc.1 in prepare-commit-msg, Reading Vendor CLI Logins Through

## Context and Problem Statement

`prepare-commit-msg` consumes `github.com/maccavelli/mcplib` for providers
(`llmprovider`), the configure wizard (`wizard`) and self-update
(`selfupdate`). It requires mcplib `v1.5.0` (`go.mod:17`). Its release job
runs mcplib's reusable workflow pinned to that tag's commit
(`.github/workflows/ci.yml:115`, `@d13f89c… # mcplib v1.5.0`).

mcplib's `main` has moved well past `v1.5.0`:
* its `0012` plans conform the providers to their reference clients: item
  fidelity, gateway conventions, the ChatGPT backend, OAuth hygiene, Grok;
* its `0013` plan surfaces catalog errors and adds model search to the wizard;
* its `0014` plan moves the Gemini provider to the Interactions API.

The owner chose `prepare-commit-msg` as the **first consumer to canary** that
work, on 2026-09-28. mcplib is published first as a release candidate,
`v1.6.0-rc.1`, on commit `4e1f9a5`. Stable `v1.6.0` follows only once the canary
works, so the other consumers never pick up a version the canary rejected.

The question is what `prepare-commit-msg` must change to consume that
candidate correctly, and what "the canary works" means.

### What was measured

On 2026-09-28, `HEAD` (`4b5dab4`) was built and tested against an archive of
mcplib `4e1f9a5`, in scratch copies only:

* **It builds, and `go vet` is clean.**
* **Three tests fail:**

  | Test | Failure | Cause |
  |---|---|---|
  | `TestRunSetupInteractive_Success` | `expected my-custom-model, got "gemini-3.7-flash"` | The wizard's model step now searches first (mcplib `0013`), so the scripted menu numbers no longer match. |
  | `TestRunSetupInteractive_ImportGrokSession` | `runSetupInteractive() error = context canceled` | These two tests pin the old import, which copied a CLI's tokens into this tool's OAuth store. |
  | `TestRunSetupInteractive_ChatGPTDoesNotCopyAccessIntoAPIKey` | `runSetupInteractive() error = API key is required` | Same cause as the row above. |

* **A user-facing regression.** mcplib's OAuth-hygiene plan (O1) replaced that
  copy with a read-through. The wizard now returns a new credential kind,
  `wizard.CredVendorCLI`, with the path of the CLI's `auth.json`. Setup
  (`internal/ui/setup.go:234-250`) knows only `CredOAuth` and API keys. So
  choosing **Use the Codex CLI login** or **Use the Grok CLI login** fails
  with "API key is required".

  mcplib made the change because a copied refresh token shares the CLI's
  refresh-token family. Both vendors revoke the whole family when one token is
  used twice, so the first refresh by either side logs the other out.
* **Unit tests reach the network.** Five setup tests make real listing calls
  with fake keys. Setup passes `Discover: true` (`setup.go:217`), and nothing
  injects an HTTP client. This already happened silently under `v1.5.0`. Since
  mcplib `0013` the wizard prints a warning (for example
  `models endpoint returned HTTP 400`), which exposed it.
* **Unaffected:**
  * `selfupdate`: its only change since `v1.5.0` is a test (`ca29b81`);
  * `WithThinkingBudget` and `Continue`: this tool uses neither, so mcplib
    `0014` changes nothing here beyond the endpoint Gemini is called on.

**Separately: `main` is already red.** `internal/ui/open.go:35`
(`go func() { _ = command.Wait() }()`, from `4b5dab4`) fails errcheck. CI run
`36256206752` on `4b5dab4` failed on it. The canary cannot pass its gate or CI
until that is fixed. On 2026-09-28 the owner chose to fold the fix into this
decision.

## Decision Drivers

* **The canary must exercise exactly the candidate.** Consume
  `v1.6.0-rc.1` in `go.mod` and in the release-workflow pin, with no local
  replace.
* **No user-visible regression.** Every choice the wizard offers must work.
* **The CLI owns its login.** Never copy or refresh a vendor CLI's refresh
  token.
* **Hermetic tests.** A unit test must never reach a live endpoint, or depend
  on credentials in the shell.
* **A green `main` first.** A canary built on a red `main` proves nothing.
* **Small and provable.** Red first, the full `make verify` gate, and mutants,
  as in the earlier plans here and in mcplib.

## Considered Options

* Adopt `v1.6.0-rc.1` and read vendor CLI logins through
* Adopt `v1.6.0-rc.1` and reject the vendor CLI choice in setup
* Adopt `v1.6.0-rc.1` and copy the CLI's tokens into this tool's store, as before
* Stay on mcplib `v1.5.0`

## Decision Outcome

Chosen option (**proposed**): **"Adopt `v1.6.0-rc.1` and read vendor CLI logins
through"**. It is the only option that exercises the candidate with every
wizard choice working, without sharing a refresh token with the CLI.

### 1. Consume the candidate

* `go.mod` requires `github.com/maccavelli/mcplib v1.6.0-rc.1`.
* The release job's pin moves to
  `@4e1f9a53e265808bbfa740e3e3b09a51ed7f56ce # mcplib v1.6.0-rc.1`, the tag's
  commit, as mcplib's workflow requires.
* **This repository's own release stays a stable tag.** The workflow rejects
  anything that is not `vX.Y.Z` ("Require a strict stable tag"). The canary
  ships as `v1.4.0`: a minor version, because vendor CLI logins now read
  through.

### 2. Vendor CLI logins read through

* **Config.** A `wizard.CredVendorCLI` result is saved as
  `auth_kind: "vendor_cli"` plus `vendor_auth_path`, the CLI's `auth.json`.
  No token is written anywhere.
* **Generation.** The provider is built on
  `llmprovider.VendorCLISession{Provider, Path}`. It re-reads the file on
  every request and never refreshes.
* **An expired login** is mcplib's `ErrAuthFailure`, telling the user to sign
  in with the CLI again. There is no need to re-run `configure`.
* **Cleanup.** Setup deletes any OAuth session this tool holds for that
  provider. It may be one an older release copied from the CLI, and it must
  never be refreshed again.
* **Validation.** A `vendor_cli` config with no path fails at generation with
  "no CLI login path …; run 'prepare-commit-msg configure'".
* **Clearing.** Choosing an API key or a browser sign-in clears
  `vendor_auth_path`.

### 3. Configure's listing is injectable, and offline in tests

* A package-level `listingClient *http.Client` (nil in production) is passed
  to the wizard (`Options.HTTPClient`) and to non-interactive discovery
  (`WithHTTPClient`).
* Every setup test installs a client that fails every request and counts the
  attempts.
* The scripted menu in `TestRunSetupInteractive_Success` follows the search
  step.
* The two import tests become read-through tests: kind, saved path, no
  session, no tokens in `config.json`.

### 4. `main` green first

`internal/ui/open.go` still reaps the browser launcher in the background, so
sign-in never blocks. A launcher that fails is reported on stderr ("could not
open a browser (…); open the URL above instead") instead of discarded. The URL
is printed before the launch, as today. This fix does not depend on the mcplib
version, so it lands first.

### Consequences

* Good, because `prepare-commit-msg` ships on mcplib's current provider work
  (the `0012`, `0013` and `0014` plans), and that work gets a real consumer
  before stable `v1.6.0`.
* Good, because the Codex and Grok CLI logins work and stay the CLI's:
  nothing here refreshes them, so neither side can revoke the other.
* Good, because unit tests stop depending on the network and on the shell's
  credentials.
* Good, because `main` returns to green independently of the canary.
* Neutral, because a session copied under an older release (`auth_kind:
  "oauth"`) keeps working until its refresh fails. Then generation fails with
  an authentication error, and the user runs `configure` and picks the CLI
  login.
  * It is not migrated automatically: the config records the same
    `auth_kind: "oauth"` for a copied session and for a browser sign-in, so
    the two cannot be told apart.
  * The release notes say so.
* Bad, because a stable release of this tool is built on an mcplib release
  candidate until promotion (§ Confirmation, promotion).

### Confirmation

**Before landing** (proven 2026-09-28; the PLAN's Appendix A):
* red: five read-through tests fail before the fix;
* all six mutants are killed, each checked passing unmutated first;
* `make verify` passes;
* the open.go fix alone passes `make verify` on `v1.5.0`.

**The canary works when all of these hold:**
1. CI passes on `main` with `v1.6.0-rc.1`: all jobs, including native tests on
   Linux, macOS and Windows.
2. The `v1.4.0` release publishes through the pinned workflow, and
   `gh release verify v1.4.0` passes.
3. `prepare-commit-msg update` moves an installed `v1.3.0` to `v1.4.0`, on
   macOS and on the owner's Windows laptop.
   * The laptop is where mcplib's `TestNativeReplaceRunningCopy` fails with
     "Access is denied" (mcplib `0012-PLAN-oauth-hygiene.md`, §9, 2026-09-28),
     so this is a real check of that path.
4. Live commit messages are generated through the installed hook:
   * with a Gemini API key, on the Interactions API;
   * through the Codex CLI login, read-through. The CLI's session is only read,
     never refreshed or revoked.

**Promotion.** Once all four hold, the owner tags mcplib `v1.6.0` on the same
commit (`4e1f9a5`). This repository then requires `v1.6.0`: the pin's SHA is
unchanged, and only its comment changes.

## Pros and Cons of the Options

### Adopt `v1.6.0-rc.1` and read vendor CLI logins through

* Good, because every wizard choice works.
* Good, because the CLI keeps sole ownership of its refresh token.
* Neutral, because it adds one config field and one credential kind.
* Bad, because sessions copied by older releases need one re-run of
  `configure`, after their refresh fails.

### Adopt `v1.6.0-rc.1` and reject the vendor CLI choice in setup

* Good, because it is smaller: no config or runtime change.
* Bad, because the wizard still offers the choice (mcplib has no option to hide
  it), so users meet an error for a listed feature.
* Bad, because the canary would not exercise mcplib's read-through at all.

### Adopt `v1.6.0-rc.1` and copy the CLI's tokens into this tool's store, as before

* Bad, because it recreates the shared refresh-token family that mcplib O1
  removed: the first refresh by either side revokes the other.
* Bad, because the wizard no longer returns the tokens, so the tool would have
  to parse vendor files itself.

### Stay on mcplib `v1.5.0`

* Good, because nothing changes here.
* Bad, because there is no canary, and the other consumers would get
  `v1.6.0` untried.

## More Information

**Related records:**
* **mcplib:**
  * `docs/0012-MADR-conform-providers-to-reference-clients.md` §5.1 (vendor
    CLI read-through) and `docs/0012-PLAN-oauth-hygiene.md` (O1, and O8 for
    the Windows path test);
  * `docs/0013-MADR-remediate-debugging-pass-findings.md` (catalog errors and
    search);
  * `docs/decisions/0014-MADR-gemini-wire-fidelity.md` (Gemini on the
    Interactions API).
* **This repository:**
  * [0002-MADR-self-update-cli-and-github-releases-integration.md](0002-MADR-self-update-cli-and-github-releases-integration.md)
    (self-update);
  * [0004-MADR-align-cicd-with-magic-cli-remote.md](../0004-MADR-align-cicd-with-magic-cli-remote.md)
    (the tag-driven release).

**Evidence:**
* The scratch probes against mcplib `4e1f9a5`, and the proof (red, mutants,
  `make verify`) are in the PLAN's Appendix A.
* CI run `36256206752` on `4b5dab4` (red on `open.go:35`).

**Plan.** [0006-PLAN-mcplib-1-6-canary.md](0006-PLAN-mcplib-1-6-canary.md).
