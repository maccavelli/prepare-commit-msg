# ADR-0001: Modernizing Gemini Provider, Model Catalog Integration, and Configure UX

* **Status:** Proposed / Under Review
* **Deciders:** Maintainers of `prepare-commit-msg` and `mcplib`
* **Date:** 2026-08-16
* **Technical Domain:** LLM Providers, CLI User Experience, Git Hook Automation, Gemini API & SDK Integration, Configuration Schema

---

## Context and Problem Statement

[`prepare-commit-msg`](file:///home/mac/gitrepos/prepare-commit-msg/README.md) is an automated Git hook written in Go that generates conventional commit messages from staged Git diffs using LLM providers (`gemini`, `openai`, `claude`). LLM integration is abstracted via the internal fleet library [`mcplib/llmprovider`](file:///data/cache/go/pkg/mod/github.com/maccavelli/mcplib@v0.2.0/llmprovider).

An architectural evaluation of the current codebase revealed key areas for modernization:

1. **Inaccurate and Stale Model Catalog:**
   * The hardcoded static catalog in [`StaticGemini`](file:///data/cache/go/pkg/mod/github.com/maccavelli/mcplib@v0.2.0/llmprovider/models_catalog.go#L28-L34) contains nonexistent/speculative model IDs (e.g., `gemini-3.1-flash-lite`) and includes reasoning-heavy Pro models (`gemini-2.5-pro`).
   * It completely omits current frontier fast models (`gemini-3.7-flash`, `gemini-3.6-flash`, and `gemini-3.5-flash-lite`).
   * Documentation in [`README.md`](file:///home/mac/gitrepos/prepare-commit-msg/README.md#L60-L61) still references deprecated and shut-down models (`gemini-2.0-flash` and `gemini-2.0-flash-lite`, decommissioned in mid-2026).

2. **Commit Message Optimization (Fast Models vs. Reasoning Models):**
   * Commit generation requires fast analysis of unified diffs and concise output with sub-2-second latency.
   * Deep reasoning models (Gemini Pro series, OpenAI `o1`/`o3`/`o4`) add 10–30+ seconds of latency and unnecessary token costs.
   * The model catalog must prioritize the **Top 6 Fast Models** for latency-sensitive Git hooks.

3. **Greenfield Configuration Schema & Legacy Cleanup:**
   * Backwards compatibility with legacy configuration schemas (e.g. `prepare-commit-msg-embedded` migration paths) is not required.
   * Configuration loading and template population should be treated as greenfield: clean, idiomatic Go structs, with template configurations pre-populated with optimal, modern default models and recommended fallbacks for each supported provider.

4. **CLI Configure Presentation & Onboarding Ergonomics:**
   * Provide real-time API key validation feedback during interactive setup.
   * Display human-readable model names with latency/tier badges.
   * Automatically pre-populate the best 2–3 fallback models.

5. **Gemini SDK Research & Functional Parity:**
   * The official unified Go SDK is [`google.golang.org/genai`](https://github.com/googleapis/go-genai), deprecating `github.com/google/generative-ai-go`.
   * Adopt SDK functional capabilities (top-level `systemInstruction`, `thinkingBudget: 0` suppression, rich model metadata introspection) while maintaining a lightweight, zero-dependency Go binary.

---

## Decision Drivers

* **Execution Latency:** Sub-2-second Git hook execution for fast developer workflows.
* **Model Freshness:** Exclusively curate the Top 6 current fast models across the Gemini 3.x and 2.5 generations.
* **Greenfield Idiomatic Config:** Clean schema without legacy migration bloat; pre-populated template provider configurations.
* **User Ergonomics:** Live API key verification, annotated menus, and auto-suggested fallbacks.
* **Binary Portability:** Zero external runtime dependencies and static compilation.

---

## Decision Outcome

**Chosen Option: Hybrid Zero-Bloat Architecture with Modern Greenfield Config & Fast-Model Engine**

### 1. Optimal Fast Models for Commit Message Generation (Top 6 Options)

| Model ID | Display Name | Category | Primary Advantage for Git Hook |
| :--- | :--- | :--- | :--- |
| **`gemini-3.7-flash`** | Gemini 3.7 Flash | Frontier Fast | **Recommended Default.** SOTA code diff understanding and concise synthesis. |
| **`gemini-3.6-flash`** | Gemini 3.6 Flash | Efficient Fast | High intelligence with ~17% reduced token overhead and fast time-to-first-token. |
| **`gemini-3.5-flash`** | Gemini 3.5 Flash | High-Speed Workhorse | Reliable, proven fast performance for code repositories. |
| **`gemini-3.5-flash-lite`** | Gemini 3.5 Flash-Lite | Ultra-Low Latency | Highest throughput and lowest latency; ideal for high-frequency committers. |
| **`gemini-2.5-flash`** | Gemini 2.5 Flash | Stable Fast | Ubiquitous fallback tier with balanced speed and accuracy. |
| **`gemini-2.5-flash-lite`** | Gemini 2.5 Flash-Lite | Ultra-Light Fast | Cost-effective and lightweight baseline fallback. |

---

### 2. Greenfield Configuration Schema

```go
type Config struct {
    ActiveProvider    string                    `json:"active_provider"`
    Providers         map[string]ProviderConfig `json:"providers"`
    TimeoutSeconds    int                       `json:"timeout_seconds"`
    MaxDiffBytes      int                       `json:"max_diff_bytes"`
    RetryCount        int                       `json:"retry_count"`
    RetryDelaySeconds int                       `json:"retry_delay_seconds"`
}

type ProviderConfig struct {
    APIKey         string   `json:"api_key,omitempty"`
    Model          string   `json:"model"`
    FallbackModels []string `json:"fallback_models,omitempty"`
}
```

#### Template Configuration Defaults
When initializing a fresh configuration, template provider entries are populated with modern defaults:

```json
{
  "active_provider": "gemini",
  "providers": {
    "gemini": {
      "api_key": "",
      "model": "gemini-3.7-flash",
      "fallback_models": [
        "gemini-3.6-flash",
        "gemini-3.5-flash",
        "gemini-2.5-flash"
      ]
    },
    "openai": {
      "api_key": "",
      "model": "gpt-4.1-mini",
      "fallback_models": [
        "gpt-4.1-nano",
        "gpt-4o-mini"
      ]
    },
    "claude": {
      "api_key": "",
      "model": "claude-haiku-4-5",
      "fallback_models": [
        "claude-sonnet-5",
        "claude-sonnet-4-6"
      ]
    }
  },
  "timeout_seconds": 120,
  "max_diff_bytes": 32000,
  "retry_count": 3,
  "retry_delay_seconds": 3
}
```

* Legacy migration logic (`prepare-commit-msg-embedded`) is purged in favor of pure standard path resolution (`os.UserConfigDir`).

---

### 3. Dynamic Discovery & Heuristic Ranking

* Dynamic discovery queries `/v1beta/models`.
* Rejects models without `generateContent`.
* Excludes reasoning/Pro models (`gemini-*-pro`, `o1/o3/o4`, `thinking`), audio, image, and deprecated `2.0`/`1.5` generations.
* Dynamically backfills newly available fast models up to `MaxListedModels = 6`.

---

### 4. Interactive CLI Configure UX

1. **Real-Time Key Verification:** Tests key against `llmprovider.ListAvailableModels` before saving.
2. **Annotated Selection Menu:** Displays human-readable labels and speed badges.
3. **Auto-Populated Fallbacks:** Suggests the next best 2–3 fast models.

---

## Action Plan & Implementation Mapping

| Target Component | File Path | Action |
| :--- | :--- | :--- |
| **Catalog & Ranking** | [`mcplib/llmprovider/models_catalog.go`](file:///data/cache/go/pkg/mod/github.com/maccavelli/mcplib@v0.2.0/llmprovider/models_catalog.go) | Update `StaticGemini` to Top 6 fast models; update `RankGeminiModel` to score 3.x Flash and penalize Pro/Reasoning; update `curateFromCatalog` with dynamic backfill. |
| **Discovery Logic** | [`mcplib/llmprovider/discovery.go`](file:///data/cache/go/pkg/mod/github.com/maccavelli/mcplib@v0.2.0/llmprovider/discovery.go) | Enhance `/v1beta/models` parsing and error handling. |
| **Config Schema & Defaults** | [`internal/config/config.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/config/config.go) | Purge legacy migration code; ensure `ApplyDefaults` populates modern default models and fallbacks for all providers in the template. |
| **Configure Wizard** | [`internal/ui/setup.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/ui/setup.go) | Add live API key verification; render annotated model menu with speed badges; auto-suggest fast fallbacks. |
| **Documentation & CLI Help** | [`README.md`](file:///home/mac/gitrepos/prepare-commit-msg/README.md) & [`main.go`](file:///home/mac/gitrepos/prepare-commit-msg/main.go) | Update documentation, help text, and examples to reference `gemini-3.7-flash`. |

---

## References

1. [Google GenAI Go SDK Repository (`googleapis/go-genai`)](https://github.com/googleapis/go-genai)
2. [Google Gemini API Models Documentation](https://ai.google.dev/gemini-api/docs/models/gemini)
3. [Conventional Commits 1.0.0 Specification](https://www.conventionalcommits.org/en/v1.0.0/)
