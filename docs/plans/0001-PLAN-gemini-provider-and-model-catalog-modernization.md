# Implementation Plan: Modernizing Gemini Provider, Model Catalog Integration, and Configure UX

* **Target Repositories:** [`prepare-commit-msg`](file:///home/mac/gitrepos/prepare-commit-msg/README.md) & [`mcplib`](file:///data/cache/go/pkg/mod/github.com/maccavelli/mcplib@v0.2.0)
* **Status:** Ready for User Review
* **Architectural Reference:** [ADR-0001: Modernizing Gemini Provider, Model Catalog Integration, and Configure UX](file:///home/mac/gitrepos/prepare-commit-msg/docs/decisions/0001-gemini-provider-and-model-catalog-modernization.md)

---

## Goal Description

This implementation plan modernizes the Gemini provider, model catalog integration, and CLI configure wizard across `prepare-commit-msg` and `mcplib/llmprovider`.

The change resolves four key requirements:
1. **Purges decommissioned models** (`gemini-2.0-flash`, `gemini-2.0-flash-lite`, and `gemini-1.5-*`) and removes speculative entries (`gemini-3.1-flash-lite`).
2. **Introduces the current frontier fast models** (`gemini-3.7-flash`, `gemini-3.6-flash`, `gemini-3.5-flash`, and `gemini-3.5-flash-lite`) while preserving `gemini-2.5-flash` and `gemini-2.5-flash-lite` as reliable fallbacks.
3. **Excludes heavy reasoning models** (Gemini Pro series, OpenAI `o1`/`o3`/`o4`) from Git hook recommendations to guarantee sub-2-second commit generation latency.
4. **Adopts a greenfield configuration schema**: Clean, idiomatic code without legacy migration shims. Pre-populates clean template provider configurations with modern default models and fallbacks.
5. **Elevates CLI Configure UX** with real-time API key verification, human-readable model labels with speed badges, and dynamic catalog backfilling.

---

## User Review Required

> [!IMPORTANT]
> **Greenfield Configuration & Legacy Cleanup:**
> Configuration schema backwards compatibility is explicitly disabled. We will remove obsolete legacy migration paths (such as `~/.config/prepare-commit-msg-embedded/config.json`) and implement a clean, idiomatic configuration loader that initializes fresh template configs with modern default models and fallbacks for all supported providers.

> [!NOTE]
> **Fleet Architecture & Module Boundary:**
> `prepare-commit-msg` depends on `github.com/maccavelli/mcplib/llmprovider` (located at `/data/gitrepos/mcplib`). The changes will update both the library (`mcplib`) and the application (`prepare-commit-msg`). We will use a local `replace` directive during development and verification to ensure seamless end-to-end testing across both repositories.

---

## Proposed Changes

### Component 1: `mcplib/llmprovider` (Provider Library)

#### [MODIFY] [`models_catalog.go`](file:///data/gitrepos/mcplib/llmprovider/models_catalog.go)
* Replace `StaticGemini` slice with the **Top 6 current fast models**:
  ```go
  var StaticGemini = []string{
      "gemini-3.7-flash",
      "gemini-3.6-flash",
      "gemini-3.5-flash",
      "gemini-3.5-flash-lite",
      "gemini-2.5-flash",
      "gemini-2.5-flash-lite",
  }
  ```
* Update `RankGeminiModel`:
  * Boost `gemini-3.7-flash` (+100), `gemini-3.6-flash` (+90), `gemini-3.5-flash` (+80).
  * Penalize reasoning/pro models (`strings.Contains(sm, "pro")`: -500) and deep-research/preview/experimental (-1000).
  * Heavily penalize deprecated generations (`2.0` / `1.5`: -2000).
* Update `curateFromCatalog`:
  * Match static catalog items first to preserve curated priority.
  * Dynamically backfill up to `MaxListedModels` (6) from remaining unlisted API models that pass `isUsableGeminiTextModel`, sorted by `RankGeminiModel`.
* Update `isUsableGeminiTextModel`:
  * Ensure models with `generateContent` and no deny keywords are accepted.

#### [MODIFY] [`models_catalog_test.go`](file:///data/gitrepos/mcplib/llmprovider/models_catalog_test.go)
* Update unit tests to validate:
  * `StaticGemini` contains only active fast models (`gemini-3.7-flash`, `gemini-3.6-flash`, etc.) and zero deprecated `2.0`, `1.5`, or `pro` models.
  * `RankGeminiModel` ranks `gemini-3.7-flash` > `gemini-3.6-flash` > `gemini-3.5-flash` > `gemini-2.5-flash` > `gemini-2.5-pro`.
  * `curateFromCatalog` correctly backfills new dynamically discovered models.

#### [MODIFY] [`gemini.go`](file:///data/gitrepos/mcplib/llmprovider/gemini.go)
* Keep `thinkingBudget` suppressed (`thinkingBudget: 0` or omitted) during standard `Generate` to prevent reasoning delays.

---

### Component 2: `prepare-commit-msg` (CLI Application & Git Hook)

#### [MODIFY] [`internal/config/config.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/config/config.go)
* **Clean Greenfield Architecture:** Remove legacy migration paths (`legacyConfigPaths`, `prepare-commit-msg-embedded`).
* **Pre-populated Template Configuration:** Update `ApplyDefaults` so that when a config is created or initialized, all supported providers (`gemini`, `openai`, `claude`) are populated with their modern recommended primary model and fallback models:
  ```go
  func ApplyDefaults(c *Config) {
      if c.Providers == nil {
          c.Providers = make(map[string]ProviderConfig)
      }
      if c.ActiveProvider == "" {
          c.ActiveProvider = llmprovider.ProviderGemini
      }
      for _, p := range SupportedProviders {
          pc, exists := c.Providers[p]
          if !exists {
              pc = ProviderConfig{}
          }
          if pc.Model == "" {
              pc.Model = DefaultModelForProvider(p)
          }
          if len(pc.FallbackModels) == 0 {
              pc.FallbackModels = DefaultFallbacksForProvider(p, pc.Model)
          }
          pc.FallbackModels = ClampFallbacks(pc.FallbackModels)
          c.Providers[p] = pc
      }
      if c.TimeoutSeconds <= 0 {
          c.TimeoutSeconds = DefaultTimeoutSeconds
      }
      if c.MaxDiffBytes <= 0 {
          c.MaxDiffBytes = DefaultMaxDiffBytes
      }
      if c.RetryCount <= 0 {
          c.RetryCount = DefaultRetryCount
      }
      if c.RetryDelaySeconds <= 0 {
          c.RetryDelaySeconds = DefaultRetryDelaySeconds
      }
  }
  ```
* Define provider default model mapping:
  * `gemini`: `gemini-3.7-flash` (fallbacks: `gemini-3.6-flash`, `gemini-3.5-flash`, `gemini-2.5-flash`)
  * `openai`: `gpt-4.1-mini` (fallbacks: `gpt-4.1-nano`, `gpt-4o-mini`)
  * `claude`: `claude-haiku-4-5` (fallbacks: `claude-sonnet-5`, `claude-sonnet-4-6`)

#### [MODIFY] [`internal/config/config_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/config/config_test.go)
* Update tests to reflect clean greenfield schema and remove tests asserting legacy migration paths.
* Verify default provider template population.

#### [MODIFY] [`internal/ui/setup.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/ui/setup.go)
* Add real-time API key verification:
  * When a key is entered or detected from environment, test it against `llmprovider.ListAvailableModels(ctx, provider, key)` with a 10s timeout.
  * Print status confirmation (`✔ API key verified`).
* Enhance `promptModel`:
  * Map model identifiers to human-readable names and latency/use-case badges:
    * `gemini-3.7-flash` ➔ `Gemini 3.7 Flash     [★ Recommended: Frontier coding intelligence]`
    * `gemini-3.6-flash` ➔ `Gemini 3.6 Flash     [High efficiency & reduced token overhead]`
    * `gemini-3.5-flash` ➔ `Gemini 3.5 Flash     [High-speed production workhorse]`
    * `gemini-3.5-flash-lite` ➔ `Gemini 3.5 Flash-Lite[Ultra-low latency & high throughput]`
    * `gemini-2.5-flash` ➔ `Gemini 2.5 Flash     [Proven balanced fast baseline]`
    * `gemini-2.5-flash-lite` ➔ `Gemini 2.5 Flash-Lite[Lightweight fast baseline]`
* Update `promptFallbacks`:
  * Recommend the next 3 fast models excluding the selected primary model.

#### [MODIFY] [`internal/ui/setup_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/ui/setup_test.go)
* Update tests to reflect the 6-model catalog for Gemini.
* Verify interactive menu choices (`1` through `6` + `7` for "Other").
* Verify fallback multi-selection with new recommended defaults.

#### [MODIFY] [`README.md`](file:///home/mac/gitrepos/prepare-commit-msg/README.md)
* Replace deprecated flags in documentation:
  * Update `--model gemini-2.5-flash` ➔ `--model gemini-3.7-flash`
  * Update `--fallback gemini-2.0-flash` ➔ `--fallback gemini-3.6-flash`
  * Update `--fallback gemini-2.0-flash-lite` ➔ `--fallback gemini-3.5-flash`

---

## Verification Plan

### Automated Tests
1. **`mcplib/llmprovider` Test Suite:**
   ```bash
   cd /data/gitrepos/mcplib && go test -v -race ./llmprovider/...
   ```
   *Verifies:* `TestIsUsableGeminiTextModel`, `TestCurateFromCatalog_Gemini`, `TestStaticModels_NoShutDownGemini20`, `TestRankGeminiModel_PrefersFlashLite`.

2. **`prepare-commit-msg` Test Suite:**
   ```bash
   cd /home/mac/gitrepos/prepare-commit-msg && go test -v -race ./...
   ```
   *Verifies:* All unit tests across `main`, `internal/ui`, `internal/config`, `internal/git`, `internal/fsutil`.

3. **Fleet Linter & Vet Check:**
   ```bash
   cd /home/mac/gitrepos/prepare-commit-msg && go vet ./...
   ```

4. **Multi-Platform Cross Build Check:**
   ```bash
   cd /home/mac/gitrepos/prepare-commit-msg && make build
   ```

### Manual Verification
1. **Interactive Configure Test:**
   * Run `./dist/prepare-commit-msg-linux-amd64 configure`
   * Select `1) gemini`
   * Confirm environment key detection and verification status.
   * Observe model choice menu displaying all 6 annotated fast models.
   * Choose option `1` (`gemini-3.7-flash`) and observe fallbacks auto-populating `gemini-3.6-flash`, `gemini-3.5-flash`, `gemini-2.5-flash`.
   * Verify config saved cleanly in `~/.config/prepare-commit-msg/config.json` with pre-populated templates for all providers.

2. **Non-Interactive Flag Configure Test:**
   * Run:
     ```bash
     ./dist/prepare-commit-msg-linux-amd64 configure --yes \
       --provider gemini \
       --model gemini-3.7-flash \
       --fallback gemini-3.6-flash \
       --fallback gemini-3.5-flash
     ```
   * Inspect saved config to ensure values match.
