package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maccavelli/mcplib/llmprovider"
	"github.com/maccavelli/prepare-commit-msg/internal/config"
	"github.com/maccavelli/prepare-commit-msg/internal/git"
)

func TestRunAnalyzer_OAuthDoesNotCallNewProviderWithAccessToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	store, err := config.NewOAuthStore()
	if err != nil {
		t.Fatalf("NewOAuthStore() error = %v", err)
	}
	session := &llmprovider.OAuthSession{
		Provider: llmprovider.ProviderOpenAI,
		Access:   "chatgpt-access",
		Issuer:   llmprovider.DefaultOpenAIIssuer,
	}
	if err := store.Save(context.Background(), llmprovider.ProviderOpenAI, session); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	originalNewProvider := newProvider
	originalNewProviderWithSource := newProviderWithSource
	originalGenerate := generateWithRetry
	t.Cleanup(func() {
		newProvider = originalNewProvider
		newProviderWithSource = originalNewProviderWithSource
		generateWithRetry = originalGenerate
	})
	newProvider = func(string, string, string, ...llmprovider.ProviderOption) (llmprovider.Provider, error) {
		t.Fatal("OAuth generation called NewProvider")
		return nil, nil
	}
	var capturedSource llmprovider.TokenSource
	newProviderWithSource = func(
		_ string,
		source llmprovider.TokenSource,
		_ string,
		_ ...llmprovider.ProviderOption,
	) (llmprovider.Provider, error) {
		capturedSource = source
		return analyzerTestProvider{}, nil
	}
	generateWithRetry = func(context.Context, llmprovider.Provider, string, int, time.Duration) (string, error) {
		return "feat: use OAuth session", nil
	}

	conf := &config.Config{
		ActiveProvider: llmprovider.ProviderOpenAI,
		Providers: map[string]config.ProviderConfig{
			llmprovider.ProviderOpenAI: {AuthKind: "oauth", Model: "gpt-5.4"},
		},
		TimeoutSeconds: 5,
	}
	messagePath := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(messagePath, nil, 0o600); err != nil {
		t.Fatalf("write message file: %v", err)
	}
	info := &git.Info{Files: []string{"main.go"}, Additions: 1}
	if err := runAnalyzer(messagePath, conf, info); err != nil {
		t.Fatalf("runAnalyzer() error = %v", err)
	}
	if _, ok := capturedSource.(*llmprovider.OAuthSession); !ok {
		t.Fatalf("NewProviderWithSource() source = %T, want *OAuthSession", capturedSource)
	}
}

// TestRunAnalyzer_VendorCLIReadsThrough: a vendor CLI login generates through
// a VendorCLISession on the saved auth file, never through an API key.
func TestRunAnalyzer_VendorCLIReadsThrough(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var pc config.ProviderConfig
	if err := json.Unmarshal([]byte(`{"auth_kind":"vendor_cli","vendor_auth_path":"/x/grok/auth.json",`+
		`"model":"grok-4.6"}`), &pc); err != nil {
		t.Fatalf("decode: %v", err)
	}

	originalNewProvider := newProvider
	originalNewProviderWithSource := newProviderWithSource
	originalGenerate := generateWithRetry
	originalEnv := osGetenv
	t.Cleanup(func() {
		newProvider = originalNewProvider
		newProviderWithSource = originalNewProviderWithSource
		generateWithRetry = originalGenerate
		osGetenv = originalEnv
	})
	osGetenv = func(string) string { return "" }
	newProvider = func(string, string, string, ...llmprovider.ProviderOption) (llmprovider.Provider, error) {
		t.Fatal("vendor CLI generation called NewProvider")
		return nil, nil
	}
	var capturedSource llmprovider.TokenSource
	newProviderWithSource = func(
		_ string,
		source llmprovider.TokenSource,
		_ string,
		_ ...llmprovider.ProviderOption,
	) (llmprovider.Provider, error) {
		capturedSource = source
		return analyzerTestProvider{}, nil
	}
	generateWithRetry = func(context.Context, llmprovider.Provider, string, int, time.Duration) (string, error) {
		return "feat: use the Grok CLI login", nil
	}

	conf := &config.Config{
		ActiveProvider: llmprovider.ProviderGrok,
		Providers:      map[string]config.ProviderConfig{llmprovider.ProviderGrok: pc},
		TimeoutSeconds: 5,
	}
	messagePath := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(messagePath, nil, 0o600); err != nil {
		t.Fatalf("write message file: %v", err)
	}
	info := &git.Info{Files: []string{"main.go"}, Additions: 1}
	if err := runAnalyzer(messagePath, conf, info); err != nil {
		t.Fatalf("runAnalyzer() error = %v", err)
	}
	session, ok := capturedSource.(*llmprovider.VendorCLISession)
	if !ok || session.Provider != llmprovider.ProviderGrok || session.Path != "/x/grok/auth.json" {
		t.Fatalf("NewProviderWithSource() source = %#v, want the Grok VendorCLISession on /x/grok/auth.json", capturedSource)
	}
}

// TestRunAnalyzer_VendorCLIWithoutPathFails: a vendor CLI login with no saved
// path sends the user back to configure, and builds no provider.
func TestRunAnalyzer_VendorCLIWithoutPathFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var pc config.ProviderConfig
	if err := json.Unmarshal([]byte(`{"auth_kind":"vendor_cli","model":"grok-4.6"}`), &pc); err != nil {
		t.Fatalf("decode: %v", err)
	}
	originalNewProvider := newProvider
	originalNewProviderWithSource := newProviderWithSource
	originalEnv := osGetenv
	t.Cleanup(func() {
		newProvider = originalNewProvider
		newProviderWithSource = originalNewProviderWithSource
		osGetenv = originalEnv
	})
	osGetenv = func(string) string { return "" }
	newProvider = func(string, string, string, ...llmprovider.ProviderOption) (llmprovider.Provider, error) {
		t.Fatal("a vendor CLI login without a path built an API-key provider")
		return nil, nil
	}
	newProviderWithSource = func(string, llmprovider.TokenSource, string, ...llmprovider.ProviderOption) (llmprovider.Provider, error) {
		t.Fatal("a vendor CLI login without a path built a provider")
		return nil, nil
	}
	conf := &config.Config{
		ActiveProvider: llmprovider.ProviderGrok,
		Providers:      map[string]config.ProviderConfig{llmprovider.ProviderGrok: pc},
		TimeoutSeconds: 5,
	}
	info := &git.Info{Files: []string{"main.go"}, Additions: 1}
	err := runAnalyzer(filepath.Join(t.TempDir(), "COMMIT_EDITMSG"), conf, info)
	if err == nil || !strings.Contains(err.Error(), "no CLI login path") {
		t.Fatalf("runAnalyzer() error = %v, want the missing CLI login path", err)
	}
}

type analyzerTestProvider struct{}

func (analyzerTestProvider) Name() string { return llmprovider.ProviderOpenAI }

func (analyzerTestProvider) Generate(context.Context, string) (string, error) {
	return "feat: use OAuth session", nil
}
