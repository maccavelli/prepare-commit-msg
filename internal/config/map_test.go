package config

import (
	"encoding/json"
	"testing"
)

func TestConfig_MapBehavior(t *testing.T) {
	// Accessing a missing map key returns a copy; assignment must write back.
	conf := &Config{
		Providers: make(map[string]ProviderConfig),
	}

	provider := "claude"
	pc := conf.Providers[provider]

	pc.APIKey = "test-key"
	pc.Model = "test-model"

	if _, ok := conf.Providers[provider]; ok {
		t.Errorf("Expected provider to be missing from map before assignment")
	}

	conf.ActiveProvider = provider
	conf.Providers[provider] = pc

	if val, ok := conf.Providers[provider]; !ok || val.APIKey != "test-key" {
		t.Errorf("Expected provider to be present after assignment")
	}

	data, _ := json.Marshal(conf)
	t.Logf("JSON: %s", string(data))
}
