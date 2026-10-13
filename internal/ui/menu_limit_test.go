package ui

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/catalog"

	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

// longZenListing serves an OpenCode Zen listing of more models than
// catalog.MaxListed, all of them known to the SDK's catalog, and routes
// configure's listing to it. Every static catalog is cut to MaxListed, so only
// a live listing can show whether configure keeps the short list
// (0014-MADR D2; 0014-PLAN deviation 1).
func longZenListing(t *testing.T) *redirectTransport {
	t.Helper()
	isolate(t)
	t.Setenv("LLMPROVIDER_DISABLE_MODELS_METADATA", "1")
	ids := []string{
		"deepseek-v4.1-flash", "qwen3.8-flash", "glm-5.3-flash", "deepseek-v4-flash",
		"gemini-3.5-flash-lite", "gemini-3.8-flash", "claude-sonnet-5", "claude-opus-5",
	}
	if len(ids) <= catalog.MaxListed {
		t.Fatalf("the listing must exceed catalog.MaxListed (%d)", catalog.MaxListed)
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, fmt.Sprintf(`{"id":%q,"object":"model","owned_by":"opencode"}`, id))
	}
	body := `{"object":"list","data":[` + strings.Join(parts, ",") + `]}`
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
	return redirect
}

// TestDiscoverModels_ShortList: configure --yes recommends the SDK's short
// list from a longer live listing.
func TestDiscoverModels_ShortList(t *testing.T) {
	redirect := longZenListing(t)
	got, err := discoverModels(context.Background(), string(llmprovider.ProviderOpencodeZen), "test-key")
	if err != nil {
		t.Fatalf("discoverModels: %v", err)
	}
	if redirect.requests.Load() == 0 {
		t.Fatal("discoverModels did not list through the injected client")
	}
	if len(got) != catalog.MaxListed {
		t.Errorf("discoverModels recommends %d models %q; want catalog.MaxListed (%d)", len(got), got, catalog.MaxListed)
	}
}

// TestRunSetupInteractive_ShortList: the wizard's model menu, over a longer
// live listing, holds MaxListed models, so "Other" is the next choice.
func TestRunSetupInteractive_ShortList(t *testing.T) {
	redirect := longZenListing(t)
	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
	config.ApplyDefaults(conf)
	pos := 0
	for i, d := range registry.Descriptors() {
		if d.ID == llmprovider.ProviderOpencodeZen {
			pos = i + 1
		}
	}
	if pos == 0 {
		t.Fatal("the registry offers no OpenCode Zen")
	}
	other := strconv.Itoa(catalog.MaxListed + 1)
	// provider, endpoint (default), key, search blank, Other, the custom id,
	// fallback search blank, none, the four operational defaults
	input := strconv.Itoa(pos) + "\n\ntest-key\n\n" + other + "\nmy-custom-model\n\n\n\n\n\n\n"
	if err := runSetupInteractive(context.Background(), conf, SetupOptions{NoEnv: true}, strings.NewReader(input)); err != nil {
		t.Fatalf("runSetupInteractive: %v", err)
	}
	if redirect.requests.Load() == 0 {
		t.Fatal("the wizard did not list through the injected client")
	}
	if got := conf.Providers[string(llmprovider.ProviderOpencodeZen)].Model; got != "my-custom-model" {
		t.Errorf("model = %q; want my-custom-model, chosen as Other at %s", got, other)
	}
}
