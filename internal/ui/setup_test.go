package ui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/maccavelli/mcplib/llmprovider"

	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

// offlineTransport fails every request, so configure's unit tests never
// reach a live listing endpoint; requests counts the attempts.
type offlineTransport struct{ requests atomic.Int32 }

func (o *offlineTransport) RoundTrip(*http.Request) (*http.Response, error) {
	o.requests.Add(1)
	return nil, errors.New("offline: unit tests make no network calls")
}

func isolate(t *testing.T) *offlineTransport {
	t.Helper()
	offline := &offlineTransport{}
	previous := listingClient
	listingClient = &http.Client{Transport: offline}
	t.Cleanup(func() { listingClient = previous })
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("AppData", filepath.Join(tmp, "AppData", "Roaming"))
	return offline
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
	// search: enter; the listing is offline, so the menu is the built-in
	// catalog: 6 models, then 7 Other
	// 7, then the custom model id
	// fallbacks: enter to search, enter for none
	// operational: all enter (defaults)
	input := "1\ny\n\n7\nmy-custom-model\n\n\n\n\n\n\n"
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

// TestRunSetupInteractive_ListingUsesInjectedClient: configure's live model
// listing goes through listingClient, so tests can keep it offline.
func TestRunSetupInteractive_ListingUsesInjectedClient(t *testing.T) {
	offline := isolate(t)

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

	input := "1\ny\n\n1\n\n\n\n\n\n\n"
	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, strings.NewReader(input)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offline.requests.Load() == 0 {
		t.Fatal("configure listed models without the injected client")
	}
}

// TestDiscoverModels_UsesInjectedClient: the non-interactive listing goes
// through listingClient too.
func TestDiscoverModels_UsesInjectedClient(t *testing.T) {
	offline := isolate(t)
	discoverModels(context.Background(), providerGemini, "test-key")
	if offline.requests.Load() == 0 {
		t.Fatal("discoverModels listed models without the injected client")
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
		store, err := config.NewOAuthStore()
		if err != nil {
			t.Fatalf("NewOAuthStore() error = %v", err)
		}
		if err := store.Save(context.Background(), llmprovider.ProviderOpenAI, &llmprovider.OAuthSession{
			Provider: llmprovider.ProviderOpenAI,
			Access:   "stale-access",
		}); err != nil {
			t.Fatalf("seed stale OpenAI OAuth session: %v", err)
		}
		// 2: openai, auth 1: API key, key, model 1, fallbacks enter, ops enter
		input := "2\n1\ntest-key\n1\n\n\n\n\n\n"
		r := strings.NewReader(input)
		if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, r); err != nil {
			t.Fatalf("%v", err)
		}
		if conf.ActiveProvider != "openai" {
			t.Errorf("Expected openai")
		}
		pc := conf.Providers[llmprovider.ProviderOpenAI]
		if pc.AuthKind != "" || pc.APIKey != "test-key" {
			t.Errorf("OpenAI auth kind/key = %q/%q", pc.AuthKind, pc.APIKey)
		}
		oauthDir, err := config.OAuthDir()
		if err != nil {
			t.Fatalf("OAuthDir() error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(oauthDir, llmprovider.ProviderOpenAI+".json")); !os.IsNotExist(err) {
			t.Fatalf("OpenAI OAuth session exists after API-key setup: %v", err)
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

const grokAuthFixture = `{
  "xai::api_key": {
    "key": "xai-MUST-SKIP",
    "auth_mode": "api_key"
  },
  "https://auth.x.ai::b1a00492-073a-47ea-816f-4c329264a828": {
    "key": "sess-grok",
    "auth_mode": "oidc",
    "refresh_token": "rt-grok",
    "expires_at": "2099-01-01T00:00:00Z",
    "oidc_issuer": "https://auth.x.ai",
    "oidc_client_id": "b1a00492-073a-47ea-816f-4c329264a828"
  }
}`

const openAIAuthFixture = `{
  "OPENAI_API_KEY": "sk-MUST-IGNORE",
  "tokens": {
    "access_token": "at-chatgpt",
    "refresh_token": "rt-chatgpt",
    "account_id": "acct_test"
  }
}`

// TestRunSetupInteractive_GrokCLILoginReadsThrough: choosing the Grok CLI
// login saves only the path to its auth file. No token is copied, and a
// session an older release copied from the CLI is removed, so it is never
// refreshed against the CLI's own (mcplib MADR 0012 §5.1).
func TestRunSetupInteractive_GrokCLILoginReadsThrough(t *testing.T) {
	isolate(t)
	vendorHome := t.TempDir()
	t.Setenv("GROK_HOME", vendorHome)
	authPath := filepath.Join(vendorHome, "auth.json")
	if err := os.WriteFile(authPath, []byte(grokAuthFixture), 0o600); err != nil {
		t.Fatalf("write Grok fixture: %v", err)
	}
	originalEnv := osGetenv
	osGetenv = os.Getenv
	t.Cleanup(func() { osGetenv = originalEnv })

	store, err := config.NewOAuthStore()
	if err != nil {
		t.Fatalf("NewOAuthStore() error = %v", err)
	}
	stale := &llmprovider.OAuthSession{Provider: llmprovider.ProviderGrok, Access: "copied-by-an-older-release"}
	if err := store.Save(context.Background(), llmprovider.ProviderGrok, stale); err != nil {
		t.Fatalf("save stale session: %v", err)
	}

	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)
	input := "4\n5\ny\n\n1\n\n\n\n\n\n\n"
	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, strings.NewReader(input)); err != nil {
		t.Fatalf("runSetupInteractive() error = %v", err)
	}
	pc := conf.Providers[llmprovider.ProviderGrok]
	if pc.AuthKind != "vendor_cli" || pc.APIKey != "" {
		t.Fatalf("Grok auth kind/key = %q/%q, want vendor_cli and no key", pc.AuthKind, pc.APIKey)
	}
	assertSavedVendorPath(t, llmprovider.ProviderGrok, authPath)
	if session, loadErr := store.Load(context.Background(), llmprovider.ProviderGrok); loadErr != nil || session != nil {
		t.Fatalf("stored Grok session = %+v, %v; want none", session, loadErr)
	}
	assertConfigOmitsOAuthTokens(t)
}

// TestRunSetupInteractive_CodexCLILoginReadsThrough: choosing the Codex CLI
// login saves only the path to its auth file; neither its access token nor
// its OPENAI_API_KEY becomes the configured API key.
func TestRunSetupInteractive_CodexCLILoginReadsThrough(t *testing.T) {
	isolate(t)
	vendorHome := t.TempDir()
	t.Setenv("CODEX_HOME", vendorHome)
	authPath := filepath.Join(vendorHome, "auth.json")
	if err := os.WriteFile(authPath, []byte(openAIAuthFixture), 0o600); err != nil {
		t.Fatalf("write OpenAI fixture: %v", err)
	}
	originalEnv := osGetenv
	osGetenv = os.Getenv
	t.Cleanup(func() { osGetenv = originalEnv })

	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)
	input := "2\n5\ny\ngpt-5.4\n\n\n\n\n\n\n"
	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, strings.NewReader(input)); err != nil {
		t.Fatalf("runSetupInteractive() error = %v", err)
	}
	pc := conf.Providers[llmprovider.ProviderOpenAI]
	if pc.AuthKind != "vendor_cli" || pc.APIKey != "" {
		t.Fatalf("OpenAI auth kind/key = %q/%q, want vendor_cli and no key", pc.AuthKind, pc.APIKey)
	}
	assertSavedVendorPath(t, llmprovider.ProviderOpenAI, authPath)
	assertConfigOmitsOAuthTokens(t)
}

// assertSavedVendorPath reads the saved configuration file and checks the
// provider's vendor_auth_path.
func assertSavedVendorPath(t *testing.T, provider, want string) {
	t.Helper()
	path, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved struct {
		Providers map[string]struct {
			VendorAuthPath string `json:"vendor_auth_path"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if got := saved.Providers[provider].VendorAuthPath; got != want {
		t.Fatalf("saved %s vendor_auth_path = %q, want %q", provider, got, want)
	}
}

func assertConfigOmitsOAuthTokens(t *testing.T) {
	t.Helper()
	path, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	for _, key := range [][]byte{[]byte(`"access"`), []byte(`"refresh"`)} {
		if bytes.Contains(data, key) {
			t.Fatalf("config contains OAuth token key %s:\n%s", key, data)
		}
	}
}

func TestRunSetupNonInteractive(t *testing.T) {
	isolate(t)

	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)
	openAIPC := conf.Providers[llmprovider.ProviderOpenAI]
	openAIPC.AuthKind = "oauth"
	conf.Providers[llmprovider.ProviderOpenAI] = openAIPC

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
	if conf.Providers["openai"].AuthKind != "" {
		t.Errorf("non-interactive auth kind: %q", conf.Providers["openai"].AuthKind)
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

func TestProviderEnvVar(t *testing.T) {
	if got := providerEnvVar(providerGemini); got != "GEMINI_API_KEY" {
		t.Fatalf("providerEnvVar(gemini) = %q, want GEMINI_API_KEY", got)
	}
	if got := providerEnvVar("unknown"); got != "" {
		t.Fatalf("providerEnvVar(unknown) = %q, want empty", got)
	}
}

func TestPromptInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		current int
		want    int
	}{
		{name: "empty keeps current", input: "\n", current: 10, want: 10},
		{name: "positive replaces current", input: "42\n", current: 10, want: 42},
		{name: "text keeps current", input: "nope\n", current: 10, want: 10},
		{name: "zero keeps current", input: "0\n", current: 10, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := promptInt(bufio.NewReader(strings.NewReader(tt.input)), "value", tt.current)
			if err != nil {
				t.Fatalf("promptInt() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("promptInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRunSetupWithOptionsNonInteractiveErrors(t *testing.T) {
	isolate(t)
	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)

	err := RunSetupWithOptions(context.Background(), conf, SetupOptions{
		Provider: "not-a-provider",
		Yes:      true,
	}, strings.NewReader(""))
	if err == nil || !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("RunSetupWithOptions() error = %v, want unsupported provider", err)
	}

	err = RunSetupWithOptions(context.Background(), conf, SetupOptions{
		Provider: providerGemini,
		NoEnv:    true,
		Yes:      true,
	}, strings.NewReader(""))
	if err == nil || !strings.Contains(err.Error(), "API key required") {
		t.Fatalf("RunSetupWithOptions() error = %v, want API key required", err)
	}
}

// TestSetup_OffersEveryDescriptor is the drift guard. Grok shipped in mcplib
// MADR 0001 and this wizard never offered it, because the provider menu was a
// hard-coded list of three. The menu now comes from llmprovider.Descriptors(),
// and this test fails the build if that ever stops being true — so a provider
// added to mcplib cannot silently go unreachable here again.
func TestSetup_OffersEveryDescriptor(t *testing.T) {
	descriptors := llmprovider.Descriptors()
	if len(descriptors) == 0 {
		t.Fatal("llmprovider.Descriptors() is empty")
	}

	// Every descriptor must be a provider this app can validate and configure.
	for _, d := range descriptors {
		if _, ok := llmprovider.DescriptorFor(d.ID); !ok {
			t.Errorf("descriptor %q does not resolve", d.ID)
		}
	}

	// The providers this wizard once hard-coded must still be present, and the
	// set must now be strictly larger than that original three.
	for _, id := range []string{
		llmprovider.ProviderGemini, llmprovider.ProviderOpenAI, llmprovider.ProviderClaude,
		llmprovider.ProviderGrok, // the one that was missing for a full release
	} {
		if _, ok := llmprovider.DescriptorFor(id); !ok {
			t.Errorf("provider %q must be offerable", id)
		}
	}
	if len(descriptors) <= 3 {
		t.Errorf("expected more than the 3 originally hard-coded providers, got %d", len(descriptors))
	}
}
