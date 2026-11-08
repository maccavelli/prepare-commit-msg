package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

// TestRunSetupNonInteractive_ReasoningEffort_0029: configure --yes saves
// --reasoning-effort, keeps it when the flag is absent, clears it with
// "default", and refuses an unknown value before saving (0029-MADR D3).
func TestRunSetupNonInteractive_ReasoningEffort_0029(t *testing.T) {
	isolate(t)
	ollama := string(llmprovider.ProviderOllama)
	run := func(flag string) (string, error) {
		conf, err := config.Load()
		if err != nil {
			t.Fatal(err)
		}
		err = runSetupNonInteractive(context.Background(), conf, SetupOptions{
			Provider: ollama, Model: "qwen2.5-coder:7b", NoEnv: true, Yes: true, ReasoningEffort: flag,
		})
		loaded, loadErr := config.Load()
		if loadErr != nil {
			t.Fatal(loadErr)
		}
		return loaded.Providers[ollama].ReasoningEffort, err
	}
	for _, step := range []struct{ flag, want string }{{"low", "low"}, {"", "low"}, {"default", ""}} {
		if got, err := run(step.flag); err != nil || got != step.want {
			t.Fatalf("flag %q: saved %q, %v; want %q", step.flag, got, err, step.want)
		}
	}
	if _, err := run("high"); err != nil {
		t.Fatal(err)
	}
	if got, err := run("turbo"); err == nil || !strings.Contains(err.Error(), `"turbo"`) || got != "high" {
		t.Fatalf("turbo: saved %q, %v; want refused, high kept", got, err)
	}
}
