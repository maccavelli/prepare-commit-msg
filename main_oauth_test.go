package main

import (
	"context"
	"os"
	"path/filepath"
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

type analyzerTestProvider struct{}

func (analyzerTestProvider) Name() string { return llmprovider.ProviderOpenAI }

func (analyzerTestProvider) Generate(context.Context, string) (string, error) {
	return "feat: use OAuth session", nil
}
