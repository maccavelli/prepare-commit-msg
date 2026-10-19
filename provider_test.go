package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/auth"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/llmtest"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/providers"
	"github.com/maccavelli/prepare-commit-msg/internal/config"
	"github.com/maccavelli/prepare-commit-msg/internal/git"
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
		return llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).ReplyText("feat: use OAuth session"), nil
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

// TestNewActiveProvider_BaseURL: a saved base_url is where the hook sends its
// request; Ollama also needs no key to get there (0014-MADR D6).
func TestNewActiveProvider_BaseURL(t *testing.T) {
	isolateUserDirs(t)
	t.Setenv("LLMPROVIDER_DISABLE_MODELS_METADATA", "1")
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"c1","object":"chat.completion","model":"llama3.3","choices":[` +
			`{"index":0,"message":{"role":"assistant","content":"feat: reach the saved endpoint"},"finish_reason":"stop"}]}`))
	}))
	t.Cleanup(srv.Close)

	ollama := string(llmprovider.ProviderOllama)
	conf := &config.Config{
		ActiveProvider: ollama,
		Providers:      map[string]config.ProviderConfig{ollama: {Model: "llama3.3", BaseURL: srv.URL}},
		TimeoutSeconds: 5,
	}
	config.ApplyDefaults(conf)
	msgPath := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(msgPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	info := &git.Info{Files: []string{"app.go"}, Additions: 1, Deletions: 1}
	if err := runAnalyzer(msgPath, conf, info); err != nil {
		t.Fatalf("runAnalyzer: %v", err)
	}
	if hits.Load() == 0 {
		t.Fatal("the saved base_url received no request")
	}
	got, err := os.ReadFile(msgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(got), "feat: reach the saved endpoint\n") {
		t.Errorf("commit message = %q; want the endpoint's answer", got)
	}
}

// TestProvider_UserAgentNamesApp: the hook's requests name this program and
// its build in User-Agent, not the SDK (0014-MADR D5).
func TestProvider_UserAgentNamesApp(t *testing.T) {
	t.Setenv("LLMPROVIDER_DISABLE_MODELS_METADATA", "1")
	var agent atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent.Store(r.UserAgent())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"c1","object":"chat.completion","model":"llama3.3","choices":[` +
			`{"index":0,"message":{"role":"assistant","content":"feat: name the app"},"finish_reason":"stop"}]}`))
	}))
	t.Cleanup(srv.Close)
	p, err := newSourceProvider(llmprovider.ProviderOllama, llmprovider.NewStaticToken(""), "llama3.3", llmprovider.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := generateText(context.Background(), p, "prompt", 0, 0); err != nil {
		t.Fatalf("generateText: %v", err)
	}
	got, _ := agent.Load().(string)
	if want := AppTitle + "/" + buildIdentity().Current() + " ("; !strings.HasPrefix(got, want) {
		t.Errorf("User-Agent = %q; want it to begin %q", got, want)
	}
}

// captureRetries points retryOut at the returned buffer for t.
func captureRetries(t *testing.T) *bytes.Buffer {
	t.Helper()
	var b bytes.Buffer
	previous := retryOut
	retryOut = &b
	t.Cleanup(func() { retryOut = previous })
	return &b
}

// retryLine is the line reportRetry writes (0015-MADR D3).
var retryLine = regexp.MustCompile(`^prepare-commit-msg: openai: retrying in [0-9.]+m?s \(attempt 1 of 2\): rate limited\n$`)

// TestGenerateText_RetryLine_0015: a retried failure writes one line before
// its wait; a success and a failure that is not retried write none
// (0015-MADR D3).
func TestGenerateText_RetryLine_0015(t *testing.T) {
	t.Run("rate limit, then success", func(t *testing.T) {
		out := captureRetries(t)
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).
			Fail(llmprovider.ErrRateLimited).ReplyText("feat: after retry")
		if got, err := generateText(context.Background(), fake, "p", 1, time.Millisecond); err != nil || got != "feat: after retry" {
			t.Fatalf("generateText() = %q, %v", got, err)
		}
		if !retryLine.MatchString(out.String()) {
			t.Fatalf("stderr = %q; want one line matching %s", out.String(), retryLine)
		}
	})
	t.Run("success", func(t *testing.T) {
		out := captureRetries(t)
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).ReplyText("feat: x")
		if _, err := generateText(context.Background(), fake, "p", 2, time.Millisecond); err != nil {
			t.Fatal(err)
		}
		if out.Len() != 0 {
			t.Fatalf("stderr = %q; want nothing for a success", out.String())
		}
	})
	t.Run("not retried", func(t *testing.T) {
		out := captureRetries(t)
		fake := llmtest.NewFake(llmprovider.ProviderOpenAI, llmprovider.Capabilities{}).
			Fail(llmprovider.ErrAuthFailure).ReplyText("never")
		if _, err := generateText(context.Background(), fake, "p", 2, time.Millisecond); !errors.Is(err, llmprovider.ErrAuthFailure) {
			t.Fatalf("err = %v; want ErrAuthFailure", err)
		}
		if out.Len() != 0 {
			t.Fatalf("stderr = %q; want nothing for a failure that is not retried", out.String())
		}
	})
}

// TestRetryKind_0015: each class the line names, and never the provider's
// message (0015-MADR D3).
func TestRetryKind_0015(t *testing.T) {
	const detail = "provider detail that must not be quoted"
	for _, c := range []struct {
		err  error
		want string
	}{
		{fmt.Errorf("%w: %s", llmprovider.ErrRateLimited, detail), "rate limited"},
		{fmt.Errorf("%w: %s", llmprovider.ErrProviderUnavailable, detail), "unavailable"},
		{&url.Error{Op: "Post", URL: "https://example.test", Err: errors.New(detail)}, "cannot reach"},
		{errors.New(detail), "error"},
	} {
		out := captureRetries(t)
		reportRetry(context.Background(), llmprovider.RetryNotice{Provider: llmprovider.ProviderOpenAI, Attempt: 1,
			MaxAttempts: 3, Wait: 1500 * time.Millisecond, Err: c.err})
		want := "prepare-commit-msg: openai: retrying in 1.5s (attempt 1 of 3): " + c.want + "\n"
		if out.String() != want {
			t.Errorf("reportRetry(%v) wrote %q; want %q", c.err, out.String(), want)
		}
	}
}
