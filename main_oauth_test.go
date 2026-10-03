package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/auth"
	"github.com/maccavelli/prepare-commit-msg/internal/config"
	"github.com/maccavelli/prepare-commit-msg/internal/git"
)

// isolateUserDirs points every per-user directory the tool or os.UserConfigDir
// can read at a temporary directory, so no test writes the live profile (P8,
// C8). It duplicates internal/config's isolateHome, which package main cannot
// import.
func isolateUserDirs(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	roaming := filepath.Join(tmp, "AppData", "Roaming")
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("APPDATA", roaming)
	t.Setenv("AppData", roaming)
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "AppData", "Local"))
	return tmp
}

// TestRedirectUserConfig_APPDATARequired: the helper sets APPDATA and AppData
// to the same temporary roaming directory.
func TestRedirectUserConfig_APPDATARequired(t *testing.T) {
	tmp := isolateUserDirs(t)
	appdata, appData := os.Getenv("APPDATA"), os.Getenv("AppData")
	if appdata == "" || appdata != appData || !strings.HasPrefix(appdata, tmp) {
		t.Fatalf("APPDATA %q, AppData %q, want the same path under %q", appdata, appData, tmp)
	}
}

// p8Fixture is this test's access token. It is access-only ChatGPT, which the
// SDK's validator accepts, and never a live token.
const p8Fixture = "p8-isolation-fixture-access"

func TestRunAnalyzer_OAuthDoesNotCallNewProviderWithAccessToken(t *testing.T) {
	liveDir, liveErr := os.UserConfigDir()
	isolateUserDirs(t)
	t.Cleanup(func() {
		if liveErr != nil {
			return
		}
		live, err := os.ReadFile(filepath.Join(liveDir, AppTitle, "oauth", "openai.json")) //nolint:gosec // a fixed path, read only
		if err == nil && strings.Contains(string(live), p8Fixture) {
			t.Errorf("the live OAuth session holds this test's fixture")
		}
	})
	store, err := config.NewOAuthStore()
	if err != nil {
		t.Fatalf("NewOAuthStore() error = %v", err)
	}
	session := &auth.OAuthSession{
		Provider: llmprovider.ProviderOpenAI,
		Access:   p8Fixture,
		Issuer:   auth.DefaultOpenAIIssuer,
		ClientID: auth.DefaultOpenAIClientID,
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
	newProvider = func(llmprovider.ProviderID, string, string, ...llmprovider.Option) (llmprovider.Provider, error) {
		t.Fatal("OAuth generation called NewProvider")
		return nil, nil
	}
	var capturedSource llmprovider.TokenSource
	newProviderWithSource = func(
		_ llmprovider.ProviderID,
		source llmprovider.TokenSource,
		_ string,
		_ ...llmprovider.Option,
	) (llmprovider.Provider, error) {
		capturedSource = source
		return analyzerTestProvider{}, nil
	}
	generateWithRetry = func(context.Context, llmprovider.Provider, string, int, time.Duration) (string, error) {
		return "feat: use OAuth session", nil
	}

	conf := &config.Config{
		ActiveProvider: string(llmprovider.ProviderOpenAI),
		Providers: map[string]config.ProviderConfig{
			string(llmprovider.ProviderOpenAI): {AuthKind: "oauth", Model: "gpt-5.4"},
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
	if _, ok := capturedSource.(*auth.OAuthSession); !ok {
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
	newProvider = func(llmprovider.ProviderID, string, string, ...llmprovider.Option) (llmprovider.Provider, error) {
		t.Fatal("vendor CLI generation called NewProvider")
		return nil, nil
	}
	var capturedSource llmprovider.TokenSource
	newProviderWithSource = func(
		_ llmprovider.ProviderID,
		source llmprovider.TokenSource,
		_ string,
		_ ...llmprovider.Option,
	) (llmprovider.Provider, error) {
		capturedSource = source
		return analyzerTestProvider{}, nil
	}
	generateWithRetry = func(context.Context, llmprovider.Provider, string, int, time.Duration) (string, error) {
		return "feat: use the Grok CLI login", nil
	}

	conf := &config.Config{
		ActiveProvider: string(llmprovider.ProviderGrok),
		Providers:      map[string]config.ProviderConfig{string(llmprovider.ProviderGrok): pc},
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
	session, ok := capturedSource.(*auth.VendorCLISession)
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
	newProvider = func(llmprovider.ProviderID, string, string, ...llmprovider.Option) (llmprovider.Provider, error) {
		t.Fatal("a vendor CLI login without a path built an API-key provider")
		return nil, nil
	}
	newProviderWithSource = func(llmprovider.ProviderID, llmprovider.TokenSource, string, ...llmprovider.Option) (llmprovider.Provider, error) {
		t.Fatal("a vendor CLI login without a path built a provider")
		return nil, nil
	}
	conf := &config.Config{
		ActiveProvider: string(llmprovider.ProviderGrok),
		Providers:      map[string]config.ProviderConfig{string(llmprovider.ProviderGrok): pc},
		TimeoutSeconds: 5,
	}
	info := &git.Info{Files: []string{"main.go"}, Additions: 1}
	err := runAnalyzer(filepath.Join(t.TempDir(), "COMMIT_EDITMSG"), conf, info)
	if err == nil || !strings.Contains(err.Error(), "no CLI login path") {
		t.Fatalf("runAnalyzer() error = %v, want the missing CLI login path", err)
	}
}

type analyzerTestProvider struct{}

func (analyzerTestProvider) ID() llmprovider.ProviderID { return llmprovider.ProviderOpenAI }

func (analyzerTestProvider) Capabilities() llmprovider.Capabilities {
	return llmprovider.Capabilities{}
}

func (analyzerTestProvider) Generate(context.Context, *llmprovider.Request) (*llmprovider.Response, error) {
	return &llmprovider.Response{Output: []llmprovider.Item{llmprovider.MessageItem{Role: llmprovider.RoleAssistant, Text: "feat: use OAuth session"}}}, nil
}
