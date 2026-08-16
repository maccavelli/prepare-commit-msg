package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

func isolate(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("AppData", filepath.Join(tmp, "AppData", "Roaming"))
}

func TestRunSetupInteractive_Success(t *testing.T) {
	isolate(t)

	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)

	oldEnv := osGetenv
	defer func() { osGetenv = oldEnv }()
	osGetenv = func(k string) string {
		if k == "GEMINI_API_KEY" {
			return "test-key"
		}
		return ""
	}

	// 1: gemini
	// y: use env key
	// Static gemini catalog has 6 models → 7 is Other
	// Enter custom model
	// fallbacks: enter = recommended
	// operational: all enter (defaults)
	input := "1\ny\n7\nmy-custom-model\n\n\n\n\n\n"
	r := strings.NewReader(input)

	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.ActiveProvider != "gemini" {
		t.Errorf("expected gemini, got %q", conf.ActiveProvider)
	}
	if conf.Providers["gemini"].Model != "my-custom-model" {
		t.Errorf("expected my-custom-model, got %q", conf.Providers["gemini"].Model)
	}
	if conf.Providers["gemini"].APIKey != "test-key" {
		t.Errorf("expected test-key")
	}
	if conf.TimeoutSeconds != config.DefaultTimeoutSeconds {
		t.Errorf("timeout defaults: %d", conf.TimeoutSeconds)
	}
}

func TestRunSetupInteractive_FallbackMultiSelect(t *testing.T) {
	isolate(t)

	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)

	oldEnv := osGetenv
	defer func() { osGetenv = oldEnv }()
	osGetenv = func(k string) string { return "" }

	// 3: claude, manual key, model 1, fallbacks "1,2", operational defaults
	input := "3\nmanual-key\n1\n1,2\n\n\n\n\n"
	r := strings.NewReader(input)

	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.ActiveProvider != "claude" {
		t.Errorf("expected claude, got %s", conf.ActiveProvider)
	}
	if conf.Providers["claude"].APIKey != "manual-key" {
		t.Errorf("expected manual-key")
	}
	fb := conf.Providers["claude"].FallbackModels
	if len(fb) == 0 || len(fb) > config.MaxFallbacks {
		t.Errorf("expected 1-%d fallbacks, got %v", config.MaxFallbacks, fb)
	}
}

func TestRunSetupInteractive_CoverageBranches(t *testing.T) {
	isolate(t)

	oldEnv := osGetenv
	defer func() { osGetenv = oldEnv }()
	osGetenv = func(k string) string { return "" }

	t.Run("OpenAI path", func(t *testing.T) {
		isolate(t)
		conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
		config.ApplyDefaults(conf)
		// 2: openai, key, model 1, fallbacks enter, ops enter
		input := "2\ntest-key\n1\n\n\n\n\n\n"
		r := strings.NewReader(input)
		if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, r); err != nil {
			t.Fatalf("%v", err)
		}
		if conf.ActiveProvider != "openai" {
			t.Errorf("Expected openai")
		}
	})

	t.Run("Empty choice defaults", func(t *testing.T) {
		isolate(t)
		conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
		config.ApplyDefaults(conf)
		input := "\ntest-key\n\n\n\n\n\n\n"
		r := strings.NewReader(input)
		if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, r); err != nil {
			t.Fatalf("%v", err)
		}
		if conf.ActiveProvider != "gemini" {
			t.Errorf("Expected gemini")
		}
	})
}

func TestRunSetupNonInteractive(t *testing.T) {
	isolate(t)

	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)

	oldEnv := osGetenv
	defer func() { osGetenv = oldEnv }()
	osGetenv = func(k string) string {
		if k == "OPENAI_API_KEY" {
			return "env-key"
		}
		return ""
	}

	opts := SetupOptions{
		Provider:  "openai",
		Model:     "gpt-4o-mini",
		Fallbacks: []string{"gpt-4o", "o3-mini", "o4-mini", "extra-ignored"},
		Yes:       true,
	}
	if err := runSetupNonInteractive(context.Background(), conf, opts); err != nil {
		t.Fatal(err)
	}
	if conf.ActiveProvider != "openai" {
		t.Fatalf("provider: %s", conf.ActiveProvider)
	}
	if conf.Providers["openai"].APIKey != "env-key" {
		t.Errorf("key from env: %q", conf.Providers["openai"].APIKey)
	}
	if conf.Providers["openai"].Model != "gpt-4o-mini" {
		t.Errorf("model: %q", conf.Providers["openai"].Model)
	}
	if len(conf.Providers["openai"].FallbackModels) != config.MaxFallbacks {
		t.Errorf("fallbacks capped: %v", conf.Providers["openai"].FallbackModels)
	}
}

func TestRecommendedFallbacks(t *testing.T) {
	models := []string{"a", "b", "c", "d"}
	fb := recommendedFallbacks(models, "a")
	if len(fb) != 3 || fb[0] != "b" {
		t.Errorf("got %v", fb)
	}
}

func TestRunSetup(t *testing.T) {
	isolate(t)
	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)

	oldEnv := osGetenv
	defer func() { osGetenv = oldEnv }()
	osGetenv = func(k string) string { return "" }

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte("\ntest-key\n\n\n\n\n\n\n"))
		w.Close()
	}()

	err := RunSetup(context.Background(), conf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultModels(t *testing.T) {
	if len(defaultModels(providerOpenAI)) == 0 {
		t.Error("expected default models for openai")
	}
	if len(defaultModels(providerGemini)) == 0 {
		t.Error("expected default models for gemini")
	}
	if len(defaultModels(providerClaude)) == 0 {
		t.Error("expected default models for claude")
	}
	if len(defaultModels("unknown")) != 0 {
		t.Error("expected no default models for unknown provider")
	}
}
