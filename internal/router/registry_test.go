package router

import (
	"testing"

	"github.com/rullopat/spider-llama/internal/config"
)

func testConfig() config.Config {
	cfg := config.Config{
		Nodes: []config.NodeConfig{
			{ID: "local", Provider: "llamacpp", BaseURL: "http://127.0.0.1:8080", Tags: []string{"local"}, MaxConcurrency: 1},
			{ID: "remote", Provider: "llamacpp", BaseURL: "http://127.0.0.1:8081", Tags: []string{"remote"}, MaxConcurrency: 1},
		},
		Models: []config.ModelConfig{
			{ID: "small", Node: "local", Aliases: []string{"light-text"}, Capabilities: []string{"text", "json"}, Tags: []string{"light"}, Priority: 50},
			{ID: "large", Node: "remote", Capabilities: []string{"text", "json", "tools"}, Tags: []string{"reasoning"}, Priority: 80},
		},
		Routes: map[string]config.Route{
			"analysis": {
				Require: config.Requirements{Capabilities: []string{"text", "json"}},
				Prefer:  config.Preferences{Tags: []string{"reasoning"}},
			},
		},
	}
	cfg.ApplyDefaults()
	return cfg
}

func TestSelectByAlias(t *testing.T) {
	reg := NewRegistry(testConfig())

	candidates, err := reg.Select(SelectionRequest{Model: "alias:light-text"})
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if len(candidates) != 1 || candidates[0].Model.ID != "small" {
		t.Fatalf("expected small model, got %#v", candidates)
	}
}

func TestSelectByTaskPreferences(t *testing.T) {
	reg := NewRegistry(testConfig())

	candidates, err := reg.Select(SelectionRequest{Model: "auto", Task: "analysis"})
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if candidates[0].Model.ID != "large" {
		t.Fatalf("expected preferred reasoning model first, got %#v", candidates[0])
	}
}

func TestTryAcquireHonorsMaxConcurrency(t *testing.T) {
	reg := NewRegistry(testConfig())

	release, ok := reg.TryAcquire("local")
	if !ok {
		t.Fatalf("expected first acquire to succeed")
	}
	defer release()

	if _, ok := reg.TryAcquire("local"); ok {
		t.Fatalf("expected second acquire to fail")
	}
}
