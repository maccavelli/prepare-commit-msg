package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maccavelli/mcplib/llmprovider"
)

// Operational defaults applied whenever a config is loaded or created.
const (
	DefaultTimeoutSeconds    = 120
	DefaultMaxDiffBytes      = 32000
	DefaultRetryCount        = 3
	DefaultRetryDelaySeconds = 3
	// MaxFallbacks is the maximum number of fallback models stored/used.
	MaxFallbacks = 3
)

// SupportedProviders is the canonical ordered list of LLM providers.
var SupportedProviders = []string{
	llmprovider.ProviderGemini,
	llmprovider.ProviderOpenAI,
	llmprovider.ProviderClaude,
}

// Config holds the application configuration including the active LLM provider,
// per-provider settings, and global operational constraints.
type Config struct {
	ActiveProvider string                    `json:"active_provider"`
	Providers      map[string]ProviderConfig `json:"providers"`
	// TimeoutSeconds is the maximum duration in seconds for LLM generation.
	TimeoutSeconds int `json:"timeout_seconds"`
	// MaxDiffBytes is the maximum size in bytes for the git diff sent to the LLM.
	MaxDiffBytes int `json:"max_diff_bytes"`
	// RetryCount is the total number of retries before giving up.
	RetryCount int `json:"retry_count"`
	// RetryDelaySeconds is the wait time in seconds between each retry.
	RetryDelaySeconds int `json:"retry_delay_seconds"`
}

// ProviderConfig stores credentials and model selection for a single LLM provider.
type ProviderConfig struct {
	APIKey         string   `json:"api_key"`
	Model          string   `json:"model"`
	FallbackModels []string `json:"fallback_models,omitempty"`
}

var (
	userHomeDir   = os.UserHomeDir
	userConfigDir = os.UserConfigDir
)

// GetConfigPath returns the primary configuration file path using the platform
// user config directory (UserConfigDir). Legacy ~/.config paths are still
// read by Load for migration.
func GetConfigPath() (string, error) {
	cfgDir, err := userConfigDir()
	if err != nil {
		// Fall back to home/.config when UserConfigDir is unavailable.
		home, homeErr := userHomeDir()
		if homeErr != nil {
			return "", fmt.Errorf("config dir: %w; home: %w", err, homeErr)
		}
		return filepath.Join(home, ".config", "prepare-commit-msg", "config.json"), nil
	}
	return filepath.Join(cfgDir, "prepare-commit-msg", "config.json"), nil
}

// DefaultModelForProvider returns the recommended primary model for a given provider.
func DefaultModelForProvider(provider string) string {
	models := llmprovider.StaticModels(provider)
	if len(models) > 0 {
		return models[0]
	}
	return ""
}

// DefaultFallbacksForProvider returns the recommended fallback models for a given provider.
func DefaultFallbacksForProvider(provider string, primary string) []string {
	models := llmprovider.StaticModels(provider)
	var out []string
	for _, m := range models {
		if m == primary {
			continue
		}
		out = append(out, m)
		if len(out) >= MaxFallbacks {
			break
		}
	}
	return out
}

// ApplyDefaults fills operational defaults and ensures all supported providers
// have complete template configurations with modern default models and fallbacks.
func ApplyDefaults(c *Config) {
	if c.Providers == nil {
		c.Providers = make(map[string]ProviderConfig)
	}
	if c.ActiveProvider == "" {
		c.ActiveProvider = llmprovider.ProviderGemini
	}
	for _, p := range SupportedProviders {
		pc, ok := c.Providers[p]
		if !ok {
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

// ClampFallbacks returns at most MaxFallbacks models, dropping empties and duplicates.
func ClampFallbacks(models []string) []string {
	seen := make(map[string]struct{}, len(models))
	out := make([]string, 0, MaxFallbacks)
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
		if len(out) >= MaxFallbacks {
			break
		}
	}
	return out
}

// Load reads configuration from the platform config path. Defaults are always applied.
func Load() (*Config, error) {
	primary, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(primary)
	if err != nil {
		if os.IsNotExist(err) {
			conf := &Config{}
			ApplyDefaults(conf)
			return conf, nil
		}
		return nil, err
	}

	return parseAndDefault(data)
}

func parseAndDefault(data []byte) (*Config, error) {
	var conf Config
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	ApplyDefaults(&conf)
	return &conf, nil
}

// Save persists the configuration to the primary path (atomic, mode 0600).
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	ApplyDefaults(c)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return atomicWriteConfig(path, data)
}

// GetActive retrieves the configuration for the active provider.
func (c *Config) GetActive() (ProviderConfig, error) {
	if c.ActiveProvider == "" {
		return ProviderConfig{}, fmt.Errorf("no active provider configured; please run 'prepare-commit-msg configure'")
	}
	pc, ok := c.Providers[c.ActiveProvider]
	if !ok {
		return ProviderConfig{}, fmt.Errorf("active provider %q not found in config", c.ActiveProvider)
	}
	return pc, nil
}

// ResolveAPIKey returns the provider API key from config, or from the matching
// environment variable when the config key is empty. useEnv controls env lookup.
func ResolveAPIKey(pc ProviderConfig, provider string, useEnv bool, getenv func(string) string) string {
	if strings.TrimSpace(pc.APIKey) != "" {
		return strings.TrimSpace(pc.APIKey)
	}
	if !useEnv {
		return ""
	}
	if getenv == nil {
		getenv = os.Getenv
	}
	if envName, ok := llmprovider.ProviderEnvVars[provider]; ok {
		return strings.TrimSpace(getenv(envName))
	}
	return ""
}

// ValidateActive checks that the active provider has a usable key and model.
// apiKey should already be resolved (config + optional env).
func ValidateActive(provider string, pc ProviderConfig, apiKey string) error {
	if strings.TrimSpace(provider) == "" {
		return fmt.Errorf("no active provider configured; please run 'prepare-commit-msg configure'")
	}
	if strings.TrimSpace(apiKey) == "" {
		envHint := ""
		if v, ok := llmprovider.ProviderEnvVars[provider]; ok {
			envHint = fmt.Sprintf(" or set %s", v)
		}
		return fmt.Errorf("no API key for provider %q; run 'prepare-commit-msg configure'%s", provider, envHint)
	}
	if strings.TrimSpace(pc.Model) == "" {
		return fmt.Errorf("no model configured for provider %q; run 'prepare-commit-msg configure'", provider)
	}
	return nil
}
