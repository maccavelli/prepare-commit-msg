package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/auth"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/llmtest"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/providers"
	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

// TestGenerateText_ResponsesRefusal: a real OpenAI provider whose Responses
// answer is only a refusal part gives errRefused, not the refusal as text.
// From SDK v1.2.0 the refusal is kept as the answer's text with the
// content_filter finish (go-llmprovider-sdk 0021-MADR W4), so only D1's check
// keeps it out of the commit message (0009-MADR D1).
func TestGenerateText_ResponsesRefusal(t *testing.T) {
	const refusal = "I'm sorry, but I can't help with that request."
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_1","status":"completed","model":"gpt-4.1-mini","output":[` +
			`{"type":"message","role":"assistant","content":[{"type":"refusal","refusal":"` + refusal + `"}]}]}`))
	}))
	t.Cleanup(srv.Close)
	p, err := providers.New(llmprovider.ProviderOpenAI, llmprovider.WithTokenSource(llmprovider.NewStaticToken("test-key")),
		llmprovider.WithModel("gpt-4.1-mini"), llmprovider.WithBaseURL(srv.URL), llmprovider.WithoutModelMetadata())
	if err != nil {
		t.Fatal(err)
	}
	got, err := generateText(context.Background(), p, "write a commit message", 0, time.Millisecond)
	if !errors.Is(err, errRefused) || got != "" {
		t.Fatalf("generateText() = %q, %v; want errRefused and no text", got, err)
	}
	if !strings.Contains(err.Error(), refusal) {
		t.Fatalf("error %q does not quote the refusal", err)
	}
}

// TestGenerateText covers generateText over the SDK's scriptable fake: the
// prompt goes as one user message, a rate limit is retried, and an
// authentication failure is not. A refused answer is an error, never text
// (0009-MADR D1).
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
	answer := func(text string, finish llmprovider.FinishReason) *llmprovider.Response {
		return &llmprovider.Response{
			Output:       []llmprovider.Item{llmprovider.MessageItem{Role: llmprovider.RoleAssistant, Text: text}},
			FinishReason: finish,
		}
	}
	t.Run("refusal is an error", func(t *testing.T) {
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).
			Reply(answer("I'm sorry, but I can't help with that request.\nSecond line.", llmprovider.FinishContentFilter))
		got, err := generateText(context.Background(), fake, "p", 3, time.Millisecond)
		if !errors.Is(err, errRefused) || got != "" || len(fake.Requests()) != 1 {
			t.Fatalf("generateText() = %q, %v after %d requests; want errRefused after 1", got, err, len(fake.Requests()))
		}
		if !strings.Contains(err.Error(), "can't help with that request.") || strings.Contains(err.Error(), "Second line") {
			t.Fatalf("error %q should quote the refusal's first line only", err)
		}
	})
	t.Run("long refusal is cut", func(t *testing.T) {
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).
			Reply(answer(strings.Repeat("x", 500), llmprovider.FinishContentFilter))
		_, err := generateText(context.Background(), fake, "p", 0, time.Millisecond)
		if !errors.Is(err, errRefused) || strings.Count(err.Error(), "x") != refusalNoteRunes {
			t.Fatalf("error %q should quote %d runes of the refusal", err, refusalNoteRunes)
		}
	})
	t.Run("stop keeps its text", func(t *testing.T) {
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).
			Reply(answer("feat: add y", llmprovider.FinishStop))
		got, err := generateText(context.Background(), fake, "p", 0, time.Millisecond)
		if err != nil || got != "feat: add y" {
			t.Fatalf("generateText() = %q, %v", got, err)
		}
	})
	t.Run("no response is incomplete", func(t *testing.T) {
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).
			Handle(func(context.Context, *llmprovider.Request) (*llmprovider.Response, error) { return nil, nil })
		_, err := generateText(context.Background(), fake, "p", 0, time.Millisecond)
		if !errors.Is(err, llmprovider.ErrIncomplete) {
			t.Fatalf("generateText() error %v, want ErrIncomplete", err)
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
