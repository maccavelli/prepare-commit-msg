package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
)

// TestReasoningFor_0029: empty asks for none; each named effort, in any
// case, is asked for; anything else is refused with the provider and the
// values named (0029-MADR D3).
func TestReasoningFor_0029(t *testing.T) {
	if r, err := ReasoningFor("openai", ProviderConfig{}); r != nil || err != nil {
		t.Fatalf("empty: %+v, %v; want none", r, err)
	}
	for _, e := range ReasoningEfforts {
		r, err := ReasoningFor("openai", ProviderConfig{ReasoningEffort: " " + strings.ToUpper(string(e)) + " "})
		if err != nil || r == nil || r.Effort != e {
			t.Errorf("%s: %+v, %v", e, r, err)
		}
	}
	r, err := ReasoningFor("openai", ProviderConfig{ReasoningEffort: "turbo"})
	if r != nil || err == nil || !strings.Contains(err.Error(), `"turbo" for provider "openai"`) ||
		!strings.Contains(err.Error(), "low, medium, high, xhigh, max") {
		t.Fatalf("turbo: %+v, %v; want an error naming it, the provider and the values", r, err)
	}
	if got := ReasoningEfforts; len(got) != 5 || got[0] != llmprovider.EffortLow || got[4] != llmprovider.EffortMax {
		t.Fatalf("ReasoningEfforts = %v", got)
	}
}

// TestSetReasoningEffort_0029: configure's flag keeps, clears, sets, or
// refuses, leaving the saved value as it was (0029-MADR D3).
func TestSetReasoningEffort_0029(t *testing.T) {
	for _, tc := range []struct {
		saved, flag, want string
		fails             bool
	}{
		{"low", "", "low", false},
		{"low", "default", "", false},
		{"low", " DEFAULT ", "", false},
		{"", "High", "high", false},
		{"medium", "xhigh", "xhigh", false},
		{"low", "turbo", "low", true},
	} {
		pc := ProviderConfig{ReasoningEffort: tc.saved}
		err := SetReasoningEffort(&pc, tc.flag)
		if (err != nil) != tc.fails || pc.ReasoningEffort != tc.want {
			t.Errorf("saved %q, flag %q: %q, %v; want %q (fails %t)", tc.saved, tc.flag, pc.ReasoningEffort, err, tc.want, tc.fails)
		}
	}
}

// TestReasoningEffort_JSON_0029: the field is saved as reasoning_effort, and
// left out when empty, so a configuration without it is written as before.
func TestReasoningEffort_JSON_0029(t *testing.T) {
	with, err := json.Marshal(ProviderConfig{Model: "m", ReasoningEffort: "low"})
	if err != nil || !strings.Contains(string(with), `"reasoning_effort":"low"`) {
		t.Fatalf("%s, %v", with, err)
	}
	without, err := json.Marshal(ProviderConfig{Model: "m"})
	if err != nil || strings.Contains(string(without), "reasoning_effort") {
		t.Fatalf("%s, %v", without, err)
	}
}
