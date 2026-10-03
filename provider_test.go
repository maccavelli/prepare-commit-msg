package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/auth"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/llmtest"
	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

// TestGenerateText covers generateText over the SDK's scriptable fake: the
// prompt goes as one user message, a rate limit is retried, and an
// authentication failure is not.
func TestGenerateText(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).ReplyText("feat: add x")
		got, err := generateText(context.Background(), fake, "the prompt", 2, time.Millisecond)
		if err != nil || got != "feat: add x" {
			t.Fatalf("generateText() = %q, %v", got, err)
		}
		reqs := fake.Requests()
		if len(reqs) != 1 || len(reqs[0].Input) != 1 {
			t.Fatalf("requests = %+v, want one with one item", reqs)
		}
		msg, ok := reqs[0].Input[0].(llmprovider.MessageItem)
		if !ok || msg.Role != llmprovider.RoleUser || msg.Text != "the prompt" {
			t.Fatalf("input = %#v, want one user message holding the prompt", reqs[0].Input[0])
		}
	})
	t.Run("rate limit retried", func(t *testing.T) {
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).
			Fail(llmprovider.ErrRateLimited).ReplyText("feat: after retry")
		got, err := generateText(context.Background(), fake, "p", 1, time.Millisecond)
		if err != nil || got != "feat: after retry" || len(fake.Requests()) != 2 {
			t.Fatalf("generateText() = %q, %v after %d requests", got, err, len(fake.Requests()))
		}
	})
	t.Run("auth failure not retried", func(t *testing.T) {
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).
			Fail(llmprovider.ErrAuthFailure).ReplyText("never")
		_, err := generateText(context.Background(), fake, "p", 3, time.Millisecond)
		if !errors.Is(err, llmprovider.ErrAuthFailure) || len(fake.Requests()) != 1 {
			t.Fatalf("generateText() error %v after %d requests, want ErrAuthFailure after 1", err, len(fake.Requests()))
		}
	})
}

// TestNewActiveProvider_KiloOrganization: a saved Kilo organization is passed
// to the provider, and nothing is added without one (MADR 0008 D3).
func TestNewActiveProvider_KiloOrganization(t *testing.T) {
	isolateUserDirs(t)
	store, err := config.NewOAuthStore()
	if err != nil {
		t.Fatal(err)
	}
	session := &auth.OAuthSession{Provider: llmprovider.ProviderKilo, Access: "kilo-fixture-access"}
	if err := store.Save(context.Background(), llmprovider.ProviderKilo, session); err != nil {
		t.Fatal(err)
	}
	original := newProviderWithSource
	t.Cleanup(func() { newProviderWithSource = original })
	var gotOpts int
	newProviderWithSource = func(_ llmprovider.ProviderID, _ llmprovider.TokenSource, _ string, opts ...llmprovider.Option) (llmprovider.Provider, error) {
		gotOpts = len(opts)
		return analyzerTestProvider{}, nil
	}
	for _, tc := range []struct {
		org  string
		want int
	}{{"org-1", 1}, {"", 0}} {
		pc := config.ProviderConfig{AuthKind: "oauth", Model: "m", Organization: tc.org}
		conf := &config.Config{ActiveProvider: string(llmprovider.ProviderKilo), Providers: map[string]config.ProviderConfig{string(llmprovider.ProviderKilo): pc}}
		if _, err := newActiveProvider(conf, pc, "m"); err != nil {
			t.Fatalf("organization %q: %v", tc.org, err)
		}
		if gotOpts != tc.want {
			t.Fatalf("organization %q: %d options, want %d", tc.org, gotOpts, tc.want)
		}
	}
	// Another provider never gets the Kilo option.
	if opts := providerOptions(llmprovider.ProviderOpenAI, config.ProviderConfig{Organization: "org-1"}); len(opts) != 0 {
		t.Fatalf("OpenAI got %d options", len(opts))
	}
}
