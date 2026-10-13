package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/auth"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/catalog"
	"github.com/maccavelli/go-llmprovider-sdk/wizard"

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
	roaming := filepath.Join(tmp, "AppData", "Roaming")
	t.Setenv("APPDATA", roaming)
	t.Setenv("AppData", roaming)
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "AppData", "Local"))
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
	// catalog's short list, then Other
	// Other, then the custom model id
	// fallbacks: enter to search, enter for none
	// operational: all enter (defaults)
	other := strconv.Itoa(min(len(catalog.Static(llmprovider.ProviderGemini)), catalog.MaxListed) + 1)
	input := "1\ny\n\n" + other + "\nmy-custom-model\n\n\n\n\n\n\n"
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
	if _, err := discoverModels(context.Background(), providerGemini, "test-key", ""); err != nil {
		t.Fatalf("an offline listing is not a refused key: %v", err)
	}
	if offline.requests.Load() == 0 {
		t.Fatal("discoverModels listed models without the injected client")
	}
}

// redirectTransport sends every request to one test server, keeping its path.
type redirectTransport struct {
	to       *url.URL
	requests atomic.Int32
}

func (r *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.requests.Add(1)
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host = r.to.Scheme, r.to.Host
	return http.DefaultTransport.RoundTrip(req)
}

// kiloTier is one Kilo listing entry for a kilo-auto tier, in the shape of
// go-llmprovider-sdk's own catalog test fixture (kiloRankEntry).
func kiloTier(id string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"created":0,"context_length":262144,`+
		`"architecture":{"input_modalities":["text"],"output_modalities":["text"]},`+
		`"pricing":{"prompt":"-1","completion":"-1"},"supported_parameters":["tools","reasoning"],`+
		`"mayTrainOnYourPrompts":false}`, id, id)
}

// TestDiscoverModels_EmptyRecommendationFallsBack: a live listing that curates
// to nothing, the SDK's 0021-MADR C13 case of a Kilo listing of only kilo-auto
// tiers, gives the curated catalog, so configure --yes still has a model
// (0009-MADR D3).
func TestDiscoverModels_EmptyRecommendationFallsBack(t *testing.T) {
	isolate(t)
	t.Setenv("LLMPROVIDER_DISABLE_MODELS_METADATA", "1")
	body := `{"data":[` + kiloTier("kilo-auto/small") + "," + kiloTier("kilo-auto/balanced") + `]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	to, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	redirect := &redirectTransport{to: to}
	listingClient = &http.Client{Transport: redirect}

	got, err := discoverModels(context.Background(), string(llmprovider.ProviderKilo), "test-key", "")
	if err != nil {
		t.Fatalf("discoverModels(kilo): %v", err)
	}
	if redirect.requests.Load() == 0 {
		t.Fatal("discoverModels did not list through the injected client")
	}
	if want := catalog.Static(llmprovider.ProviderKilo); len(want) == 0 || !slices.Equal(got, want) {
		t.Fatalf("discoverModels(kilo) = %q, want the curated catalog %q", got, want)
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
		if err := store.Save(context.Background(), llmprovider.ProviderOpenAI, &auth.OAuthSession{
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
		pc := conf.Providers[string(llmprovider.ProviderOpenAI)]
		if pc.AuthKind != "" || pc.APIKey != "test-key" {
			t.Errorf("OpenAI auth kind/key = %q/%q", pc.AuthKind, pc.APIKey)
		}
		oauthDir, err := config.OAuthDir()
		if err != nil {
			t.Fatalf("OAuthDir() error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(oauthDir, string(llmprovider.ProviderOpenAI)+".json")); !os.IsNotExist(err) {
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
	stale := &auth.OAuthSession{Provider: llmprovider.ProviderGrok, Access: "copied-by-an-older-release"}
	if err := store.Save(context.Background(), llmprovider.ProviderGrok, stale); err != nil {
		t.Fatalf("save stale session: %v", err)
	}

	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)
	input := "4\n5\ny\n\n1\n\n\n\n\n\n\n"
	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, strings.NewReader(input)); err != nil {
		t.Fatalf("runSetupInteractive() error = %v", err)
	}
	pc := conf.Providers[string(llmprovider.ProviderGrok)]
	if pc.AuthKind != "vendor_cli" || pc.APIKey != "" {
		t.Fatalf("Grok auth kind/key = %q/%q, want vendor_cli and no key", pc.AuthKind, pc.APIKey)
	}
	assertSavedVendorPath(t, string(llmprovider.ProviderGrok), authPath)
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
	pc := conf.Providers[string(llmprovider.ProviderOpenAI)]
	if pc.AuthKind != "vendor_cli" || pc.APIKey != "" {
		t.Fatalf("OpenAI auth kind/key = %q/%q, want vendor_cli and no key", pc.AuthKind, pc.APIKey)
	}
	assertSavedVendorPath(t, string(llmprovider.ProviderOpenAI), authPath)
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
	openAIPC := conf.Providers[string(llmprovider.ProviderOpenAI)]
	openAIPC.AuthKind = "oauth"
	conf.Providers[string(llmprovider.ProviderOpenAI)] = openAIPC

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
		{name: "end of input keeps current", input: "", current: 10, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &wizard.TextPrompter{In: strings.NewReader(tt.input), Out: io.Discard}
			got, err := promptInt(p, "value", tt.current)
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
// hard-coded list of three. The menu now comes from the SDK registry's
// Descriptors(), and this test fails the build if that ever stops being true —
// so a provider added to go-llmprovider-sdk cannot silently go unreachable here
// again.
func TestSetup_OffersEveryDescriptor(t *testing.T) {
	descriptors := registry.Descriptors()
	if len(descriptors) == 0 {
		t.Fatal("registry.Descriptors() is empty")
	}

	// Every descriptor must be a provider this app can validate and configure.
	for _, d := range descriptors {
		if _, ok := registry.Descriptor(d.ID); !ok {
			t.Errorf("descriptor %q does not resolve", d.ID)
		}
	}

	// The providers this wizard once hard-coded must still be present, and the
	// set must now be strictly larger than that original three.
	for _, id := range []llmprovider.ProviderID{
		llmprovider.ProviderGemini, llmprovider.ProviderOpenAI, llmprovider.ProviderClaude,
		llmprovider.ProviderGrok, // the one that was missing for a full release
	} {
		if _, ok := registry.Descriptor(id); !ok {
			t.Errorf("provider %q must be offerable", id)
		}
	}
	if len(descriptors) <= 3 {
		t.Errorf("expected more than the 3 originally hard-coded providers, got %d", len(descriptors))
	}
}

// TestRunSetupInteractive_OperationalTyped: operational values typed after
// the wizard's answers are saved. A second reader on the same input found the
// wizard's buffer had drained it, so piped values were lost (0014-MADR D3).
func TestRunSetupInteractive_OperationalTyped(t *testing.T) {
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
	// 1: gemini; y: the env key; search blank; model 1; fallback search
	// blank; none; then the four operational values
	input := "1\ny\n\n1\n\n\n99\n12345\n7\n9\n"
	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, strings.NewReader(input)); err != nil {
		t.Fatalf("runSetupInteractive: %v", err)
	}
	got := []int{conf.TimeoutSeconds, conf.MaxDiffBytes, conf.RetryCount, conf.RetryDelaySeconds}
	if want := []int{99, 12345, 7, 9}; !slices.Equal(got, want) {
		t.Errorf("operational values = %v; want the typed %v", got, want)
	}
}

// geminiKeyInvalid is Gemini's refusal of an invalid key, the body of
// go-llmprovider-sdk's own fixture (geminiInteractionsKeyInvalid) without the
// array the Interactions API wraps it in.
const geminiKeyInvalid = `{"error":{"code":400,"message":"API key not valid. Please pass a valid API key.",` +
	`"status":"INVALID_ARGUMENT","details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo",` +
	`"reason":"API_KEY_INVALID","domain":"googleapis.com"}]}}`

// TestRunSetupNonInteractive_RefusedKey: configure --yes fails, and saves
// nothing, when the listing refuses the key; a 403, which refuses a
// permission rather than the key, still saves with the static catalog
// (0014-MADR D4).
func TestRunSetupNonInteractive_RefusedKey(t *testing.T) {
	for _, c := range []struct {
		name    string
		status  int
		body    string
		refused bool
	}{
		{"invalid key", http.StatusBadRequest, geminiKeyInvalid, true},
		{"not permitted", http.StatusForbidden, `{"error":{"code":403,"message":"denied","status":"PERMISSION_DENIED"}}`, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			isolate(t)
			t.Setenv("LLMPROVIDER_DISABLE_MODELS_METADATA", "1")
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(c.status)
				_, _ = w.Write([]byte(c.body))
			}))
			t.Cleanup(srv.Close)
			to, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			redirect := &redirectTransport{to: to}
			listingClient = &http.Client{Transport: redirect}

			conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
			config.ApplyDefaults(conf)
			err = runSetupNonInteractive(context.Background(), conf, SetupOptions{
				Provider: providerGemini, APIKey: "bad-key", NoEnv: true, Yes: true,
			})
			if redirect.requests.Load() == 0 {
				t.Fatal("configure --yes did not list through the injected client")
			}
			path, pathErr := config.GetConfigPath()
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			_, statErr := os.Stat(path)
			if c.refused {
				if err == nil || !errors.Is(err, llmprovider.ErrAuthFailure) || !strings.Contains(err.Error(), "gemini refused the API key") {
					t.Fatalf("err = %v; want gemini's refusal, ErrAuthFailure", err)
				}
				if !errors.Is(statErr, os.ErrNotExist) {
					t.Errorf("a refused key was saved to %s (stat: %v)", path, statErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("a 403 is not a refused key: %v", err)
			}
			if statErr != nil {
				t.Fatalf("config not saved: %v", statErr)
			}
			if want := catalog.Static(llmprovider.ProviderGemini)[0]; conf.Providers[providerGemini].Model != want {
				t.Errorf("model = %q; want the static catalog's %q", conf.Providers[providerGemini].Model, want)
			}
		})
	}
}

// hostLog records the Host each fixture request named.
type hostLog struct {
	mu    sync.Mutex
	hosts []string
}

func (h *hostLog) add(host string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hosts = append(h.hosts, host)
}

func (h *hostLog) all() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.hosts)
}

// ollamaServer answers Ollama's reachability check and model listing, and
// redirects configure's requests to itself, whatever endpoint is typed. The
// log holds the host each request was addressed to.
func ollamaServer(t *testing.T) (*redirectTransport, *hostLog) {
	t.Helper()
	isolate(t)
	t.Setenv("LLMPROVIDER_DISABLE_MODELS_METADATA", "1")
	log := &hostLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.add(r.Host)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/version":
			_, _ = w.Write([]byte(`{"version":"0.12.0"}`))
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models":[{"name":"llama3.3"},{"name":"qwen3"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	to, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	redirect := &redirectTransport{to: to}
	listingClient = &http.Client{Transport: redirect}
	return redirect, log
}

// TestRunSetupInteractive_SavesBaseURL: a typed endpoint that is not the
// provider's default is saved as base_url; accepting the default saves none;
// a saved one is offered again and kept (0014-MADR D6); --base-url, without
// --yes, is the offered default (0014-MADR Amendment 2, B2).
func TestRunSetupInteractive_SavesBaseURL(t *testing.T) {
	ollama := llmprovider.ProviderOllama
	d, ok := registry.Descriptor(ollama)
	if !ok || d.DefaultBaseURL == "" {
		t.Fatal("the registry offers no Ollama endpoint default")
	}
	pos := slices.IndexFunc(registry.Descriptors(), func(x llmprovider.Descriptor) bool { return x.ID == ollama }) + 1
	for _, tc := range []struct {
		name, saved, flag, typed, want string
	}{
		{"typed endpoint", "", "", "http://ollama.test:11434", "http://ollama.test:11434"},
		{"default endpoint", "", "", "", ""},
		{"saved endpoint kept", "http://ollama.test:11434", "", "", "http://ollama.test:11434"},
		{"flag seeds the endpoint", "", "http://ollama.test:11434", "", "http://ollama.test:11434"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			redirect, _ := ollamaServer(t)
			conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
			config.ApplyDefaults(conf)
			if tc.saved != "" {
				pc := conf.Providers[string(ollama)]
				pc.BaseURL = tc.saved
				conf.Providers[string(ollama)] = pc
				conf.ActiveProvider = string(ollama)
			}
			// provider, endpoint, search blank, the first model, fallback
			// search blank, none, the four operational defaults
			input := strconv.Itoa(pos) + "\n" + tc.typed + "\n\n1\n\n\n\n\n\n\n"
			opts := SetupOptions{NoEnv: true}
			if tc.flag != "" {
				opts.Provider, opts.BaseURL = string(ollama), tc.flag
			}
			if err := runSetupInteractive(context.Background(), conf, opts, strings.NewReader(input)); err != nil {
				t.Fatalf("runSetupInteractive: %v", err)
			}
			if redirect.requests.Load() == 0 {
				t.Fatal("configure never reached the Ollama fixture")
			}
			if conf.ActiveProvider != string(ollama) {
				t.Fatalf("active provider = %q; want ollama", conf.ActiveProvider)
			}
			if got := conf.Providers[string(ollama)].BaseURL; got != tc.want {
				t.Errorf("base_url = %q; want %q", got, tc.want)
			}
		})
	}
}

// assertNotSaved fails when configure wrote a config file.
func assertNotSaved(t *testing.T) {
	t.Helper()
	path, err := config.GetConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("config was saved to %s (stat: %v)", path, err)
	}
}

// TestRunSetupNonInteractive_KeylessProvider: README's Example 2, Ollama with
// no key, is saved; Gemini with no key is still refused (0014-MADR
// Amendment 2, B1).
func TestRunSetupNonInteractive_KeylessProvider(t *testing.T) {
	isolate(t)
	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)
	ollama := string(llmprovider.ProviderOllama)
	if err := runSetupNonInteractive(context.Background(), conf, SetupOptions{
		Provider: ollama, Model: "qwen2.5-coder:7b", NoEnv: true, Yes: true,
	}); err != nil {
		t.Fatalf("Ollama with no key: %v; want it saved", err)
	}
	loaded, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if pc := loaded.Providers[ollama]; loaded.ActiveProvider != ollama || pc.Model != "qwen2.5-coder:7b" || pc.APIKey != "" {
		t.Errorf("saved %q (%q, key set %t); want ollama (qwen2.5-coder:7b), no key", loaded.ActiveProvider, pc.Model, pc.APIKey != "")
	}

	isolate(t)
	conf = &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)
	err = runSetupNonInteractive(context.Background(), conf, SetupOptions{
		Provider: providerGemini, Model: "m", NoEnv: true, Yes: true,
	})
	if err == nil || !strings.Contains(err.Error(), "API key required") {
		t.Errorf("Gemini with no key: %v; want API key required", err)
	}
	assertNotSaved(t)
}

// TestRunSetupNonInteractive_BaseURL: --base-url is where configure --yes
// lists, and is saved unless it is the default; with no flag a saved
// endpoint is listed and kept (0014-MADR Amendment 2, B2, B3).
func TestRunSetupNonInteractive_BaseURL(t *testing.T) {
	ollama := llmprovider.ProviderOllama
	d, ok := registry.Descriptor(ollama)
	if !ok || d.DefaultBaseURL == "" {
		t.Fatal("the registry offers no Ollama endpoint default")
	}
	defaultHost := strings.TrimPrefix(d.DefaultBaseURL, "http://")
	for _, tc := range []struct {
		name, saved, flag, host, want string
	}{
		{"flag", "", "http://ollama.test:11434", "ollama.test:11434", "http://ollama.test:11434"},
		{"default flag", "", d.DefaultBaseURL, defaultHost, ""},
		{"saved endpoint kept", "http://ollama.test:11434", "", "ollama.test:11434", "http://ollama.test:11434"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, hosts := ollamaServer(t)
			conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
			config.ApplyDefaults(conf)
			if tc.saved != "" {
				pc := conf.Providers[string(ollama)]
				pc.BaseURL = tc.saved
				conf.Providers[string(ollama)] = pc
			}
			if err := runSetupNonInteractive(context.Background(), conf, SetupOptions{
				Provider: string(ollama), BaseURL: tc.flag, NoEnv: true, Yes: true,
			}); err != nil {
				t.Fatalf("runSetupNonInteractive: %v", err)
			}
			got := hosts.all()
			if len(got) == 0 || slices.ContainsFunc(got, func(h string) bool { return h != tc.host }) {
				t.Errorf("listing hosts = %q; want only %q", got, tc.host)
			}
			pc := conf.Providers[string(ollama)]
			if pc.Model != "llama3.3" {
				t.Errorf("model = %q; want the listing's first, llama3.3", pc.Model)
			}
			if pc.BaseURL != tc.want {
				t.Errorf("base_url = %q; want %q", pc.BaseURL, tc.want)
			}
		})
	}
}

// TestRunSetupNonInteractive_BaseURLRefused: each --base-url the wizard's rule
// refuses, and plain http to a remote host for a provider that sends a key,
// fails with nothing saved and no password quoted; plain http to loopback is
// allowed (0014-MADR Amendment 2, B2).
func TestRunSetupNonInteractive_BaseURLRefused(t *testing.T) {
	together := string(llmprovider.ProviderTogether)
	for _, tc := range []struct {
		name, provider, baseURL, want string
	}{
		{"provider without an endpoint", providerGemini, "https://example.test/v1", "takes no --base-url"},
		{"user name and password", together, "https://user:hunter2@example.test/v1", "no user name or password"},
		{"query", together, "https://example.test/v1?x=1", "query or a fragment"},
		{"fragment", together, "https://example.test/v1#x", "query or a fragment"},
		{"ftp scheme", together, "ftp://example.test/v1", "needs http:// or https://"},
		{"plain http to a remote host", together, "http://example.test/v1", "plain http to a remote host"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			isolate(t)
			conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
			config.ApplyDefaults(conf)
			err := runSetupNonInteractive(context.Background(), conf, SetupOptions{
				Provider: tc.provider, APIKey: "test-key", Model: "m", Fallbacks: []string{"f"}, BaseURL: tc.baseURL, Yes: true,
			})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v; want %q", err, tc.want)
			}
			if strings.Contains(err.Error(), "hunter2") {
				t.Errorf("err quotes the password: %v", err)
			}
			assertNotSaved(t)
		})
	}

	isolate(t)
	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)
	if err := runSetupNonInteractive(context.Background(), conf, SetupOptions{
		Provider: together, APIKey: "test-key", Model: "m", Fallbacks: []string{"f"}, BaseURL: "http://127.0.0.1:8080/v1", Yes: true,
	}); err != nil {
		t.Fatalf("plain http to loopback: %v; want it allowed", err)
	}
	if got := conf.Providers[together].BaseURL; got != "http://127.0.0.1:8080/v1" {
		t.Errorf("base_url = %q; want the loopback endpoint", got)
	}
}
