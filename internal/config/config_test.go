package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func isolateHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	// Windows UserConfigDir uses AppData.
	t.Setenv("AppData", filepath.Join(tmp, "AppData", "Roaming"))
	// macOS UserConfigDir uses HOME/Library/Application Support — HOME is set.
	return tmp
}

func TestConfig_SaveAndLoad(t *testing.T) {
	isolateHome(t)

	conf := &Config{
		ActiveProvider: "test-provider",
		Providers: map[string]ProviderConfig{
			"test-provider": {
				APIKey: "test-key",
				Model:  "test-model",
			},
		},
	}

	if err := conf.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.ActiveProvider != conf.ActiveProvider {
		t.Errorf("expected active provider %q, got %q", conf.ActiveProvider, loaded.ActiveProvider)
	}

	pc, ok := loaded.Providers["test-provider"]
	if !ok || pc.APIKey != "test-key" {
		t.Errorf("expected provider config not found or incorrect")
	}
	if loaded.TimeoutSeconds != DefaultTimeoutSeconds {
		t.Errorf("expected default timeout %d, got %d", DefaultTimeoutSeconds, loaded.TimeoutSeconds)
	}
}

func TestLoad_AppliesDefaultsOnEmpty(t *testing.T) {
	isolateHome(t)

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TimeoutSeconds != DefaultTimeoutSeconds {
		t.Errorf("timeout: got %d want %d", loaded.TimeoutSeconds, DefaultTimeoutSeconds)
	}
	if loaded.MaxDiffBytes != DefaultMaxDiffBytes {
		t.Errorf("maxdiff: got %d want %d", loaded.MaxDiffBytes, DefaultMaxDiffBytes)
	}
	if loaded.RetryCount != DefaultRetryCount {
		t.Errorf("retry: got %d want %d", loaded.RetryCount, DefaultRetryCount)
	}
	if loaded.RetryDelaySeconds != DefaultRetryDelaySeconds {
		t.Errorf("delay: got %d want %d", loaded.RetryDelaySeconds, DefaultRetryDelaySeconds)
	}
	for _, p := range SupportedProviders {
		if _, ok := loaded.Providers[p]; !ok {
			t.Errorf("missing provider slot %s", p)
		}
	}
}

func TestConfig_GetActive(t *testing.T) {
	conf := &Config{
		ActiveProvider: "p1",
		Providers: map[string]ProviderConfig{
			"p1": {APIKey: "k1", Model: "m1"},
		},
	}

	pc, err := conf.GetActive()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if pc.APIKey != "k1" {
		t.Errorf("expected k1, got %s", pc.APIKey)
	}

	conf.ActiveProvider = "non-existent"
	if _, err := conf.GetActive(); err == nil {
		t.Error("expected error for non-existent provider, got nil")
	}

	conf.ActiveProvider = ""
	if _, err := conf.GetActive(); err == nil {
		t.Error("expected error for empty active provider, got nil")
	}
}

func TestConfig_TemplateDefaults(t *testing.T) {
	isolateHome(t)

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ActiveProvider != "gemini" {
		t.Errorf("expected default active provider 'gemini', got %q", loaded.ActiveProvider)
	}

	g := loaded.Providers["gemini"]
	if g.Model != "gemini-3.7-flash" {
		t.Errorf("expected gemini model 'gemini-3.7-flash', got %q", g.Model)
	}
	if len(g.FallbackModels) < 2 || g.FallbackModels[0] != "gemini-3.6-flash" {
		t.Errorf("expected gemini fallbacks, got %v", g.FallbackModels)
	}

	o := loaded.Providers["openai"]
	if o.Model != "gpt-4.1-mini" {
		t.Errorf("expected openai model 'gpt-4.1-mini', got %q", o.Model)
	}

	c := loaded.Providers["claude"]
	if c.Model != "claude-haiku-4-5" {
		t.Errorf("expected claude model 'claude-haiku-4-5', got %q", c.Model)
	}
}

func TestConfig_UserRepro(t *testing.T) {
	isolateHome(t)

	// Seed via Save path so primary location is correct on all OS.
	seed := &Config{
		ActiveProvider: "gemini",
		Providers: map[string]ProviderConfig{
			"gemini": {APIKey: "REDACTED_API_KEY", Model: "gemini-2.5-flash-lite"},
			"openai": {APIKey: "", Model: "gpt-4o"},
		},
		TimeoutSeconds:    120,
		MaxDiffBytes:      32000,
		RetryCount:        3,
		RetryDelaySeconds: 3,
	}
	if err := seed.Save(); err != nil {
		t.Fatal(err)
	}

	conf, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	provider := "claude"
	pc := conf.Providers[provider]
	pc.APIKey = "new-claude-key"
	pc.Model = "claude-3-5-haiku-latest"
	conf.ActiveProvider = provider
	conf.Providers[provider] = pc

	if err := conf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if loaded.ActiveProvider != "claude" {
		t.Errorf("Expected active provider 'claude', got %q", loaded.ActiveProvider)
	}
	if pc, ok := loaded.Providers["claude"]; !ok || pc.APIKey != "new-claude-key" {
		t.Error("Claude provider missing or incorrect after reload")
	}
	if pc, ok := loaded.Providers["gemini"]; !ok || pc.APIKey != "REDACTED_API_KEY" {
		t.Error("Gemini provider lost or incorrect after reload")
	}
}

func TestClampFallbacks(t *testing.T) {
	in := []string{"a", "b", "c", "d", "a", "", "e"}
	out := ClampFallbacks(in)
	if len(out) != MaxFallbacks {
		t.Fatalf("expected %d fallbacks, got %d: %v", MaxFallbacks, len(out), out)
	}
	if out[0] != "a" || out[1] != "b" || out[2] != "c" {
		t.Errorf("unexpected order/content: %v", out)
	}
}

func TestResolveAPIKey(t *testing.T) {
	pc := ProviderConfig{APIKey: "cfg"}
	if got := ResolveAPIKey(pc, "openai", true, func(string) string { return "env" }); got != "cfg" {
		t.Errorf("config should win: %q", got)
	}
	pc.APIKey = ""
	if got := ResolveAPIKey(pc, "openai", true, func(k string) string {
		if k == "OPENAI_API_KEY" {
			return "env"
		}
		return ""
	}); got != "env" {
		t.Errorf("env fallback: %q", got)
	}
	if got := ResolveAPIKey(pc, "openai", false, func(string) string { return "env" }); got != "" {
		t.Errorf("no-env should block: %q", got)
	}
}

func TestValidateActive(t *testing.T) {
	if err := ValidateActive("openai", ProviderConfig{Model: "m"}, ""); err == nil {
		t.Error("expected missing key error")
	}
	if err := ValidateActive("openai", ProviderConfig{}, "key"); err == nil {
		t.Error("expected missing model error")
	}
	if err := ValidateActive("openai", ProviderConfig{Model: "m"}, "key"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestGetConfigPath_Error(t *testing.T) {
	oldHome := userHomeDir
	oldCfg := userConfigDir
	defer func() {
		userHomeDir = oldHome
		userConfigDir = oldCfg
	}()

	userConfigDir = func() (string, error) {
		return "", fmt.Errorf("mock configdir error")
	}
	userHomeDir = func() (string, error) {
		return "", fmt.Errorf("mock homedir error")
	}

	if _, err := GetConfigPath(); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestLoad_Errors(t *testing.T) {
	isolateHome(t)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		// path is a file path; create parent then a directory with that name
		_ = os.RemoveAll(path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := Load(); err == nil {
		t.Error("expected error reading a directory, got nil")
	}
	_ = os.RemoveAll(path)

	if err := os.WriteFile(path, []byte(`{ bad, json }`), 0o644); err != nil {
		// parent may be missing after RemoveAll
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, []byte(`{ bad, json }`), 0o644)
	}
	if _, err := Load(); err == nil {
		t.Error("expected unmarshal error, got nil")
	}
}

func TestSave_Errors(t *testing.T) {
	c := &Config{}
	oldHome := userHomeDir
	oldCfg := userConfigDir
	defer func() {
		userHomeDir = oldHome
		userConfigDir = oldCfg
	}()

	userConfigDir = func() (string, error) {
		return "", fmt.Errorf("mock error")
	}
	userHomeDir = func() (string, error) {
		return "", fmt.Errorf("mock error")
	}
	if err := c.Save(); err == nil {
		t.Error("expected save error due to home dir failure")
	}

	userHomeDir = oldHome
	userConfigDir = oldCfg

	tmpDir := isolateHome(t)
	// File where config directory should be → MkdirAll fails inside atomic write.
	cfgPath, err := GetConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Dir(filepath.Dir(cfgPath)) // prepare-commit-msg parent
	// Make the app config parent a file.
	_ = os.RemoveAll(filepath.Dir(cfgPath))
	if err := os.MkdirAll(filepath.Dir(parent), 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows AppData path layout differs; block the immediate parent of config.json's parent.
	block := filepath.Dir(cfgPath)
	if err := os.WriteFile(block, []byte("file-not-dir"), 0o644); err != nil {
		// parent chain may not exist
		_ = os.MkdirAll(filepath.Dir(block), 0o755)
		if err := os.WriteFile(block, []byte("file-not-dir"), 0o644); err != nil {
			t.Skipf("could not create blocking file at %s: %v (GOOS=%s tmp=%s)", block, err, runtime.GOOS, tmpDir)
		}
	}

	if err := c.Save(); err == nil {
		t.Error("expected save error due to mkdir failure")
	}
}

func TestFallbackModels(t *testing.T) {
	c := Config{
		Providers: map[string]ProviderConfig{
			"openai": {Model: "gpt-4o", FallbackModels: []string{"gpt-4o-mini"}},
		},
	}
	pc := c.Providers["openai"]
	if len(pc.FallbackModels) != 1 || pc.FallbackModels[0] != "gpt-4o-mini" {
		t.Errorf("expected 1 fallback model 'gpt-4o-mini', got %v", pc.FallbackModels)
	}
}
