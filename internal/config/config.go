package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/auth"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/catalog"
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
	string(llmprovider.ProviderGemini),
	string(llmprovider.ProviderOpenAI),
	string(llmprovider.ProviderClaude),
	string(llmprovider.ProviderGrok),
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

// AuthKindVendorCLI selects a vendor CLI's own login (Codex or Grok), read in
// place through auth.VendorCLISession (mcplib MADR 0012 §5.1).
const AuthKindVendorCLI = "vendor_cli"

// ProviderConfig stores credentials and model selection for a single LLM provider.
type ProviderConfig struct {
	// AuthKind is empty for legacy/API-key auth, "oauth" for a saved session,
	// and AuthKindVendorCLI for a vendor CLI's own login.
	AuthKind string `json:"auth_kind,omitempty"`
	// VendorAuthPath is the vendor CLI's auth file when AuthKind is
	// AuthKindVendorCLI. It holds no token: the file is read on every request,
	// and the CLI keeps refreshing its own login.
	VendorAuthPath string   `json:"vendor_auth_path,omitempty"`
	APIKey         string   `json:"api_key"`
	Model          string   `json:"model"`
	FallbackModels []string `json:"fallback_models,omitempty"`
	// Organization is the Kilo organization a device login chose. Generation
	// sends it; empty is the personal account.
	Organization string `json:"organization,omitempty"`
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

// OAuthDir returns the private token directory beside the main config file.
func OAuthDir() (string, error) {
	path, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(path), "oauth"), nil
}

// NewOAuthStore creates the file-backed store for provider OAuth sessions.
func NewOAuthStore() (*auth.FileTokenStore, error) {
	dir, err := OAuthDir()
	if err != nil {
		return nil, err
	}
	return auth.NewFileTokenStore(dir)
}

// DefaultModelForProvider returns the recommended primary model for a given provider.
func DefaultModelForProvider(provider string) string {
	models := catalog.Static(llmprovider.ProviderID(provider))
	if len(models) > 0 {
		return models[0]
	}
	return ""
}

// DefaultFallbacksForProvider returns the recommended fallback models for a given provider.
func DefaultFallbacksForProvider(provider string, primary string) []string {
	models := catalog.Static(llmprovider.ProviderID(provider))
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
		c.ActiveProvider = string(llmprovider.ProviderGemini)
	}
	for provider, pc := range c.Providers {
		switch {
		case IsOAuth(pc):
			pc.AuthKind = "oauth"
		case IsVendorCLI(pc):
			pc.AuthKind = AuthKindVendorCLI
		default:
			pc.AuthKind = ""
		}
		c.Providers[provider] = pc
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

// ClaudeKeyFallback is the variable earlier releases read for Claude. It is
// still read when ANTHROPIC_API_KEY is unset, so an existing setup keeps
// working.
const ClaudeKeyFallback = "CLAUDE_API_KEY"

// LookupEnv returns getenv, or os.Getenv when it is nil, with
// ClaudeKeyFallback read for ANTHROPIC_API_KEY when that is unset.
func LookupEnv(getenv func(string) string) func(string) string {
	if getenv == nil {
		getenv = os.Getenv
	}
	anthropic := llmprovider.ProviderEnvVars()[llmprovider.ProviderClaude]
	return func(name string) string {
		value := getenv(name)
		if name == anthropic && strings.TrimSpace(value) == "" {
			return getenv(ClaudeKeyFallback)
		}
		return value
	}
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
	if envName, ok := llmprovider.ProviderEnvVars()[llmprovider.ProviderID(provider)]; ok {
		return strings.TrimSpace(LookupEnv(getenv)(envName))
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
		if v, ok := llmprovider.ProviderEnvVars()[llmprovider.ProviderID(provider)]; ok {
			envHint = fmt.Sprintf(" or set %s", v)
		}
		return fmt.Errorf("no API key for provider %q; run 'prepare-commit-msg configure'%s", provider, envHint)
	}
	if strings.TrimSpace(pc.Model) == "" {
		return fmt.Errorf("no model configured for provider %q; run 'prepare-commit-msg configure'", provider)
	}
	return nil
}

// IsOAuth reports whether a provider config selects subscription authentication.
func IsOAuth(pc ProviderConfig) bool {
	return strings.EqualFold(pc.AuthKind, "oauth")
}

// IsVendorCLI reports whether a provider config reads a vendor CLI's login.
func IsVendorCLI(pc ProviderConfig) bool {
	return strings.EqualFold(pc.AuthKind, AuthKindVendorCLI)
}

// ValidateVendorCLI checks that a vendor CLI login has a model and an auth
// file path. The file, and the token in it, are read on use.
func ValidateVendorCLI(provider string, pc ProviderConfig) error {
	if strings.TrimSpace(provider) == "" {
		return fmt.Errorf("no active provider configured; please run 'prepare-commit-msg configure'")
	}
	if strings.TrimSpace(pc.Model) == "" {
		return fmt.Errorf("no model configured for provider %q; run 'prepare-commit-msg configure'", provider)
	}
	if strings.TrimSpace(pc.VendorAuthPath) == "" {
		return fmt.Errorf("no CLI login path for provider %q; run 'prepare-commit-msg configure'", provider)
	}
	return nil
}

// ValidateOAuth checks that an OAuth provider has both a model and a saved session.
func ValidateOAuth(
	ctx context.Context,
	provider string,
	pc ProviderConfig,
	store auth.TokenStore,
) error {
	if strings.TrimSpace(provider) == "" {
		return fmt.Errorf("no active provider configured; please run 'prepare-commit-msg configure'")
	}
	if strings.TrimSpace(pc.Model) == "" {
		return fmt.Errorf("no model configured for provider %q; run 'prepare-commit-msg configure'", provider)
	}
	if store == nil {
		return fmt.Errorf("no OAuth session for provider %q; run 'prepare-commit-msg configure'", provider)
	}
	session, err := store.Load(ctx, llmprovider.ProviderID(provider))
	if err != nil {
		return fmt.Errorf("load OAuth session for provider %q: %w", provider, err)
	}
	if session == nil {
		return fmt.Errorf("no OAuth session for provider %q; run 'prepare-commit-msg configure'", provider)
	}
	if err := auth.ValidateOAuthSession(session); err != nil {
		return fmt.Errorf("OAuth session for provider %q is not usable: %w; run 'prepare-commit-msg configure'", provider, err)
	}
	return nil
}
